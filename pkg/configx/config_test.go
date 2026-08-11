package configx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type sampleConfig struct {
	Server struct {
		HTTP struct {
			Addr string `yaml:"addr"`
		} `yaml:"http"`
	} `yaml:"server"`
}

func writeFile(t *testing.T, path string, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

func TestLoadYAMLUsesFallbackWhenNacosDisabled(t *testing.T) {
	dir := t.TempDir()
	nacosPath := filepath.Join(dir, "nacos.yaml")
	fallbackPath := filepath.Join(dir, "config.yaml")
	writeFile(t, nacosPath, "nacos:\n  enabled: false\n")
	writeFile(t, fallbackPath, "server:\n  http:\n    addr: 0.0.0.0:9000\n")

	var cfg sampleConfig
	source, err := LoadYAML(&cfg, LoadOptions{
		NacosPath:    nacosPath,
		FallbackPath: fallbackPath,
	})

	require.NoError(t, err)
	require.Equal(t, SourceFallback, source)
	require.Equal(t, "0.0.0.0:9000", cfg.Server.HTTP.Addr)
}

func TestLoadYAMLFallsBackWhenRemoteFailsAndFailFastFalse(t *testing.T) {
	dir := t.TempDir()
	nacosPath := filepath.Join(dir, "nacos.yaml")
	fallbackPath := filepath.Join(dir, "config.yaml")
	writeFile(t, nacosPath, "nacos:\n  enabled: true\n  fail_fast: false\n  data_id: ride-car\n")
	writeFile(t, fallbackPath, "server:\n  http:\n    addr: 0.0.0.0:9040\n")

	var cfg sampleConfig
	source, err := LoadYAML(&cfg, LoadOptions{
		NacosPath:    nacosPath,
		FallbackPath: fallbackPath,
		RemoteLoader: func(NacosConfig) ([]byte, error) {
			return nil, errors.New("nacos offline")
		},
	})

	require.NoError(t, err)
	require.Equal(t, SourceFallback, source)
	require.Equal(t, "0.0.0.0:9040", cfg.Server.HTTP.Addr)
}

func TestLoadYAMLReturnsErrorWhenRemoteFailsAndFailFastTrue(t *testing.T) {
	dir := t.TempDir()
	nacosPath := filepath.Join(dir, "nacos.yaml")
	fallbackPath := filepath.Join(dir, "config.yaml")
	writeFile(t, nacosPath, "nacos:\n  enabled: true\n  fail_fast: true\n  data_id: ride-car\n")
	writeFile(t, fallbackPath, "server:\n  http:\n    addr: 0.0.0.0:9040\n")

	var cfg sampleConfig
	source, err := LoadYAML(&cfg, LoadOptions{
		NacosPath:    nacosPath,
		FallbackPath: fallbackPath,
		RemoteLoader: func(NacosConfig) ([]byte, error) {
			return nil, errors.New("nacos offline")
		},
	})

	require.Error(t, err)
	require.Equal(t, SourceNone, source)
	require.Contains(t, err.Error(), "load nacos config")
}
