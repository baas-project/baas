package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/baas-project/baas/pkg/api"

	log "github.com/sirupsen/logrus"

	"github.com/baas-project/baas/control_server/httpserver"
	"github.com/baas-project/baas/control_server/machines"
	"github.com/baas-project/baas/control_server/pixieserver"
)

var (
	static   = flag.String("static", "control_server/static", "Static file dir to server under /static/.")
	diskpath = flag.String("disks", "control_server/disks", "Location to store disk images.")
)

func init() {
	lvlstring := os.Getenv("LOG_LEVEL")

	loglevel, err := log.ParseLevel(lvlstring)
	if err != nil {
		loglevel = log.DebugLevel
	}

	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(loglevel)

	// log error after the logger is initialised
	if err != nil && lvlstring != "" {
		log.Errorf("loglevel string %s could not be parsed, defaulting to Info: %v", lvlstring, err)
	}
}

func main() {
	flag.Parse()

	log.Info("Starting BAAS control server")

	machineStore := machines.InMemoryStore()
	err := machineStore.UpdateMachine(machines.Machine{
		MacAddress:   "52:54:00:ae:a3:b3",
		Architecture: machines.X86_64,
	})
	if err != nil {
		log.Fatal(err)
	}

	go pixieserver.StartPixiecore(fmt.Sprintf("http://localhost:%s", strconv.Itoa(api.Port)))
	httpserver.StartServer(machineStore, *static, *diskpath, "0.0.0.0", api.Port)
}
