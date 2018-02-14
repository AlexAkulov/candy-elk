package amqp

import (
	"bytes"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"github.com/AlexAkulov/candy-elk"
)

// Publish publishes bulk to AMQP
func (b *Publisher) Publish(bulk []*elkstreams.LogMessage) error {
	msg := b.CreateAMQPBulk(bulk)

	if b.channel == nil {
		return fmt.Errorf("channel is nil")
	}
	start := time.Now()
	err := b.channel.Publish(
		b.Config.Exchange,
		b.Config.RoutingKey,
		false,
		false,
		msg)
	// timer := time.NewTimer(time.Duration(b.Config.PublishTimeout) * time.Second)
	t := float64(time.Since(start) / time.Millisecond)
	b.metrics.publushTimeTotal.Observe(t)
	b.metrics.publishBulksTotal.Add(1)
	b.metrics.publishMessagesTotal.Add(float64(len(bulk)))
	return err
}

// CreateAMQPBulk return AMQP message in legacy format
func (b *Publisher) CreateAMQPBulk(bulk []*elkstreams.LogMessage) (amqpBulk amqp.Publishing) {
	var rawBulk bytes.Buffer
	for _, message := range bulk {
		rawBulk.WriteString(fmt.Sprintf("{\"index\": {\"_index\": \"%s\", \"_type\": \"%s\"}}\n", message.IndexName, message.IndexType))
		rawBulk.Write(message.Body)
		rawBulk.WriteByte('\n')
	}
	amqpBulk.Body = rawBulk.Bytes()
	return
}

// CreateNewAMQPBulk return AMQP message in new format
func (b *Publisher) CreateNewAMQPBulk(bulk []*elkstreams.LogMessage) (amqpBulk amqp.Publishing) {
	amqpBulk.Headers = amqp.Table{
		"index": bulk[0].IndexName,
		"type": bulk[0].IndexType,
	}
	var body bytes.Buffer
	for i := range bulk {
		body.Write(bulk[i].Body)
		body.WriteByte('\n')
	}
	amqpBulk.Body = body.Bytes()
	return
}
