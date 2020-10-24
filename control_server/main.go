package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"baas/pkg/api"

	log "github.com/sirupsen/logrus"

	"baas/control_server/httpserver"
	"baas/control_server/machines"
	"baas/control_server/pixieserver"
)

var (
	static = flag.String("static", "control_server/static", "Static file dir to server under /static/")
)

func init() {
	lvlstring := os.Getenv("LOG_LEVEL")

	loglevel, err := log.ParseLevel(lvlstring)
	if err != nil {
		loglevel = log.InfoLevel
	}

	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(loglevel)

	// log error after the logger is initialised
	if err != nil && lvlstring != ""{
		log.Errorf("loglevel string %s could not be parsed, defaulting to Info: %v", lvlstring, err)
	}
}

func main() {
	flag.Parse()

	log.Info("Starting BAAS control server")

	machineStore := machines.InMemoryStore()
	err := machineStore.UpdateMachine(machines.Machine{

		Architecture: machines.X86_64,
	})
	if err != nil {
		log.Fatal(err)
	}

	go pixieserver.StartPixiecore(fmt.Sprintf("http://localhost:%s", strconv.Itoa(api.Port)))
	httpserver.StartServer(machineStore, *static, "0.0.0.0", api.Port)
}
