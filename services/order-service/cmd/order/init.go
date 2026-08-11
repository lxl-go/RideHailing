package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"go.uber.org/zap"

	"ride-hailing/services/order-service/internal/biz"
	"ride-hailing/services/order-service/internal/conf"
	"ride-hailing/services/order-service/internal/data"
	"ride-hailing/services/order-service/internal/server"
	"ride-hailing/services/order-service/internal/service"
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
	repo := data.NewOrderRepo(db, logger)
	uc := biz.NewOrderUsecase(node, logger, repo)
	orderSvc := service.NewOrderService(uc)
	httpSrv := server.NewHTTPServer(cfg.Server, orderSvc)
	grpcSrv := server.NewGRPCServer(cfg.Server, orderSvc)
	return newApp(httpSrv, grpcSrv, registrar), nil
}
