package profiler

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/alexakulov/candy-elk"
	"github.com/alexakulov/candy-elk/helpers"
)

// Profiler is implementation of Profiler
type Profiler struct {
	Config Config
	Log    elkstreams.Logger
}

// Start is
func (p *Profiler) Start() error {
	if !helpers.ToBool(p.Config.Enabled) {
		p.Log.Info("msg", "profiler disabled")
		return nil
	}
	go func() {
		if err := http.ListenAndServe(p.Config.Listen, nil); err != nil {
			p.Log.Error("msg", "Error starting profiling", "error", err)
		}
	}()
	return nil
}

// Stop is
func (p *Profiler) Stop() error {
	return nil
}
