package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/AlexAkulov/candy-elk/amqp"
	"github.com/AlexAkulov/candy-elk/elastic"
	"github.com/AlexAkulov/candy-elk/metrics"
	"github.com/AlexAkulov/candy-elk/profiler"
)

type config struct {
	Logfile   string              `yaml:"logfile"`
	LogLevel  string              `yaml:"loglevel"`
	Consumer  amqp.ConfigConsumer `yaml:"amqp"`
	Publisher elastic.Config      `yaml:"elastic"`
	Metrics   metrics.Config      `yaml:"metrics"`
	Profiling profiler.Config     `yaml:"pprof"`
}

func defaultConfig() *config {
	return &config{
		LogLevel: "debug",
		Logfile:  "stdout",
		Consumer: amqp.ConfigConsumer{
			Connections: []amqp.ConnectionConfig{
				amqp.ConnectionConfig{
					URL:               "amqp://guest:guest@localhost:5672",
					Exchange:          "guest",
					RoutingKey:        "guest",
					Queue:             "guest",
					ReconnectInterval: 2,
					WaitAck:           "yes",
				},
			},
		},
		Publisher: elastic.Config{
			ElasticUrls:         []string{"http://localhost:9200"},
			BulkSize:            1000,
			BulkRefreshInterval: 30,
			ConcurentWrites:     10,
		},
		Metrics: metrics.Config{
			Enabled:                  true,
			GraphiteConnectionString: "",
			GraphitePrefix:           "DevOps",
		},
		Profiling: profiler.Config{
			Enabled: "false",
			Listen:  ":6060",
		},
	}
}

func loadConfig(configLocation string) (*config, error) {
	config := defaultConfig()
	configYaml, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configYaml), &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func printDefaultConfig() {
	c := defaultConfig()
	d, _ := yaml.Marshal(&c)
	fmt.Print(string(d))
}
