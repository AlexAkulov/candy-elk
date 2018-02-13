package amqp

import (
	"time"
)

// ConfigPublisher setting
type ConfigPublisher struct {
	URL               string `yaml:"url"`
	Exchange          string `yaml:"exchange"`
	RoutingKey        string `yaml:"key"`
	PublishTimeout    int64  `yaml:"publish_timeout"`
	ReconnectInterval int64  `yaml:"reconnect_interval"`
}

// ConnectionConfig settings
type ConnectionConfig struct {
	URL               string `yaml:"url"`
	Exchange          string `yaml:"exchange"`
	RoutingKey        string `yaml:"key"`
	Queue             string `yaml:"queue"`
	PrefetchCount     int    `yaml:"prefetch_count"`
	ReconnectInterval int64  `yaml:"reconnect_interval"`
	reconnectInterval time.Duration
	WaitAck           string `yaml:"wait_ack"`
	waitAck           bool
}

// ConfigConsumer settings
type ConfigConsumer struct {
	Connections []ConnectionConfig `yaml:"connections"`
}
