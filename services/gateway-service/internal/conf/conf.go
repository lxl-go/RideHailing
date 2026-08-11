package conf

type Bootstrap struct {
	Server   *Server   `yaml:"server"`
	Registry *Registry `yaml:"registry"`
	Clients  *Clients  `yaml:"clients"`
	Auth     *Auth     `yaml:"auth"`
	Alipay   *Alipay   `yaml:"alipay"`
	Amap     *Amap     `yaml:"amap"`
}

type Server struct {
	Http *Server_HTTP `yaml:"http"`
}

type Server_HTTP struct {
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
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

type Clients struct {
	Auth      *Clients_Auth      `yaml:"auth"`
	Trip      *Clients_Trip      `yaml:"trip"`
	Order     *Clients_Order     `yaml:"order"`
	Review    *Clients_Review    `yaml:"review"`
	Passenger *Clients_Passenger `yaml:"passenger"`
	Driver    *Clients_Driver    `yaml:"driver"`
}

type Clients_Auth struct {
	Endpoint    string `yaml:"endpoint"`
	HTTPBaseURL string `yaml:"http_base_url"`
}

type Clients_Trip struct {
	Endpoint    string `yaml:"endpoint"`
	HTTPBaseURL string `yaml:"http_base_url"`
}

type Clients_Order struct {
	Endpoint    string `yaml:"endpoint"`
	HTTPBaseURL string `yaml:"http_base_url"`
}

type Clients_Review struct {
	Endpoint    string `yaml:"endpoint"`
	HTTPBaseURL string `yaml:"http_base_url"`
}

type Clients_Passenger struct {
	Endpoint    string `yaml:"endpoint"`
	HTTPBaseURL string `yaml:"http_base_url"`
}

type Clients_Driver struct {
	Endpoint    string `yaml:"endpoint"`
	HTTPBaseURL string `yaml:"http_base_url"`
}

type Auth struct {
	Jwt        *Auth_JWT        `yaml:"jwt"`
	Permission *Auth_Permission `yaml:"permission"`
}

type Auth_JWT struct {
	Enabled                 bool   `yaml:"enabled"`
	Secret                  string `yaml:"secret"`
	Issuer                  string `yaml:"issuer"`
	ExpireSeconds           int64  `yaml:"expire_seconds"`
	CompatibleHeaderEnabled bool   `yaml:"compatible_header_enabled"`
}

type Auth_Permission struct {
	CacheTTL         string `yaml:"cache_ttl"`
	FailureThreshold int    `yaml:"failure_threshold"`
	CircuitOpenTTL   string `yaml:"circuit_open_ttl"`
}

type Alipay struct {
	AppID           string `yaml:"app_id"`
	PrivateKey      string `yaml:"private_key"`
	AlipayPublicKey string `yaml:"alipay_public_key"`
	Production      bool   `yaml:"production"`
	NotifyURL       string `yaml:"notify_url"`
	ReturnURL       string `yaml:"return_url"`
}

type Amap struct {
	WebKey  string `yaml:"web_key"`
	Timeout string `yaml:"timeout"`
}
