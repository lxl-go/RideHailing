package server

import (
	"time"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	"ride-hailing/pkg/grpcx"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
	"ride-hailing/services/order-service/internal/conf"
	"ride-hailing/services/order-service/internal/service"
)

func NewGRPCServer(c *conf.Server, orderSvc *service.OrderService) *kgrpc.Server {
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
	orderv1.RegisterOrderServiceServer(srv, orderSvc)
	return srv
}
