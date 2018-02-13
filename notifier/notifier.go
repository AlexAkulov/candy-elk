package notifier

import (
	"fmt"
	"time"


	"gopkg.in/olivere/elastic.v3"

	"github.com/alexakulov/candy-elk"
	"github.com/alexakulov/candy-elk/notifier/meta"
	"github.com/alexakulov/candy-elk/notifier/scheduler"
)

// Publisher is an implementation of elkstreams.Publisher interface for sending Notifications
type Publisher struct {
	Config Config
	Log    elkstreams.Logger
	es     *elastic.Client

	metaUpdater *meta.Updater
	s *scheduler.Scheduler


}

// Start initializes Elasticsearch connection
func (p *Publisher) Start() error {
	var err error
	p.es, err = elastic.NewClient(
		elastic.SetURL(p.Config.ElasticUrls...),
		elastic.SetErrorLog(p.Log),
		elastic.SetHealthcheck(true),
		elastic.SetHealthcheckTimeoutStartup(time.Second),
		)
	if err != nil {
		return err
	}
	p.Log.Debug("msg", "elasticsearch connected")

	p.metaUpdater = &meta.Updater{
		Log: p.Log,
		ESClient: p.es,
	}

	if err := p.metaUpdater.Start(); err != nil {
		return fmt.Errorf("Can't start meta updater: %v", err)
	}

	p.s := &scheduler.Scheduler{
		Log: p.Log,
	}

	if err := p.s.Start(); err != nil {
		return fmt.Errorf("Can't start sheduller: %v", err)
	}

	return err
}

// Stop flushes and stops publishing
func (p *Publisher) Stop() error {
	if err := p.metaUpdater.Stop(); err != nil {
		p.Log.Debug("msg", "stop metaUpdater failed", "err", err)
	}

	if err := p.s.Stop(); err != nil {
		p.Log.Debug("msg", "stop sheduller failed", "err", err)
	}

	p.Log.Debug("msg", "stop elastic")
	p.es.Stop()
	p.Log.Debug("msg", "elastic stopped")
	return nil
}

// Publish add messages to bulk in Elastic
func (p *Publisher) Publish(bulk []*elkstreams.LogMessage) error {
	for i := range bulk {
		metas = p.metaUpdater.MatchEvent(bulk[i])
		if len(metas) > 0 {
			alert := Alert{
				Metas:     metas,
				Message:   m,
				Timestamp: time.Now(),
			}
			p.s.Add(alert)
		}
	 }
	return nil
}

