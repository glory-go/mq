package hibikenasynq

import (
	"sync"

	"github.com/hibiken/asynq"
	"github.com/mitchellh/mapstructure"
)

const (
	HibikenAsynqPubName = "hibiken_asynq"
)

type asynqPub struct {
	config  map[string]*hibikenAsynqPubConfig
	clients map[string]*asynq.Client
}

var (
	asynqPubInstance *asynqPub
	pubOnce          sync.Once
)

func GetAsynqPub() *asynqPub {
	pubOnce.Do(func() {
		asynqPubInstance = &asynqPub{
			config:  make(map[string]*hibikenAsynqPubConfig),
			clients: make(map[string]*asynq.Client),
		}
	})
	return asynqPubInstance
}

func (q *asynqPub) GetClient(name string) *asynq.Client {
	return q.clients[name]
}

func (q *asynqPub) Name() string { return HibikenAsynqPubName }

func (q *asynqPub) Init(config map[string]any) error {
	for name, raw := range config {
		conf := &hibikenAsynqPubConfig{}
		if err := mapstructure.Decode(raw, conf); err != nil {
			return err
		}
		q.config[name] = conf

		client := asynq.NewClient(asynq.RedisClientOpt{
			Addr:     conf.Addr,
			Username: conf.Username,
			Password: conf.Password,
			DB:       conf.DB,
		})
		q.clients[name] = client
	}

	return nil
}
