package aliyunfc

import (
	"sync"

	"github.com/glory-go/glory/v2/sub"
)

var (
	registerSubOnce sync.Once
)

func registerSub() {
	registerSubOnce.Do(func() {
		sub.GetSub().RegisterSubProvider(getAliyunFCSub())
	})
}
