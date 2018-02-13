package main

import (
	"os"
	"reflect"

	"github.com/alexakulov/candy-elk"
)

func mustStart(service elkstreams.Service) {
	name := reflect.TypeOf(service)

	log.Debug("msg", "starting service", "name", name)
	if err := service.Start(); err != nil {
		log.Error("msg", "error starting service", "name", name, "error", err)
		os.Exit(1)
	}
	log.Debug("msg", "started service", "name", name)
}

func mustStop(service elkstreams.Service) {
	name := reflect.TypeOf(service)

	log.Debug("msg", "stopping service", "name", name)
	if err := service.Stop(); err != nil {
		log.Error("msg", "error stopping service", "name", name, "error", err)
		os.Exit(1)
	}
	log.Debug("msg", "stopped service", "name", name)
}
