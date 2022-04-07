package redismq

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/glory-go/glory/log"
	"github.com/glory-go/glory/mq"
	dqueue "github.com/go-online-public/delay-queue"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
)

const (
	defaultBuckCnt = 3
	defaultTTR     = 3
)

type RedisMQService struct {
	config *Config
	ttr    int64
	client *dqueue.DelayRedisQueue
}

func conn(conf *Config) (*redis.Client, error) {
	db, err := strconv.Atoi(conf.DB)
	if err != nil {
		return nil, err
	}
	redisclient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", conf.Host, conf.Port),
		Username: conf.Username,
		Password: conf.Password,
		DB:       db,
	})

	return redisclient, nil
}

func (s *RedisMQService) Connect() error {
	buckCnt, err := strconv.Atoi(s.config.BuckCnt)
	if err != nil {
		return err
	}
	if buckCnt <= 0 {
		buckCnt = defaultBuckCnt
	}
	ttr, err := strconv.ParseInt(s.config.TTR, 10, 64)
	if err != nil {
		return err
	}
	if ttr <= 0 {
		ttr = defaultTTR
	}
	s.ttr = ttr
	redisclient, err := conn(s.config)
	if err != nil {
		return err
	}
	s.client = dqueue.New(context.Background(), s.config.Name, buckCnt, redisclient)

	return nil
}

func (s *RedisMQService) Send(topic string, msg []byte) (msgID string, err error) {
	ctx := context.Background()
	return s.send(ctx, topic, msg, time.Now().Add(4*time.Second))
}

func (s *RedisMQService) DelaySend(topic string, msg []byte, handleTime time.Time) (msgID string, err error) {
	ctx := context.Background()
	return s.send(ctx, topic, msg, handleTime)
}

func (s *RedisMQService) Publish(topic string, msg []byte) (msgID string, err error) {
	return "", errors.New("invalid mod, don't use Pub() in normal mod")
}

func (s *RedisMQService) send(ctx context.Context, topic string, msg []byte, handleTime time.Time) (msgID string, err error) {
	id := xid.New().String()
	if err := s.client.Push(ctx, dqueue.Job{
		Topic: topic,
		Id:    id,
		Delay: int64(time.Until(handleTime).Seconds()),
		TTR:   s.ttr,
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
		if job == nil {
			continue
		}
		if err := handler(ctx, []byte(job.Body)); err != nil {
			continue // 消息会一直被处理，直到处理成功或TTR时间到达
		}
		s.client.Remove(ctx, job.Id)
	}
}
