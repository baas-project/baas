// Package httpserver provides functions for handling http requests on the control server.
// This is used to respond to requests from pixiecore, to serve files (kernel, initramfs, disk images)
// and to communicate with machines running the management os.
package httpserver

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"baas/control_server/machines"
)

// StartServer defines all routes and then starts listening for HTTP requests.
// TODO: Config struct
func StartServer(machineStore machines.MachineStore, staticDir string, diskpath string, address string, port int) {
	r := mux.NewRouter()

	r.StrictSlash(true)
	r.Use(logging)

	// Serve boot configurations to pixiecore
	bch := BootConfigHandler{machineStore}
	r.HandleFunc("/v1/boot/{mac}", bch.ServeBootConfigurations)

	// Serve static files (kernel, initramfs, disk images)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Routes for communicating with the management os
	mmosh := ManagementOsHandler{machineStore, diskpath}
	mmosr := r.PathPrefix("/mmos").Subrouter()

	mmosr.HandleFunc("/inform", mmosh.BootInform).Methods(http.MethodPost)
	mmosr.HandleFunc("/disk/{uuid}", mmosh.UploadDiskImage).Methods(http.MethodPost)
	mmosr.HandleFunc("/disk/{uuid}", mmosh.DownloadDiskImage).Methods(http.MethodGet)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%s", address, strconv.Itoa(port)),
	}

	log.Fatal(srv.ListenAndServe())
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s request on %s", r.Method, r.URL)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
