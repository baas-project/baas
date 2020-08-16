package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)


var (
	port = flag.Int("port", 4242, "Port to listen on")
	static = flag.String("static", "./static/", "Static file dir to server under /static/")
)

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/v1/boot/{mac}", api)
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

func api(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving boot config for %s", filepath.Base(r.URL.Path))
	resp := struct {
		K string	`json:"kernel"`
		I []string	`json:"initrd"`
		//C string	`json:"cmdline"`
	}{
		K: "http://localhost:4242/static/vmlinuz/",
		I: []string{
			"http://localhost:4242/static/initramfs",
		},
		//C: "squashfs,sd-mod,usb-storage quiet nomodeset",
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		panic(err)
	}
}
