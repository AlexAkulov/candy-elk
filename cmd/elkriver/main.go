package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"

	"github.com/alexakulov/candy-elk/amqp"
	"github.com/alexakulov/candy-elk/elastic"
	"github.com/alexakulov/candy-elk/logger"
	"github.com/alexakulov/candy-elk/profiler"
)

var (
	version   = "unknown"
	goVersion = "unknown"
)

func main() {
	var opts struct {
		ConfigLocation     string `short:"c" long:"config" description:"path to config yaml file" default:"elkriver.yml"`
		PrintDefaultConfig bool   `long:"print-default-config" description:"Print default config and exit"`
		Version            bool   `short:"v" long:"version" description:"print version and exit"`
	}
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	if opts.Version {
		fmt.Println("elkriver:", version)
		fmt.Println("Golang:", goVersion)
		os.Exit(0)
	}

	if opts.PrintDefaultConfig {
		printDefaultConfig()
		os.Exit(0)
	}

	config, err := loadConfig(opts.ConfigLocation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lvl=fatal msg=\"can't read config\" file=\"%s\" err=\"%s\"\n", opts.ConfigLocation, err)
		os.Exit(1)
	}

	writer := os.Stdout
	if config.Logfile != "" && config.Logfile != "stdout" {
		var err error
		if writer, err = os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			fmt.Fprintf(os.Stderr, "lvl=fatal msg=\"can't open logfile\" file=\"%s\" err=\"%s\"\n", config.Logfile, err)
			os.Exit(1)
		}
		defer writer.Close()
	}

	log := logger.New(config.LogLevel, writer)

	// Start

	p := &profiler.Profiler{
		Config: config.Profiling,
		Log:    logger.With(log, "component", "profiler"),
	}
	p.Start()

	es := &elastic.Publisher{
		Config: config.Publisher,
		Log:    logger.With(log, "component", "publisher"),
	}
	if err := es.Start(); err != nil {
		log.Error("msg", "can't start publisher", "err", err)
		os.Exit(1)
	}

	c := &amqp.Consumer{
		Config:    config.Consumer,
		Log:       logger.With(log, "component", "consumer"),
		Publisher: es,
	}
	if err := c.Start(); err != nil {
		log.Error("msg", "can't start consumer", "err", err)
		os.Exit(1)
	}

	log.Info("msg", "started", "pid", os.Getpid(), "version", version)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	log.Info("msg", "received signal", "signal", <-signalChannel)

	// Stop

	if err := c.Stop(); err != nil {
		log.Error("msg", "stop consumer", "err", err)
	}
	if err := es.Stop(); err != nil {
		log.Error("msg", "stop publusher", "err", err)
	}
	p.Stop()

	log.Info("msg", "stopped", "pid", os.Getpid(), "version", version)
}
