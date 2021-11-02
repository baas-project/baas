package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// filePathFmt is the format string to create the path the image should be written to
const filePathFmt = "/%s/%v.img"

// maximumFileSize is the maximum file size allowed to be uploaded as an image
const maximumFileSize = 5 * 1024 * 1024 * 1024 // 5 GiB
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

// CreateImageFile creates the actual image on disk with a given size.
func CreateImageFile(imageSize uint, image *model.ImageModel, api *API) error {
	f, err := os.OpenFile(fmt.Sprintf(api.diskpath+filePathFmt, image.UUID, image.Versions[0].Version),
		os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return err
	}

	// Create an image with a size of 512 MiB
	size := int64(imageSize * 1024 * 1024)

	_, err = f.Seek(size-1, 0)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte{0})
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
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
func (api *API) CreateImage(w http.ResponseWriter, r *http.Request) {
	name, err := GetName(w, r)
	if err != nil {
		return
	}

	image := model.ImageModel{}
	err = json.NewDecoder(r.Body).Decode(&image)

	// Input validation
	if image.Name == "" {
		http.Error(w, "Name is not allowed to be empty", http.StatusBadRequest)
		return
	}

	if len(image.Versions) != 0 {
		http.Error(w, "There shouldn't be a version", http.StatusBadRequest)
		return
	}

	if image.DiskUUID == "" {
		http.Error(w, "DiskUUID is not allowed to be empty", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "couldn't decode image model", http.StatusBadRequest)
		log.Errorf("decode image model: %v", err)
		return
	}

	// Generate the UUID and create the entry in the database.
	// We don't actually make an image file yet.
	image.UUID = model.ImageUUID(uuid.New().String())
	err = api.store.CreateImage(name, &image)
	if err != nil {
		http.Error(w, "couldn't create image model", http.StatusInternalServerError)
		log.Errorf("decode create model: %v", err)
		return
	}

	// Create the actual image together with the first empty version which a user may or may not use.
	err = os.Mkdir(fmt.Sprintf(api.diskpath+"/%s", image.UUID), os.ModePerm)
	if err != nil {
		http.Error(w, "could not create image", http.StatusInternalServerError)
		log.Errorf("cannot create image directory: %v", err)
		return
	}

	err = CreateImageFile(imageFileSize, &image, api)

	if err != nil {
		http.Error(w, "Cannot create the image file", http.StatusInternalServerError)
		log.Errorf("image creation failed: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&image)
}

// GetImagesByUser fetches all the images of the given user
// Example request: user/Jan/images
// Example result: [
//  {
//    "Name": "Windows",
//    "Versions "a9c11954-6161-410b-b238-c03df5c529e9",
//    "DiskUUID": "30DF-844C",
//    "UserModelID": 2
//  },
//  {
//    "Name": "Arch Linux",
//    "Versions": null,
//    "UUID": "341b2c69-8776-4e54-9330-7c9692f7ed28",
//    "DiskUUID": "30DF-844C",
//    "UserModelID": 2
//  }
//]
func (api *API) GetImagesByUser(w http.ResponseWriter, r *http.Request) {
	name, err := GetName(w, r)
	if err != nil {
		return
	}

	images, err := api.store.GetImagesByUsername(name)

	if err != nil {
		http.Error(w, "couldn't get images", http.StatusInternalServerError)
		log.Errorf("get images by users: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(images)
}

// GetImagesByName gets any image based on the user who created it and human-readable name assigned to it.
// Example Request: user/Jan/images/Gentoo
// Example Response: [
//  {
//    "Name": "Gentoo",
//    "Versions": null,
//    "UUID": "57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf",
//    "DiskUUID": "30DF-844C",
//    "UserModelID": 1
//  }
//]
func (api *API) GetImagesByName(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		return
	}

	imageName, err := GetTag("image_name", w, r)
	if err != nil {
		return
	}

	// TODO: Security needs to be done using auth system instead, no role checking in this route code.
	images, err := api.store.GetImagesByNameAndUsername(imageName, username)

	if err != nil {
		http.Error(w, "couldn't get image", http.StatusInternalServerError)
		log.Errorf("get image by name: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(images)
}

// GetImage gets any image based on it's unique id.
// Example request: image/57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf
// Example response: {
//  "Name": "Gentoo",
//  "Versions": null,
//  "UUID": "57bf0cd3-c2bf-4257-acdd-b7f1c8633fcf",
//  "DiskUUID": "30DF-844C",
//  "UserModelID": 1
//}
func (api *API) GetImage(w http.ResponseWriter, r *http.Request) {
	uniqueID, err := GetTag("uuid", w, r)
	if err != nil {
		return
	}

	res, err := api.store.GetImageByUUID(model.ImageUUID(uniqueID))
	if err != nil {
		http.Error(w, "couldn't get image", http.StatusInternalServerError)
		log.Errorf("get image: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(res)
}

// DownloadImageFile gets the specified version of the image off the disk and offers it to the client
func DownloadImageFile(uniqueID string, version string, api *API, w http.ResponseWriter) {
	f, err := os.Open(fmt.Sprintf(api.diskpath+filePathFmt, uniqueID, version))
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
func (api *API) DownloadImage(w http.ResponseWriter, r *http.Request) {
	version, err := GetTag("version", w, r)
	if err != nil {
		http.Error(w, "Invalid version in the URI", http.StatusInternalServerError)
		log.Errorf("Download image: %v", err)
		return
	}

	uniqueID, err := GetTag("uuid", w, r)
	if err != nil {
		http.Error(w, "Invalid uuid in the URI", http.StatusInternalServerError)
		log.Errorf("Download image: %v", err)
		return
	}

	DownloadImageFile(uniqueID, version, api, w)
}

// DownloadLatestImage offers the latest version
// Example request: image/87f58936-9540-4dad-aba6-253f06142166/latest
func (api *API) DownloadLatestImage(w http.ResponseWriter, r *http.Request) {
	uniqueID, err := GetTag("uuid", w, r)
	if err != nil {
		http.Error(w, "Invalid uuid in the URI", http.StatusInternalServerError)
		log.Errorf("Download image: %v", err)
		return
	}

	image, err := api.store.GetImageByUUID(model.ImageUUID(uniqueID))
	if err != nil {
		http.Error(w, "Invalid uuid in the URI", http.StatusInternalServerError)
		log.Errorf("Download latest image: %v", err)
		return
	}

	version := image.Versions[len(image.Versions)-1]
	DownloadImageFile(uniqueID, strconv.FormatInt(version.Version, 10), api, w)
}

// UploadImage takes the uploaded file and stores as a new version of the image
// Example request: image/87f58936-9540-4dad-aba6-253f06142166 -H "Content-Type: multipart/form-data"
//                     -F "file=@/tmp/test3.img"
// Example return: Successfully uploaded image: 134251234
func (api *API) UploadImage(w http.ResponseWriter, r *http.Request) {
	uniqueID, err := GetTag("uuid", w, r)
	if err != nil {
		return
	}

	// Accept maximum file of 5 Gigabytes (wowzers~)
	// https://stackoverflow.com/questions/40684307/how-can-i-receive-an-uploaded-file-using-a-golang-net-http-server
	err = r.ParseMultipartForm(maximumFileSize)

	if err != nil {
		http.Error(w, "Cannot parse POST form", http.StatusBadRequest)
		log.Errorf("Cannot parse POST form: %v", err)
		return
	}

	f, _, err := r.FormFile("file")

	// Go error's handling is pain
	if err != nil {
		http.Error(w, "File upload failed", http.StatusInternalServerError)
		log.Errorf("Cannot upload image: %v", err)
		return
	}

	// One liner which closes the file at the end of the call.
	defer func() {
		if err = f.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()

	// First fetch the image from the database, so we can get the id using the unique id.
	// Do not ask me why this is needed, revamp of the database might be needed.
	image, err := api.store.GetImageByUUID(model.ImageUUID(uniqueID))

	if err != nil {
		http.Error(w, "cannot fetch the image from the database", http.StatusNotFound)
		log.Errorf("cannot fetch image from database: %v", err)
		return
	}

	version := model.Version{
		Version:      time.Now().Unix(),
		ImageModelID: image.ID,
	}

	api.store.CreateNewImageVersion(version)

	// Write the file onto the disk
	dest, err := os.OpenFile(fmt.Sprintf(api.diskpath+filePathFmt, uniqueID, version.Version), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		http.Error(w, "Cannot open destination file", http.StatusInternalServerError)
		log.Errorf("Cannot open destination file: %v", err)
		return
	}

	_, err = io.Copy(dest, f)
	if err != nil {
		http.Error(w, "Cannot copy over the contents of the file", http.StatusInternalServerError)
		log.Errorf("Cannot copy the contents of the file: %v", err)
		return
	}

	defer func() {
		if err := dest.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()
	http.Error(w, "Successfully uploaded image: "+strconv.FormatInt(version.Version, 10), http.StatusOK)
}
