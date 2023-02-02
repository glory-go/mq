package hibikenasynq

import (
	"sync"

	"github.com/glory-go/glory/v2/config"
	"github.com/glory-go/glory/v2/sub"
)

var (
	registerPubOnce, registerSubOnce sync.Once
)

func registerPub() {
	registerPubOnce.Do(func() {
		config.RegisterComponent(getAsynqPub())
	})
}

func registerSub() {
	registerSubOnce.Do(func() {
		sub.GetSub().RegisterSubProvider(getAsynqSub())
	})
}
