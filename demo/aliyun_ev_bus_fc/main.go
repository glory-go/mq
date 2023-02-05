package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alibabacloud-go/eventbridge-sdk/eventbridge"
	geventbridge "github.com/glory-go/mq/v2/aliyun_event_bridge"
	fc "github.com/glory-go/mq/v2/aliyun_fc"
	"github.com/hibiken/asynq"
)

const (
	Source    = "dev"
	Subject   = "dev:glory_demo"
	EventType = "glory_demo:aliyun_ev_bus_fc"
)

func main() {
	// if err := publishMsg(); err != nil {
	// 	panic(err)
	// }
	subscribeMsg()
}

func publishMsg() error {
	pubClient := geventbridge.GetClient("default")
	msg := "hello, glory, " + time.Now().String()
	event := geventbridge.NewEvent(Source, Subject, []byte(msg)).
		GenerateEventBusEvent(EventType, geventbridge.GetConf("default").EventBusName)
	res, err := pubClient.PutEvents([]*eventbridge.CloudEvent{event})
	if err != nil {
		return err
	}
	fmt.Println(res.String())
	return nil
}

func subscribeMsg() {
	fc.GetMux("default").HandleFunc(EventType, func(ctx context.Context, t *asynq.Task) error {
		fmt.Printf("receive msg with payload %s", string(t.Payload()))
		return nil
	})
}
