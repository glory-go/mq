package hibikenasynq

import (
	"github.com/glory-go/glory/v2/config"
	"github.com/glory-go/glory/v2/sub"
)

func init() {
	sub.GetSub().RegisterSubProvider(GetAsynqSub())
	config.RegisterComponent(GetAsynqPub())
}
