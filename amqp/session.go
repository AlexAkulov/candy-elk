package amqp

import (
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"gopkg.in/tomb.v2"

	"github.com/alexakulov/candy-elk"
	"github.com/alexakulov/candy-elk/helpers"
)

const (
	// ConsumerName AMQP ConsumerName
	ConsumerName = "elkriver"
)

type Session struct {
	Config     ConnectionConfig
	Log        elkstreams.Logger
	Publisher  elkstreams.Publisher
	connection *amqp.Connection
	channel    *amqp.Channel
	delivery   <-chan amqp.Delivery
	amqpErrors chan *amqp.Error
	tomb       tomb.Tomb
	active     bool
}

func (session *Session) tryConnect() {
	var err error
	defer func() {
		if err != nil && session.connection != nil {
			session.connection.Close()
		}
	}()

	if session.connection, err = amqp.Dial(session.Config.URL); err != nil {
		session.Log.Error("msg", "can't connect to rabbitmq", "err", err)
		return
	}
	if session.channel, err = session.connection.Channel(); err != nil {
		session.Log.Error("msg", "can't create channel", "err", err)
		return
	}

	if err = PreparePipe(
		session.channel,
		session.Config.Exchange,
		session.Config.RoutingKey,
		session.Config.Queue,
	); err != nil {
		session.Log.Error("msg", "prepare rabbitmq failed, maybe queue, exchange or bind alreary exists with bad settings", "err", err)
		return
	}
	if session.Config.PrefetchCount < 1 {
		session.Log.Error("msg", "prefetch_count can't equal 0")
		return
	}
	if err = session.channel.Qos(session.Config.PrefetchCount, 0, false); err != nil {
		session.Log.Error("msg", "can't set qos", "err", err)
		return
	}

	if session.delivery, err = session.channel.Consume(
		session.Config.Queue,    // queue
		ConsumerName,            // consumer
		!session.Config.waitAck, // autoAck
		false, // exclusive
		true,  // noLocal - separate connections for Channel.Consume and ACKs
		false, // noWait
		nil,   // args
	); err != nil {
		session.Log.Error("msg", "can't delivery channel", "err", err)
		return
	}

	session.amqpErrors = make(chan *amqp.Error, 1)
	session.channel.NotifyClose(session.amqpErrors)

	go session.get()
}

func (session *Session) createStableConnect() error {
	ticker := time.NewTicker(session.Config.reconnectInterval)
	session.tryConnect()
	for {
		select {
		case <-session.tomb.Dying(): // Exit
			if session.channel != nil {
				err := session.channel.Cancel(ConsumerName, session.Config.waitAck)
				return err
			}
			// session.channel.Close()
			// session.connection.Close()
			return nil
		case err := <-session.amqpErrors:
			if err != nil {
				session.Log.Error("msg", "connection lost", "err", err)
			}

			// session.channel.Close()
			// session.connection.Close()
			// time.Sleep(session.Config.reconnectInterval)
			// session.tryConnect()
		case <-ticker.C: // if rabbitmq not avaliable on start
			// session.Log.Debug("msg", "check connection")
			if !session.active {
				session.tryConnect()
			}
		}
	}
}

func (session *Session) Start() error {
	session.Config.reconnectInterval = time.Duration(session.Config.ReconnectInterval) * time.Second
	session.Config.waitAck = helpers.ToBool(session.Config.WaitAck)
	session.tomb.Go(session.createStableConnect)
	return nil
}

func (session *Session) get() {
	session.Log.Debug("msg", "start delivery", "queue", session.Config.Queue)
	session.active = true
	for message := range session.delivery {
		// fmt.Println(string(message.Body))
		go func(m amqp.Delivery) {
			if len(m.Body) == 0 {
				if session.Config.waitAck {
					// удаляем из рэббита пустые сообщения
					if err := m.Ack(false); err != nil {
						session.Log.Warn("msg", "can't send ack for empty message", "err", err)
					}
				}
				return
			}
			bulk, err := decodeAMQPBulkLegacy(&m)
			if err != nil {
				session.Log.Warn("msg", "bad message", "err", err, "body", string(m.Body))
				if session.Config.waitAck {
					// удаляем из рэббита плохие сообщения
					if err := m.Ack(false); err != nil {
						session.Log.Warn("msg", "can't send ack for bad message", "err", err)
					}
				}
				return
			}
			if session.Config.waitAck {
				var ack sync.WaitGroup
				ack.Add(len(bulk))
				for i, _ := range bulk {
					bulk[i].Ack = &ack
				}
				session.Publisher.Publish(bulk)
				ack.Wait()
				if err := m.Ack(false); err != nil {
					session.Log.Warn("msg", "can't send ack", "err", err)
				}
			} else {
				session.Publisher.Publish(bulk)
			}
		}(message)

		// message.Ack(false)
		// go func(m amqp.Delivery) {
		// 	if session.Config.waitAck {
		// 		time.Sleep(time.Second * 4)
		// 		// fmt.Println(string(m.Body), "delivered")
		// 		// m.Ack(false)
		// 		fmt.Println(string(m.Body), "undelivered")
		// 		m.Nack(false, true)
		// 	}
		// }(message)
	}
	session.Log.Debug("msg", "stop delivery", "queue", session.Config.Queue)
	session.connection.Close()
	session.active = false
	// TODO: где-то тут нужно обработать удаление очереди в рэббите что-бы переподключиться немедленно
	// сейчас он переключится только после session.Config.ReconnectInterval
	// session.amqpErrors <- amqp.ErrClosed приводит к panic: send on closed channel
}

func (session *Session) Stop() error {
	session.Log.Debug("msg", "stop")
	session.tomb.Go(func() error {
		timer := time.NewTimer(30 * time.Second)
		select {
		case <-session.tomb.Dying():
			return nil
		case <-timer.C:
			return fmt.Errorf("Consumer did not closed after 30 seconds")
		}
	})
	session.tomb.Kill(nil)
	err := session.tomb.Wait()
	session.Log.Debug("msg", "stopped", "err", err)
	return err
}
