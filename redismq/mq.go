package redismq

import (
	"context"
	"fmt"
	"time"

	"github.com/glory-go/glory/log"
	"github.com/glory-go/glory/mq"
	"github.com/glory-go/glory/tools"
	dqueue "github.com/go-online-public/delay-queue"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
)

const (
	defaultBuckCnt = 3
)

type RedisMQService struct {
	config *Config
	client *dqueue.DelayRedisQueue
}

func AliyunRocketMQServiceFactory(rawConfig map[string]string) (mq.MQService, error) {
	srv := &RedisMQService{
		config: &Config{},
	}
	if err := tools.YamlStructConverter(rawConfig, srv.config); err != nil {
		return nil, err
	}
	if srv.config.Port == "" {
		srv.config.Port = "6379"
	}
	if srv.config.BuckCnt <= 0 {
		srv.config.BuckCnt = defaultBuckCnt
	}
	if srv.config.QueueName == "" {
		srv.config.QueueName = xid.New().String()
	}

	return srv, nil
}

func (s *RedisMQService) Connect() error {
	redisclient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", s.config.Host, s.config.Port),
		Username: s.config.Username,
		Password: s.config.Password,
		DB:       s.config.DB,
	})
	s.client = dqueue.New(context.Background(), s.config.QueueName, s.config.BuckCnt, redisclient)

	return nil
}

func (s *RedisMQService) Send(topic string, msg []byte) (msgID string, err error) {
	ctx := context.Background()
	return s.send(ctx, topic, msg, time.Now().Add(0))
}

func (s *RedisMQService) DelaySend(topic string, msg []byte, handleTime time.Time) (msgID string, err error) {
	ctx := context.Background()
	return s.send(ctx, topic, msg, handleTime)
}

func (s *RedisMQService) send(ctx context.Context, topic string, msg []byte, handleTime time.Time) (msgID string, err error) {
	id := xid.New().String()
	if err := s.client.Push(ctx, dqueue.Job{
		Topic: topic,
		Id:    id,
		Delay: int64(handleTime.Second()),
		TTR:   int64(handleTime.Sub(time.Now()).Seconds()),
		Body:  string(msg),
	}); err != nil {
		return "", err
	}
	return id, nil
}

func (s *RedisMQService) RegisterHandler(topic string, handler mq.MQMsgHandler) {
	funcName := "RegisterHandler"
	ctx := context.Background()
	for {
		job, err := s.client.Pop(ctx, []string{topic})
		if err != nil {
			log.CtxErrorf(ctx, "[%s] fail to get job from redis queue, err: %v", funcName, err)
			continue
		}
		handler(ctx, []byte(job.Body))
	}
}