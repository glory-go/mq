package redisstream

import (
	"github.com/glory-go/glory/v2/config"
	"github.com/glory-go/glory/v2/sub"
)

func init() {
	sub.GetSub().RegisterSubProvider(getAsynqSub())
	config.RegisterComponent(getAsynqPub())
}
