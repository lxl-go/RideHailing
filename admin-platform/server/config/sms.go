package config

type SMS struct {
	Enabled    bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Provider   string `mapstructure:"provider" json:"provider" yaml:"provider"`
	AccessKey  string `mapstructure:"access-key" json:"access-key" yaml:"access-key"`
	SecretKey  string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key"`
	SignName   string `mapstructure:"sign-name" json:"sign-name" yaml:"sign-name"`
	TemplateID string `mapstructure:"template-id" json:"template-id" yaml:"template-id"`
}
