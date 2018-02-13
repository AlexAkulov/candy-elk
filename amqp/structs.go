package amqp

// BulkHeader is a header of ES bulk insert request
type BulkHeader struct {
	Index struct {
		MessageIndex string `json:"_index"`
		MessageType  string `json:"_type"`
	} `json:"index"`
}
