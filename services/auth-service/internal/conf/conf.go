package conf

type Bootstrap struct {
	Server   *Server   `yaml:"server"`
	Data     *Data     `yaml:"data"`
	Registry *Registry `yaml:"registry"`
	Auth     *Auth     `yaml:"auth"`
}

type Server struct {
	Http *Server_HTTP `yaml:"http"`
	Grpc *Server_GRPC `yaml:"grpc"`
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
	Redis    *Data_Redis    `yaml:"redis"`
}

type Data_Database struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type Data_Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
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

type Auth struct {
	Jwt                       *Auth_JWT `yaml:"jwt"`
	Sms                       *Auth_SMS `yaml:"sms"`
	RefreshTokenExpireSeconds int64     `yaml:"refresh_token_expire_seconds"`
	RequireSmsCodeVerify      *bool     `yaml:"require_sms_code_verify"`
}

type Auth_JWT struct {
	Enabled       bool   `yaml:"enabled"`
	Secret        string `yaml:"secret"`
	Issuer        string `yaml:"issuer"`
	ExpireSeconds int64  `yaml:"expire_seconds"`
}

type Auth_SMS struct {
	CodeExpireSeconds   int64       `yaml:"code_expire_seconds"`
	SendCooldownSeconds int64       `yaml:"send_cooldown_seconds"`
	MaxVerifyAttempts   int         `yaml:"max_verify_attempts"`
	VerifyLockSeconds   int64       `yaml:"verify_lock_seconds"`
	Ihuyi               *Auth_Ihuyi `yaml:"ihuyi"`
}

type Auth_Ihuyi struct {
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
	Endpoint string `yaml:"endpoint"`
}
