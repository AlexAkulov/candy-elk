package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/alexakulov/candy-elk/amqp"
	"github.com/alexakulov/candy-elk/http"
	"github.com/alexakulov/candy-elk/metrics"
	"github.com/alexakulov/candy-elk/profiler"
)

type config struct {
	Logfile   string               `yaml:"logfile"`
	LogLevel  string               `yaml:"loglevel"`
	AMQP      amqp.ConfigPublisher `yaml:"amqp"`
	Metrics   metrics.Config       `yaml:"metrics"`
	HTTP      http.Config          `yaml:"http"`
	Profiling profiler.Config      `yaml:"pprof"`
}

func defaultConfig() *config {
	return &config{
		LogLevel: "debug",
		Logfile:  "stdout",
		AMQP: amqp.ConfigPublisher{
			URL:               "amqp://guest:guest@localhost:5672",
			PublishTimeout:    5,
			ReconnectInterval: 2,
		},
		Metrics: metrics.Config{
			Enabled:                  true,
			GraphiteConnectionString: "",
			GraphitePrefix:           "DevOps",
		},
		HTTP: http.Config{
			Address:     ":8080",
			Timeout:     30,
			IdleTimeout: 300,
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
		return nil, fmt.Errorf("Can't read file [%s] [%s]", configLocation, err)
	}
	err = yaml.Unmarshal([]byte(configYaml), &config)
	if err != nil {
		return nil, fmt.Errorf("Can't parse config file [%s] [%s]", configLocation, err)
	}
	return config, nil
}

func printDefaultConfig() {
	c := defaultConfig()
	d, _ := yaml.Marshal(&c)
	fmt.Print(string(d))
}
