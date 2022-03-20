package redismq

import (
	"github.com/glory-go/glory/mq"
)

const (
	MQType = "redismq"
)

func init() {
	mq.RegisterMQType(MQType, RedisMQServiceFactory)
}
