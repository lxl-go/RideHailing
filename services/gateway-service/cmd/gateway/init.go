package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"

	"ride-hailing/services/gateway-service/internal/biz"
	"ride-hailing/services/gateway-service/internal/conf"
	"ride-hailing/services/gateway-service/internal/data"
	"ride-hailing/services/gateway-service/internal/server"
	"ride-hailing/services/gateway-service/internal/service"
)

func initApp(cfg *conf.Bootstrap, registrar registry.Registrar, discovery registry.Discovery) (*kratos.App, error) {
	authClient, err := data.NewAuthClient(cfg.Clients, discovery)
	if err != nil {
		return nil, err
	}
	tripClient, err := data.NewTripClient(cfg.Clients, discovery)
	if err != nil {
		return nil, err
	}
	orderClient, err := data.NewOrderClient(cfg.Clients, discovery)
	if err != nil {
		return nil, err
	}
	reviewClient, err := data.NewReviewClient(cfg.Clients, discovery)
	if err != nil {
		return nil, err
	}
	passengerClient, err := data.NewPassengerClient(cfg.Clients, discovery)
	if err != nil {
		return nil, err
	}
	driverClient, err := data.NewDriverClient(cfg.Clients, discovery)
	if err != nil {
		return nil, err
	}
	authUsecase := biz.NewAuthUsecase(authClient)
	tripUsecase := biz.NewTripUsecase(tripClient)
	orderUsecase := biz.NewOrderUsecase(orderClient)
	reviewUsecase := biz.NewReviewUsecase(reviewClient)
	passengerUsecase := biz.NewPassengerUsecase(passengerClient)
	driverUsecase := biz.NewDriverUsecase(driverClient)
	authService := service.NewAuthService(authUsecase)
	tripService := service.NewTripService(tripUsecase)
	orderService := service.NewOrderService(orderUsecase)
	reviewService := service.NewReviewService(reviewUsecase)
	passengerService := service.NewPassengerService(passengerUsecase)
	driverService := service.NewDriverService(driverUsecase)
	httpSrv := server.NewHTTPServer(cfg.Server, cfg.Auth, cfg.Alipay, cfg.Amap, authService, tripService, orderService, reviewService, passengerService, driverService)
	return newApp(httpSrv, registrar), nil
}
