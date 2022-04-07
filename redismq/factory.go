package redismq

import (
	"fmt"

	"github.com/glory-go/glory/config"
	"github.com/glory-go/glory/mq"
	"github.com/glory-go/glory/tools"
	"github.com/rs/xid"
)

func RedisMQServiceFactory(mod config.ChannelType, rawConfig map[string]string) (mq.MQService, error) {
	conf := &Config{}
	if err := tools.YamlStructConverter(rawConfig, conf); err != nil {
		return nil, err
	}
	if conf.Port == "" {
		conf.Port = "6379"
	}
	if conf.BuckCnt == "" {
		conf.BuckCnt = "0"
	}
	if conf.Name == "" {
		conf.Name = xid.New().String()
	}
	if conf.TTR == "" {
		conf.TTR = "0"
	}

	var srv mq.MQService
	switch mod {
	case config.Direct:
		srv = &RedisMQService{
			config: conf,
		}
	case config.PubSub:
		srv = &PubSubRedisMQService{
			config: conf,
		}
	default:
		return nil, fmt.Errorf("mod type not support in redismq: %s", mod)
	}

	return srv, nil
}
