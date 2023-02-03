package aliyunfc

type aliyunFCSubConfig struct {
	Addr         string `mapstructure:"addr"`
	ReadTimeout  int    `mapstructure:"read_timeout"`  // 单位：ms
	WriteTimeout int    `mapstructure:"write_timeout"` // 单位：ms
}
