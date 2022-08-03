// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/baas-project/baas/pkg/model/images"
	"github.com/baas-project/baas/pkg/model/user"

	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func _getImageSetup(w http.ResponseWriter, r *http.Request, api *API) (*images.ImageSetup, error) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to create image setup", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return nil, err
	}

	tagUUID, err := GetTag("setup_uuid", w, r)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("UUID not found in URI: %v", err)
		return nil, err
	}

	setup, err := api.store.GetImageSetup(tagUUID)
	if err != nil {
		http.Error(w, "Failed to find image setup", http.StatusBadRequest)
		log.Errorf("Cannot find image setup: %v", err)
		return nil, err
	}

	if setup.Username != username {
		http.Error(w, "Image not owned by this user", http.StatusUnauthorized)
		log.Errorf("Image not owned by requesting user: %v", err)
		return nil, err
	}

	return &setup, nil
}

// createImageSetup defines an endpoint which creates an ImageSetup in the database
// Example request: POST /user/[name]/image_setup
// Example response: Successfully created image setup.
func (api_ *API) createImageSetup(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to create image setup", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	// Create an ImageSetup and associate it with an user
	imageSetup := images.ImageSetup{}
	err = json.NewDecoder(r.Body).Decode(&imageSetup)
	imageSetup.Username = username
	imageSetup.UUID = images.ImageUUID(uuid.New().String())

	if imageSetup.Name == "" {
		http.Error(w, "Did not set image setup name", http.StatusBadRequest)
		log.Errorf("Did not sent image setup name: %v", err)
		return
	}

	err = api_.store.CreateImageSetup(username, &imageSetup)
	if err != nil {
		http.Error(w, "Failed to create image setup", http.StatusBadRequest)
		log.Errorf("Error creating database entry: %v", err)
		return
	}

	imageSetup.Images = []images.ImageFrozen{}

	_ = json.NewEncoder(w).Encode(imageSetup)
}

// findImageSetupsByUsername returns all ImageSetups associated with a specific user
// Example request: GET /user/{name}/image_setup
// Example response:
// [{"Name": "Linux Kernel",
//   "Images": [],
//   "User": "ValentijnvdBeek",
//    "UUID": "fcc0ed46-1f55-4366-a1fd-73b61473bbd5"},
//  {"Name": "Linux Kernel 2",
//   "Images": [],
//   "User": "ValentijnvdBeek",
//   "UUID": "2b59ff94-7fb6-4239-b2e6-82f1e30f4355"}]
func (api_ *API) findImageSetupsByUsername(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	// TODO: Better unique error returns
	imageSetup, err := api_.store.FindImageSetupsByUsername(username)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("Find image setups cannot be found: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(imageSetup)
}

// getImageSetup returns the ImageSetup associated with an UUID
// Example request: GET /[name]/image_setup/[uuid]
// Example response:
// { "Name": "Linux Kernel 2",
//   "Images": [{"UUIDImage": "3a760707-c160-40fa-81be-430b75131ddc",
//               "VersionNumber": 3 }],
//   "User": "ValentijnvdBeek",
//   "UUID": "2b59ff94-7fb6-4239-b2e6-82f1e30f4355" }
func (api_ *API) getImageSetup(w http.ResponseWriter, r *http.Request) {
	setup, err := _getImageSetup(w, r, api_)
	if err != nil {
		return
	}

	_ = json.NewEncoder(w).Encode(setup)
}

// addImageSetup deletes an image from a setup
// Request: DELETE /user/{name}/image_setup/{setup_uuid}/image/{image_uuid}
// Response: Successfully deleted image from setup
func (api_ *API) removeImageFromImageSetup(w http.ResponseWriter, r *http.Request) {
	setup, err := _getImageSetup(w, r, api_)
	if err != nil {
		return
	}

	imageMsg := model.ImageSetupMessage{}
	err = json.NewDecoder(r.Body).Decode(&imageMsg)
	if err != nil {
		http.Error(w, "Cannot find image UUID in JSON message.", http.StatusBadRequest)
		log.Errorf("Cannot find image UUID: %v", err)
		return
	}

	image, err := api_.store.GetImageByUUID(images.ImageUUID(imageMsg.UUID))

	if err != nil {
		http.Error(w, "Failed to add image to image setups", http.StatusBadRequest)
		log.Errorf("Cannot find images: %v", err)
		return
	}

	version := images.Version{
		Version:        imageMsg.Version,
		ImageModelUUID: image.UUID,
	}

	err = api_.store.RemoveImageFromImageSetup(setup, image, version, imageMsg.Update)
	if err != nil {
		http.Error(w, "Cannot remove image from setup", http.StatusBadRequest)
		log.Errorf("Cannot delete image from setup: %s, %v", imageMsg.UUID, err)
		return
	}

	http.Error(w, "Successfully deleted image from setup", http.StatusOK)
}

// getImagesFromImageSetup gets all the images from an image setup
// Example request: GET /user/{name}/image_setup/{setup_uuid}/images
// Example response: A list of all the images
func (api_ *API) getImagesFromImageSetup(w http.ResponseWriter, r *http.Request) {
	setup, err := _getImageSetup(w, r, api_)
	if err != nil {
		return
	}

	_ = json.NewEncoder(w).Encode(setup.Images)
}

// getImageSetups fetches all the image setups related to the user
func (api_ *API) getImageSetups(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	imageSetups, err := api_.store.GetImageSetups(username)

	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(imageSetups)
}

// addImageToImageSetup add an ImageModel to the associated ImageSetup
// Example request: POST /[name]/image_setup/[uuid]
// Example body: {"Uuid": "3a760707-c160-40fa-81be-430b75131ddc", "Version": 3}
// Example response:
//  {"Name": "Linux Kernel 2",
//   "Images": [{"UUIDImage":"3a760707-c160-40fa-81be-430b75131ddc","VersionNumber":3}],
//   "User":"ValentijnvdBeek",
//   "UUID":"2b59ff94-7fb6-4239-b2e6-82f1e30f4355"}
func (api_ *API) addImageToImageSetup(w http.ResponseWriter, r *http.Request) {
	imageSetup, err := _getImageSetup(w, r, api_)
	if err != nil {
		return
	}

	imageMsg := model.ImageSetupMessage{}
	err = json.NewDecoder(r.Body).Decode(&imageMsg)
	if err != nil {
		http.Error(w, "Cannot find image UUID in JSON message.", http.StatusBadRequest)
		log.Errorf("Cannot find image UUID: %v", err)
		return
	}

	image, err := api_.store.GetImageByUUID(images.ImageUUID(imageMsg.UUID))

	if err != nil {
		http.Error(w, "Failed to add image to image setups", http.StatusBadRequest)
		log.Errorf("Cannot find images: %v", err)
		return
	}

	var targetVersion images.Version
	for _, version := range image.Versions {
		if version.Version == imageMsg.Version {
			targetVersion = version
		}
	}

	if targetVersion.ImageModelUUID == "" {
		http.Error(w, "Failed to add image to image setups", http.StatusBadRequest)
		log.Errorf("Cannot find images: version %d not found", imageMsg.Version)
		return
	}

	api_.store.AddImageToImageSetup(imageSetup, image, targetVersion, imageMsg.Update)

	_ = json.NewEncoder(w).Encode(imageSetup)
}

// deleteImageSetup deletes the image setup from the database
// Request: DELETE /user/{name}/image_setup/{setup_uuid}
// Response: Successfully deleted image setup
func (api_ *API) deleteImageSetup(w http.ResponseWriter, r *http.Request) {
	setup, err := _getImageSetup(w, r, api_)
	if err != nil {
		return
	}

	err = api_.store.DeleteImageSetup(setup)
	if err != nil {
		http.Error(w, "Failed to delete the image setup.", http.StatusBadRequest)
		log.Errorf("Delete image setup: %v", err)
		return
	}

	http.Error(w, "Successfully deleted image setup", http.StatusOK)
}

// modifyImageSetup modifies the metadata of the image setup, but
// cannot be used to change the images in the setup.
// Request: PUT /user/{name}/image_setup/{setup_uuid}
// Response: the modified image setup
func (api_ *API) modifyImageSetup(w http.ResponseWriter, r *http.Request) {
	oldSetup, err := _getImageSetup(w, r, api_)
	if err != nil {
		return
	}

	newSetup := images.ImageSetup{}
	err = json.NewDecoder(r.Body).Decode(&newSetup)
	if err != nil {
		http.Error(w, "Cannot decode the request body.", http.StatusBadRequest)
		log.Errorf("Modify image setup: %v", err)
		return
	}

	// Allows for easier objects to be sent over and ensures you
	// cannot secretly modify a different setup.
	newSetup.UUID = oldSetup.UUID
	err = api_.store.ModifyImageSetup(&newSetup)
	if err != nil {
		http.Error(w, "Failed to modify the image setup.", http.StatusBadRequest)
		log.Errorf("Modify image setup: %v", err)
		return
	}
	_ = json.NewEncoder(w).Encode(newSetup)
}

// RegisterImageSetupHandlers sets the metadata for each of the routes and registers them to the global handler
func (api_ *API) RegisterImageSetupHandlers() {
	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.createImageSetup,
		Method:      http.MethodPost,
		Description: "Creates an image setup",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setups",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.getImageSetups,
		Method:      http.MethodGet,
		Description: "Gets the image setups associated with a user",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.findImageSetupsByUsername,
		Method:      http.MethodGet,
		Description: "Find image setups by username",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup/{setup_uuid}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.getImageSetup,
		Method:      http.MethodGet,
		Description: "Get a specific image setup",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup/{setup_uuid}/images",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.getImagesFromImageSetup,
		Method:      http.MethodGet,
		Description: "Gets the images associated with an image setup",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup/{setup_uuid}/images",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.addImageToImageSetup,
		Method:      http.MethodPost,
		Description: "Add image to the setup system",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup/{setup_uuid}/images",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.removeImageFromImageSetup,
		Method:      http.MethodDelete,
		Description: "Deletes an image from the setup",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup/{setup_uuid}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.deleteImageSetup,
		Method:      http.MethodDelete,
		Description: "Deletes the image setup",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/user/{name}/image_setup/{setup_uuid}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.modifyImageSetup,
		Method:      http.MethodPut,
		Description: "Modifies the image setup",
	})
}
