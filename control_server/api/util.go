package api

import (
	"errors"
	"github.com/baas-project/baas/pkg/database"
	"github.com/baas-project/baas/pkg/images"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// imageFileSize is the size of the standard image that is created in MiB.
const imageFileSize = 512 // size in MiB

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

func (api_ *API) RegisterImagePackageHandlers() {
	api_.RegisterImageDockerHandlers()
	api_.RegisterImageHandlers()
	api_.RegisterImageSetupHandlers()
}
