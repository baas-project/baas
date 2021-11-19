// Package api provides functions for handling http requests on the control server.
// This is used to respond to requests from pixiecore, to serve files (kernel, initramfs, disk images)
// and to communicate with machines running the management os.
package api

import (
	"fmt"
	"math/rand"

	"github.com/baas-project/baas/pkg/database"
	"github.com/gorilla/sessions"
)

// API is a struct on which functions are defined that respond to requests
// from either the management OS, or the end user (through some kind of interface).
// This struct holds state necessary for the request handlers.
type API struct {
	store    database.Store
	diskpath string
	session  *sessions.CookieStore
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
