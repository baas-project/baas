package httpserver

import (
	"baas/control_server/machines"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

func StartServer(machineStore machines.MachineStore, staticDir string, address string, port int) {
	r := mux.NewRouter()
	r.Handle("/v1/boot/{mac}", BootConfigHandler{machineStore})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	r.Handle("/mmos/test", TestHandler{machineStore}).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%s", address, strconv.Itoa(port)),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}