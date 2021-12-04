package rocketmq

type Config struct {
	Endpoint      string `yaml:"endpoint"`
	AccessKey     string `yaml:"access_key"`
	SecretKey     string `yaml:"secret_key"`
	InstanceID    string `yaml:"instance_id"`
	SecurityToken string `yaml:"secority_token"`
	AutoACK       bool   `yaml:"auto_ack"`

	ConsumerGroupID string `yaml:"consumer_group_id"`

	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
