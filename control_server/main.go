package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/baas-project/baas/pkg/database/sqlite"

	log "github.com/sirupsen/logrus"

	"github.com/baas-project/baas/control_server/api"
	"github.com/baas-project/baas/control_server/pixieserver"
	api_pkg "github.com/baas-project/baas/pkg/api"
)

var (
	static   = flag.String("static", "control_server/static", "Static file dir to server under /static/.")
	diskpath = flag.String("disks", "control_server/disks", "Location to store disk images.")
)

func init() {
	levelString := os.Getenv("LOG_LEVEL")

	loglevel, err := log.ParseLevel(levelString)
	if err != nil {
		loglevel = log.DebugLevel
	}

	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(loglevel)

	// log error after the logger is initialised
	if err != nil && levelString != "" {
		log.Errorf("loglevel string %s could not be parsed, defaulting to Info: %v", levelString, err)
	}
}

func main() {
	flag.Parse()

	log.Info("Starting BAAS control server")

	store, err := sqlite.NewSqliteStore("store.db")
	if err != nil {
		log.Fatal(err)
	}

	//mac := "52:54:00:d9:71:93"

	// mac
	if err != nil {
		log.Fatal(err)
	}

	go pixieserver.StartPixiecore(fmt.Sprintf("http://localhost:%s", strconv.Itoa(api_pkg.Port)))
	api.StartServer(store, *static, *diskpath, "0.0.0.0", api_pkg.Port)
}
