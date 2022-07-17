// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package api provides functions for handling http requests on the control server.
// This is used to respond to request from pixiecore, to serve files (kernel, initramfs, disk images)
// and to communicate with machines running the management os.
package api

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/baas-project/baas/pkg/database"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

/*func NewAPI(store database.Store, diskpath path) *API {
	fmt.Println(a ...interface{})
}
*/

// API is a struct on which functions are defined that respond to request
// from either the management OS, or the end user (through some kind of interface).
// This struct holds state necessary for the request handlers.
type API struct {
	store    database.Store
	diskpath string
	session  *sessions.CookieStore
	Routes   []Route
}

// NewAPI creates a new API struct.
func NewAPI(store database.Store, diskpath string) *API {
	session := sessions.NewCookieStore([]byte(fmt.Sprint(rand.Intn(2_000_000))))
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
	}

	return &API{
		store:    store,
		diskpath: diskpath,
		session:  session,
	}
}

// CheckRole verifies whether a user is allowed to use this particular route or not.
// lint:
func (api_ *API) CheckRole(route Route, next http.HandlerFunc) http.HandlerFunc { // nolint
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: dear god, security here needs to be done beter.
		if r.Header.Get("type") == "system" {
			next.ServeHTTP(w, r)
			return
		} else if !route.UserAllowed {
			http.Error(w, "Users are not allowed to access this endpoint", http.StatusBadRequest)
			return
		}
		session, _ := api_.session.Get(r, "session-name")
		role, ok := session.Values["Role"].(string)
		if !ok {
			http.Error(w, "User's role not found", http.StatusNotFound)
			return
		}

		found := false
		for _, b := range route.Permissions {
			if role == string(b) {
				found = true
			}
		}

		// If this resource is from the same user they might be able to access it
		if !found && !checkSameUser(route, w, r, api_) {
			http.Error(w, fmt.Sprintf("User role '%s' not permitted to access this resource.", role), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// checkSameUser checks if this resource is owned by the same issue. It only works for when the user is in the URI, database needs to be checked manually
func checkSameUser(route Route, _ http.ResponseWriter, r *http.Request, api *API) bool {
	// Check if the same user exception applies this method
	if !route.UserAllowed {
		return false
	}

	session, _ := api.session.Get(r, "session-name")

	username, ok := session.Values["Username"].(string)

	if !ok {
		return false
	}

	// Get the username from the URI
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok || name == "" {
		return false
	}

	return username == name
}
