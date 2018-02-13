// +build integration

package main

import (
	"bytes"
	_http "net/http"
	"testing"

	"github.com/go-kit/kit/log"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/alexakulov/candy-elk/mock/amqp"
	"github.com/alexakulov/candy-elk/http"
	"github.com/alexakulov/candy-elk/metrics"
	"fmt"
)

var (
	metricsServer  *metrics.MetricStorage
	broker         *amqp.MockBroker
	handler        *http.Server
)

func TestIntegration(t *testing.T) {
	Convey("Metrics must be started", t, func() {
		metricsServer = &metrics.MetricStorage{
			Config: metrics.Config{
				Enabled: false,
			},
			Logger: log.NewNopLogger(),
		}
		So(metricsServer.Start(), ShouldBeNil)
	})
	Convey("MockBrocker must be started",t,func(){
		broker = &amqp.MockBroker{}
		So(broker.Start(),ShouldBeNil)
	})
	Convey("HTTP Server must be started", t, func() {
		apiKeys := make(map[string][]string)
		apiKeys["test-apikey"] = []string{"index1-????.??.??", "index2-*"}

		handler = &http.Server{
			Config: http.Config{
				Address: ":8181",
				APIKeys: apiKeys,
			},
			MetricStorage: metricsServer,
			Broker:        broker,
			Logger:        log.NewNopLogger(),
		}
		So(handler.Start(), ShouldBeNil)
	})
	Convey("Response 401 when Authorization header doesn't exists", t, func() {
		res, err := _http.Post(
			"http://localhost:8181/logs/index2-1",
			"",
			bytes.NewBufferString(""),
		)
		So(err, ShouldBeNil)
		So(res.StatusCode, ShouldEqual, _http.StatusUnauthorized)
	})
	Convey("Response 401 when Authorization header is bad", t, func() {
		req, _ := _http.NewRequest(
			"POST",
			"http://localhost:8181/logs/index2-1",
			bytes.NewBufferString(""))
		req.Header.Set("Authorization", "bad")
		res, err := _http.DefaultClient.Do(req)
		So(err, ShouldBeNil)
		So(res.StatusCode, ShouldEqual, _http.StatusUnauthorized)
	})
	Convey("One message without EOL must be applied", t, func() {
		body := `{"field1": "text1", "field2": "text2"}`
		req, _ := _http.NewRequest(
			"POST",
			"http://localhost:8181/logs/index2-1",
			bytes.NewBufferString(body))
		req.Header.Set("Authorization", "ELK test-apikey")
		res, err := _http.DefaultClient.Do(req)
		fmt.Println(broker.GetMessage())
		So(err, ShouldBeNil)
		So(res.StatusCode, ShouldEqual, _http.StatusOK)
	})
	Convey("One message with linux EOL must be applied", t, func() {
		body := `{"field1": "text1", "field2": "text2"}\n`
		req, _ := _http.NewRequest(
			"POST",
			"http://localhost:8181/logs/index2-1",
			bytes.NewBufferString(body))
		req.Header.Set("Authorization", "ELK test-apikey")
		res, err := _http.DefaultClient.Do(req)
		fmt.Println(broker.GetMessage())
		So(err, ShouldBeNil)
		So(res.StatusCode, ShouldEqual, _http.StatusOK)
	})
	Convey("One message with windows EOL must be applied", t, func() {
		body := `{"field1": "text1", "field2": "text2"}\n\r`
		req, _ := _http.NewRequest(
			"POST",
			"http://localhost:8181/logs/index2-1",
			bytes.NewBufferString(body))
		req.Header.Set("Authorization", "ELK test-apikey")
		res, err := _http.DefaultClient.Do(req)
		fmt.Println(broker.GetMessage())
		So(err, ShouldBeNil)
		So(res.StatusCode, ShouldEqual, _http.StatusOK)
	})
	Convey("Bulk message with linux EOL must be applied", t, func() {
		body := `{"field1": "text1", "field2": "text2"}\n
				 {"field1": "text1", "field2": "text2"}\n
				 {"field1": "text1", "field2": "text2"}`
		req, _ := _http.NewRequest(
			"POST",
			"http://localhost:8181/logs/index2-1",
			bytes.NewBufferString(body))
		req.Header.Set("Authorization", "ELK test-apikey")
		res, err := _http.DefaultClient.Do(req)
		fmt.Println(string(*broker.GetMessage()))
		So(err, ShouldBeNil)
		So(res.StatusCode, ShouldEqual, _http.StatusOK)
	})
	Convey("Bulk message with windows EOL must be applied", t, func() {
		body := `{"field1": "text1", "field2": "text2"}\r\n
				{"field1": "text1", "field2": "text2"}\r\n
				{"field1": "text1", "field2": "text2"}\r\n`
		req, _ := _http.NewRequest(
			"POST",
			"http://localhost:8181/logs/index2-1",
			bytes.NewBufferString(body))
		req.Header.Set("Authorization", "ELK test-apikey")
		res, err := _http.DefaultClient.Do(req)
		fmt.Println(string(*broker.GetMessage()))
		So(err, ShouldBeNil)
		So(res.StatusCode, ShouldEqual, _http.StatusOK)
	})

	Convey("All Servers must be stopped", t, func() {
		So(handler.Stop(), ShouldBeNil)
		So(metricsServer.Stop(), ShouldBeNil)
	})
}
