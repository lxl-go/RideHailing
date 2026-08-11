package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ride-hailing/services/trip-service/internal/conf"
)

type Location struct {
	POIID            string
	Name             string
	FormattedAddress string
	Longitude        float64
	Latitude         float64
}

type Route struct {
	DistanceMeters  int
	DurationSeconds int
}

type AMapClient struct {
	key     string
	baseURL string
	http    *http.Client
}

func NewAMapClient(config *conf.AMap) *AMapClient {
	baseURL, timeout := "https://restapi.amap.com", 3*time.Second
	if config != nil {
		if strings.TrimSpace(config.BaseURL) != "" {
			baseURL = strings.TrimRight(config.BaseURL, "/")
		}
		if parsed, err := time.ParseDuration(config.Timeout); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	key := ""
	if config != nil {
		key = strings.TrimSpace(config.WebServiceKey)
	}
	return &AMapClient{key: key, baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

func (c *AMapClient) Search(ctx context.Context, keyword, city string) ([]Location, error) {
	if c.key == "" {
		return nil, fmt.Errorf("amap web service key is not configured")
	}
	values := url.Values{"key": {c.key}, "keywords": {keyword}, "city": {city}, "offset": {"10"}, "extensions": {"base"}}
	var response struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Pois   []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Address  string `json:"address"`
			Location string `json:"location"`
		} `json:"pois"`
	}
	if err := c.get(ctx, "/v3/place/text", values, &response); err != nil {
		return nil, err
	}
	if response.Status != "1" {
		return nil, fmt.Errorf("amap location search failed: %s", response.Info)
	}
	locations := make([]Location, 0, len(response.Pois))
	for _, poi := range response.Pois {
		longitude, latitude, err := parseLocation(poi.Location)
		if err != nil {
			continue
		}
		locations = append(locations, Location{POIID: poi.ID, Name: poi.Name, FormattedAddress: poi.Address, Longitude: longitude, Latitude: latitude})
	}
	if len(locations) == 0 {
		return nil, fmt.Errorf("amap returned no valid locations")
	}
	return locations, nil
}

func (c *AMapClient) DrivingRoute(ctx context.Context, origin, destination Location) (Route, error) {
	if c.key == "" {
		return Route{}, fmt.Errorf("amap web service key is not configured")
	}
	values := url.Values{"key": {c.key}, "origin": {fmt.Sprintf("%f,%f", origin.Longitude, origin.Latitude)}, "destination": {fmt.Sprintf("%f,%f", destination.Longitude, destination.Latitude)}, "extensions": {"base"}}
	var response struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Route  struct {
			Paths []struct {
				Distance string `json:"distance"`
				Duration string `json:"duration"`
			} `json:"paths"`
		} `json:"route"`
	}
	if err := c.get(ctx, "/v3/direction/driving", values, &response); err != nil {
		return Route{}, err
	}
	if response.Status != "1" || len(response.Route.Paths) == 0 {
		return Route{}, fmt.Errorf("amap route unavailable: %s", response.Info)
	}
	distance, err := strconv.Atoi(response.Route.Paths[0].Distance)
	if err != nil {
		return Route{}, err
	}
	duration, err := strconv.Atoi(response.Route.Paths[0].Duration)
	if err != nil {
		return Route{}, err
	}
	return Route{DistanceMeters: distance, DurationSeconds: duration}, nil
}

func (c *AMapClient) get(ctx context.Context, path string, values url.Values, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+values.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("amap http status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func parseLocation(value string) (float64, float64, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid amap location")
	}
	longitude, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, err
	}
	latitude, err := strconv.ParseFloat(parts[1], 64)
	return longitude, latitude, err
}
