package rabbitmq

import (
	"fmt"
	"time"

	"github.com/glory-go/glory/mq"
	"github.com/glory-go/glory/tools"
	"github.com/streadway/amqp"
	"google.golang.org/appengine/log"
)

type RabbitMQService struct {
	config *Config
	conn   *amqp.Connection
}

func RabbitMQServiceFactory(rawConfig map[string]string) (mq.MQService, error) {
	srv := &RabbitMQService{
		config: &Config{},
	}
	tools.YamlStructConverter(rawConfig, srv.config)

	return srv, nil
}

func (s *RabbitMQService) Connect() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s", s.config.Username, s.config.Password, s.config.Host, s.config.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	s.conn = conn

	return nil
}

func (s *RabbitMQService) Send(topic string, msg []byte) (msgID string, err error) {
	panic("not implement")
}

func (s *RabbitMQService) DelaySend(topic string, msg []byte, handleTime time.Time) (msgID string, err error) {
	panic("not implement")
}

func (s *RabbitMQService) RegisterHandler(topic string, handler mq.MQMsgHandler) {
	panic("not implement")
}

func (s *RabbitMQService) send(chName string, msg []byte, delayTime *int64) error {
	chConfig, ok := s.config.Channels[chName]
	if !ok {
		return fmt.Errorf("channel %v not found", chName)
	}
	// 检查格式
	if chConfig.Type == Delay || chConfig.Type == PubSub {
		if chConfig.ExchangeName == "" {
			return fmt.Errorf("invalid channel %v, delay queue must has exchange name", chName)
		}
	}
	// 建立连接
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 初始化queue
	q, err := ch.QueueDeclare(
		chConfig.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		log.Errorf("init channel queue %v for %v meets error: %v", chConfig.QueueName, chName, err)
		return err
	}

	// 初始化exchange
	if chConfig.ExchangeName != "" {
		args := amqp.Table{}
		// TODO: 未考虑direct模式下应该如何定义
		exchangeType := "fanout"
		if chConfig.Type == Delay {
			args["x-delayed-type"] = "direct"
			exchangeType = "x-delayed-message"
		}
		if err := ch.ExchangeDeclare(
			chConfig.ExchangeName, // name
			exchangeType,          // type
			true,                  // durable
			false,                 // auto-deleted
			false,                 // internal
			false,                 // no-wait
			args,                  // arguments
		); err != nil {
			log.Errorf("init channel exchange %v for %v meets error", chConfig.ExchangeName, chName)
			return err
		}
	}

	headers := amqp.Table{}
	routingKey := q.Name
	if chConfig.Type == Delay {
		if delayTime == nil {
			tmp := int64(0)
			delayTime = &tmp
		}
		headers["x-delay"] = *delayTime
		routingKey = ""
	}
	if err = ch.Publish(
		chConfig.ExchangeName, // exchange
		routingKey,            // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			Headers:     headers,
			ContentType: "text/plain",
			Body:        msg,
		}); err != nil {
		return err
	}
	return nil
}
