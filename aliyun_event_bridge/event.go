package aliyuneventbridge

import (
	"time"

	"github.com/alibabacloud-go/eventbridge-sdk/eventbridge"
	"github.com/rs/xid"
)

type event struct {
	ID      string
	Source  string
	Subject string
	Payload []byte

	EventTime time.Time
}

func NewEvent(source, subject string, payload []byte) *event {
	return &event{
		Source:  source,
		Subject: subject,
		Payload: payload,
	}
}

func (e *event) GenerateEventBusEvent(eventType, eventBusName string) *eventbridge.CloudEvent {
	if e.ID == "" {
		e.ID = xid.New().String()
	}
	if e.EventTime.IsZero() {
		e.EventTime = time.Now()
	}

	event := new(eventbridge.CloudEvent).
		SetDatacontenttype("application/json").
		SetData(e.Payload).
		SetId(e.ID).
		SetSource(e.Source).
		SetTime(e.EventTime.Format("2006-01-02T15:04:05-07:00")).
		SetSubject(e.Subject).
		SetType(eventType).
		SetExtensions(map[string]interface{}{
			"aliyuneventbusname": eventBusName,
		})
	return event
}
