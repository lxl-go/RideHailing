package config

import (
	"fmt"
	"os"
	"path/filepath"

	"ride-hailing/admin-server/pkg/nacosx"

	"gopkg.in/yaml.v3"
)

type NacosCfg struct {
	Nacos nacosx.Config `mapstructure:"nacos" yaml:"nacos"`
}

// LoadConfig 读取本地nacos.yaml → 判断enabled → 返回远端或本地YAML字节
func LoadConfig(configDir string) ([]byte, string, error) {
	nacosPath := filepath.Join(configDir, "nacos.yaml")
	localPath := filepath.Join(configDir, "config.yaml")

	nacosBytes, err := os.ReadFile(nacosPath)
	if err != nil {
		return nil, "", fmt.Errorf("read nacos.yaml failed: %w", err)
	}

	var nc NacosCfg
	if err := yaml.Unmarshal(nacosBytes, &nc); err != nil {
		return nil, "", fmt.Errorf("parse nacos.yaml failed: %w", err)
	}

	if !nc.Nacos.Enabled {
		data, err := os.ReadFile(localPath)
		if err != nil {
			return nil, "", fmt.Errorf("read local config.yaml failed: %w", err)
		}
		return data, localPath, nil
	}

	data, err := nacosx.LoadConfig(&nc.Nacos)
	if err != nil {
		if nc.Nacos.FailFast {
			return nil, "", fmt.Errorf("nacos load config (fail-fast): %w", err)
		}
		data, err2 := os.ReadFile(localPath)
		if err2 != nil {
			return nil, "", fmt.Errorf("nacos failed and local fallback also failed: %w (nacos: %v)", err2, err)
		}
		return data, localPath, nil
	}

	return data, "nacos:" + nc.Nacos.DataID + "@" + nc.Nacos.Group, nil
}
