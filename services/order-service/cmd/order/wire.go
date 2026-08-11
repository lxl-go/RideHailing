//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"go.uber.org/zap"

	"ride-hailing/services/order-service/internal/biz"
	"ride-hailing/services/order-service/internal/conf"
	"ride-hailing/services/order-service/internal/data"
	"ride-hailing/services/order-service/internal/server"
	"ride-hailing/services/order-service/internal/service"
)

func wireApp(*conf.Bootstrap, *zap.Logger, registry.Registrar) (*kratos.App, error) {
	panic(wire.Build(data.ProviderSet, biz.ProviderSet, service.ProviderSet, server.ProviderSet, newApp))
}
