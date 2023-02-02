package aliyuneventbridge

import (
	"sync"

	"github.com/alibabacloud-go/eventbridge-sdk/eventbridge"
	"github.com/mitchellh/mapstructure"
)

const (
	AliyunEventBridgePubName = "aliyun_event_bridge"
)

type aliyunEventBridgePub struct {
	config  map[string]*aliyunEventBridgePubConfig
	clients map[string]*eventbridge.Client
}

var (
	aliyunEventBridgePubInstance *aliyunEventBridgePub
	pubOnce                      sync.Once
)

func getAliyunEventBridgePub() *aliyunEventBridgePub {
	pubOnce.Do(func() {
		aliyunEventBridgePubInstance = &aliyunEventBridgePub{
			config:  make(map[string]*aliyunEventBridgePubConfig),
			clients: make(map[string]*eventbridge.Client),
		}
	})
	return aliyunEventBridgePubInstance
}

func GetClient(name string) *eventbridge.Client {
	registerPub()
	return getAliyunEventBridgePub().clients[name]
}

func GetConf(name string) *aliyunEventBridgePubConfig {
	registerPub()
	return getAliyunEventBridgePub().config[name]
}

func (q *aliyunEventBridgePub) Name() string { return AliyunEventBridgePubName }

func (q *aliyunEventBridgePub) Init(config map[string]any) error {
	for name, raw := range config {
		conf := &aliyunEventBridgePubConfig{}
		if err := mapstructure.Decode(raw, conf); err != nil {
			return err
		}
		q.config[name] = conf

		config := new(eventbridge.Config).
			SetAccessKeyId(conf.AccessKeyID).
			SetAccessKeySecret(conf.AccessKeySecret).
			SetEndpoint(conf.Endpoint)
		client, err := eventbridge.NewClient(config)
		if err != nil {
			return err
		}
		q.clients[name] = client
	}

	return nil
}
