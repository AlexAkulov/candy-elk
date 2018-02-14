package http

import (
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/AlexAkulov/candy-elk"
	"github.com/AlexAkulov/candy-elk/logger"
)

func TestHTTP(t *testing.T) {
	apiKeys := make(map[string][]string)
	apiKeys["test-apikey"] = []string{"index1-????.??.??", "index2-*"}
	h := Server{
		Config: Config{
			APIKeys: apiKeys,
		},
		Log: logger.NewNopLogger(),
	}

	Convey("URI path ends with slash hould be read correct", t, func() {
		indexName, typeName, err := h.readPath("/logs/index/type")
		So(err, ShouldBeNil)
		So(indexName, ShouldEqual, "index")
		So(typeName, ShouldEqual, "type")
	})
	Convey("URI path ends without slash hould be read correct", t, func() {
		indexName, typeName, err := h.readPath("/logs/index/type/")
		So(err, ShouldBeNil)
		So(indexName, ShouldEqual, "index")
		So(typeName, ShouldEqual, "type")
	})
	Convey("URI path contains excess slashes should be return error", t, func() {
		_, _, err := h.readPath("/logs/index/type/a/b/c")
		So(err, ShouldNotBeNil)
	})
	Convey("URI path contains double slashes should be read correct", t, func() {
		indexName, typeName, err := h.readPath("//logs///index////type//")
		So(err, ShouldBeNil)
		So(indexName, ShouldEqual, "index")
		So(typeName, ShouldEqual, "type")
	})
	Convey("URI path without type should be read default type", t, func() {
		indexName, typeName, err := h.readPath("/logs/index/")
		So(err, ShouldBeNil)
		So(indexName, ShouldEqual, "index")
		So(typeName, ShouldEqual, DefaultType)
	})

	Convey("Authorization header is empty then should return 401 error", t, func() {
		code, err := h.authorize("", "")
		So(err, ShouldNotBeNil)
		So(code, ShouldEqual, http.StatusUnauthorized)
	})
	Convey("Authorization header is valid then should not return error", t, func() {
		_, err := h.authorize("ELK test-apikey", "index1-yyyy.mm.dd")
		So(err, ShouldBeNil)
	})
	Convey("Authorization header is valid should not then return error", t, func() {
		_, err := h.authorize("ELK test-apikey", "index2-abc")
		So(err, ShouldBeNil)
	})
	Convey("Authorization exists, but header format is bad then should return 401", t, func() {
		code, err := h.authorize("test-apikey", "index2-abc")
		So(err, ShouldNotBeNil)
		So(code, ShouldEqual, http.StatusUnauthorized)
	})
	Convey("Authorization key is not exists then should return 401", t, func() {
		code, err := h.authorize("ELK bad-apikey", "index2-abc")
		So(err, ShouldNotBeNil)
		So(code, ShouldEqual, http.StatusUnauthorized)
	})
	Convey("Authorization key exists, but index has not then should return 403 error", t, func() {
		code, err := h.authorize("ELK test-apikey", "index3")
		So(err, ShouldNotBeNil)
		So(code, ShouldEqual, http.StatusForbidden)
	})
	Convey("Decode one line body", t, func() {
		indexName := "index"
		indexType := "type"
		expectedMessage := []*elkstreams.LogMessage{
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message\":\"content\"}"),
			},
		}
		Convey("with Unix EOL", func() {
			body := strings.NewReader("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message\":\"content\"}\n")
			m, err := h.decodeMessages(indexName, indexType, body)
			So(err, ShouldBeNil)
			So(len(m), ShouldEqual, 1)
			So(m, ShouldResemble, expectedMessage)
		})
		Convey("with Windows EOL", func() {
			body := strings.NewReader("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message\":\"content\"}\r\n")
			m, err := h.decodeMessages(indexName, indexType, body)
			So(err, ShouldBeNil)
			So(len(m), ShouldEqual, 1)
			So(m, ShouldResemble, expectedMessage)
		})
		Convey("without EOL", func() {
			body := strings.NewReader("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message\":\"content\"}")
			m, err := h.decodeMessages(indexName, indexType, body)
			So(err, ShouldBeNil)
			So(len(m), ShouldEqual, 1)
			So(m, ShouldResemble, expectedMessage)
		})
	})
	Convey("Decode bad message", t, func() {
		body := strings.NewReader("bad line")
		_, err := h.decodeMessages("index", "type", body)
		So(err, ShouldNotBeNil)
	})
	Convey("Decode message without @timestamp field", t, func() {
		body := strings.NewReader("{\"message1\":\"content1\"}")
		_, err := h.decodeMessages("index", "type", body)
		So(err, ShouldNotBeNil)
	})
	Convey("Decode bulk with bad line", t, func() {
		body := "{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message1\":\"content1\"}\n" +
			"bad line\n" +
			"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message3\":\"content3\"}\n"
		m, err := h.decodeMessages("index", "type", strings.NewReader(body))
		So(err, ShouldBeNil)
		So(len(m), ShouldEqual, 2)
	})
	Convey("Decode bad bulk", t, func() {
		body := "{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message1\"\"content1\"}\n" +
			"bad line\n" +
			"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message3\":\"content3\"\n"
		_, err := h.decodeMessages("index", "type", strings.NewReader(body))
		So(err, ShouldNotBeNil)
	})
	Convey("Decode bulk message", t, func() {
		indexName := "index"
		indexType := "type"
		expectedMessage := []*elkstreams.LogMessage{
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message1\":\"content1\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message2\":\"content2\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message3\":\"content3\"}"),
			},
		}
		Convey("with Unix EOL", func() {
			body := "{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message1\":\"content1\"}\n" +
				"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message2\":\"content2\"}\n" +
				"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message3\":\"content3\"}\n"
			m, err := h.decodeMessages(indexName, indexType, strings.NewReader(body))
			So(err, ShouldBeNil)
			So(len(m), ShouldEqual, 3)
			So(m, ShouldResemble, expectedMessage)
		})
		Convey("with Windows EOL", func() {
			body := "{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message1\":\"content1\"}\r\n" +
				"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message2\":\"content2\"}\r\n" +
				"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message3\":\"content3\"}\r\n"
			m, err := h.decodeMessages(indexName, indexType, strings.NewReader(body))
			So(err, ShouldBeNil)
			So(len(m), ShouldEqual, 3)
			So(m, ShouldResemble, expectedMessage)
		})
	})
	Convey("Decode large bulk message", t, func() {
		indexName := "index"
		indexType := "type"
		content := make([]byte, 1024*1024)
		const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := range content {
			content[i] = chars[r.Intn(len(chars))]
		}
		expectedMessage := []*elkstreams.LogMessage{
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message1\":\"content1\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message2\":\"" + string(content) + "\"}"),
			},
			&elkstreams.LogMessage{
				IndexName: indexName,
				IndexType: indexType,
				Body:      []byte("{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message3\":\"content3\"}"),
			},
		}
		body := "{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message1\":\"content1\"}\n" +
			"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message2\":\"" + string(content) + "\"}\n" +
			"{\"@timestamp\":\"2017-06-28T01:00:00.000Z\",\"message3\":\"content3\"}\n"
		m, err := h.decodeMessages(indexName, indexType, strings.NewReader(body))
		So(err, ShouldBeNil)
		So(len(m), ShouldEqual, 3)
		So(m, ShouldResemble, expectedMessage)
	})
}
