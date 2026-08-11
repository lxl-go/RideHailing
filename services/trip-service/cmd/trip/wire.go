//go:build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"go.uber.org/zap"

	"ride-hailing/services/trip-service/internal/biz"
	"ride-hailing/services/trip-service/internal/conf"
	"ride-hailing/services/trip-service/internal/data"
	"ride-hailing/services/trip-service/internal/server"
	"ride-hailing/services/trip-service/internal/service"
)

func initAppWithWire(*conf.Bootstrap, *zap.Logger, registry.Registrar) (*kratos.App, error) {
	panic(wire.Build(
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
		newApp,
	))
}
