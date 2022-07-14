package hibikenasynq

import (
	"sync"

	"github.com/hibiken/asynq"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/sync/errgroup"
)

const (
	HibikenAsynqSubName = "hibiken_asynq"
)

type asynqSub struct {
	servers map[string]*asynq.Server
	config  map[string]*hibikenAsynqSubConfig
	mux     map[string]*asynq.ServeMux

	g errgroup.Group
}

var (
	asyncSubInstance *asynqSub
	subOnce          sync.Once
)

func GetAsynqSub() *asynqSub {
	subOnce.Do(func() {
		asyncSubInstance = &asynqSub{
			servers: make(map[string]*asynq.Server),
			config:  make(map[string]*hibikenAsynqSubConfig),
			mux:     make(map[string]*asynq.ServeMux),
		}
	})
	return asyncSubInstance
}

func (q *asynqSub) GetMux(name string) *asynq.ServeMux {
	return q.mux[name]
}

func (q *asynqSub) Name() string { return HibikenAsynqSubName }

func (q *asynqSub) Init(config map[string]any) error {
	for name, raw := range config {
		conf := &hibikenAsynqSubConfig{}
		if err := mapstructure.Decode(raw, conf); err != nil {
			return err
		}
		q.config[name] = conf

		srv := asynq.NewServer(asynq.RedisClientOpt{
			Addr:     conf.Addr,
			Username: conf.Username,
			Password: conf.Password,
			DB:       conf.DB,
		}, asynq.Config{
			Concurrency: conf.MaxWorkerCnt,
			Queues:      conf.QueuePriority,
		})
		q.servers[name] = srv

		q.mux[name] = asynq.NewServeMux()
	}

	return nil
}

func (q *asynqSub) Run() error {
	for name := range q.servers {
		srv := q.servers[name]
		mux, ok := q.mux[name]
		if !ok {
			continue
		}
		q.g.Go(func() error {
			return srv.Run(mux)
		})
	}

	return q.g.Wait()
}
