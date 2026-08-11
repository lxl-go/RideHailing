package server

import (
	"time"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	"ride-hailing/pkg/grpcx"
	reviewv1 "ride-hailing/services/review-service/api/review/v1"
	"ride-hailing/services/review-service/internal/conf"
	"ride-hailing/services/review-service/internal/service"
)

func NewGRPCServer(c *conf.Server, reviewSvc *service.ReviewService) *kgrpc.Server {
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
	reviewv1.RegisterReviewServiceServer(srv, reviewSvc)
	return srv
}
