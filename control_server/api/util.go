// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
	"github.com/baas-project/baas/pkg/model/images"
	"net/http"
	"os"

	"github.com/baas-project/baas/pkg/database"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// GetTag helper function which gets the name out of the request
// Returns the name in the URI
func GetTag(tag string, w http.ResponseWriter, r *http.Request) (string, error) {
	vars := mux.Vars(r)
	res, ok := vars[tag]

	if !ok || res == "" {
		http.Error(w, tag+" not found", http.StatusBadRequest)
		log.Errorf(tag + " not provided")
		return "", errors.New(tag + " not found")
	}

	return res, nil
}

// GetName is a shorthand for GetTag(name, r, w)
func GetName(w http.ResponseWriter, r *http.Request) (string, error) {
	return GetTag("name", w, r)
}

// CreateNewVersion creates a new version for a specified image
func CreateNewVersion(uuid string, store database.Store) (images.Version, error) {
	// First fetch the image from the database, so we can get the id using the unique id.
	// Do not ask me why this is needed, revamp of the database might be needed.
	image, err := store.GetImageByUUID(images.ImageUUID(uuid))

	if err != nil {
		return images.Version{}, errors.New("cannot fetch image from database")
	}

	version := images.Version{Version: image.Versions[len(image.Versions)-1].Version + 1,
		ImageModelUUID: image.UUID}

	store.CreateNewImageVersion(version)
	return version, nil
}

// ErrorWrite writes the same error message on the HTTP stream and log
func ErrorWrite(w http.ResponseWriter, err error, msg string) error {
	if err != nil {
		http.Error(w, msg, http.StatusInternalServerError)
		log.Errorf("Invalid machine: %v", err)
	}

	return err
}

// OpenImageFile opens an image file so it can be read by the system
func OpenImageFile(uniqueID string, version string, api *API) (*os.File, error) {
	f, err := os.Open(fmt.Sprintf(api.diskpath+images.FilePathFmt, uniqueID, version))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// RegisterImagePackageHandlers runs the handlers which install the routes for the modules
func (api_ *API) RegisterImagePackageHandlers() {
	api_.RegisterImageDockerHandlers()
	api_.RegisterImageHandlers()
	api_.RegisterImageSetupHandlers()
}
