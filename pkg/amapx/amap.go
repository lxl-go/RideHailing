package amapx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://restapi.amap.com"
	defaultTimeout = 5 * time.Second
)

type Config struct {
	WebKey     string
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type Client struct {
	cfg    Config
	client *http.Client
}

type Location struct {
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	FormattedAddress string  `json:"formattedAddress"`
	Province         string  `json:"province,omitempty"`
	City             string  `json:"city,omitempty"`
	District         string  `json:"district,omitempty"`
	AdCode           string  `json:"adCode,omitempty"`
}

type RoutePoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Route struct {
	Origin          Location     `json:"origin"`
	Destination     Location     `json:"destination"`
	DistanceMeters  int64        `json:"distanceMeters"`
	DurationSeconds int64        `json:"durationSeconds"`
	Polyline        []RoutePoint `json:"polyline"`
}

type Weather struct {
	Province      string `json:"province"`
	City          string `json:"city"`
	AdCode        string `json:"adCode"`
	Weather       string `json:"weather"`
	Temperature   string `json:"temperature"`
	WindDirection string `json:"windDirection"`
	WindPower     string `json:"windPower"`
	Humidity      string `json:"humidity"`
	ReportTime    string `json:"reportTime"`
}

type StaticMarker struct {
	Size      string    // small | mid | large，空为 small
	Color     string    // 例如 0xFF0000，空为默认
	Label     string    // 0-9 / A-Z / 单个中文字
	Locations [][2]float64 // [lng, lat] 列表
}

type StaticMapRequest struct {
	Location string         // 中心点 "lng,lat"，覆盖物存在时可选
	Zoom     int            // [1,17]
	Size     string         // 宽*高，如 "750*400"，默认 400*400
	Scale    int            // 1 普通 / 2 高清
	Markers  []StaticMarker // 最大10个点
}

type StaticMap struct {
	ContentType string // 如 image/png
	Data        []byte
}

func NewClient(cfg Config) (*Client, error) {
	cfg.WebKey = strings.TrimSpace(cfg.WebKey)
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	if cfg.WebKey == "" {
		return nil, errors.New("amap web key is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	} else if client.Timeout <= 0 {
		client.Timeout = cfg.Timeout
	}
	return &Client{cfg: cfg, client: client}, nil
}

func (c *Client) Geocode(ctx context.Context, address, city string) (*Location, error) {
	reqURL, err := c.buildURL("/v3/geocode/geo", map[string]string{
		"address": address,
		"city":    city,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Status   string `json:"status"`
		Info     string `json:"info"`
		Geocodes []struct {
			FormattedAddress string `json:"formatted_address"`
			Location         string `json:"location"`
			Province         string `json:"province"`
			City             string `json:"city"`
			District         string `json:"district"`
			AdCode           string `json:"adcode"`
		} `json:"geocodes"`
	}
	if err := c.getJSON(ctx, reqURL, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "1" {
		return nil, fmt.Errorf("amap geocode failed: %s", resp.Info)
	}
	if len(resp.Geocodes) == 0 {
		return nil, errors.New("amap geocode returned no result")
	}
	item := resp.Geocodes[0]
	lng, lat, err := parseLocation(item.Location)
	if err != nil {
		return nil, err
	}
	return &Location{
		Latitude:         lat,
		Longitude:        lng,
		FormattedAddress: item.FormattedAddress,
		Province:         item.Province,
		City:             item.City,
		District:         item.District,
		AdCode:           item.AdCode,
	}, nil
}

func (c *Client) Regeo(ctx context.Context, lat, lng float64) (*Location, error) {
	reqURL, err := c.buildURL("/v3/geocode/regeo", map[string]string{
		"location":   fmt.Sprintf("%.6f,%.6f", lng, lat),
		"extensions": "base",
		"radius":     "1000",
		"roadlevel":  "1",
		"homeorcorp": "0",
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Status    string `json:"status"`
		Info      string `json:"info"`
		Regeocode struct {
			FormattedAddress string `json:"formatted_address"`
			AddressComponent struct {
				Province string `json:"province"`
				City     string `json:"city"`
				District string `json:"district"`
				AdCode   string `json:"adcode"`
			} `json:"addressComponent"`
		} `json:"regeocode"`
	}
	if err := c.getJSON(ctx, reqURL, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "1" {
		return nil, fmt.Errorf("amap regeo failed: %s", resp.Info)
	}
	return &Location{
		Latitude:         lat,
		Longitude:        lng,
		FormattedAddress: resp.Regeocode.FormattedAddress,
		Province:         resp.Regeocode.AddressComponent.Province,
		City:             resp.Regeocode.AddressComponent.City,
		District:         resp.Regeocode.AddressComponent.District,
		AdCode:           resp.Regeocode.AddressComponent.AdCode,
	}, nil
}

func (c *Client) DrivingRoute(ctx context.Context, origin, destination, strategy string) (*Route, error) {
	reqURL, err := c.buildURL("/v3/direction/driving", map[string]string{
		"origin":      strings.TrimSpace(origin),
		"destination": strings.TrimSpace(destination),
		"strategy":    strings.TrimSpace(strategy),
		"extensions":  "all",
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Route  struct {
			Paths []struct {
				Distance string `json:"distance"`
				Duration string `json:"duration"`
				Steps    []struct {
					Polyline string `json:"polyline"`
				} `json:"steps"`
			} `json:"paths"`
		} `json:"route"`
	}
	if err := c.getJSON(ctx, reqURL, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "1" {
		return nil, fmt.Errorf("amap driving route failed: %s", resp.Info)
	}
	if len(resp.Route.Paths) == 0 {
		return nil, errors.New("amap driving route returned no path")
	}
	path := resp.Route.Paths[0]
	points := make([]RoutePoint, 0, 32)
	for _, step := range path.Steps {
		for _, rawPoint := range strings.Split(step.Polyline, ";") {
			rawPoint = strings.TrimSpace(rawPoint)
			if rawPoint == "" {
				continue
			}
			lng, lat, err := parseLocation(rawPoint)
			if err != nil {
				return nil, err
			}
			if len(points) > 0 {
				last := points[len(points)-1]
				if last.Latitude == lat && last.Longitude == lng {
					continue
				}
			}
			points = append(points, RoutePoint{Latitude: lat, Longitude: lng})
		}
	}
	distanceMeters, err := strconv.ParseInt(strings.TrimSpace(path.Distance), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse amap driving distance: %w", err)
	}
	durationSeconds, err := strconv.ParseInt(strings.TrimSpace(path.Duration), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse amap driving duration: %w", err)
	}
	return &Route{
		DistanceMeters:  distanceMeters,
		DurationSeconds: durationSeconds,
		Polyline:        points,
	}, nil
}

func (c *Client) RouteByAddress(ctx context.Context, originAddress, destinationAddress, city, strategy string) (*Route, *Location, *Location, error) {
	origin, err := c.Geocode(ctx, originAddress, city)
	if err != nil {
		return nil, nil, nil, err
	}
	destination, err := c.Geocode(ctx, destinationAddress, city)
	if err != nil {
		return nil, nil, nil, err
	}
	route, err := c.DrivingRoute(ctx, coordString(origin.Longitude, origin.Latitude), coordString(destination.Longitude, destination.Latitude), strategy)
	if err != nil {
		return nil, nil, nil, err
	}
	route.Origin = *origin
	route.Destination = *destination
	return route, origin, destination, nil
}

func (c *Client) Distance(ctx context.Context, origins []string, destination string) (int64, error) {
	if len(origins) == 0 {
		return 0, errors.New("origins is required")
	}
	route, err := c.DrivingRoute(ctx, origins[0], destination, "")
	if err != nil {
		return 0, err
	}
	return route.DistanceMeters, nil
}

func (c *Client) Weather(ctx context.Context, city string) (*Weather, error) {
	reqURL, err := c.buildURL("/v3/weather/weatherInfo", map[string]string{
		"city":       strings.TrimSpace(city),
		"extensions": "base",
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Lives  []struct {
			Province      string `json:"province"`
			City          string `json:"city"`
			AdCode        string `json:"adcode"`
			Weather       string `json:"weather"`
			Temperature   string `json:"temperature"`
			WindDirection string `json:"winddirection"`
			WindPower     string `json:"windpower"`
			Humidity      string `json:"humidity"`
			ReportTime    string `json:"reporttime"`
		} `json:"lives"`
	}
	if err := c.getJSON(ctx, reqURL, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "1" {
		return nil, fmt.Errorf("amap weather failed: %s", resp.Info)
	}
	if len(resp.Lives) == 0 {
		return nil, errors.New("amap weather returned no result")
	}
	item := resp.Lives[0]
	return &Weather{
		Province:      item.Province,
		City:          item.City,
		AdCode:        item.AdCode,
		Weather:       item.Weather,
		Temperature:   item.Temperature,
		WindDirection: item.WindDirection,
		WindPower:     item.WindPower,
		Humidity:      item.Humidity,
		ReportTime:    item.ReportTime,
	}, nil
}

func (c *Client) StaticMap(ctx context.Context, req StaticMapRequest) (*StaticMap, error) {
	params := map[string]string{}
	if strings.TrimSpace(req.Location) != "" {
		params["location"] = strings.TrimSpace(req.Location)
	}
	if req.Zoom > 0 {
		params["zoom"] = strconv.Itoa(req.Zoom)
	}
	if strings.TrimSpace(req.Size) != "" {
		params["size"] = strings.TrimSpace(req.Size)
	}
	if req.Scale > 0 {
		params["scale"] = strconv.Itoa(req.Scale)
	}
	if markers := buildMarkersParam(req.Markers); markers != "" {
		params["markers"] = markers
	}
	reqURL, err := c.buildURL("/v3/staticmap", params)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("amap staticmap http status %d", resp.StatusCode)
	}
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "image") {
		var errResp struct {
			Status string `json:"status"`
			Info   string `json:"info"`
		}
		if jsonErr := json.NewDecoder(resp.Body).Decode(&errResp); jsonErr == nil && errResp.Status == "0" {
			return nil, fmt.Errorf("amap staticmap failed: %s", errResp.Info)
		}
		return nil, errors.New("amap staticmap returned non-image response")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &StaticMap{ContentType: contentType, Data: data}, nil
}

func buildMarkersParam(markers []StaticMarker) string {
	groups := make([]string, 0, len(markers))
	for _, marker := range markers {
		if len(marker.Locations) == 0 {
			continue
		}
		style := strings.TrimSpace(marker.Size)
		if color := strings.TrimSpace(marker.Color); color != "" {
			if style != "" {
				style += ","
			}
			style += color
		}
		if label := strings.TrimSpace(marker.Label); label != "" {
			if style != "" {
				style += ","
			}
			style += label
		}
		locations := make([]string, 0, len(marker.Locations))
		for _, point := range marker.Locations {
			locations = append(locations, fmt.Sprintf("%.6f,%.6f", point[0], point[1]))
		}
		groups = append(groups, style+":"+strings.Join(locations, ";"))
	}
	return strings.Join(groups, "|")
}

func (c *Client) getJSON(ctx context.Context, reqURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("amap http status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) buildURL(path string, params map[string]string) (string, error) {
	base, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	query := base.Query()
	query.Set("key", c.cfg.WebKey)
	for key, value := range params {
		if strings.TrimSpace(value) == "" {
			continue
		}
		query.Set(key, value)
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func parseLocation(value string) (float64, float64, error) {
	parts := strings.Split(strings.TrimSpace(value), ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid location: %s", value)
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse longitude: %w", err)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse latitude: %w", err)
	}
	return lng, lat, nil
}

func coordString(lng, lat float64) string {
	return fmt.Sprintf("%.6f,%.6f", lng, lat)
}
