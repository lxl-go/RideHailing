package server

import (
	"time"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	"ride-hailing/pkg/grpcx"
	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
	"ride-hailing/services/passenger-service/internal/conf"
	"ride-hailing/services/passenger-service/internal/service"
)

func NewGRPCServer(c *conf.Server, passengerSvc *service.PassengerService) *kgrpc.Server {
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
	passengerv1.RegisterPassengerServiceServer(srv, passengerSvc)
	return srv
}
