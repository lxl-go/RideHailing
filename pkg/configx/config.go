package configx

import (
	"fmt"
	"os"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

type Source string

const (
	SourceNone     Source = ""
	SourceNacos    Source = "nacos"
	SourceFallback Source = "fallback"
)

type NacosConfig struct {
	Enabled     bool   `yaml:"enabled"`
	ServerAddr  string `yaml:"server_addr"`
	ServerPort  uint64 `yaml:"server_port"`
	NamespaceID string `yaml:"namespace_id"`
	Group       string `yaml:"group"`
	DataID      string `yaml:"data_id"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	TimeoutMs   uint64 `yaml:"timeout_ms"`
	FailFast    bool   `yaml:"fail_fast"`
}

type LoadOptions struct {
	NacosPath    string
	FallbackPath string
	RemoteLoader RemoteLoader
}

type RemoteLoader func(NacosConfig) ([]byte, error)

type nacosFile struct {
	Nacos NacosConfig `yaml:"nacos"`
}

func LoadYAML(target any, opts LoadOptions) (Source, error) {
	nacosCfg, err := readNacosFile(opts.NacosPath)
	if err != nil {
		return SourceNone, err
	}
	nacosCfg = withDefaults(nacosCfg)
	if !nacosCfg.Enabled {
		if err := readYAMLFile(opts.FallbackPath, target); err != nil {
			return SourceNone, err
		}
		return SourceFallback, nil
	}

	loader := opts.RemoteLoader
	if loader == nil {
		loader = LoadRemoteYAML
	}
	body, err := loader(nacosCfg)
	if err != nil {
		if nacosCfg.FailFast {
			return SourceNone, fmt.Errorf("load nacos config: %w", err)
		}
		if err := readYAMLFile(opts.FallbackPath, target); err != nil {
			return SourceNone, err
		}
		return SourceFallback, nil
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		if nacosCfg.FailFast {
			return SourceNone, fmt.Errorf("load nacos config: empty config")
		}
		if err := readYAMLFile(opts.FallbackPath, target); err != nil {
			return SourceNone, err
		}
		return SourceFallback, nil
	}
	if err := yaml.Unmarshal(body, target); err != nil {
		return SourceNone, fmt.Errorf("parse nacos yaml: %w", err)
	}
	return SourceNacos, nil
}

func LoadRemoteYAML(cfg NacosConfig) ([]byte, error) {
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"clientConfig": constant.ClientConfig{
			NamespaceId:         namespaceForSDK(cfg.NamespaceID),
			TimeoutMs:           cfg.TimeoutMs,
			NotLoadCacheAtStart: true,
			Username:            cfg.Username,
			Password:            cfg.Password,
		},
		"serverConfigs": []constant.ServerConfig{{
			IpAddr: cfg.ServerAddr,
			Port:   cfg.ServerPort,
		}},
	})
	if err != nil {
		return nil, err
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	})
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func readNacosFile(path string) (NacosConfig, error) {
	if strings.TrimSpace(path) == "" {
		path = "configs/nacos.yaml"
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return NacosConfig{}, err
	}
	var file nacosFile
	if err := yaml.Unmarshal(body, &file); err != nil {
		return NacosConfig{}, err
	}
	return file.Nacos, nil
}

func readYAMLFile(path string, target any) error {
	if strings.TrimSpace(path) == "" {
		path = "configs/config.yaml"
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(body, target); err != nil {
		return fmt.Errorf("parse fallback yaml: %w", err)
	}
	return nil
}

func withDefaults(cfg NacosConfig) NacosConfig {
	if strings.TrimSpace(cfg.ServerAddr) == "" {
		cfg.ServerAddr = "127.0.0.1"
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = 8848
	}
	if strings.TrimSpace(cfg.NamespaceID) == "" {
		cfg.NamespaceID = "public"
	}
	if strings.TrimSpace(cfg.Group) == "" {
		cfg.Group = "DEFAULT_GROUP"
	}
	if strings.TrimSpace(cfg.DataID) == "" {
		cfg.DataID = "ride-car"
	}
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = 5000
	}
	return cfg
}

func namespaceForSDK(namespace string) string {
	if namespace == "" || namespace == "public" {
		return ""
	}
	return namespace
}
