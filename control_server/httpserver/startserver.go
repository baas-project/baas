// Package httpserver provides functions for handling http requests on the control server.
// This is used to respond to requests from pixiecore, to serve files (kernel, initramfs, disk images)
// and to communicate with machines running the management os.
package httpserver

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/baas-project/baas/pkg/httplog"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/baas-project/baas/control_server/machines"
)

// StartServer defines all routes and then starts listening for HTTP requests.
// TODO: Config struct
func StartServer(machineStore machines.MachineStore, staticDir string, diskpath string, address string, port int) {
	r := mux.NewRouter()

	r.StrictSlash(true)
	r.Use(logging)

	r.HandleFunc("/log", httplog.CreateLogHandler(log.StandardLogger()))

	// Serve static files (kernel, initramfs, disk images)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Routes for communicating with the management os
	routes := NewRoutes(machineStore, diskpath)
	mmosr := r.PathPrefix("/mmos").Subrouter()

	// Serve boot configurations to pixiecore (this url is hardcoded in pixiecore)
	r.HandleFunc("/v1/boot/{mac}", routes.ServeBootConfigurations)

	mmosr.HandleFunc("/inform", routes.BootInform).Methods(http.MethodPost)
	mmosr.HandleFunc("/disk/{uuid}", routes.UploadDiskImage).Methods(http.MethodPost)
	mmosr.HandleFunc("/disk/{uuid}", routes.DownloadDiskImage).Methods(http.MethodGet)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%s", address, strconv.Itoa(port)),
	}

	log.Fatal(srv.ListenAndServe())
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We don't want to log the fact that we are logging.
		if r.URL.Path != "/log" {
			log.Debugf("%s request on %s", r.Method, r.URL)
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
