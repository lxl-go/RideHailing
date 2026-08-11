package server

import (
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/driver-service/internal/conf"
	"ride-hailing/services/driver-service/internal/service"
)

func NewHTTPServer(c *conf.Server, driverSvc *service.DriverService) *khttp.Server {
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
	driverv1.RegisterDriverServiceHTTPServer(srv, driverSvc)
	return srv
}
