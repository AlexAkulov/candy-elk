package es6x

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AlexAkulov/candy-elk"
	"github.com/AlexAkulov/candy-elk/elastic"
	"github.com/facebookgo/muster"
	es6x "gopkg.in/olivere/elastic.v6"
)

// Publisher is an implementation of elkstreams.Publisher interface for publishing to Elasticsearch
type Publisher struct {
	Config elastic.Config
	Log    elkstreams.Logger
	es     *es6x.Client
	muster *muster.Client
}

// Start initializes Elasticsearch connection
func (p *Publisher) Start() error {
	var err error
	p.es, err = es6x.NewClient(
		es6x.SetURL(p.Config.ElasticUrls...),
		es6x.SetErrorLog(p.Log),
		es6x.SetHealthcheck(true),
		es6x.SetHealthcheckTimeoutStartup(time.Second),
	)
	if err != nil {
		return err
	}
	p.Log.Debug("msg", "elasticsearch connected")

	p.muster = &muster.Client{
		MaxBatchSize:         p.Config.BulkSize,
		MaxConcurrentBatches: p.Config.ConcurentWrites,
		BatchTimeout:         time.Duration(p.Config.BulkRefreshInterval) * time.Second,
		BatchMaker:           p.batchMaker,
	}
	err = p.muster.Start()

	return err
}

func (p *Publisher) batchMaker() muster.Batch {
	return &bulk{
		Publisher: p,
	}
}

// Stop flushes and stops publishing
func (p *Publisher) Stop() error {
	p.Log.Debug("msg", "stop muster")
	err := p.muster.Stop()
	p.Log.Debug("msg", "muster stopped", "err", err)
	p.Log.Debug("msg", "stop elastic")
	p.es.Stop()
	p.Log.Debug("msg", "elastic stopped")
	return nil
}

type bulk struct {
	Publisher *Publisher
	Items     []*elkstreams.LogMessage
}

// Publish add messages to bulk in Elastic
func (p *Publisher) Publish(bulk []*elkstreams.LogMessage) error {
	for i := range bulk {
		p.muster.Work <- bulk[i]
	}
	return nil
}

func (b *bulk) Add(item interface{}) {
	b.Items = append(b.Items, item.(*elkstreams.LogMessage))
}

func (b *bulk) Fire(notifier muster.Notifier) {
	defer notifier.Done()
	bulkRequest := b.Publisher.es.Bulk()
	for i, _ := range b.Items {
		bulkRequest.Add(
			es6x.NewBulkIndexRequest().Index(b.Items[i].IndexName).Type(b.Items[i].IndexType).Doc(string(b.Items[i].Body)),
		)
	}
	var (
		res *es6x.BulkResponse
		err error
	)
	for {
		ctx := context.Background()
		res, err = bulkRequest.Do(ctx)
		if err != nil {
			b.Publisher.Log.Error("msg", "failed write bulk to es", "err", err, "count", len(b.Items))
			// Return messages to RabbitMQ
			// for i, _ := range b.Items {
			// 	// b.Items[i].Nack()
			// }
			time.Sleep(time.Second * 10)
			continue
		}
		break
	}
	failed := res.Failed()
	b.Publisher.processLostMessages(failed)
	for i, _ := range b.Items {
		if b.Items[i].Ack != nil {
			b.Items[i].Ack.Done()
		}
	}
	b.Publisher.Log.Debug("msg", "bulk writed", "size", len(b.Items), "took", res.Took, "failed", len(failed))
}

func (p *Publisher) processLostMessages(failed []*es6x.BulkResponseItem) {
	for i, res := range failed {
		if i > 5 {
			p.Log.Debug("msg", "Others response error details are omitted")
			return
		}

		response, err := json.Marshal(res.Error)
		if err != nil {
			p.Log.Warn("msg", "Can not decode response error", "err", err)
			continue
		}
		p.Log.Warn("msg", "fail details", "err", string(response), "index", res.Index, "type", res.Type, "status", res.Status)
	}
}
