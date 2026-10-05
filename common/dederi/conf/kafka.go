package conf

type KafkaAuthConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type KafkaProducerConfig struct {
	Brokers []string        `mapstructure:"brokers"`
	Auth    KafkaAuthConfig `mapstructure:"auth"`
	Topics  []string        `mapstructure:"topics"`
}

type KafkaConsumerConfig struct {
	Brokers []string        `mapstructure:"brokers"`
	GroupID string          `mapstructure:"groupId"`
	Topics  []string        `mapstructure:"topics"`
	Auth    KafkaAuthConfig `mapstructure:"auth"`
}
