package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"

	"github.com/alexakulov/candy-elk/amqp"
	"github.com/alexakulov/candy-elk/http"
	"github.com/alexakulov/candy-elk/logger"
	"github.com/alexakulov/candy-elk/metrics"
	"github.com/alexakulov/candy-elk/profiler"
)

var (
	version   = "unknown"
	goVersion = "unknown"
	log       *logger.Logger
)

func main() {
	var opts struct {
		ConfigLocation     string `short:"c" long:"config" description:"path to config yaml file" default:"elkgate.yml"`
		PrintDefaultConfig bool   `long:"print-default-config" description:"Print default config and exit"`
		Version            bool   `short:"v" long:"version" description:"print version and exit"`
	}
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	if opts.Version {
		fmt.Println("elkgate:", version)
		fmt.Println("Golang:", goVersion)
		os.Exit(0)
	}

	if opts.PrintDefaultConfig {
		printDefaultConfig()
		os.Exit(0)
	}

	config, err := loadConfig(opts.ConfigLocation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open config %s: %s\n", opts.ConfigLocation, err)
		os.Exit(1)
	}

	writer := os.Stdout
	if config.Logfile != "" && config.Logfile != "stdout" {
		var err error
		if writer, err = os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			fmt.Fprintf(os.Stderr, "failed to open logfile %s: %s\n", config.Logfile, err)
			os.Exit(1)
		}
		defer writer.Close()
	}

	log = logger.New(config.LogLevel, writer)

	metrics := &metrics.MetricStorage{
		Config: config.Metrics,
		Log:    logger.With(log, "component", "metrics"),
	}

	publisher := &amqp.Publisher{
		Config:        config.AMQP,
		MetricStorage: metrics,
		Log:           logger.With(log, "component", "amqp"),
	}

	handler := &http.Server{
		Config:        config.HTTP,
		MetricStorage: metrics,
		Publisher:     publisher,
		Log:           logger.With(log, "component", "http"),
	}

	pprof := &profiler.Profiler{
		Config: config.Profiling,
		Log:    logger.With(log, "component", "profiler"),
	}

	mustStart(metrics)
	mustStart(publisher)
	mustStart(handler)
	mustStart(pprof)

	log.Info("msg", "started", "pid", os.Getpid(), "version", version)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		s := <-signalChannel
		log.Info("msg", "received signal", "signal", s)
		if s != syscall.SIGHUP {
			break
		}
		config, err := loadConfig(opts.ConfigLocation)
		if err != nil {
			log.Error("msg", "reload failed", "err", err)
			continue
		}
		handler.Config = config.HTTP
		log.SetLevel(config.LogLevel)
		metrics.Log.SetLevel(config.LogLevel)
		handler.Log.SetLevel(config.LogLevel)
		publisher.Log.SetLevel(config.LogLevel)
		log.Info("msg", "api-keys and loglevel was be reloaded")
	}

	mustStop(handler)
	mustStop(publisher)
	mustStop(metrics)
	mustStart(pprof)
	log.Info("msg", "stopped", "pid", os.Getpid(), "version", version)
}
