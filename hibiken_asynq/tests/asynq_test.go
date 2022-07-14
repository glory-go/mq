package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glory-go/glory/v2/config"
	hibikenasynq "github.com/glory-go/mq/v2/hibiken_asynq"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func Test_Asynq(t *testing.T) {
	setup()
	// 注册handler
	called := false
	hibikenasynq.GetAsynqSub().GetMux("mock_server_1").HandleFunc("test_topic_1", func(ctx context.Context, t *asynq.Task) error {
		called = true
		return nil
	})
	// TODO 等待pub实现完成，测试收发的能力

	assert.True(t, called)
}

func setup() {
	// 初始化redis连接
	redis := miniredis.NewMiniRedis()
	// 初始化配置
	content := fmt.Sprintf(`
service:
mq:
hibiken_asynq:
  mock_server_1:
	addr: %s
`, redis.Addr())
	path := fmt.Sprintf("/tmp/test_sub_%d.yaml", time.Now().Unix())
	config.ChangeDefaultConfigPath(path)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(content)
	if err != nil {
		panic(err)
	}
}
