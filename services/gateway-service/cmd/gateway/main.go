package main

import (
	"log"

	"github.com/go-kratos/kratos/v2/registry"
	"go.uber.org/zap"

	"ride-hailing/pkg/configx"
	pkgregistry "ride-hailing/pkg/registry"
	"ride-hailing/services/gateway-service/internal/conf"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("[%s] init logger: %v", serviceName, err)
	}
	defer func() {
		_ = logger.Sync()
	}()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	var cfg conf.Bootstrap
	source, err := configx.LoadYAML(&cfg, configx.LoadOptions{
		NacosPath:    "configs/nacos.yaml",
		FallbackPath: "configs/config.yaml",
	})
	if err != nil {
		log.Fatalf("[%s] load config: %v", serviceName, err)
	}
	logger.Info("config loaded", zap.String("source", string(source)))

	var registrar registry.Registrar
	var discovery registry.Discovery
	if cfg.Registry != nil && cfg.Registry.Nacos != nil && cfg.Registry.Nacos.Enabled {
		nacosCfg := configx.NacosConfig{
			ServerAddr:  cfg.Registry.Nacos.ServerAddr,
			ServerPort:  cfg.Registry.Nacos.ServerPort,
			NamespaceID: cfg.Registry.Nacos.NamespaceId,
			Group:       cfg.Registry.Nacos.Group,
			Username:    cfg.Registry.Nacos.Username,
			Password:    cfg.Registry.Nacos.Password,
			TimeoutMs:   cfg.Registry.Nacos.TimeoutMs,
		}
		registrar, err = pkgregistry.NewRegistrar(nacosCfg)
		if err != nil {
			log.Fatalf("[%s] init registrar: %v", serviceName, err)
		}
		discovery, err = pkgregistry.NewDiscovery(nacosCfg)
		if err != nil {
			log.Fatalf("[%s] init discovery: %v", serviceName, err)
		}
	}

	app, err := initApp(&cfg, registrar, discovery)
	if err != nil {
		log.Fatalf("[%s] init app: %v", serviceName, err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("[%s] run: %v", serviceName, err)
	}
}
