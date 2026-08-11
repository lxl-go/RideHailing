package nacosx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type Config struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
	FailFast   bool   `mapstructure:"fail-fast" yaml:"fail-fast"`
	ServerAddr string `mapstructure:"server-addr" yaml:"server-addr"`
	Namespace  string `mapstructure:"namespace" yaml:"namespace"`
	DataID     string `mapstructure:"data-id" yaml:"data-id"`
	Group      string `mapstructure:"group" yaml:"group"`
	Username   string `mapstructure:"username" yaml:"username"`
	Password   string `mapstructure:"password" yaml:"password"`
}

func parseServerAddr(addr string) (ip string, port uint64, err error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid server-addr format %q, expect ip:port", addr)
	}
	ip = parts[0]
	p, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in server-addr %q: %w", addr, err)
	}
	return ip, p, nil
}

func LoadConfig(cfg *Config) ([]byte, error) {
	ip, port, err := parseServerAddr(cfg.ServerAddr)
	if err != nil {
		return nil, fmt.Errorf("nacos config: %w", err)
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(ip, port),
	}

	opts := []constant.ClientOption{
		constant.WithNamespaceId(cfg.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithLogDir("./log/nacos"),
		constant.WithCacheDir("./cache/nacos"),
		constant.WithLogLevel("info"),
	}
	if cfg.Username != "" {
		opts = append(opts, constant.WithUsername(cfg.Username))
	}
	if cfg.Password != "" {
		opts = append(opts, constant.WithPassword(cfg.Password))
	}
	cc := *constant.NewClientConfig(opts...)

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("nacos create config client failed: %w", err)
	}

	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	})
	if err != nil {
		return nil, fmt.Errorf("nacos get config failed: %w", err)
	}

	return []byte(content), nil
}
