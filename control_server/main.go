package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)


var (
	port = flag.Int("port", 4242, "Port to listen on")
)

func main() {
	flag.Parse()

	http.HandleFunc("/v1/boot/", api)
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}

func api(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving boot config for %s", filepath.Base(r.URL.Path))
	resp := struct {
		K string   `json:"kernel"`
		I []string `json:"initrd"`
	}{
		K: "http://localhost:8000/vmlinuz64",
		I: []string{
			"http://localhost:8000/rootfs.gz",
			"http://localhost:8000/modules64.gz",
		},
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		panic(err)
	}
}
