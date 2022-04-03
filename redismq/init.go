package redismq

import (
	"github.com/glory-go/glory/mq"
)

const (
	MQType = "redismq"
)

func Init() {
	mq.RegisterMQType(MQType, RedisMQServiceFactory)
}
