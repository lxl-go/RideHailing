package conf

type Bootstrap struct {
	Server   *Server   `yaml:"server"`
	Data     *Data     `yaml:"data"`
	RealName *RealName `yaml:"real_name"`
	Registry *Registry `yaml:"registry"`
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

type RealName struct {
	Tencent *RealName_Tencent `yaml:"tencent"`
}

type RealName_Tencent struct {
	Enabled   bool   `yaml:"enabled"`
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Endpoint  string `yaml:"endpoint"`
	Timeout   string `yaml:"timeout"`
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
