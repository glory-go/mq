package rabbitmq

import (
	"fmt"
	"time"

	"github.com/glory-go/glory/mq"
	"github.com/glory-go/glory/tools"
	"github.com/streadway/amqp"
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
