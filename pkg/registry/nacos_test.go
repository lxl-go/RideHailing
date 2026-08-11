package registry

import (
	"testing"

	"github.com/stretchr/testify/require"

	"ride-hailing/pkg/configx"
)

func TestServerConfigsUseDefaults(t *testing.T) {
	cfg := configx.NacosConfig{}
	servers := serverConfigs(cfg)

	require.Len(t, servers, 1)
	require.Equal(t, "127.0.0.1", servers[0].IpAddr)
	require.Equal(t, uint64(8848), servers[0].Port)
}

func TestClientConfigMapsPublicNamespaceToEmpty(t *testing.T) {
	cfg := configx.NacosConfig{
		NamespaceID: "public",
		TimeoutMs:   3000,
		Username:    "nacos",
		Password:    "nacos",
	}
	client := clientConfig(cfg)

	require.Equal(t, "", client.NamespaceId)
	require.Equal(t, uint64(3000), client.TimeoutMs)
	require.Equal(t, "nacos", client.Username)
	require.Equal(t, "nacos", client.Password)
}
