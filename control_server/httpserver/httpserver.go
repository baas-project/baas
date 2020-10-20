// Package httpserver provides functions for handling http requests on the control server.
// This is used to respond to requests from pixiecore, to serve files (kernel, initramfs, disk images)
// and to communicate with machines running the management os.
package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"baas/control_server/machines"
)

// StartServer defines all routes and then starts listening for HTTP requests.
func StartServer(machineStore machines.MachineStore, staticDir string, address string, port int) {
	r := mux.NewRouter()

	// Serve boot configurations to pixiecore
	bch := BootConfigHandler{machineStore}
	r.HandleFunc("/v1/boot/{mac}", bch.ServeBootConfigurations)

	// Serve static files (kernel, initramfs, disk images)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// communicate with the management OS
	mmosh := ManagementOsHandler{machineStore}

	// More functions can be added to mmosh and more handlefuncs can be added here
	r.HandleFunc("/mmos/test", mmosh.RespondToTestPostRequest).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%s", address, strconv.Itoa(port)),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
