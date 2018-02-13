package metrics

import (
	"net"
	"time"

	"github.com/go-kit/kit/metrics/graphite"

	"github.com/alexakulov/candy-elk"
)

// MetricStorage is a Graphite implementation of elkstreams.MetricStorage interface
type MetricStorage struct {
	Config   Config
	Log      elkstreams.Logger
	registry *graphite.Graphite
}

func (ms *MetricStorage) run() {
	t := time.NewTicker(60 * time.Second)
	go func(c <-chan time.Time) {
		for range c {
			addr, err := net.ResolveTCPAddr("tcp", ms.Config.GraphiteConnectionString)
			if err != nil {
				ms.Log.Warn("msg", "can't resolve graphite uri", "error", err)
				continue
			}
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				ms.Log.Warn("msg", "can't connect to graphite", "error", err)
				continue
			}
			if _, err := ms.registry.WriteTo(conn); err != nil {
				ms.Log.Warn("msg", "can't send metrics to graphite", "err", err)
			}
			conn.Close()
			ms.Log.Debug("msg", "sent stats to graphite successfully")
		}
	}(t.C)

}

// Start initializes Graphite reporter
func (ms *MetricStorage) Start() error {
	ms.registry = graphite.New(ms.Config.GraphitePrefix, nil)
	if ms.Config.Enabled {
		ms.Log.Debug("msg", "Graphite enabled")
		ms.run()
		return nil
	}
	ms.Log.Debug("msg", "Graphite disabled")
	return nil
}

// Stop does nothing - there is no way to gracefully flush Graphite reporter
func (ms *MetricStorage) Stop() error {
	return nil
}

// RegisterHistogram creates a uniform-sampled histogram of integers
func (ms *MetricStorage) RegisterHistogram(name string) elkstreams.MetricHistogram {
	return ms.registry.NewHistogram(name, 50)
}

// RegisterCounter creates a counter
func (ms *MetricStorage) RegisterCounter(name string) elkstreams.MetricCounter {
	return ms.registry.NewCounter(name)
}

// RegisterGauge creates a gauge
func (ms *MetricStorage) RegisterGauge(name string) elkstreams.MetricGauge {
	return ms.registry.NewGauge(name)
}
