// Package api provides functions for handling http requests on the control server.
// This is used to respond to requests from pixiecore, to serve files (kernel, initramfs, disk images)
// and to communicate with machines running the management os.
package api

import (
	"github.com/baas-project/baas/pkg/database"
)

// API is a struct on which functions are defined that respond to requests
// from either the management OS, or the end user (through some kind of interface).
// This struct holds state necessary for the request handlers.
type API struct {
	store    database.Store
	diskpath string
}

// NewAPI creates a new API struct.
func NewAPI(store database.Store, diskpath string) *API {
	return &API{
		store:    store,
		diskpath: diskpath,
	}
}
