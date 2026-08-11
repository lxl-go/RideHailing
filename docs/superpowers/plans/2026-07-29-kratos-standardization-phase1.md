# Kratos Standardization Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first production-grade Kratos migration slice: shared Nacos support, `trip-service`, and `gateway-service`, while leaving `admin-platform` unchanged.

**Architecture:** Introduce `services/` as the new backend source of truth and migrate trip capability first because it spans passenger search and driver publishing. `trip-service` owns trip business/data logic and exposes Kratos HTTP/gRPC from proto; `gateway-service` exposes frontend-facing HTTP APIs and calls `trip-service` by Nacos discovery.

**Tech Stack:** Go 1.26, Kratos v2, Google Wire, protobuf, gRPC, Kratos HTTP transport, GORM, MySQL database `01ride`, Redis-ready config, Nacos config and registry/discovery.

## Global Constraints

- `admin-platform` remains unchanged.
- New backend services live under `services/`.
- Nacos is both configuration center and service registry/discovery.
- Nacos Data ID is `ride-car`, Group is `DEFAULT_GROUP`, Namespace is `public`, format is YAML.
- Each service has `configs/nacos.yaml` for Nacos connection and `configs/config.yaml` as complete local fallback.
- MySQL database name is `01ride`.
- Business services must not listen on or register `8080` or `8848`.
- Generated files are not edited manually.
- No git commit is required during this implementation.

---

## File Structure

Create:

- `pkg/configx/config.go`: shared Nacos-first config loader with local fallback.
- `pkg/configx/config_test.go`: loader tests for disabled Nacos, fallback, and fail-fast behavior.
- `pkg/registry/nacos.go`: Nacos registry/discovery helper used by Kratos apps.
- `pkg/registry/nacos_test.go`: validates Nacos config normalization without contacting a server.
- `services/trip-service/go.mod`: independent service module.
- `services/trip-service/Makefile`: generation and test commands.
- `services/trip-service/api/trip/v1/trip.proto`: source proto with HTTP annotations.
- `services/trip-service/internal/conf/conf.proto`: service config schema.
- `services/trip-service/configs/nacos.yaml`: Nacos connection example.
- `services/trip-service/configs/config.yaml`: complete fallback config using MySQL `01ride`.
- `services/trip-service/cmd/trip/main.go`: process entry.
- `services/trip-service/cmd/trip/app.go`: Kratos app constructor.
- `services/trip-service/cmd/trip/wire.go`: Wire injector.
- `services/trip-service/internal/server/server.go`: provider set.
- `services/trip-service/internal/server/http.go`: Kratos HTTP server.
- `services/trip-service/internal/server/grpc.go`: Kratos gRPC server.
- `services/trip-service/internal/service/service.go`: provider set.
- `services/trip-service/internal/service/trip.go`: generated API implementation.
- `services/trip-service/internal/biz/biz.go`: provider set.
- `services/trip-service/internal/biz/trip.go`: trip domain and use case.
- `services/trip-service/internal/biz/repo.go`: trip repo interface.
- `services/trip-service/internal/biz/errors.go`: domain errors.
- `services/trip-service/internal/biz/trip_test.go`: use case tests.
- `services/trip-service/internal/data/data.go`: database and snowflake providers.
- `services/trip-service/internal/data/trip.go`: GORM repo implementation.
- `services/trip-service/internal/data/trip_test.go`: repo tests with sqlite fallback.
- `services/gateway-service/go.mod`: independent gateway module.
- `services/gateway-service/Makefile`: generation and test commands.
- `services/gateway-service/api/gateway/v1/trip.proto`: frontend gateway trip API.
- `services/gateway-service/internal/conf/conf.proto`: gateway config schema.
- `services/gateway-service/configs/nacos.yaml`: Nacos connection example.
- `services/gateway-service/configs/config.yaml`: complete fallback config.
- `services/gateway-service/cmd/gateway/main.go`: process entry.
- `services/gateway-service/cmd/gateway/app.go`: Kratos app constructor.
- `services/gateway-service/cmd/gateway/wire.go`: Wire injector.
- `services/gateway-service/internal/server/server.go`: provider set.
- `services/gateway-service/internal/server/http.go`: Kratos HTTP server.
- `services/gateway-service/internal/service/service.go`: provider set.
- `services/gateway-service/internal/service/trip.go`: gateway API implementation.
- `services/gateway-service/internal/biz/biz.go`: provider set.
- `services/gateway-service/internal/biz/trip.go`: gateway trip use case.
- `services/gateway-service/internal/data/data.go`: gRPC client providers.
- `services/gateway-service/internal/data/trip_client.go`: trip client creation.
- `services/gateway-service/internal/biz/trip_test.go`: gateway use case tests.

Modify:

- `go.work`: add `./services/trip-service` and `./services/gateway-service`.
- `pkg/go.mod`: add Kratos/Nacos config dependencies needed by `pkg/configx` and `pkg/registry`.
- `passenger-platform/uni-app/src/utils/request.js`: switch base URL to configurable gateway URL with fallback `http://localhost:9000`.
- `driver-platform/uni-app/src/utils/request.js`: switch base URL to configurable gateway URL with fallback `http://localhost:9000` and use `X-User-Id` consistently.

Generated after implementation:

- `services/trip-service/api/trip/v1/*.pb.go`
- `services/trip-service/api/trip/v1/*_grpc.pb.go`
- `services/trip-service/api/trip/v1/*_http.pb.go`
- `services/trip-service/internal/conf/conf.pb.go`
- `services/trip-service/cmd/trip/wire_gen.go`
- `services/gateway-service/api/gateway/v1/*.pb.go`
- `services/gateway-service/api/gateway/v1/*_http.pb.go`
- `services/gateway-service/internal/conf/conf.pb.go`
- `services/gateway-service/cmd/gateway/wire_gen.go`

---

### Task 1: Shared Config Loader

**Files:**
- Create: `pkg/configx/config.go`
- Create: `pkg/configx/config_test.go`
- Modify: `pkg/go.mod`

**Interfaces:**
- Produces: `type NacosConfig struct`, `type LoadOptions struct`, `func LoadYAML(target any, opts LoadOptions) (Source, error)`, `type Source string`
- Consumes: local YAML files and optional injected remote loader

- [ ] **Step 1: Add failing tests for local config, fallback, and fail-fast**

Create `pkg/configx/config_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test ./pkg/configx
```

Expected: FAIL because package `pkg/configx` does not exist.

- [ ] **Step 3: Implement config loader**

Create `pkg/configx/config.go`:

```go
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

type nacosFile struct {
	Nacos NacosConfig `yaml:"nacos"`
}

type RemoteLoader func(NacosConfig) ([]byte, error)

type LoadOptions struct {
	NacosPath    string
	FallbackPath string
	RemoteLoader RemoteLoader
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
	clientConfig := constant.ClientConfig{
		NamespaceId:         namespaceForSDK(cfg.NamespaceID),
		TimeoutMs:           cfg.TimeoutMs,
		NotLoadCacheAtStart: true,
		Username:            cfg.Username,
		Password:            cfg.Password,
	}
	serverConfig := []constant.ServerConfig{{
		IpAddr: cfg.ServerAddr,
		Port:   cfg.ServerPort,
	}}
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"clientConfig":  clientConfig,
		"serverConfigs": serverConfig,
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
	if namespace == "public" {
		return ""
	}
	return namespace
}
```

- [ ] **Step 4: Add dependencies**

Run:

```bash
go get github.com/nacos-group/nacos-sdk-go/v2@v2.3.5 github.com/stretchr/testify@v1.11.1
```

Expected: `pkg/go.mod` contains Nacos SDK and Testify.

- [ ] **Step 5: Verify configx tests pass**

Run:

```bash
go test ./pkg/configx
```

Expected: PASS.

---

### Task 2: Shared Nacos Registry Helper

**Files:**
- Create: `pkg/registry/nacos.go`
- Create: `pkg/registry/nacos_test.go`
- Modify: `pkg/go.mod`

**Interfaces:**
- Consumes: `configx.NacosConfig`
- Produces: `func NewRegistrar(cfg configx.NacosConfig) (registry.Registrar, error)`, `func NewDiscovery(cfg configx.NacosConfig) (registry.Discovery, error)`

- [ ] **Step 1: Add tests for default normalization**

Create `pkg/registry/nacos_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test ./pkg/registry
```

Expected: FAIL because package `pkg/registry` does not exist.

- [ ] **Step 3: Implement registry helper**

Create `pkg/registry/nacos.go`:

```go
package registry

import (
	"strings"

	kratosregistry "github.com/go-kratos/kratos/v2/registry"
	nacosregistry "github.com/go-kratos/kratos/contrib/registry/nacos/v2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"

	"ride-hailing/pkg/configx"
)

func NewRegistrar(cfg configx.NacosConfig) (kratosregistry.Registrar, error) {
	client, err := clients.NewNamingClient(voMap(cfg))
	if err != nil {
		return nil, err
	}
	return nacosregistry.New(client), nil
}

func NewDiscovery(cfg configx.NacosConfig) (kratosregistry.Discovery, error) {
	client, err := clients.NewNamingClient(voMap(cfg))
	if err != nil {
		return nil, err
	}
	return nacosregistry.New(client), nil
}

func voMap(cfg configx.NacosConfig) map[string]interface{} {
	return map[string]interface{}{
		"clientConfig":  clientConfig(cfg),
		"serverConfigs": serverConfigs(cfg),
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
```

- [ ] **Step 4: Add dependencies**

Run:

```bash
go get github.com/go-kratos/kratos/v2@v2.9.2 github.com/go-kratos/kratos/contrib/registry/nacos/v2@latest
```

Expected: `pkg/go.mod` includes Kratos and Kratos Nacos registry contrib.

- [ ] **Step 5: Verify registry tests pass**

Run:

```bash
go test ./pkg/registry
```

Expected: PASS.

---

### Task 3: Trip Service Proto And Config

**Files:**
- Create: `services/trip-service/go.mod`
- Create: `services/trip-service/Makefile`
- Create: `services/trip-service/api/trip/v1/trip.proto`
- Create: `services/trip-service/internal/conf/conf.proto`
- Create: `services/trip-service/configs/nacos.yaml`
- Create: `services/trip-service/configs/config.yaml`
- Modify: `go.work`

**Interfaces:**
- Produces proto service `trip.v1.TripService`
- Produces config root `conf.Bootstrap`

- [ ] **Step 1: Create service module files**

Create `services/trip-service/go.mod`:

```go
module ride-hailing/services/trip-service

go 1.26

require (
	github.com/bwmarrin/snowflake v0.3.0
	github.com/go-kratos/kratos/v2 v2.9.2
	github.com/google/wire v0.7.0
	go.uber.org/zap v1.28.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.2
	ride-hailing/pkg v0.0.0-00010101000000-000000000000
)

replace ride-hailing/pkg => ../../pkg
```

Create `services/trip-service/Makefile`:

```makefile
.PHONY: api config wire test

api:
	protoc --proto_path=./api \
		--proto_path=../../third_party \
		--go_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		api/trip/v1/trip.proto

config:
	protoc --proto_path=./internal \
		--go_out=paths=source_relative:./internal \
		internal/conf/conf.proto

wire:
	cd cmd/trip && go run github.com/google/wire/cmd/wire

test:
	go test ./...
```

- [ ] **Step 2: Create trip proto with passenger and driver HTTP APIs**

Create `services/trip-service/api/trip/v1/trip.proto`:

```proto
syntax = "proto3";

package trip.v1;

import "google/api/annotations.proto";

option go_package = "ride-hailing/services/trip-service/api/trip/v1;tripv1";

service TripService {
  rpc SearchTrips(SearchTripsRequest) returns (SearchTripsReply) {
    option (google.api.http) = {
      get: "/v1/trips"
    };
  }

  rpc GetTripDetail(GetTripDetailRequest) returns (GetTripDetailReply) {
    option (google.api.http) = {
      get: "/v1/trips/{id}"
    };
  }

  rpc PublishTrip(PublishTripRequest) returns (PublishTripReply) {
    option (google.api.http) = {
      post: "/v1/driver/trips"
      body: "*"
    };
  }

  rpc ListDriverTrips(ListDriverTripsRequest) returns (ListDriverTripsReply) {
    option (google.api.http) = {
      get: "/v1/driver/trips"
    };
  }

  rpc UpdateTripStatus(UpdateTripStatusRequest) returns (UpdateTripStatusReply) {
    option (google.api.http) = {
      put: "/v1/driver/trips/{id}/status"
      body: "*"
    };
  }
}

message TripItem {
  int64 id = 1;
  int64 driver_id = 2;
  string origin = 3;
  string destination = 4;
  string depart_time = 5;
  string arrive_time = 6;
  int32 seats_total = 7;
  int32 seats_available = 8;
  double price = 9;
  int32 status = 10;
  string created_at = 11;
}

message SearchTripsRequest {
  string origin = 1;
  string destination = 2;
  string depart_date = 3;
  int32 page = 4;
  int32 page_size = 5;
}

message SearchTripsReply {
  int64 total = 1;
  repeated TripItem items = 2;
}

message GetTripDetailRequest {
  int64 id = 1;
}

message GetTripDetailReply {
  TripItem trip = 1;
}

message PublishTripRequest {
  string origin = 1;
  string destination = 2;
  string depart_time = 3;
  string arrive_time = 4;
  int32 seats_total = 5;
  double price = 6;
  int64 driver_id = 7;
}

message PublishTripReply {
  int64 trip_id = 1;
}

message ListDriverTripsRequest {
  int32 status = 1;
  int32 page = 2;
  int32 page_size = 3;
  int64 driver_id = 4;
}

message ListDriverTripsReply {
  int64 total = 1;
  repeated TripItem items = 2;
}

message UpdateTripStatusRequest {
  int64 id = 1;
  int32 status = 2;
}

message UpdateTripStatusReply {}
```

- [ ] **Step 3: Create config proto and local config files**

Create `services/trip-service/internal/conf/conf.proto`:

```proto
syntax = "proto3";

package trip.internal.conf;

option go_package = "ride-hailing/services/trip-service/internal/conf;conf";

message Bootstrap {
  Server server = 1;
  Data data = 2;
  Registry registry = 3;
}

message Server {
  message HTTP {
    string addr = 1;
    string timeout = 2;
  }
  message GRPC {
    string addr = 1;
    string timeout = 2;
  }
  HTTP http = 1;
  GRPC grpc = 2;
  int64 snowflake_node = 3;
}

message Data {
  message Database {
    string driver = 1;
    string source = 2;
  }
  Database database = 1;
}

message Registry {
  message Nacos {
    bool enabled = 1;
    string server_addr = 2;
    uint64 server_port = 3;
    string namespace_id = 4;
    string group = 5;
    string username = 6;
    string password = 7;
    uint64 timeout_ms = 8;
  }
  Nacos nacos = 1;
}
```

Create `services/trip-service/configs/nacos.yaml`:

```yaml
nacos:
  enabled: true
  server_addr: 127.0.0.1
  server_port: 8848
  namespace_id: public
  group: DEFAULT_GROUP
  data_id: ride-car
  username: nacos
  password: nacos
  timeout_ms: 5000
  fail_fast: false
```

Create `services/trip-service/configs/config.yaml`:

```yaml
server:
  http:
    addr: 0.0.0.0:9040
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9140
    timeout: 1s
  snowflake_node: 31
data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/01ride?charset=utf8mb4&parseTime=True&loc=Local
registry:
  nacos:
    enabled: true
    server_addr: 127.0.0.1
    server_port: 8848
    namespace_id: public
    group: DEFAULT_GROUP
    username: nacos
    password: nacos
    timeout_ms: 5000
```

- [ ] **Step 4: Update go.work**

Modify `go.work` and add:

```go
	./services/trip-service
```

inside the `use (...)` block.

- [ ] **Step 5: Run generation commands**

Run:

```bash
cd services/trip-service
make api
make config
```

Expected: generated proto and conf files exist.

---

### Task 4: Trip Service Business And Data Layers

**Files:**
- Create: `services/trip-service/internal/biz/biz.go`
- Create: `services/trip-service/internal/biz/trip.go`
- Create: `services/trip-service/internal/biz/repo.go`
- Create: `services/trip-service/internal/biz/errors.go`
- Create: `services/trip-service/internal/biz/trip_test.go`
- Create: `services/trip-service/internal/data/data.go`
- Create: `services/trip-service/internal/data/trip.go`
- Create: `services/trip-service/internal/data/trip_test.go`

**Interfaces:**
- Produces: `TripUsecase` with `SearchTrips`, `GetTripDetail`, `PublishTrip`, `ListDriverTrips`, `UpdateTripStatus`
- Produces: `TripRepo` implementation backed by GORM

- [ ] **Step 1: Write use case tests**

Create `services/trip-service/internal/biz/trip_test.go`:

```go
package biz

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeTripRepo struct {
	items []Trip
}

func (r *fakeTripRepo) SearchTrips(context.Context, string, string, string, int, int) ([]Trip, int64, error) {
	return r.items, int64(len(r.items)), nil
}

func (r *fakeTripRepo) GetByID(_ context.Context, id int64) (*Trip, error) {
	for _, item := range r.items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, ErrTripNotFound
}

func (r *fakeTripRepo) Create(_ context.Context, trip *Trip) error {
	r.items = append(r.items, *trip)
	return nil
}

func (r *fakeTripRepo) ListByDriver(_ context.Context, driverID int64, _ int, _ int, _ int) ([]Trip, int64, error) {
	var out []Trip
	for _, item := range r.items {
		if item.DriverID == driverID {
			out = append(out, item)
		}
	}
	return out, int64(len(out)), nil
}

func (r *fakeTripRepo) UpdateStatus(_ context.Context, id int64, status int) error {
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Status = status
			return nil
		}
	}
	return ErrTripNotFound
}

func TestPublishTripDefaultsSeatsAndStatus(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	repo := &fakeTripRepo{}
	uc := NewTripUsecase(node, zap.NewNop(), repo)

	trip, err := uc.PublishTrip(context.Background(), PublishTripCommand{
		DriverID:    2001,
		Origin:      "A",
		Destination: "B",
		DepartTime:  time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC),
		ArriveTime:  time.Date(2026, 7, 29, 11, 0, 0, 0, time.UTC),
		SeatsTotal:  3,
		Price:       19.9,
	})

	require.NoError(t, err)
	require.NotZero(t, trip.ID)
	require.Equal(t, 3, trip.SeatsAvailable)
	require.Equal(t, TripStatusRecruiting, trip.Status)
	require.Len(t, repo.items, 1)
}
```

- [ ] **Step 2: Implement biz layer**

Create `services/trip-service/internal/biz/errors.go`:

```go
package biz

import "errors"

var ErrTripNotFound = errors.New("trip not found")
var ErrInvalidTrip = errors.New("invalid trip")
```

Create `services/trip-service/internal/biz/repo.go`:

```go
package biz

import "context"

type TripRepo interface {
	SearchTrips(ctx context.Context, origin, destination, departDate string, page, pageSize int) ([]Trip, int64, error)
	GetByID(ctx context.Context, id int64) (*Trip, error)
	Create(ctx context.Context, trip *Trip) error
	ListByDriver(ctx context.Context, driverID int64, status int, page, pageSize int) ([]Trip, int64, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
}
```

Create `services/trip-service/internal/biz/trip.go`:

```go
package biz

import (
	"context"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
)

const TripStatusRecruiting = 1

type Trip struct {
	ID             int64
	DriverID       int64
	Origin         string
	Destination    string
	DepartTime     time.Time
	ArriveTime     time.Time
	SeatsTotal     int
	SeatsAvailable int
	Price          float64
	Status         int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PublishTripCommand struct {
	DriverID    int64
	Origin      string
	Destination string
	DepartTime  time.Time
	ArriveTime  time.Time
	SeatsTotal  int
	Price       float64
}

type TripUsecase struct {
	node *snowflake.Node
	log  *zap.Logger
	repo TripRepo
}

func NewTripUsecase(node *snowflake.Node, log *zap.Logger, repo TripRepo) *TripUsecase {
	return &TripUsecase{node: node, log: log, repo: repo}
}

func (uc *TripUsecase) SearchTrips(ctx context.Context, origin, destination, departDate string, page, pageSize int) ([]Trip, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.SearchTrips(ctx, strings.TrimSpace(origin), strings.TrimSpace(destination), strings.TrimSpace(departDate), page, pageSize)
}

func (uc *TripUsecase) GetTripDetail(ctx context.Context, id int64) (*Trip, error) {
	if id <= 0 {
		return nil, ErrInvalidTrip
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *TripUsecase) PublishTrip(ctx context.Context, cmd PublishTripCommand) (*Trip, error) {
	if cmd.DriverID <= 0 || strings.TrimSpace(cmd.Origin) == "" || strings.TrimSpace(cmd.Destination) == "" || cmd.SeatsTotal <= 0 || cmd.Price < 0 {
		return nil, ErrInvalidTrip
	}
	trip := &Trip{
		ID:             uc.node.Generate().Int64(),
		DriverID:       cmd.DriverID,
		Origin:         strings.TrimSpace(cmd.Origin),
		Destination:    strings.TrimSpace(cmd.Destination),
		DepartTime:     cmd.DepartTime,
		ArriveTime:     cmd.ArriveTime,
		SeatsTotal:     cmd.SeatsTotal,
		SeatsAvailable: cmd.SeatsTotal,
		Price:          cmd.Price,
		Status:         TripStatusRecruiting,
	}
	if err := uc.repo.Create(ctx, trip); err != nil {
		uc.log.Error("publish trip failed", zap.Error(err))
		return nil, err
	}
	return trip, nil
}

func (uc *TripUsecase) ListDriverTrips(ctx context.Context, driverID int64, status int, page, pageSize int) ([]Trip, int64, error) {
	if driverID <= 0 {
		return nil, 0, ErrInvalidTrip
	}
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.ListByDriver(ctx, driverID, status, page, pageSize)
}

func (uc *TripUsecase) UpdateTripStatus(ctx context.Context, id int64, status int) error {
	if id <= 0 || status <= 0 {
		return ErrInvalidTrip
	}
	return uc.repo.UpdateStatus(ctx, id, status)
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
```

Create `services/trip-service/internal/biz/biz.go`:

```go
package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewTripUsecase)
```

- [ ] **Step 3: Implement data layer**

Create `services/trip-service/internal/data/data.go`:

```go
package data

import (
	"github.com/bwmarrin/snowflake"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ride-hailing/services/trip-service/internal/biz"
	"ride-hailing/services/trip-service/internal/conf"
)

type Data struct {
	DB *gorm.DB
}

func NewData(db *gorm.DB) *Data {
	return &Data{DB: db}
}

func NewDB(c *conf.Data, logger *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&tripModel{}); err != nil {
		return nil, err
	}
	return db, nil
}

func NewSnowflakeNode(c *conf.Server) (*snowflake.Node, error) {
	return snowflake.NewNode(c.SnowflakeNode)
}

var ProviderSet = wire.NewSet(
	NewDB,
	NewData,
	NewSnowflakeNode,
	NewTripRepo,
	wire.Bind(new(biz.TripRepo), new(*TripRepo)),
)
```

Create `services/trip-service/internal/data/trip.go` by adapting the existing passenger/driver `tripdata` repos and include search, create, list, and status update methods against table `carpool_trip`.

- [ ] **Step 4: Run biz tests**

Run:

```bash
cd services/trip-service
go test ./internal/biz
```

Expected: PASS.

---

### Task 5: Trip Service Kratos App

**Files:**
- Create: `services/trip-service/internal/service/service.go`
- Create: `services/trip-service/internal/service/trip.go`
- Create: `services/trip-service/internal/server/server.go`
- Create: `services/trip-service/internal/server/http.go`
- Create: `services/trip-service/internal/server/grpc.go`
- Create: `services/trip-service/cmd/trip/app.go`
- Create: `services/trip-service/cmd/trip/main.go`
- Create: `services/trip-service/cmd/trip/wire.go`

**Interfaces:**
- Consumes: `TripUsecase`, generated `tripv1.TripServiceServer`
- Produces: runnable Kratos app with HTTP `9040`, gRPC `9140`, Nacos registration

- [ ] **Step 1: Implement service adapter**

Create `services/trip-service/internal/service/trip.go`:

```go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/transport/grpc/status"
	"google.golang.org/grpc/codes"

	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
	"ride-hailing/services/trip-service/internal/biz"
)

type TripService struct {
	tripv1.UnimplementedTripServiceServer
	uc *biz.TripUsecase
}

func NewTripService(uc *biz.TripUsecase) *TripService {
	return &TripService{uc: uc}
}

func (s *TripService) SearchTrips(ctx context.Context, req *tripv1.SearchTripsRequest) (*tripv1.SearchTripsReply, error) {
	items, total, err := s.uc.SearchTrips(ctx, req.Origin, req.Destination, req.DepartDate, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.SearchTripsReply{Total: total, Items: tripsToProto(items)}, nil
}

func (s *TripService) GetTripDetail(ctx context.Context, req *tripv1.GetTripDetailRequest) (*tripv1.GetTripDetailReply, error) {
	item, err := s.uc.GetTripDetail(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.GetTripDetailReply{Trip: tripToProto(item)}, nil
}

func (s *TripService) PublishTrip(ctx context.Context, req *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error) {
	depart, err := parseTime(req.DepartTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid depart_time")
	}
	arrive, err := parseTime(req.ArriveTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid arrive_time")
	}
	item, err := s.uc.PublishTrip(ctx, biz.PublishTripCommand{
		DriverID:    req.DriverId,
		Origin:      req.Origin,
		Destination: req.Destination,
		DepartTime:  depart,
		ArriveTime:  arrive,
		SeatsTotal:  int(req.SeatsTotal),
		Price:       req.Price,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.PublishTripReply{TripId: item.ID}, nil
}

func (s *TripService) ListDriverTrips(ctx context.Context, req *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error) {
	items, total, err := s.uc.ListDriverTrips(ctx, req.DriverId, int(req.Status), int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.ListDriverTripsReply{Total: total, Items: tripsToProto(items)}, nil
}

func (s *TripService) UpdateTripStatus(ctx context.Context, req *tripv1.UpdateTripStatusRequest) (*tripv1.UpdateTripStatusReply, error) {
	if err := s.uc.UpdateTripStatus(ctx, req.Id, int(req.Status)); err != nil {
		return nil, mapError(err)
	}
	return &tripv1.UpdateTripStatusReply{}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidTrip):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrTripNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func tripsToProto(items []biz.Trip) []*tripv1.TripItem {
	out := make([]*tripv1.TripItem, len(items))
	for i := range items {
		out[i] = tripToProto(&items[i])
	}
	return out
}

func tripToProto(item *biz.Trip) *tripv1.TripItem {
	if item == nil {
		return nil
	}
	return &tripv1.TripItem{
		Id:             item.ID,
		DriverId:       item.DriverID,
		Origin:         item.Origin,
		Destination:    item.Destination,
		DepartTime:     formatTime(item.DepartTime),
		ArriveTime:     formatTime(item.ArriveTime),
		SeatsTotal:     int32(item.SeatsTotal),
		SeatsAvailable: int32(item.SeatsAvailable),
		Price:          item.Price,
		Status:         int32(item.Status),
		CreatedAt:      formatTime(item.CreatedAt),
	}
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
```

Create `services/trip-service/internal/service/service.go`:

```go
package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewTripService)
```

- [ ] **Step 2: Implement servers and app**

Create Kratos HTTP and gRPC servers that register `tripv1.RegisterTripServiceHTTPServer` and `tripv1.RegisterTripServiceServer`. Use configured addresses `9040` and `9140`.

- [ ] **Step 3: Implement main with Nacos-first config**

`cmd/trip/main.go` must:

1. Load `configs/nacos.yaml`.
2. Load Nacos `ride-car` config or fallback `configs/config.yaml`.
3. Create zap logger.
4. Create Nacos registrar from `registry.nacos`.
5. Call `initApp`.
6. Run the app.

- [ ] **Step 4: Generate Wire**

Run:

```bash
cd services/trip-service
make wire
```

Expected: `cmd/trip/wire_gen.go` exists.

- [ ] **Step 5: Test trip service**

Run:

```bash
cd services/trip-service
go test ./...
```

Expected: PASS.

---

### Task 6: Gateway Service Skeleton And Trip Client

**Files:**
- Create all `services/gateway-service/...` files listed in File Structure
- Modify: `go.work`

**Interfaces:**
- Consumes: `trip-service` generated gRPC client
- Produces: HTTP gateway API compatible with current uni-app paths:
  - `GET /carpool/trips`
  - `GET /carpool/trips/{id}`
  - `POST /carpool/trips`
  - `GET /carpool/trips/mine`
  - `PUT /carpool/trips/{id}/status`

- [ ] **Step 1: Create gateway module and proto**

Create `services/gateway-service/go.mod` with module `ride-hailing/services/gateway-service`, Kratos, Wire, and replace directives for `ride-hailing/pkg` and `ride-hailing/services/trip-service`.

Create `api/gateway/v1/trip.proto` exposing the five HTTP routes above with `google.api.http` annotations.

- [ ] **Step 2: Create gateway fallback config**

Create `services/gateway-service/configs/config.yaml`:

```yaml
server:
  http:
    addr: 0.0.0.0:9000
    timeout: 1s
registry:
  nacos:
    enabled: true
    server_addr: 127.0.0.1
    server_port: 8848
    namespace_id: public
    group: DEFAULT_GROUP
    username: nacos
    password: nacos
    timeout_ms: 5000
clients:
  trip:
    endpoint: discovery:///trip-service
```

- [ ] **Step 3: Implement trip client provider**

Create a data provider that creates `tripv1.TripServiceClient` using Kratos gRPC client with Nacos discovery and endpoint `discovery:///trip-service`.

- [ ] **Step 4: Implement gateway trip service**

Map current frontend request semantics to `trip-service`:

- passenger search calls `SearchTrips`
- passenger detail calls `GetTripDetail`
- driver publish calls `PublishTrip`
- driver my trips calls `ListDriverTrips`
- driver status update calls `UpdateTripStatus`

For temporary local development, extract identity from `X-User-Id`. Do not use `X-Driver-Id`.

- [ ] **Step 5: Generate and test**

Run:

```bash
cd services/gateway-service
make api
make config
make wire
go test ./...
```

Expected: PASS.

---

### Task 7: Uni-App Gateway Routing Cleanup

**Files:**
- Modify: `passenger-platform/uni-app/src/utils/request.js`
- Modify: `driver-platform/uni-app/src/utils/request.js`

**Interfaces:**
- Consumes: gateway HTTP `http://localhost:9000`
- Produces: consistent frontend requests using `X-User-Id`

- [ ] **Step 1: Update passenger request base URL**

Modify `passenger-platform/uni-app/src/utils/request.js`:

```js
const BASE_URL = import.meta?.env?.VITE_GATEWAY_API || 'http://localhost:9000'
```

Keep header:

```js
'X-User-Id': getUserId()
```

- [ ] **Step 2: Update driver request base URL and identity header**

Modify `driver-platform/uni-app/src/utils/request.js`:

```js
const BASE_URL = import.meta?.env?.VITE_GATEWAY_API || 'http://localhost:9000'
```

Change header from:

```js
'X-Driver-Id': getDriverId()
```

to:

```js
'X-User-Id': getDriverId()
```

- [ ] **Step 3: Verify no old uni-app gateway ports remain**

Run:

```bash
rg "localhost:8081|localhost:8082|X-Driver-Id" passenger-platform/uni-app/src driver-platform/uni-app/src
```

Expected: no matches.

---

### Task 8: Phase 1 Verification

**Files:**
- Read only unless failures require small fixes in files created above.

**Interfaces:**
- Verifies first Kratos migration slice.

- [ ] **Step 1: Run shared package tests**

Run:

```bash
go test ./pkg/configx ./pkg/registry
```

Expected: PASS.

- [ ] **Step 2: Run trip service tests**

Run:

```bash
cd services/trip-service
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Run gateway service tests**

Run:

```bash
cd services/gateway-service
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Verify forbidden ports are not service listen ports**

Run:

```bash
rg "0.0.0.0:8080|:8080|0.0.0.0:8848" services
```

Expected: no service listener matches. `server_port: 8848` inside `nacos.yaml` is allowed because it is the Nacos server port.

- [ ] **Step 5: Verify MySQL database name**

Run:

```bash
rg "01ride" services
```

Expected: service fallback configs contain MySQL DSNs using `01ride`.

- [ ] **Step 6: Verify service discovery endpoint**

Run:

```bash
rg "127.0.0.1:9[0-9]{3}|discovery:///trip-service" services/gateway-service
```

Expected: gateway uses `discovery:///trip-service` and does not hard-code a trip gRPC port.

---

## Follow-up Plans

After Phase 1 passes:

- Phase 2 migrates `order-service` and `review-service`.
- Phase 3 adds passenger/driver identity and auth through `user-service`, `passenger-service`, and `driver-service`.
- Phase 4 deprecates and removes `passenger-platform/service` and `driver-platform/service` after equivalent flows pass through `gateway-service`.
