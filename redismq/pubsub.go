package redismq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glory-go/glory/log"
	"github.com/glory-go/glory/mq"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
)

const (
	StreamMaxLen = 10000 // stream最大长度
)

type PubSubRedisMQService struct {
	config *Config
	client *redis.Client
}

func (s *PubSubRedisMQService) Connect() error {
	client, err := conn(s.config)
	if err != nil {
		return err
	}
	s.client = client

	return nil
}

func (s *PubSubRedisMQService) getStreamName(topic string) string {
	return fmt.Sprintf("%v:%v", s.config.Name, topic)
}

func (s *PubSubRedisMQService) Send(topic string, msg []byte) (msgID string, err error) {
	return "", errors.New("invalid mod, don't use Send() in pubsub mod")
}

func (s *PubSubRedisMQService) DelaySend(topic string, msg []byte, handleTime time.Time) (msgID string, err error) {
	return "", errors.New("invalid mod, don't use DelaySend() in pubsub mod")
}

func (s *PubSubRedisMQService) Publish(topic string, msg []byte) (msgID string, err error) {
	ctx := context.Background()
	// 使用redis的stream
	streamName := s.getStreamName(topic)
	msgID, err = s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
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

func (s *PubSubRedisMQService) RegisterHandler(topic string, handler mq.MQMsgHandler) {
	ctx := context.Background()
	streamName := s.getStreamName(topic)
	// 若未提供组名，则报错
	if s.config.GroupName == "" {
		panic("group name is empty")
	}
	// 检查group是否存在，不存在则创建一个
	_, err := s.client.XGroupCreateMkStream(ctx, streamName, s.config.GroupName, "$").Result()
	if err != nil {
		panic(err) // 注意：这里可能为重复创建group
	}
	// 生成consumer name，要求消费组内唯一
	consumer := xid.New().String()

	for {
		// 获取消息
		msgs, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    s.config.GroupName,
			Consumer: consumer,
			Streams: []string{
				streamName,
			},
			Count: 1,
			Block: time.Second,
			NoAck: true, // 默认不ack
		}).Result()
		if err != nil {
			log.CtxErrorf(ctx, "redis mq service xreadgroup error: %v", err)
			continue
		}
		if len(msgs) == 0 {
			continue
		}
		msg := msgs[0].Messages[0]
		log.CtxInfof(ctx, "redis pubsub service receive msg id: %v", msg.ID)
		// 处理消息
		if err := handler(ctx, msg.Values["msg"].([]byte)); err != nil {
			continue
		}
		// 删除消息
		if err := s.client.XAck(ctx, streamName, s.config.GroupName, msg.ID).Err(); err != nil {
			log.CtxErrorf(ctx, "redis mq service xack error: %v", err)
		}
	}
}
