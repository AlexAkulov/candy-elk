package elkstreams

import (
	"sync"
)

type DecodedLogMessage struct {
	IndexName string
	IndexType string
	Fields    map[string]interface{}
}

// LogMessage is a single log line
type LogMessage struct {
	IndexName string
	IndexType string
	Body      []byte
	Ack       *sync.WaitGroup
}

// Publisher is a way to publish logs
type Publisher interface {
	Start() error
	Stop() error
	Publish([]*LogMessage) error
}
