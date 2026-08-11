package data

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"

	"ride-hailing/pkg/grpcx"
	"ride-hailing/services/gateway-service/internal/conf"
	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

func NewTripClient(c *conf.Clients, discovery registry.Discovery) (TripClient, error) {
	baseURL := "http://127.0.0.1:9040"
	endpoint := ""
	if c != nil && c.Trip != nil && c.Trip.HTTPBaseURL != "" {
		baseURL = c.Trip.HTTPBaseURL
	}
	if c != nil && c.Trip != nil && c.Trip.Endpoint != "" {
		endpoint = c.Trip.Endpoint
	}
	if discovery != nil && strings.HasPrefix(endpoint, "discovery:///") {
		conn, err := grpcx.DialInsecure(context.Background(), endpoint, discovery, grpcx.ClientOptions{})
		if err != nil {
			return nil, err
		}
		return NewTripGRPCClient(tripv1.NewTripServiceClient(conn)), nil
	}
	return NewTripHTTPClient(baseURL), nil
}

var ProviderSet = wire.NewSet(NewAuthClient, NewTripClient, NewOrderClient, NewReviewClient, NewPassengerClient, NewDriverClient)
