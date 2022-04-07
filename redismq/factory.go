package redismq

import (
	"github.com/glory-go/glory/config"
	"github.com/glory-go/glory/mq"
	"github.com/glory-go/glory/tools"
	"github.com/rs/xid"
)

func RedisMQServiceFactory(mod config.ChannelType, rawConfig map[string]string) (mq.MQService, error) {
	srv := &RedisMQService{
		config: &Config{},
	}
	if err := tools.YamlStructConverter(rawConfig, srv.config); err != nil {
		return nil, err
	}
	if srv.config.Port == "" {
		srv.config.Port = "6379"
	}
	if srv.config.BuckCnt == "" {
		srv.config.BuckCnt = "0"
	}
	if srv.config.Name == "" {
		srv.config.Name = xid.New().String()
	}
	if srv.config.TTR == "" {
		srv.config.TTR = "0"
	}

	return srv, nil
}
