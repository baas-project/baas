// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"net/http"

	"github.com/baas-project/baas/pkg/database"
	"github.com/baas-project/baas/pkg/httplog"
	"github.com/baas-project/baas/pkg/model"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

// Route stores the data about each of the routes and any related metadata.
type Route struct {
	URI         string
	Permissions []model.UserRole
	UserAllowed bool
	Handler     func(w http.ResponseWriter, r *http.Request)
	Method      string

	// Cute little feature
	Description string
}

func getHandler(machineStore database.Store, staticDir string, diskpath string) http.Handler {
	// API for communicating with the management os
	api := NewAPI(machineStore, diskpath)

	r := mux.NewRouter()

	r.StrictSlash(true)
	r.Use(logging)

	// Applications (in particular, the management OS) can send logs here to be logged on the control server.
	r.HandleFunc("/log", httplog.CreateLogHandler(log.StandardLogger()))

	// TODO: we may want to split this up, especially the disk images part
	// TODO: isn't this already the case?
	// Serve static files (kernel, initramfs, disk images)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	api.RegisterMachineHandlers()
	api.RegisterUserHandlers()
	api.RegisterImagePackageHandlers()

	for _, route := range api.Routes {
		r.HandleFunc(route.URI, api.CheckRole(route, route.Handler)).Methods(route.Method)
	}

	// OAuth login handlers, we deal with these separately since they should always be available.
	r.HandleFunc("/user/login/github", api.LoginGithub).Methods(http.MethodGet)
	r.HandleFunc("/user/login/github/callback", api.LoginGithubCallback).Methods(http.MethodGet)

	// Serve boot configurations to pixiecore (this url is hardcoded in pixiecore)
	r.HandleFunc("/v1/boot/{mac}", api.ServeBootConfigurations)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:9090"},
		AllowedHeaders:   []string{"Authorization", "Set-Cookie"},
		AllowCredentials: true,
		Debug:            true,
	})

	return c.Handler(r)
}

// StartServer defines all routes and then starts listening for HTTP requests.
func StartServer(machineStore database.Store, staticDir string, diskPath string, address string, port int) {
	srv := http.Server{
		Handler: getHandler(machineStore, staticDir, diskPath),
		Addr:    fmt.Sprintf("%s:%d", address, port),
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
