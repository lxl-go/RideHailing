package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"go.uber.org/zap"

	"ride-hailing/services/trip-service/internal/biz"
	"ride-hailing/services/trip-service/internal/conf"
	"ride-hailing/services/trip-service/internal/data"
	"ride-hailing/services/trip-service/internal/server"
	"ride-hailing/services/trip-service/internal/service"
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
	redisClient, err := data.NewRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}
	repo := data.NewTripRepo(db, logger, redisClient)
	amap := data.NewAMapClient(cfg.AMap)
	uc := biz.NewTripUsecase(node, logger, repo)
	tripSvc := service.NewTripService(uc, amap)
	httpSrv := server.NewHTTPServer(cfg.Server, tripSvc)
	grpcSrv := server.NewGRPCServer(cfg.Server, tripSvc)
	return newApp(httpSrv, grpcSrv, registrar), nil
}
