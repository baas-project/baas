package main

import (
	"baas/control_server/machines"
	"baas/control_server/pixieserver"
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	port   = flag.Int("port", 4848, "Port to listen on")
	static = flag.String("static", "control_server/static", "Static file dir to server under /static/")
)

func main() {
	flag.Parse()

	machineStore := machines.InMemoryStore()

	log.Printf("Started Architecture Watcher")
	go machines.WatchArchitecturesDhcp(machineStore)

	go pixieserver.StartPixiecore("http://localhost:4848")

	r := mux.NewRouter()
	r.Handle("/v1/boot/{mac}", pixieCoreHandler{
		machineStore,
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(*static))))

	srv := &http.Server{
		Handler: r,
		Addr:    ":" + strconv.Itoa(*port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

type pixieCoreHandler struct {
	machineStore machines.MachineStore
}

func (p pixieCoreHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	mac := vars["mac"]
	addr, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		log.Printf("An error occured: %v", err)
		return
	}
	// Remove port

	log.Printf("Serving boot config for %v at ip: %v", mac, addr)

	m, err := p.machineStore.GetMachine(mac)
	if err != nil {
		log.Printf("An error occured: %v", err)
		return
	}

	type pixieCoreResponse struct {
		K string   `json:"kernel"`
		I []string `json:"initrd"`
	}

	var resp pixieCoreResponse

	switch m.Architecture {
	case machines.X86_64:
		resp = pixieCoreResponse{
			K: "http://localhost:4848/static/vmlinuz",
			I: []string{
				"http://localhost:4848/static/initramfs",
			},
		}
	case machines.Arm64:
		fallthrough
	case machines.Unknown:
		writer.WriteHeader(404)
		return
	}

	log.Printf("Sending boot config %v", resp)

	if err := json.NewEncoder(writer).Encode(&resp); err != nil {
		panic(err)
	}

	//go protoClient(context.Background(), addr)
}
