package redisstream

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
)

const (
	RedisStreamPubName = "redis_stream"
)

type asynqPub struct {
	config  map[string]*redisStreamPubConfig
	clients map[string]*redis.Client
}

var (
	redisStreamPubInstance *asynqPub
	pubOnce                sync.Once
)

func getAsynqPub() *asynqPub {
	pubOnce.Do(func() {
		redisStreamPubInstance = &asynqPub{
			config:  make(map[string]*redisStreamPubConfig),
			clients: make(map[string]*redis.Client),
		}
	})
	return redisStreamPubInstance
}

func Publish(ctx context.Context, name, topic string, msg []byte) (msgID string, err error) {
	// 使用redis的stream
	msgID, err = redisStreamPubInstance.clients[name].XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		MaxLen: StreamMaxLen,
		ID:     "*", // 自动生成
		Values: map[string]interface{}{
			"msg": msg,
		},
	}).Result()
	if err != nil {
		return "", err
	}

	return msgID, nil
}

func (q *asynqPub) Name() string { return RedisStreamPubName }

func (q *asynqPub) Init(config map[string]any) error {
	for name, raw := range config {
		conf := &redisStreamPubConfig{}
		if err := mapstructure.Decode(raw, conf); err != nil {
			return err
		}
		q.config[name] = conf

		redisclient := redis.NewClient(&redis.Options{
			Addr:     conf.Addr,
			Username: conf.Username,
			Password: conf.Password,
			DB:       conf.DB,
		})
		q.clients[name] = redisclient
	}

	return nil
}
