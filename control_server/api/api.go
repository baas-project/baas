// Package api provides functions for handling http requests on the control server.
// This is used to respond to requests from pixiecore, to serve files (kernel, initramfs, disk images)
// and to communicate with machines running the management os.
package api

import (
	"github.com/baas-project/baas/pkg/database"
)

// Api is a struct on which functions are defined that respond to requests
// from either the management OS, or the end user (through some kind of interface).
//This struct holds state necessary for the request handlers.
type Api struct {
	store    database.Store
	diskpath string
}

// NewApi creates a new Api struct.
func NewApi(store database.Store, diskpath string) *Api {
	return &Api{
		store:    store,
		diskpath: diskpath,
	}
}
