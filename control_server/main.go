package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"baas/pkg/api"

	"baas/control_server/httpserver"
	"baas/control_server/machines"
	"baas/control_server/pixieserver"
)

var (
	static = flag.String("static", "control_server/static", "Static file dir to server under /static/")
)

func main() {
	flag.Parse()

	machineStore := machines.InMemoryStore()
	err := machineStore.UpdateMachine(machines.Machine{
		MacAddress:   "06:99:2b:9b:3a:22",
		Architecture: machines.X86_64,
	})
	if err != nil {
		log.Fatal(err)
	}

	go pixieserver.StartPixiecore(fmt.Sprintf("http://localhost:%s", strconv.Itoa(api.Port)))
	httpserver.StartServer(machineStore, *static, "0.0.0.0", api.Port)
}
