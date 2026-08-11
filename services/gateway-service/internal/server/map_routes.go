package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"

	"ride-hailing/pkg/amapx"
	"ride-hailing/services/gateway-service/internal/conf"
)

type amapRouteClient interface {
	Geocode(ctx context.Context, address, city string) (*amapx.Location, error)
	Regeo(ctx context.Context, lat, lng float64) (*amapx.Location, error)
	RouteByAddress(ctx context.Context, originAddress, destinationAddress, city, strategy string) (*amapx.Route, *amapx.Location, *amapx.Location, error)
	Weather(ctx context.Context, city string) (*amapx.Weather, error)
	StaticMap(ctx context.Context, req amapx.StaticMapRequest) (*amapx.StaticMap, error)
}

func registerMapRoutes(srv *khttp.Server, amapCfg *conf.Amap) {
	client, err := newAmapClient(amapCfg)
	if err != nil {
		zap.L().Warn("amap config unavailable", zap.Error(err))
		registerMapRoutesWithClient(srv, nil)
		return
	}
	registerMapRoutesWithClient(srv, client)
}

func registerMapRoutesWithClient(srv *khttp.Server, client amapRouteClient) {
	router := srv.Route("/")
	router.GET("/api/v1/maps/geocode", func(ctx khttp.Context) error {
		if client == nil {
			return returnUnavailable(ctx, "amap config unavailable")
		}
		query := ctx.Query()
		address := strings.TrimSpace(query.Get("address"))
		if address == "" {
			return returnBadRequest(ctx, "address is required")
		}
		point, err := client.Geocode(ctx, address, strings.TrimSpace(query.Get("city")))
		if err != nil {
			zap.L().Warn("amap geocode failed", zap.String("address", address), zap.Error(err))
			return err
		}
		zap.L().Info("amap geocode ok", zap.String("address", address))
		return returnData(ctx, mapLocationResponse(point), nil)
	})
	router.GET("/api/v1/maps/regeo", func(ctx khttp.Context) error {
		if client == nil {
			return returnUnavailable(ctx, "amap config unavailable")
		}
		lat, lng, err := parseLatLngQuery(ctx)
		if err != nil {
			return returnBadRequest(ctx, err.Error())
		}
		point, err := client.Regeo(ctx, lat, lng)
		if err != nil {
			zap.L().Warn("amap regeo failed", zap.Float64("lat", lat), zap.Float64("lng", lng), zap.Error(err))
			return err
		}
		zap.L().Info("amap regeo ok", zap.Float64("lat", lat), zap.Float64("lng", lng))
		return returnData(ctx, mapLocationResponse(point), nil)
	})
	router.GET("/api/v1/maps/route", func(ctx khttp.Context) error {
		if client == nil {
			return returnUnavailable(ctx, "amap config unavailable")
		}
		query := ctx.Query()
		origin := strings.TrimSpace(query.Get("origin"))
		destination := strings.TrimSpace(query.Get("destination"))
		if origin == "" || destination == "" {
			return returnBadRequest(ctx, "origin and destination are required")
		}
		route, _, _, err := client.RouteByAddress(ctx, origin, destination, strings.TrimSpace(query.Get("city")), strings.TrimSpace(query.Get("strategy")))
		if err != nil {
			zap.L().Warn("amap route failed", zap.String("origin", origin), zap.String("destination", destination), zap.Error(err))
			return err
		}
		zap.L().Info("amap route ok", zap.String("origin", origin), zap.String("destination", destination))
		return returnData(ctx, mapRouteResponse(route), nil)
	})
	router.GET("/api/v1/maps/weather", func(ctx khttp.Context) error {
		if client == nil {
			return returnUnavailable(ctx, "amap config unavailable")
		}
		city := strings.TrimSpace(ctx.Query().Get("city"))
		if city == "" {
			return returnBadRequest(ctx, "city is required")
		}
		weather, err := client.Weather(ctx, city)
		if err != nil {
			zap.L().Warn("amap weather failed", zap.String("city", city), zap.Error(err))
			return err
		}
		zap.L().Info("amap weather ok", zap.String("city", city))
		return returnData(ctx, weather, nil)
	})
	router.GET("/api/v1/maps/static", func(ctx khttp.Context) error {
		if client == nil {
			return returnUnavailable(ctx, "amap config unavailable")
		}
		query := ctx.Query()
		origin := strings.TrimSpace(query.Get("origin"))
		destination := strings.TrimSpace(query.Get("destination"))
		if origin == "" || destination == "" {
			return returnBadRequest(ctx, "origin and destination are required")
		}
		originPoint, err := parseStaticCoord(origin)
		if err != nil {
			return returnBadRequest(ctx, "origin is invalid")
		}
		destinationPoint, err := parseStaticCoord(destination)
		if err != nil {
			return returnBadRequest(ctx, "destination is invalid")
		}
		markers := []amapx.StaticMarker{
			{Size: "large", Color: "0x00C853", Label: "起", Locations: [][2]float64{originPoint}},
			{Size: "large", Color: "0xFF0000", Label: "终", Locations: [][2]float64{destinationPoint}},
		}
		req := amapx.StaticMapRequest{
			Markers: markers,
			Scale:   parsePositiveInt(query.Get("scale"), 2),
		}
		if size := strings.TrimSpace(query.Get("size")); size != "" {
			req.Size = size
		}
		if zoom := parseInt(query.Get("zoom")); zoom > 0 {
			req.Zoom = zoom
		} else {
			req.Zoom = 12
		}
		image, err := client.StaticMap(ctx, req)
		if err != nil {
			zap.L().Warn("amap staticmap failed", zap.String("origin", origin), zap.String("destination", destination), zap.Error(err))
			return err
		}
		ctx.Response().Header().Set("Content-Type", image.ContentType)
		ctx.Response().Header().Set("Cache-Control", "public, max-age=3600")
		ctx.Response().WriteHeader(http.StatusOK)
		_, _ = ctx.Response().Write(image.Data)
		return nil
	})
}

func parseStaticCoord(value string) ([2]float64, error) {
	parts := strings.Split(strings.TrimSpace(value), ",")
	if len(parts) != 2 {
		return [2]float64{}, errors.New("invalid coordinate")
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil || lng < -180 || lng > 180 {
		return [2]float64{}, errors.New("invalid longitude")
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil || lat < -90 || lat > 90 {
		return [2]float64{}, errors.New("invalid latitude")
	}
	return [2]float64{lng, lat}, nil
}

func parsePositiveInt(value string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func newAmapClient(cfg *conf.Amap) (*amapx.Client, error) {
	if cfg == nil {
		cfg = &conf.Amap{}
	}
	timeout := 5 * time.Second
	if d, err := time.ParseDuration(strings.TrimSpace(cfg.Timeout)); err == nil && d > 0 {
		timeout = d
	}
	return amapx.NewClient(amapx.Config{
		WebKey:  configOrEnv(cfg.WebKey, "AMAP_WEB_KEY"),
		Timeout: timeout,
	})
}

func parseLatLngQuery(ctx khttp.Context) (float64, float64, error) {
	lat, err := strconv.ParseFloat(strings.TrimSpace(ctx.Query().Get("lat")), 64)
	if err != nil || lat < -90 || lat > 90 {
		return 0, 0, errors.New("lat is invalid")
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(ctx.Query().Get("lng")), 64)
	if err != nil || lng < -180 || lng > 180 {
		return 0, 0, errors.New("lng is invalid")
	}
	return lat, lng, nil
}

func mapRouteResponse(route *amapx.Route) map[string]any {
	if route == nil {
		return map[string]any{}
	}
	points := make([]map[string]any, 0, len(route.Polyline))
	for _, point := range route.Polyline {
		points = append(points, map[string]any{
			"latitude":  point.Latitude,
			"longitude": point.Longitude,
			"lat":       point.Latitude,
			"lng":       point.Longitude,
		})
	}
	return map[string]any{
		"origin":           mapLocationResponse(&route.Origin),
		"destination":      mapLocationResponse(&route.Destination),
		"distanceMeters":   route.DistanceMeters,
		"distance_meters":  route.DistanceMeters,
		"durationSeconds":  route.DurationSeconds,
		"duration_seconds": route.DurationSeconds,
		"polyline":         points,
		"points":           points,
	}
}

func mapLocationResponse(point *amapx.Location) map[string]any {
	if point == nil {
		return map[string]any{}
	}
	return map[string]any{
		"latitude":          point.Latitude,
		"longitude":         point.Longitude,
		"lat":               point.Latitude,
		"lng":               point.Longitude,
		"formattedAddress":  point.FormattedAddress,
		"formatted_address": point.FormattedAddress,
		"province":          point.Province,
		"city":              point.City,
		"district":          point.District,
		"adCode":            point.AdCode,
		"adcode":            point.AdCode,
	}
}
