package meta

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/quipo/statsd"
	"gopkg.in/olivere/elastic.v3"
	"gopkg.in/tomb.v2"

	"github.com/AlexAkulov/candy-elk"
)

// CompareFunc is a function to compare field value and alter limit
type CompareFunc func(float64, float64) bool

// CompareFunctions is map for known compare functions by name
var CompareFunctions = map[string]CompareFunc{
	"less": func(v1, v2 float64) bool {
		return v1 < v2
	},
	"less_or_equal": func(v1, v2 float64) bool {
		return v1 <= v2
	},
	"is_equal": func(v1, v2 float64) bool {
		return v1 == v2
	},
	"greater_or_equal": func(v1, v2 float64) bool {
		return v1 >= v2
	},
	"greater": func(v1, v2 float64) bool {
		return v1 > v2
	},
}

// Updater read settigs from elasticsearch every minute
type Matcher struct {
	Log      elkstreams.Logger
	ESClient *elastic.Client

	alertMetas map[string][]*AlertMeta
	eventMetas map[string]map[string]*EventMeta

	tomb *tomb.Tomb

	stats statsd.Statsd
}

// Start Updater
func (m *Updater) Start() error {
	if err := m.readAlertMetas(); err != nil {
		return err
	}
	if err := m.readEventMetas(); err != nil {
		return err
	}

	m.tomb.Go(func() error {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-m.tomb.Dying(): // Exit
				return nil
			case <-ticker.C:
				if err := m.readAlertMetas(); err != nil {
					m.Log.Error("err", err)
				}
				if err := m.readEventMetas(); err != nil {
					m.Log.Error("err", err)
				}
			}
		}
	})
	return nil
}

// Stop Updater
func (m *Updater) Stop() error {
	m.tomb.Kill(nil)
	return m.tomb.Wait()
}

// Filter describes field and matching condition (regexp or limit)
type Filter struct {
	CompareFunctionName string `json:"compare_function"`
	Field               string `json:"field"`
	Limit               string `json:"limit"`
	Regexp              string `json:"regexp"`
	compiledRegexp      *regexp.Regexp
	compareFunction     CompareFunc
	LimitValue          float64
}

// AlertMeta config for alerting criteria
type AlertMeta struct {
	ID             string
	Filters        []*Filter `json:"filters"`
	IndexTemplate  string    `json:"index_template"`
	Name           string    `json:"name"`
	Recipient      string    `json:"recipient"`
	SenderType     string    `json:"sender_type"`
	ApplyToIgnored bool      `json:"ignored"`
}

// EventMeta config for display logging event
type EventMeta struct {
	ExcTraceHash  string `json:"exc_trace_hash"`
	Ignored       bool   `json:"ignored"`
	IndexTemplate string `json:"index_template"`
	Metric        string `json:"metric"`
}

// Alert contains matched message and AlertMetas
type Alert struct {
	Metas     []*AlertMeta
	Message   *elkstreams.LogMessage
	Timestamp time.Time
}

// Parse compile regexp and init limit value
func (alertMeta *AlertMeta) parse() error {
	var err error
	for _, filter := range alertMeta.Filters {
		if len(filter.Regexp) > 0 {
			if filter.compiledRegexp, err = regexp.Compile(filter.Regexp); err != nil {
				return fmt.Errorf("can not compile regexp %s: %s", filter.Regexp, err)
			}
		}
		if len(filter.CompareFunctionName) > 0 {
			cf, ok := CompareFunctions[filter.CompareFunctionName]
			if !ok {
				return fmt.Errorf("compare function %s is not defined", filter.CompareFunctionName)
			}
			filter.compareFunction = cf
			filter.LimitValue, err = getFloatValue(filter.Field, filter.Limit)
			if err != nil {
				return fmt.Errorf("can not get limit value %s: %s", filter.Limit, err)
			}
		}
	}
	return nil
}

// ReadAlertMetas scroll all AlertMeta documents in esd index and store it in Metas map
func (m *Updater) readAlertMetas() error {
	newMetas := make(map[string][]*AlertMeta)
	scrollID := ""
	for {
		searchResult, err := m.ESClient.Scroll("esd").Type("AlertMeta").Size(100).ScrollId(scrollID).Do()
		if err == elastic.EOS {
			break
		}
		if err != nil {
			return fmt.Errorf("Can not scroll alert meta: %s", err)
		}
		for _, hit := range searchResult.Hits.Hits {
			item := &AlertMeta{}
			err := json.Unmarshal(*hit.Source, &item)
			if err != nil {
				m.Log.Error("msg", "Can't parse alert meta", "meta", *hit.Source, "err", err)
				continue
			}
			if err := item.parse(); err != nil {
				m.Log.Error("msg", "Can't parse alert meta", "meta", *hit.Source, "err", err)
				continue
			}
			item.ID = hit.Id
			list, ok := newMetas[item.IndexTemplate]
			if !ok {
				list = make([]*AlertMeta, 0)
			}
			list = append(list, item)
			newMetas[item.IndexTemplate] = list
		}
		scrollID = searchResult.ScrollId
	}
	m.alertMetas = newMetas
	return nil
}

// ReadEventMetas scroll all AlertMeta documents in esd index and store it in Metas map
func (m *Updater) readEventMetas() error {
	newMetas := make(map[string]map[string]*EventMeta)
	scrollID := ""

	for {
		searchResult, err := m.ESClient.Scroll("esd").Type("EventMeta").Size(100).ScrollId(scrollID).Do()
		if err == elastic.EOS {
			break
		}
		if err != nil {
			return fmt.Errorf("Can not scroll alert meta: %s", err)
		}
		for _, hit := range searchResult.Hits.Hits {
			item := &EventMeta{}
			err := json.Unmarshal(*hit.Source, &item)
			if err != nil {
				m.Log.Error("msg", "Can not parse event meta", "meta", *hit.Source, "err", err)
				continue
			}
			metas, ok := newMetas[item.IndexTemplate]
			if !ok {
				metas = make(map[string]*EventMeta)
			}
			metas[item.ExcTraceHash] = item
			newMetas[item.IndexTemplate] = metas
		}
		scrollID = searchResult.ScrollId
	}
	m.eventMetas = newMetas
	return nil
}
