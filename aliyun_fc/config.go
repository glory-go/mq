package aliyunfc

type aliyunFCPubConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`

	ServiceName  string `mapstructure:"service"`
	FunctionName string `mapstructure:"function"`
}

type aliyunFCSubConfig struct {
	Addr         string `mapstructure:"addr"`
	ReadTimeout  int    `mapstructure:"read_timeout"`  // 单位：ms
	WriteTimeout int    `mapstructure:"write_timeout"` // 单位：ms
}
