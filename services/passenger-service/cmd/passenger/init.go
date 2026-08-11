package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"go.uber.org/zap"

	"ride-hailing/services/passenger-service/internal/biz"
	"ride-hailing/services/passenger-service/internal/conf"
	"ride-hailing/services/passenger-service/internal/data"
	"ride-hailing/services/passenger-service/internal/server"
	"ride-hailing/services/passenger-service/internal/service"
)

func initApp(cfg *conf.Bootstrap, logger *zap.Logger, registrar registry.Registrar) (*kratos.App, error) {
	db, err := data.NewDB(cfg.Data, logger)
	if err != nil {
		return nil, err
	}
	repo := data.NewPassengerRepo(db, logger)
	uc := biz.NewPassengerUsecase(logger, repo)
	passengerSvc := service.NewPassengerService(uc)
	httpSrv := server.NewHTTPServer(cfg.Server, passengerSvc)
	grpcSrv := server.NewGRPCServer(cfg.Server, passengerSvc)
	return newApp(httpSrv, grpcSrv, registrar), nil
}
