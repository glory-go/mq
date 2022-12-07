package redisstream

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/xid"
	"golang.org/x/sync/errgroup"
)

const (
	RedisStreamSubName = "redis_stream"
)

const (
	StreamMaxLen = 10000           // stream最大长度
	BlockTime    = time.Minute     // 阻塞读取时间
	ErrWaitTime  = time.Second * 3 // 请求错误后的重试时间
)

type MsgHandler func(ctx context.Context, data []byte) error

type redisStreamSub struct {
	redisClient map[string]*redis.Client
	config      map[string]*redisStreamSubConfig
	handler     map[string]map[string]MsgHandler // name -> topic -> handler

	g errgroup.Group
}

var (
	redisStreamSubInstance *redisStreamSub
	subOnce                sync.Once
)

func getAsynqSub() *redisStreamSub {
	subOnce.Do(func() {
		redisStreamSubInstance = &redisStreamSub{
			redisClient: make(map[string]*redis.Client),
			config:      make(map[string]*redisStreamSubConfig),
			handler:     make(map[string]map[string]MsgHandler),
		}
	})
	return redisStreamSubInstance
}

func (s *redisStreamSub) Name() string { return RedisStreamSubName }

func (s *redisStreamSub) Init(config map[string]any) error {
	for name, raw := range config {
		conf := &redisStreamSubConfig{}
		if err := mapstructure.Decode(raw, conf); err != nil {
			return err
		}
		s.config[name] = conf

		redisclient := redis.NewClient(&redis.Options{
			Addr:     conf.Addr,
			Username: conf.Username,
			Password: conf.Password,
			DB:       conf.DB,
		})
		s.redisClient[name] = redisclient
		s.handler[name] = make(map[string]MsgHandler)
	}

	return nil
}

func RegisterHandler(name, topic string, handler MsgHandler) {
	if redisStreamSubInstance.handler[name] == nil {
		panic("sub " + name + " has no config")
	}
	if redisStreamSubInstance.handler[name][topic] != nil {
		panic("topic " + topic + " registered twice")
	}
	// 若未提供组名，则报错
	if redisStreamSubInstance.config[name].GroupName == "" {
		panic("group name of " + name + " is empty")
	}
	redisStreamSubInstance.handler[name][topic] = handler
	ctx := context.Background()
	// 检查group是否存在，不存在则创建一个
	_, err := redisStreamSubInstance.redisClient[name].XGroupCreateMkStream(ctx, topic, redisStreamSubInstance.config[name].GroupName, "$").Result()
	if err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			fmt.Printf("group %s already exists", redisStreamSubInstance.config[name].GroupName)
		} else {
			fmt.Printf("create group %s failed, %s", redisStreamSubInstance.config[name].GroupName, err.Error())
			panic(err)
		}
	}
}

func (q *redisStreamSub) Run() error {
	// 生成consumer name，要求消费组内唯一
	consumerName := xid.New().String()
	for name := range q.redisClient {
		client := q.redisClient[name]
		topicHandler, ok := q.handler[name]
		if !ok {
			continue
		}
		for topic := range topicHandler {
			localTopic := topic
			localName := name
			q.g.Go(func() error {
				for {
					ctx := context.Background()
					// 获取消息
					msgs, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
						Group:    q.config[localName].GroupName,
						Consumer: consumerName,
						Streams: []string{
							localTopic,
							">", // 获取最新的消息
						},
						Count: 1,
						Block: BlockTime,
						NoAck: !q.config[localName].AutoAck, // 默认不ack
					}).Result()
					if err != nil {
						if err == redis.Nil {
							continue
						}
						fmt.Printf("redis mq service xreadgroup error: %v", err)
						time.Sleep(ErrWaitTime)
						continue
					}
					if len(msgs) == 0 {
						continue
					}
					msg := msgs[0].Messages[0]
					fmt.Printf("redis pubsub service receive msg id: %v", msg.ID)
					// 处理消息
					if err := topicHandler[localTopic](ctx, []byte(msg.Values["msg"].(string))); err != nil {
						continue
					}
					// 删除消息
					if !q.config[name].AutoAck {
						if err := client.XAck(ctx, localTopic, q.config[name].GroupName, msg.ID).Err(); err != nil {
							fmt.Printf("redis mq service xack error: %v", err)
						}
					}
				}
			})
		}
	}

	return q.g.Wait()
}
