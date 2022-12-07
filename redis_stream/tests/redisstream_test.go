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
	redisstream "github.com/glory-go/mq/v2/redis_stream"
	"github.com/stretchr/testify/assert"
)

func Test_RedisStream(t *testing.T) {
	setup()
	// 注册handler
	called1, called2 := false, false
	redisstream.RegisterHandler("mock_server_1", "mock_topic_1", func(ctx context.Context, data []byte) error {
		called1 = true
		return nil
	})
	redisstream.RegisterHandler("mock_server_2", "mock_topic_1", func(ctx context.Context, data []byte) error {
		called2 = true
		return nil
	})
	go func() { service.GetService().Run() }()
	// 测试收发的能力
	_, err := redisstream.Publish(context.Background(), "mock_client_1", "mock_topic_1", []byte{})
	assert.Nil(t, err)

	// 等一会进行判断
	time.Sleep(time.Second * 1)
	assert.True(t, called1)
	assert.True(t, called2)
}

func setup() *miniredis.Miniredis {
	// 初始化redis连接
	redis := miniredis.NewMiniRedis()
	redis.Start()
	// 初始化配置
	content := fmt.Sprintf(`
service:
  sub:
    redis_stream:
      mock_server_1:
        addr: %s
        group_name: mock_case_1
      mock_server_2:
        addr: %s
        group_name: mock_case_2
redis_stream:
  mock_client_1:
    addr: %s
`, redis.Addr(), redis.Addr(), redis.Addr())
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

	return redis
}
