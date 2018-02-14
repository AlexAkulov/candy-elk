package amqp

import (
	"bytes"
	"encoding/json"

	"github.com/streadway/amqp"

	"github.com/AlexAkulov/candy-elk"
	"fmt"
)

func decodeAMQPBulkLegacy(message *amqp.Delivery) ([]*elkstreams.LogMessage, error){
	var (
		decodedBulk []*elkstreams.LogMessage
		header BulkHeader
	)
	for i, line := range bytes.Split(message.Body, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		if i%2 == 0 { // Header
			if err := json.Unmarshal(line, &header); err != nil {
				return nil, err
				// consumer.Log.Warn("msg", "Can't parse json header", "body", string(line), "err", err)
			}
			continue
		}
		// Body
		if len(header.Index.MessageIndex) == 0 {
			// consumer.Log.Warn("msg", "Can't parse json", "body", string(line), "err", "empty index name")
			continue
		}
		decodedBulk = append(decodedBulk, &elkstreams.LogMessage{
			IndexName: header.Index.MessageIndex,
			IndexType: header.Index.MessageType,
			Body:      line,
		})

	}
	if len(decodedBulk) < 1 {
		return nil, fmt.Errorf("empty message")
	}
	return decodedBulk, nil
}

func (consumer *Consumer) decodeAMQPBulkNew(message amqp.Delivery) []*elkstreams.LogMessage {
	return nil
}
