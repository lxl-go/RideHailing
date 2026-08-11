package registry

import (
	"strings"

	kratosregistry "github.com/go-kratos/kratos/v2/registry"
	nacosregistry "github.com/go-kratos/kratos/contrib/registry/nacos/v2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"ride-hailing/pkg/configx"
)

func NewRegistrar(cfg configx.NacosConfig) (kratosregistry.Registrar, error) {
	client, err := clients.NewNamingClient(clientOptions(cfg))
	if err != nil {
		return nil, err
	}
	return nacosregistry.New(client), nil
}

func NewDiscovery(cfg configx.NacosConfig) (kratosregistry.Discovery, error) {
	client, err := clients.NewNamingClient(clientOptions(cfg))
	if err != nil {
		return nil, err
	}
	return nacosregistry.New(client), nil
}

func clientOptions(cfg configx.NacosConfig) vo.NacosClientParam {
	clientCfg := clientConfig(cfg)
	return vo.NacosClientParam{
		ClientConfig:  &clientCfg,
		ServerConfigs: serverConfigs(cfg),
	}
}

func clientConfig(cfg configx.NacosConfig) constant.ClientConfig {
	timeout := cfg.TimeoutMs
	if timeout == 0 {
		timeout = 5000
	}
	return constant.ClientConfig{
		NamespaceId:         namespaceForSDK(cfg.NamespaceID),
		TimeoutMs:           timeout,
		NotLoadCacheAtStart: true,
		Username:            cfg.Username,
		Password:            cfg.Password,
	}
}

func serverConfigs(cfg configx.NacosConfig) []constant.ServerConfig {
	addr := strings.TrimSpace(cfg.ServerAddr)
	if addr == "" {
		addr = "127.0.0.1"
	}
	port := cfg.ServerPort
	if port == 0 {
		port = 8848
	}
	return []constant.ServerConfig{{IpAddr: addr, Port: port}}
}

func namespaceForSDK(namespace string) string {
	if namespace == "" || namespace == "public" {
		return ""
	}
	return namespace
}
