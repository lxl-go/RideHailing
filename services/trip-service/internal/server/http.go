package server

import (
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
	"ride-hailing/services/trip-service/internal/conf"
	"ride-hailing/services/trip-service/internal/service"
)

func NewHTTPServer(c *conf.Server, tripSvc *service.TripService) *khttp.Server {
	opts := []khttp.ServerOption{}
	if c != nil && c.Http != nil {
		if c.Http.Addr != "" {
			opts = append(opts, khttp.Address(c.Http.Addr))
		}
		if d, err := time.ParseDuration(c.Http.Timeout); err == nil && d > 0 {
			opts = append(opts, khttp.Timeout(d))
		}
	}
	srv := khttp.NewServer(opts...)
	tripv1.RegisterTripServiceHTTPServer(srv, tripSvc)
	return srv
}
