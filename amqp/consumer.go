package amqp

import (
	"github.com/alexakulov/candy-elk"
	"github.com/alexakulov/candy-elk/logger"
)

// Consumer type
type Consumer struct {
	Config    ConfigConsumer
	Publisher elkstreams.Publisher
	Log       elkstreams.Logger
	sessions  []*Session
}

// Start consumer
func (consumer *Consumer) Start() error {
	consumer.sessions = make([]*Session, len(consumer.Config.Connections))
	for i, sessionConfig := range consumer.Config.Connections {
		consumer.sessions[i] = &Session{
			Config: sessionConfig,
			Publisher: consumer.Publisher,
			Log:    logger.With(consumer.Log.(*logger.Logger), "session", i),
		}
		if err := consumer.sessions[i].Start(); err != nil {
			return err
		}
	}
	consumer.Log.Debug("msg", "started")
	return nil
}

// Stop consumer
func (consumer *Consumer) Stop() error {
	for i := range consumer.sessions {
		// TODO: тут нужно завершать сессии одновременно и жидать когда все закроются
		if err := consumer.sessions[i].Stop(); err != nil {
			consumer.Log.Error("session", i, "msg", "don't clean stop", "err", err)
		}
	}
	consumer.Log.Debug("msg", "stop")
	return nil
}
