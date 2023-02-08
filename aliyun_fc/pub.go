package aliyunfc

import (
	"sync"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/fc-open-20210406/v2/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/mitchellh/mapstructure"
)

const (
	AliyunFCPubName = "aliyun_fc_pub"
)

type aliyunFCPub struct {
	config  map[string]*aliyunFCPubConfig
	clients map[string]*client.Client
}

var (
	aliyunFCPubInstance *aliyunFCPub
	pubOnce             sync.Once
)

func getAliyunFCPub() *aliyunFCPub {
	pubOnce.Do(func() {
		aliyunFCPubInstance = &aliyunFCPub{
			config:  make(map[string]*aliyunFCPubConfig),
			clients: make(map[string]*client.Client),
		}
	})
	return aliyunFCPubInstance
}

func Involke(name string, request *client.InvokeFunctionRequest, headers *client.InvokeFunctionHeaders,
	runtime *service.RuntimeOptions) (*client.InvokeFunctionResponse, error) {
	registerPub()
	conf := getAliyunFCPub().config[name]

	c := getAliyunFCPub().clients[name]
	return c.InvokeFunctionWithOptions(&conf.ServiceName, &conf.FunctionName, request, headers, runtime)
}

func (q *aliyunFCPub) Name() string { return AliyunFCPubName }

func (q *aliyunFCPub) Init(config map[string]any) error {
	for name, raw := range config {
		conf := &aliyunFCPubConfig{}
		if err := mapstructure.Decode(raw, conf); err != nil {
			return err
		}
		q.config[name] = conf

		c, err := client.NewClient(&openapi.Config{
			AccessKeyId:     &conf.AccessKeyID,
			AccessKeySecret: &conf.AccessKeySecret,
			Endpoint:        &conf.Endpoint,
		})
		if err != nil {
			return err
		}
		q.clients[name] = c
	}

	return nil
}
