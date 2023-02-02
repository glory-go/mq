package aliyuneventbridge

type aliyunEventBridgePubConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`

	EventBusName string `mapstructure:"event_bus_name"`
}
