package amqp

import (
	"fmt"

	"github.com/streadway/amqp"
)


// PreparePipe - will be created exchange, queue and binding
func PreparePipe(channel *amqp.Channel, exchange, key, queue string) error {
	// Exchange must be created
	if err := channel.ExchangeDeclare(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // autoDelete
		false,    // internal
		false,    // noWait
		nil,      // args
	); err != nil {
		return fmt.Errorf("cannot declare fanout exchange: %v", err)
	}

	// Queue must be created
	if _, err := channel.QueueDeclare(
		queue, // name
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	); err != nil {
		return fmt.Errorf("cannot declare queue: %v", err)
	}

	// Binging must be created
	if err := channel.QueueBind(
		exchange, // destination
		key,      // key
		queue,    // source
		false,    // noWait
		nil,      // args
	); err != nil {
		return fmt.Errorf("cannot create binding: %v", err)
	}
	return nil
}
