package amqp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"

	"github.com/alexakulov/candy-elk"
	"github.com/alexakulov/candy-elk/logger"
)

func TestAMQP(t *testing.T) {
	b := Publisher{
		Log: logger.NewNopLogger(),
	}
	// c := Consumer{
		// Log: logger.NewNopLogger(),
	// 	Log: logger.New("debug", os.Stdout),
	// }

	Convey("Once message bulk", t, func() {
		testBulk := []*elkstreams.LogMessage{
			&elkstreams.LogMessage{
				IndexName: "index-abc",
				IndexType: "type-abc",
				Body:      []byte("{\"message\":\"content\"}"),
			}}

		expectedMessage := "{\"index\": {\"_index\": \"index-abc\", \"_type\": \"type-abc\"}}\n" +
			"{\"message\":\"content\"}\n"
		Convey("Publish", func() {
			m := b.CreateAMQPBulk(testBulk)
			So(string(m.Body), ShouldEqual, expectedMessage)
		})
		Convey("Receive", func() {
			m, err := decodeAMQPBulkLegacy(&amqp.Delivery{
				Body: []byte(expectedMessage),
			})
			So(err, ShouldBeNil)
			So(m, ShouldResemble, testBulk)
		})

	})
	Convey("Once message new bulk", t, func() {
		testBulk := []*elkstreams.LogMessage{
			&elkstreams.LogMessage{
				IndexName: "index-abc",
				IndexType: "type-abc",
				Body:      []byte("{\"message\":\"content\"}"),
			}}

		expectedMessage := amqp.Publishing{
			Headers: amqp.Table{
				"index": "index-abc",
				"type":  "type-abc",
			},
			Body: []byte("{\"message\":\"content\"}\n"),
		}
		m := b.CreateNewAMQPBulk(testBulk)
		So(m, ShouldResemble, expectedMessage)
	})
	Convey("A few messages in bulk", t, func() {
		indexName := "index"
		indexType := "type"
		testBulk := []*elkstreams.LogMessage{
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"message1\":\"content1\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"message2\":\"content2\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"message3\":\"content3\"}"),
			},
		}
		expectedMessage := "{\"index\": {\"_index\": \"index\", \"_type\": \"type\"}}\n" +
			"{\"message1\":\"content1\"}\n" +
			"{\"index\": {\"_index\": \"index\", \"_type\": \"type\"}}\n" +
			"{\"message2\":\"content2\"}\n" +
			"{\"index\": {\"_index\": \"index\", \"_type\": \"type\"}}\n" +
			"{\"message3\":\"content3\"}\n"
		Convey("Publish", func() {
			m := b.CreateAMQPBulk(testBulk)
			So(string(m.Body), ShouldEqual, expectedMessage)
		})
		Convey("Receive", func() {
			m, err := decodeAMQPBulkLegacy(&amqp.Delivery{
				Body: []byte(expectedMessage),
			})
			So(err, ShouldBeNil)
			So(m, ShouldResemble, testBulk)
		})

	})
	Convey("A few messages in bulk for different indices", t, func() {
		testBulk := []*elkstreams.LogMessage{
			&elkstreams.LogMessage{
				IndexName: "index-1",
				IndexType: "type1",
				Body:      []byte("{\"message1\":\"content1\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: "index-2",
				IndexType: "type2",
				Body:      []byte("{\"message2\":\"content2\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: "index-3",
				IndexType: "type3",
				Body:      []byte("{\"message3\":\"content3\"}"),
			},
		}
		expectedMessage := "{\"index\": {\"_index\": \"index-1\", \"_type\": \"type1\"}}\n" +
			"{\"message1\":\"content1\"}\n" +
			"{\"index\": {\"_index\": \"index-2\", \"_type\": \"type2\"}}\n" +
			"{\"message2\":\"content2\"}\n" +
			"{\"index\": {\"_index\": \"index-3\", \"_type\": \"type3\"}}\n" +
			"{\"message3\":\"content3\"}\n"
		Convey("Publish", func() {
			m := b.CreateAMQPBulk(testBulk)
			So(string(m.Body), ShouldEqual, expectedMessage)
		})
		Convey("Receive", func() {
			m, err := decodeAMQPBulkLegacy(&amqp.Delivery{
				Body: []byte(expectedMessage),
			})
			So(err, ShouldBeNil)
			So(m, ShouldResemble, testBulk)
		})
	})
}
