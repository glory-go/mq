package rocketmq

import (
	"github.com/glory-go/glory/mq"
)

const (
	MQType = "aliyun_rocketmq"
)

func init() {
	mq.RegisterMQType(MQType, AliyunRocketMQServiceFactory)
}
