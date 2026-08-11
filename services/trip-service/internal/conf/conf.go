package conf

type Bootstrap struct {
	Server   *Server   `yaml:"server"`
	Data     *Data     `yaml:"data"`
	Redis    *Redis    `yaml:"redis"`
	AMap     *AMap     `yaml:"amap"`
	Registry *Registry `yaml:"registry"`
}

// Redis contains local development connection and lock defaults.
type Redis struct {
	Addr         string     `yaml:"addr"`
	Username     string     `yaml:"username"`
	Password     string     `yaml:"password"`
	DB           int        `yaml:"db"`
	PoolSize     int        `yaml:"pool_size"`
	DialTimeout  string     `yaml:"dial_timeout"`
	ReadTimeout  string     `yaml:"read_timeout"`
	WriteTimeout string     `yaml:"write_timeout"`
	Lock         *RedisLock `yaml:"lock"`
}

type RedisLock struct {
	KeyPrefix  string `yaml:"key_prefix"`
	DefaultTTL string `yaml:"default_ttl"`
}

type AMap struct {
	WebServiceKey string `yaml:"web_service_key"`
	BaseURL       string `yaml:"base_url"`
	Timeout       string `yaml:"timeout"`
}

type Server struct {
	Http          *Server_HTTP `yaml:"http"`
	Grpc          *Server_GRPC `yaml:"grpc"`
	SnowflakeNode int64        `yaml:"snowflake_node"`
}

type Server_HTTP struct {
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

type Server_GRPC struct {
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

type Data struct {
	Database *Data_Database `yaml:"database"`
}

type Data_Database struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type Registry struct {
	Nacos *Registry_Nacos `yaml:"nacos"`
}

type Registry_Nacos struct {
	Enabled     bool   `yaml:"enabled"`
	ServerAddr  string `yaml:"server_addr"`
	ServerPort  uint64 `yaml:"server_port"`
	NamespaceId string `yaml:"namespace_id"`
	Group       string `yaml:"group"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	TimeoutMs   uint64 `yaml:"timeout_ms"`
}
