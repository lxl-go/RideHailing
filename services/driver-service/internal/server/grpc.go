package server

import (
	"time"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	"ride-hailing/pkg/grpcx"
	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/driver-service/internal/conf"
	"ride-hailing/services/driver-service/internal/service"
)

func NewGRPCServer(c *conf.Server, driverSvc *service.DriverService) *kgrpc.Server {
	opts := []kgrpc.ServerOption{kgrpc.UnaryInterceptor(grpcx.UnaryServerInterceptor())}
	if c != nil && c.Grpc != nil {
		if c.Grpc.Addr != "" {
			opts = append(opts, kgrpc.Address(c.Grpc.Addr))
		}
		if d, err := time.ParseDuration(c.Grpc.Timeout); err == nil && d > 0 {
			opts = append(opts, kgrpc.Timeout(d))
		}
	}
	srv := kgrpc.NewServer(opts...)
	driverv1.RegisterDriverServiceServer(srv, driverSvc)
	return srv
}
