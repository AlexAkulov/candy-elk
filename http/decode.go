package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AlexAkulov/candy-elk"
)

func (h *Server) decodeMessages(indexName string, indexType string, body io.Reader) ([]*elkstreams.LogMessage, error) {
	var (
		bulk       []*elkstreams.LogMessage
		badLines   int
		totalLines int
	)

	bodyReader := bufio.NewReader(body)
	var line bytes.Buffer
	for {
		part, isPrefix, err := bodyReader.ReadLine()
		if err == nil {
			line.Write(part)
			if !isPrefix {
				totalLines++
				if err := h.checkMessage(line.Bytes()); err != nil {
					h.Log.Debug("msg", err, "body", line)
					badLines++
				} else {
					l := make([]byte, line.Len())
					copy(l, line.Bytes())

					bulk = append(bulk, &elkstreams.LogMessage{
						IndexName: indexName,
						IndexType: indexType,
						Body:      l,
					})
				}
				line.Reset()
			}
			continue
		}
		if err != io.EOF {
			return nil, fmt.Errorf("can't read body: %s", err)
		}
		break
	}
	if len(bulk) < 1 {
		return nil, fmt.Errorf("bad request. %d of %d lines is bad", badLines, totalLines)
	}
	if badLines > 0 {
		h.Log.Warn("index", indexName, "type", indexType, "msg", "bulk contains bad lines", "bad_count", badLines, "total_count", totalLines)
	}
	return bulk, nil
}

func (h *Server) checkMessage(message []byte) error {
	var decodedMessage map[string]interface{}
	if err := json.Unmarshal(message, &decodedMessage); err != nil {
		return fmt.Errorf("cannot unmarshal json")
	}

	if _, ok := decodedMessage["@timestamp"]; !ok {
		return fmt.Errorf("@timestamp field doesn't exist")
	}

	return nil
}
