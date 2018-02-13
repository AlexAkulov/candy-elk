package amqp

import (
	"github.com/streadway/amqp"

	"github.com/alexakulov/candy-elk"
	publisheramqp "github.com/alexakulov/candy-elk/amqp"

)

type MockBroker struct {
	channel chan amqp.Publishing
}

func (mb *MockBroker) Start() error {
	mb.channel = make(chan amqp.Publishing, 1000)
	return nil
}

func (mb *MockBroker) PublishLogMessages(bulk []*elkstreams.LogMessage) error {
	b := publisheramqp.Publisher{}
	mb.channel <- b.CreateAMQPBulk(bulk)
	return nil
}

func (mb *MockBroker) GetMessage() (m amqp.Publishing) {
	if len(mb.channel) > 0 {
		m = <- mb.channel
	}
	return
}
