package redisstream

type redisStreamSubConfig struct {
	Addr      string `mapstructure:"addr"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db"`

	GroupName string `mapstructure:"group_name"`
	AutoAck   bool   `mapstructure:"auto_ack"`
}

type redisStreamPubConfig struct {
	Addr     string `mapstructure:"addr"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
