package amqp

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"gopkg.in/tomb.v2"

	"github.com/alexakulov/candy-elk"
)

// Publisher is an AMQP implementation of elkstreams.Publisher interface
type Publisher struct {
	Config        ConfigPublisher
	Log           elkstreams.Logger
	MetricStorage elkstreams.MetricStorage
	connection    *amqp.Connection
	channel       *amqp.Channel
	amqpErrors    chan *amqp.Error

	tomb    tomb.Tomb
	metrics struct {
		publushTimeTotal     elkstreams.MetricHistogram
		publishBulksTotal    elkstreams.MetricCounter
		publishMessagesTotal elkstreams.MetricCounter
		// publishBulksByIndex
		// publishMessagesByIndex
	}
}

func (b *Publisher) Close() {
	if b.channel != nil {
		b.channel.Close()
	}
	if b.connection != nil {
		b.connection.Close()
	}
}

func (b *Publisher) makeConnection() error {
	b.Close()
	var err error
	defer func(err error) {
		if err != nil {
			b.Log.Error("msg", "connection failed", "err", err)
		} else {
			b.Log.Info("msg", "connection established")
		}
	}(err)

	if b.connection, err = amqp.Dial(b.Config.URL); err != nil {
		return err
	}

	if b.channel, err = b.connection.Channel(); err != nil {
		return err
	}
	b.amqpErrors = make(chan *amqp.Error, 1)
	b.channel.NotifyClose(b.amqpErrors)
	if err = b.channel.ExchangeDeclare(
		b.Config.Exchange, // name
		"direct",          // type
		true,              // durable
		false,             // autoDelete
		false,             // internal
		false,             // noWait
		nil,               // args
	); err != nil {
		return fmt.Errorf("cannot declare fanout exchange: %v", err)
	}

	return nil
}

// Start initializes AMQP connections
func (b *Publisher) Start() error {
	b.metrics.publushTimeTotal = b.MetricStorage.RegisterHistogram("amqp.publish_time.total")
	b.metrics.publishBulksTotal = b.MetricStorage.RegisterCounter("amqp.bulks.total")
	b.metrics.publishMessagesTotal = b.MetricStorage.RegisterCounter("amqp.messages.total")

	b.tomb.Go(func() error {
		b.makeConnection()
		ticker := time.NewTicker(time.Second * time.Duration(b.Config.ReconnectInterval))
		for {
			select {
			case <-b.tomb.Dying(): // Exit
				b.Close()
				return nil
			case err := <-b.amqpErrors:
				b.Log.Error("msg", "error communicating with remote server", "error", err)
				time.Sleep(time.Second * time.Duration(b.Config.ReconnectInterval))
				b.makeConnection()
			case <-ticker.C:
				if b.connection == nil || b.channel == nil {
					b.makeConnection()
				}
			}
		}
	})
	return nil
}

// Stop flushes and stops publishing
func (b *Publisher) Stop() error {
	b.tomb.Go(func() error {
		timer := time.NewTimer(10 * time.Second)
		select {
		case <-b.tomb.Dying():
			return nil
		case <-timer.C:
			return fmt.Errorf("at least one publishing timed out, had to cancel")
		}
	})

	b.tomb.Kill(nil)
	return b.tomb.Wait()
}
