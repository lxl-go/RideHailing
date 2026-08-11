package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"go.uber.org/zap"

	"ride-hailing/services/review-service/internal/biz"
	"ride-hailing/services/review-service/internal/conf"
	"ride-hailing/services/review-service/internal/data"
	"ride-hailing/services/review-service/internal/server"
	"ride-hailing/services/review-service/internal/service"
)

func initApp(cfg *conf.Bootstrap, logger *zap.Logger, registrar registry.Registrar) (*kratos.App, error) {
	node, err := data.NewSnowflakeNode(cfg.Server)
	if err != nil {
		return nil, err
	}
	db, err := data.NewDB(cfg.Data, logger)
	if err != nil {
		return nil, err
	}
	repo := data.NewReviewRepo(db, logger)
	uc := biz.NewReviewUsecase(node, logger, repo)
	reviewSvc := service.NewReviewService(uc)
	httpSrv := server.NewHTTPServer(cfg.Server, reviewSvc)
	grpcSrv := server.NewGRPCServer(cfg.Server, reviewSvc)
	return newApp(httpSrv, grpcSrv, registrar), nil
}
