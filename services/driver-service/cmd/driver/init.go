package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"go.uber.org/zap"

	"ride-hailing/services/driver-service/internal/biz"
	"ride-hailing/services/driver-service/internal/conf"
	"ride-hailing/services/driver-service/internal/data"
	"ride-hailing/services/driver-service/internal/server"
	"ride-hailing/services/driver-service/internal/service"
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
	repo := data.NewDriverRepo(db, logger)
	verifier := data.NewRealNameVerifier(cfg.RealName)
	uc := biz.NewDriverUsecase(node, logger, repo, verifier)
	driverSvc := service.NewDriverService(uc)
	httpSrv := server.NewHTTPServer(cfg.Server, driverSvc)
	grpcSrv := server.NewGRPCServer(cfg.Server, driverSvc)
	return newApp(httpSrv, grpcSrv, registrar), nil
}
