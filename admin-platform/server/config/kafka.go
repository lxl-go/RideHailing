package config

type Kafka struct {
	Brokers  []string `mapstructure:"brokers" json:"brokers" yaml:"brokers"`
	Topic    string   `mapstructure:"topic" json:"topic" yaml:"topic"`
	ClientID string   `mapstructure:"client-id" json:"client-id" yaml:"client-id"`
}
