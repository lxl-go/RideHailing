package main

import (
	"fmt"
	"os"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	// Read config content
	configData, err := os.ReadFile("../../config.yaml")
	if err != nil {
		panic(fmt.Errorf("read config.yaml: %w", err))
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}

	cc := *constant.NewClientConfig(
		constant.WithNamespaceId("public"),
		constant.WithTimeoutMs(5000),
		constant.WithLogDir("./log"),
		constant.WithCacheDir("./cache"),
		constant.WithLogLevel("info"),
		constant.WithUsername("nacos"),
		constant.WithPassword("nacos"),
	)

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(fmt.Errorf("create client: %w", err))
	}

	pub, err := client.PublishConfig(vo.ConfigParam{
		DataId:  "ride-car",
		Group:   "DEFAULT_GROUP",
		Content: string(configData),
		Type:    "yaml",
	})
	if err != nil {
		panic(fmt.Errorf("publish config: %w", err))
	}

	fmt.Printf("Published: %v (len=%d)\n", pub, len(configData))
}
