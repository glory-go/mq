package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glory-go/glory/v2/config"
	"github.com/glory-go/glory/v2/service"
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
	go func() { service.GetService().Run() }()
	// 测试收发的能力
	_, err := hibikenasynq.GetAsynqPub().GetClient("mock_client_1").Enqueue(asynq.NewTask("test_topic_1", []byte{}))

	// 等一会进行判断
	time.Sleep(time.Second * 1)
	assert.Nil(t, err)
	assert.True(t, called)
}

func setup() {
	// 初始化redis连接
	redis := miniredis.NewMiniRedis()
	redis.Start()
	// 初始化配置
	content := fmt.Sprintf(`
service:
  sub:
    hibiken_asynq:
      mock_server_1:
        addr: %s
hibiken_asynq:
  mock_client_1:
    addr: %s
`, redis.Addr(), redis.Addr())
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
	config.Init()
}
