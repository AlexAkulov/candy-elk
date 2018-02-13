package elkstreams

// Logger is a simple logging wrapper interface
type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Printf(format string, v ...interface{})
	SetLevel(string)
}



// MetricStorage is a way to store internal application metrics
type MetricStorage interface {
	RegisterHistogram(string) MetricHistogram
	RegisterCounter(string) MetricCounter
	RegisterGauge(string) MetricGauge
}

// MetricHistogram is a simple histogram
type MetricHistogram interface {
	Observe(float64)
}

// MetricCounter is a simple counter
type MetricCounter interface {
	Add(float64)
}

// MetricGauge is a simple Gauge
type MetricGauge interface {
	Add(float64)
	Set(float64)
}

// Service is started and stopped in main function, which assembles services into a working application
type Service interface {
	Start() error
	Stop() error
}
