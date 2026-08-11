package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYAML 加载 YAML 配置文件（对标文档 config 通用工具）
func LoadYAML(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

// LoadYAMLWithEnv 加载 YAML 并支持环境变量覆盖（对标文档 Nacos 降级策略）
// envPrefix 为环境变量前缀，如 "PASSENGER" 对应 PASSENGER_PORT 覆盖 cfg.Server.Port
func LoadYAMLWithEnv(path, envPrefix string, out interface{}) error {
	if err := LoadYAML(path, out); err != nil {
		return err
	}
	// 环境变量覆盖（简单实现，复杂场景建议 Viper）
	if port := os.Getenv(envPrefix + "_PORT"); port != "" {
		if m, ok := out.(map[string]interface{}); ok {
			m["port"] = port
		}
	}
	return nil
}

// GetEnv 读取环境变量，带默认值（对标文档配置回退链）
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// MustGetEnv 读取环境变量，空值时 panic
func MustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required env " + key + " not set")
	}
	return v
}
