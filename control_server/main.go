package main

import (
	"baas/control_server/httpserver"
	"baas/control_server/machines"
	"baas/control_server/pixieserver"
	"flag"
	"fmt"
	"strconv"
)

var (
	port   = flag.Int("port", 4848, "Port to listen on")
	static = flag.String("static", "control_server/static", "Static file dir to server under /static/")
)

func main() {
	flag.Parse()

	machineStore := machines.InMemoryStore()

	go machines.WatchArchitecturesDhcp(machineStore)
	go pixieserver.StartPixiecore(fmt.Sprintf("http://localhost:%s", strconv.Itoa(*port)))
	httpserver.StartServer(machineStore, *static, "0.0.0.0", *port)
}


