package server

import (
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	authv1 "ride-hailing/services/auth-service/api/auth/v1"
	"ride-hailing/services/auth-service/internal/conf"
	"ride-hailing/services/auth-service/internal/service"
)

func NewHTTPServer(c *conf.Server, authSvc *service.AuthService) *khttp.Server {
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
	authv1.RegisterAuthServiceHTTPServer(srv, authSvc)
	return srv
}
