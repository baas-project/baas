package main

import (
	"control_server/machines"
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)


var (
	port = flag.Int("port", 4242, "Port to listen on")
	static = flag.String("static", "./static/", "Static file dir to server under /static/")
)

func main() {
	flag.Parse()

	machineStore := machines.InMemoryStore()

	go machines.WatchArchitecturesDhcp(&machineStore)

	r := mux.NewRouter()
	r.Handle("/v1/boot/{mac}", pixieCoreHandler {
		&machineStore,
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(*static))))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":"+strconv.Itoa(*port),
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

	log.Printf("Serving boot config for %v", mac)

	m, err := p.machineStore.GetMachine(mac)
	if err != nil {
		log.Printf("An error occured: %v", err)
		return
	}

	type pixieCoreResponse struct {
		K string	`json:"kernel"`
		I []string	`json:"initrd"`
	}

	var resp pixieCoreResponse

	switch m.Architecture {
	case machines.X86_64:
		resp = pixieCoreResponse {
			K: "http://localhost:4242/static/vmlinuz/",
			I: []string{
				"http://localhost:4242/static/initramfs",
			},
		}
	case machines.Arm64:
		fallthrough
	case machines.Unknown:
		writer.WriteHeader(404)
		return
	}


	if err := json.NewEncoder(writer).Encode(&resp); err != nil {
		panic(err)
	}
}

