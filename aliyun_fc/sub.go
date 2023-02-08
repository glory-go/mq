package aliyunfc

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	AliyunFCSubName = "aliyun_fc"
)

type aliyunFCSub struct {
	routers map[string]*gin.Engine
	config  map[string]*aliyunFCSubConfig
	mux     map[string]*asynq.ServeMux

	g errgroup.Group
}

var (
	aliyunFCSubInstance *aliyunFCSub
	subOnce             sync.Once
)

func getAliyunFCSub() *aliyunFCSub {
	subOnce.Do(func() {
		aliyunFCSubInstance = &aliyunFCSub{
			routers: make(map[string]*gin.Engine),
			config:  make(map[string]*aliyunFCSubConfig),
			mux:     make(map[string]*asynq.ServeMux),
		}
	})
	return aliyunFCSubInstance
}

func GetMux(name string) *asynq.ServeMux {
	registerSub()
	return getAliyunFCSub().mux[name]
}

func (q *aliyunFCSub) Name() string { return AliyunFCSubName }

func (q *aliyunFCSub) Init(config map[string]any) error {
	for name, raw := range config {
		r := gin.Default()
		r.ContextWithFallback = true
		mux := asynq.NewServeMux()
		tmpName := name
		r.POST("/involke", involkeHandler(tmpName, mux))
		q.routers[name] = r
		// 解析配置
		conf := &aliyunFCSubConfig{}
		if err := mapstructure.Decode(raw, conf); err != nil {
			return err
		}
		q.config[name] = conf
		q.mux[name] = mux
	}

	return nil
}

func (q *aliyunFCSub) Run() error {
	for k := range q.routers {
		name := k
		q.g.Go(func() error {
			server := http.Server{
				Addr:         q.config[name].Addr,
				Handler:      q.routers[name],
				ReadTimeout:  time.Duration(q.config[name].ReadTimeout) * time.Millisecond,
				WriteTimeout: time.Duration(q.config[name].WriteTimeout) * time.Millisecond,
			}
			return server.ListenAndServe()
		})
	}

	return q.g.Wait()
}

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func involkeHandler(name string, mux *asynq.ServeMux) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		logger := logrus.WithContext(ctx)
		event := &Event{}
		if err := ctx.BindJSON(event); err != nil {
			logger.Errorf("fail to parse body to event, err: %v", err)
			ctx.JSON(http.StatusNotFound, err)
			ctx.Header(FCStatusHeader, "404")
			return
		}
		logger.Infof("mux %s receive event with type %s", name, event.Type)
		logger.Debugf("mux %s receive event, type %s with payload %s", name, event.Type, string(event.Payload))

		task := asynq.NewTask(event.Type, event.Payload)
		if err := mux.ProcessTask(ctx, task); err != nil {
			logger.Errorf("fail to parse body to event, err: %v", err)
			ctx.JSON(http.StatusNotFound, err)
			ctx.Header(FCStatusHeader, "404")
			return
		}
		ctx.JSON(http.StatusOK, gin.H{})
		ctx.Header(FCStatusHeader, "200")
	}
}
