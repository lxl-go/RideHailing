package server

import (
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	reviewv1 "ride-hailing/services/review-service/api/review/v1"
	"ride-hailing/services/review-service/internal/conf"
	"ride-hailing/services/review-service/internal/service"
)

func NewHTTPServer(c *conf.Server, reviewSvc *service.ReviewService) *khttp.Server {
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
	reviewv1.RegisterReviewServiceHTTPServer(srv, reviewSvc)
	return srv
}
