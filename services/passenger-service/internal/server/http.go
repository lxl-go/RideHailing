package server

import (
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
	"ride-hailing/services/passenger-service/internal/conf"
	"ride-hailing/services/passenger-service/internal/service"
)

func NewHTTPServer(c *conf.Server, passengerSvc *service.PassengerService) *khttp.Server {
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
	passengerv1.RegisterPassengerServiceHTTPServer(srv, passengerSvc)
	return srv
}
