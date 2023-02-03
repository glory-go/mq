package aliyuneventbridge

import (
	"sync"

	"github.com/glory-go/glory/v2/config"
)

var (
	registerPubOnce sync.Once
)

func registerPub() {
	registerPubOnce.Do(func() {
		config.RegisterComponent(getAliyunEventBridgePub())
	})
}
