package rocketmq

import (
	"fmt"
	"strings"
	"time"

	mq_http_sdk "github.com/aliyunmq/mq-http-go-sdk"
	"github.com/glory-go/glory/log"
	"github.com/glory-go/glory/mq"
	"github.com/glory-go/glory/tools"
	"github.com/gogap/errors"
)

type AliyunRocketMQService struct {
	config *Config
	client mq_http_sdk.MQClient
}

func AliyunRocketMQServiceFactory(rawConfig map[string]string) (mq.MQService, error) {
	srv := &AliyunRocketMQService{
		config: &Config{},
	}
	tools.YamlStructConverter(rawConfig, srv.config)

	return srv, nil
}

func (s *AliyunRocketMQService) Connect() error {
	client := mq_http_sdk.NewAliyunMQClient(s.config.Endpoint, s.config.AccessKey, s.config.SecretKey,
		s.config.SecurityToken)
	s.client = client

	return nil
}

func (s *AliyunRocketMQService) Send(topic string, msg []byte) (msgID string, err error) {
	producer := s.client.GetProducer(s.config.InstanceID, topic)
	res, err := producer.PublishMessage(mq_http_sdk.PublishMessageRequest{
		MessageBody: string(msg),
	})
	if err != nil {
		return "", err
	}

	return res.MessageId, nil
}

func (s *AliyunRocketMQService) DelaySend(topic string, msg []byte, handleTime time.Time) (msgID string, err error) {
	producer := s.client.GetProducer(s.config.InstanceID, topic)
	res, err := producer.PublishMessage(mq_http_sdk.PublishMessageRequest{
		MessageBody:      string(msg),
		StartDeliverTime: handleTime.Sub(time.Now()).Milliseconds(),
	})
	if err != nil {
		return "", err
	}

	return res.MessageId, nil
}

func (s *AliyunRocketMQService) RegisterHandler(topic string, handler mq.MQMsgHandler) {
	consumer := s.client.GetConsumer(s.config.InstanceID, topic, s.config.ConsumerGroupID, "")
	for {
		endChan := make(chan int)
		respChan := make(chan mq_http_sdk.ConsumeMessageResponse)
		errChan := make(chan error)
		go func() {
			select {
			case resp := <-respChan:
				// 处理业务逻辑
				var handles []string
				log.Infof("Consume %d messages---->", len(resp.Messages))
				for _, v := range resp.Messages {
					handles = append(handles, v.ReceiptHandle)
					log.Infof("\tMessageID: %s, PublishTime: %d, MessageTag: %s\n"+
						"\tConsumedTimes: %d, FirstConsumeTime: %d, NextConsumeTime: %d\n"+
						"\tProps: %s\n",
						v.MessageId, v.PublishTime, v.MessageTag, v.ConsumedTimes,
						v.FirstConsumeTime, v.NextConsumeTime, v.Properties)
				}

				// NextConsumeTime前若不确认消息消费成功，则消息会重复消费
				// 消息句柄有时间戳，同一条消息每次消费拿到的都不一样
				if s.config.AutoACK {
					ack(consumer, handles)
				}

				// 消费消息
				for idx, respMsg := range resp.Messages {
					msg := respMsg
					go func() {
						ctx := tools.GetCtxWithLogID()
						if err := handler(ctx, []byte(msg.Message)); err != nil {
							log.Warn(err)
						} else if !s.config.AutoACK {
							ack(consumer, []string{handles[idx]})
						}
					}()
				}
				endChan <- 1
			case err := <-errChan:
				// 没有消息
				if strings.Contains(err.(errors.ErrCode).Error(), "MessageNotExist") {
					log.Info("no new message, continue!")
				} else {
					log.Error(err)
					time.Sleep(time.Duration(3) * time.Second)
				}
				endChan <- 1
			case <-time.After(35 * time.Second):
				log.Warn("Timeout of consumer message ??")
				endChan <- 1
			}
		}()

		// 长轮询消费消息
		// 长轮询表示如果topic没有消息则请求会在服务端挂住3s，3s内如果有消息可以消费则立即返回
		consumer.ConsumeMessage(respChan, errChan,
			3, // 一次最多消费3条(最多可设置为16条)
			3, // 长轮询时间3秒（最多可设置为30秒）
		)
		<-endChan
	}
}

func ack(consumer mq_http_sdk.MQConsumer, handlers []string) {
	ackerr := consumer.AckMessage(handlers)
	if ackerr != nil {
		// 某些消息的句柄可能超时了会导致确认不成功
		fmt.Println(ackerr)
		for _, errAckItem := range ackerr.(errors.ErrCode).Context()["Detail"].([]mq_http_sdk.ErrAckItem) {
			log.Infof("\tErrorHandle:%s, ErrorCode:%s, ErrorMsg:%s\n",
				errAckItem.ErrorHandle, errAckItem.ErrorCode, errAckItem.ErrorMsg)
		}
		time.Sleep(time.Duration(3) * time.Second)
	} else {
		log.Infof("Ack ---->\n\t%s\n", handlers)
	}
}
