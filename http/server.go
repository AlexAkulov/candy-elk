package http

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/tomb.v2"

	"github.com/alexakulov/candy-elk"
)

const (
	// DefaultType used when type don't defined
	DefaultType = "LogEvent"
)

// Server processes incoming log messages
type Server struct {
	Config        Config
	Publisher     elkstreams.Publisher
	Log           elkstreams.Logger
	MetricStorage elkstreams.MetricStorage
	tomb          tomb.Tomb
	metrics       struct {
		response    map[string]elkstreams.MetricCounter
		requestTime map[string]elkstreams.MetricHistogram
	}
}

// Start initializes HTTP request handling
func (h *Server) Start() error {
	h.metrics.response = make(map[string]elkstreams.MetricCounter)
	h.metrics.requestTime = make(map[string]elkstreams.MetricHistogram)
	for _, n := range []string{"total", "200", "400", "401", "403", "405", "500"} {
		h.metrics.response[n] = h.MetricStorage.RegisterCounter("http.response." + n)
		h.metrics.requestTime[n] = h.MetricStorage.RegisterHistogram("http.request_time." + n)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleRequest)

	server := &http.Server{
		Addr:         h.Config.Address,
		Handler:      mux,
		ReadTimeout:  time.Duration(h.Config.Timeout) * time.Second,
		WriteTimeout: time.Duration(h.Config.Timeout) * time.Second,
		IdleTimeout:  time.Duration(h.Config.IdleTimeout) * time.Second,
	}
	server.SetKeepAlivesEnabled(true)

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}

	h.tomb.Go(func() error {
		err := server.Serve(listener)
		select {
		case <-h.tomb.Dying():
			return nil
		default:
			return err
		}
	})

	h.tomb.Go(func() error {
		<-h.tomb.Dying()
		return listener.Close()
	})

	return nil
}

// Stop finishes listening to HTTP
func (h *Server) Stop() error {
	h.tomb.Kill(nil)
	return h.tomb.Wait()
}

func (h *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	start := time.Now()

	statusCode, indexName, indexType, err := h.pipe(r)

	t := float64(time.Since(start) / time.Millisecond)
	h.metrics.response["total"].Add(1)
	h.metrics.requestTime["total"].Observe(t)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		h.Log.Warn("status", statusCode, "index", indexName, "type", indexType, "msg", "JSON processing pipeline failed", "error", err)
		h.metrics.response[strconv.Itoa(statusCode)].Add(1)
		h.metrics.requestTime[strconv.Itoa(statusCode)].Observe(t)
		return
	}
	h.metrics.response["200"].Add(1)
	h.metrics.requestTime["200"].Observe(t)
	h.Log.Debug("status", statusCode, "index", indexName, "type", indexType, "duration", t)
	w.WriteHeader(statusCode)
}

func (h *Server) pipe(r *http.Request) (statusCode int, indexName, indexType string, err error) {
	if r.Method != http.MethodPost {
		statusCode = http.StatusMethodNotAllowed
		err = fmt.Errorf("only POST method supported")
		return
	}

	if indexName, indexType, err = h.readPath(r.URL.Path); err != nil {
		statusCode = http.StatusBadRequest
		return
	}

	if statusCode, err = h.authorize(r.Header.Get("Authorization"), indexName); err != nil {
		return
	}

	var decodedMessages []*elkstreams.LogMessage
	if decodedMessages, err = h.decodeMessages(indexName, indexType, r.Body); err != nil {
		statusCode = http.StatusBadRequest
		return
	}

	if err = h.Publisher.Publish(decodedMessages); err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	return
}
