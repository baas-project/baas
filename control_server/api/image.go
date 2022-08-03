// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/baas-project/baas/pkg/model/images"
	"github.com/baas-project/baas/pkg/model/user"

	"github.com/baas-project/baas/pkg/fs"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (api_ *API) checkUserImage(w http.ResponseWriter, r *http.Request) (*images.ImageModel, error) {
	uniqueID, err := GetTag("uuid", w, r)
	if err != nil {
		http.Error(w, "uri does not include UUID", http.StatusInternalServerError)
		log.Errorf("image error: %v", err)
		return nil, errors.New("failed to get image")
	}

	image, err := api_.store.GetImageByUUID(images.ImageUUID(uniqueID))
	if err != nil {
		http.Error(w, "cannot get image", http.StatusInternalServerError)
		log.Errorf("could not get image: %v", err)
		return nil, errors.New("failed to get image")
	}

	session, _ := api_.session.Get(r, "session-name")
	username, ok := session.Values["Username"].(string)

	// TODO: Hardcoded nonesense
	if r.Header.Get("type") == "system" {
		return image, nil
	}

	if !ok || username != image.Username {
		http.Error(w, "user does not own this image", http.StatusForbidden)
		log.Errorf("access denied: %v", ok)
		return nil, errors.New("failed to get image")
	}

	return image, nil
}

// CreateImage creates an image based on a name
// Example request: POST user/Jan/image
// Example body: {"DiskUUID": "30DF-844C", "Name": "Fedora"}
// Example response: {"Name": "Fedora",
//                    "Versions": [{"Version": "2021-11-01T00:11:22.38125222+01:00",
//                                  "ImageModelID": 0}],
//                    "UUID": "eed13670-5974-4c98-b044-347e1f630bc5",
//                    "DiskUUID": "30DF-844C",
//                    "UserModelID": 0}
func (api_ *API) CreateImage(w http.ResponseWriter, r *http.Request) {
	image := images.ImageModel{}
	err := json.NewDecoder(r.Body).Decode(&image)

	// Input validation
	if image.Name == "" {
		http.Error(w, "Name is not allowed to be empty", http.StatusBadRequest)
		return
	}

	if image.Username == "" {
		http.Error(w, "Username is not allowed to be empty", http.StatusBadRequest)
		return
	}

	if len(image.Versions) != 0 {
		http.Error(w, "There shouldn't be a version", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "couldn't decode image model", http.StatusBadRequest)
		log.Errorf("decode image model: %v", err)
		return
	}

	// Generate the UUID and create the entry in the database.
	// We don't actually make an image file yet.
	image.UUID = images.ImageUUID(uuid.New().String())

	if image.Type == "" {
		image.Type = "base"
	}

	api_.store.CreateImage(&image)

	if err != nil {
		http.Error(w, "couldn't create image model", http.StatusInternalServerError)
		log.Errorf("decode create model: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&image)
}

// GetImage gets any image based on its unique id.
// Example request: image/57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf
// Example response: {
//  "Name": "Gentoo",
//  "Versions": null,
//  "UUID": "57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf",
//  "DiskUUID": "30DF-844C",
//  "UserModelID": 1
//}
func (api_ *API) GetImage(w http.ResponseWriter, r *http.Request) {
	image, err := api_.checkUserImage(w, r)
	if err != nil {
		return
	}

	_ = json.NewEncoder(w).Encode(image)
}

// UpdateImage changes some of the parameters of the image
// Example request: PUT image/57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf
// Example response: the updated image
func (api_ *API) UpdateImage(w http.ResponseWriter, r *http.Request) {
	oldImage, err := api_.checkUserImage(w, r)
	if err != nil {
		return
	}

	var newImage images.ImageModel
	err = json.NewDecoder(r.Body).Decode(&newImage)
	if err != nil || oldImage.UUID != newImage.UUID {
		http.Error(w, "invalid machine given", http.StatusBadRequest)
		log.Errorf("Invalid machine given: %v", err)
		return
	}

	api_.store.UpdateImage(&newImage)

	_ = json.NewEncoder(w).Encode(newImage)
}

// DeleteImage removes an image based on its UUID
// Example request: DELETE image/57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf
// Example response: Successfully deleted image
func (api_ *API) DeleteImage(w http.ResponseWriter, r *http.Request) {
	image, err := api_.checkUserImage(w, r)
	if err != nil {
		return
	}

	// Delete the images and versions from the database
	if api_.store.DeleteImage(image) != nil {
		http.Error(w, "couldn't delete image", http.StatusInternalServerError)
		log.Errorf("delete image: %v", err)
		return
	}

	http.Error(w, "Successfully deleted image", http.StatusOK)
}

// DownloadImageFile gets the specified version of the image off the disk and offers it to the client
func DownloadImageFile(image *images.ImageModel, version string, w http.ResponseWriter) {
	val, err := strconv.ParseUint(version, 10, 64)
	if err != nil {
		http.Error(w, "Cannot download the image", http.StatusNotFound)
		log.Errorf("Download image: %v", err)
		return
	}

	f, err := image.OpenImageFile(val)
	if err != nil {
		http.Error(w, "Cannot download the image", http.StatusNotFound)
		log.Errorf("Download image: %v", err)
		return
	}

	// Defer closing the file until the end of the program
	defer func() {
		err = f.Close()
		if err != nil {
			http.Error(w, "Cannot close image file", http.StatusInternalServerError)
			log.Errorf("Cannot close image file: %v", err)
		}
	}()

	// We set the Content-Type to disk/raw as a placeholder, but it does not actually exist. It might be nice to change
	// this at some later date some more common value.
	w.Header().Set("Content-Type", "disk/raw")
	_, err = io.Copy(w, f)

	if err != nil {
		http.Error(w, "Cannot serve image", http.StatusInternalServerError)
		log.Errorf("Cannot serve image: %v", err)
		return
	}
}

// DownloadImage offers the requested image to the respective client
// Example request: image/87f58936-9540-4dad-aba6-253f06142166/1635888918
func (api_ *API) DownloadImage(w http.ResponseWriter, r *http.Request) {
	version, err := GetTag("version", w, r)
	if err != nil {
		http.Error(w, "Invalid version in the URI", http.StatusInternalServerError)
		log.Errorf("Download image: %v", err)
		return
	}

	image, err := api_.checkUserImage(w, r)
	if err != nil {
		return
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("filename=%s-%s.img", image.UUID, version))

	DownloadImageFile(image, version, w)
}

// DownloadLatestImage offers the latest version
// Example request: image/87f58936-9540-4dad-aba6-253f06142166/latest
func (api_ *API) DownloadLatestImage(w http.ResponseWriter, r *http.Request) {
	image, err := api_.checkUserImage(w, r)
	if err != nil {
		return
	}

	versionTxt := image.Versions[len(image.Versions)-1]
	version := strconv.FormatUint(versionTxt.Version, 10)

	w.Header().Add("Content-Disposition", fmt.Sprintf("filename=%s-%s.img", image.UUID, version))
	DownloadImageFile(image, version, w)
}

func createNewVersion(api *API, uniqueID string) (*images.Version, error) {
	v, err := CreateNewVersion(uniqueID, api.store)
	return &v, err
}

func updateVersion(api *API, uniqueID string) (*images.Version, error) {
	image, err := api.store.GetImageByUUID(images.ImageUUID(uniqueID))
	if err != nil {
		return nil, err
	}

	// Update the latest version
	return &image.Versions[len(image.Versions)-1], nil
}

func manageVersion(api *API, versionPart io.Reader, uniqueID string) (*images.Version, error) {
	data, err := ioutil.ReadAll(versionPart)
	val := strings.ToLower(string(data))

	if err != nil || (val != "true" && val != "false") {
		return nil, err
	}

	var version *images.Version
	if val == "true" {
		version, err = createNewVersion(api, uniqueID)
	} else {
		version, err = updateVersion(api, uniqueID)
	}

	return version, err
}

// UploadImage takes the uploaded file and stores as a new version of the image
// Example request: image/87f58936-9540-4dad-aba6-253f06142166 -H "Content-Type: multipart/form-data"
//                     -F "file=@/tmp/test3.img"
// Example return: Successfully uploaded image: 134251234
func (api_ *API) UploadImage(w http.ResponseWriter, r *http.Request) {
	log.Info("Started with upload")
	image, err := api_.checkUserImage(w, r)
	if err != nil {
		return
	}

	// Get the reader to the multireader
	mr, err := r.MultipartReader()
	if ErrorWrite(w, err, "Cannot parse POST form") != nil {
		return
	}

	// Get the parameters for this update
	// TODO: Bad design. Write a new endpoint or use a header for this.
	versionPart, err := mr.NextPart()
	if ErrorWrite(w, err, "File upload failed") != nil {
		return
	}

	// One liner which closes the file at the end of the call.
	defer func() {
		if err = versionPart.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()

	version, err := manageVersion(api_, versionPart, string(image.UUID))
	if err != nil {
		http.Error(w, "cannot fetch the image from the database", http.StatusNotFound)
		log.Errorf("cannot fetch image from database: %v", err)
		return
	}

	// We only use the first part right now, but this might change
	p, err := mr.NextPart()
	if ErrorWrite(w, err, "File upload failed") != nil {
		return
	}

	// One liner which closes the file at the end of the call.
	defer func() {
		if err = p.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()

	// Write the file onto the disk
	dest, err := os.OpenFile(fmt.Sprintf(api_.diskpath+images.FilePathFmt, image.UUID, version.Version),
		os.O_WRONLY|os.O_CREATE, 0644)
	if ErrorWrite(w, err, "Cannot open destination file") != nil {
		return
	}

	err = fs.CopyStream(p, dest)

	if ErrorWrite(w, err, "Cannot copy over the contents of the file") != nil {
		return
	}

	defer func() {
		if err := dest.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()
	http.Error(w, "Successfully uploaded image: "+strconv.FormatUint(version.Version, 10), http.StatusOK)
}

// RegisterImageHandlers sets the metadata for each of the routes and registers them to the global handler
func (api_ *API) RegisterImageHandlers() {
	api_.Routes = append(api_.Routes, Route{
		URI:         "/image",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.CreateImage,
		Method:      http.MethodPost,
		Description: "Creates a new image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.GetImage,
		Method:      http.MethodGet,
		Description: "Gets information about an image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.DeleteImage,
		Method:      http.MethodDelete,
		Description: "Removes an image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.UpdateImage,
		Method:      http.MethodPut,
		Description: "Updates an image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}/latest",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.DownloadLatestImage,
		Method:      http.MethodPost,
		Description: "Offers the latest version of the image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}/{version}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.DownloadImage,
		Method:      http.MethodGet,
		Description: "Requests a particular version of the image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.UploadImage,
		Method:      http.MethodPost,
		Description: "Uploads a new version of the image",
	})
}
