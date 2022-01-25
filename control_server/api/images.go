package api

import (
	"encoding/json"
	"fmt"
	"github.com/baas-project/baas/pkg/images"
	"github.com/baas-project/baas/pkg/util"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/baas-project/baas/pkg/fs"
	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// imageFileSize is the size of the standard image that is created in MiB.
const imageFileSize = 6 // size in Gib

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

// createNewVersion creates a new version for a specified image
func createNewVersion(uuid string, api *API) (images.Version, error) {
	// First fetch the image from the database, so we can get the id using the unique id.
	// Do not ask me why this is needed, revamp of the database might be needed.
	image, err := api.store.GetImageByUUID(images.ImageUUID(uuid))

	if err != nil {
		return images.Version{}, errors.New("Cannot fetch image from database")
	}

	version := images.Version{Version: image.Versions[len(image.Versions)-1].Version + 1,
		ImageModelUUID: image.UUID}

	api.store.CreateNewImageVersion(version)
	return version, nil
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

	image := images.ImageModel{}
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

	if err != nil {
		http.Error(w, "couldn't decode image model", http.StatusBadRequest)
		log.Errorf("decode image model: %v", err)
		return
	}

	// Generate the UUID and create the entry in the database.
	// We don't actually make an image file yet.
	image.UUID = images.ImageUUID(uuid.New().String())
	image.Username = name
	api.store.CreateImage(&image)

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

	err = image.CreateImageFile(imageFileSize, api.diskpath, images.SizeMegabyte)

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

	userImages, err := api.store.GetImagesByUsername(name)

	if err != nil {
		http.Error(w, "couldn't get userImages", http.StatusInternalServerError)
		log.Errorf("get userImages by users: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(userImages)
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

	userImages, err := api.store.GetImagesByNameAndUsername(imageName, username)

	if err != nil {
		http.Error(w, "couldn't get image", http.StatusInternalServerError)
		log.Errorf("get image by name: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(userImages)
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

	res, err := api.store.GetImageByUUID(images.ImageUUID(uniqueID))
	if err != nil {
		http.Error(w, "couldn't get image", http.StatusInternalServerError)
		log.Errorf("get image: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(res)
}

// DownloadImageFile gets the specified version of the image off the disk and offers it to the client
func DownloadImageFile(uniqueID string, version string, api *API, w http.ResponseWriter) {
	f, err := os.Open(fmt.Sprintf(api.diskpath+images.FilePathFmt, uniqueID, version))
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

	image, err := api.store.GetImageByUUID(images.ImageUUID(uniqueID))
	if err != nil {
		http.Error(w, "Invalid uuid in the URI", http.StatusInternalServerError)
		log.Errorf("Download latest image: %v", err)
		return
	}

	version := image.Versions[len(image.Versions)-1]

	DownloadImageFile(uniqueID, strconv.FormatUint(version.Version, 10), api, w)
}

// UploadImage takes the uploaded file and stores as a new version of the image
// Example request: image/87f58936-9540-4dad-aba6-253f06142166 -H "Content-Type: multipart/form-data"
//                     -F "file=@/tmp/test3.img"
// Example return: Successfully uploaded image: 134251234
func (api *API) UploadImage(w http.ResponseWriter, r *http.Request) {
	log.Info("Started with upload")
	uniqueID, err := GetTag("uuid", w, r)
	if err != nil {
		return
	}

	// Get the reader to the multireader
	mr, err := r.MultipartReader()

	if err != nil {
		http.Error(w, "Cannot parse POST form", http.StatusBadRequest)
		log.Errorf("Cannot parse POST form: %v", err)
		return
	}

	// We only use the first part right now, but this might change
	p, err := mr.NextPart()

	// Go error's handling is pain
	if err != nil {
		http.Error(w, "File upload failed", http.StatusInternalServerError)
		log.Errorf("Cannot upload image: %v", err)
		return
	}

	// One liner which closes the file at the end of the call.
	defer func() {
		if err = p.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()

	version, err := createNewVersion(uniqueID, api)
	if err != nil {
		http.Error(w, "cannot fetch the image from the database", http.StatusNotFound)
		log.Errorf("cannot fetch image from database: %v", err)
		return
	}

	// Write the file onto the disk
	dest, err := os.OpenFile(fmt.Sprintf(api.diskpath+images.FilePathFmt, uniqueID, version.Version), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		http.Error(w, "Cannot open destination file", http.StatusInternalServerError)
		log.Errorf("Cannot open destination file: %v", err)
		return
	}

	err = fs.CopyStream(p, dest)
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
	http.Error(w, "Successfully uploaded image: "+strconv.FormatUint(version.Version, 10), http.StatusOK)
}

// createImageSetup defines an endpoint which creates an ImageSetup in the database
// Example request: POST /user/[name]/image_setup
// Example response: Succesfully created image setup.
func (api *API) createImageSetup(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to create image setup", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	// Create an ImageSetup and associate it with an user
	image := images.ImageSetup{}
	err = json.NewDecoder(r.Body).Decode(&image)
	image.User = username
	image.UUID = images.ImageUUID(uuid.New().String())

	if image.Name == "" {
		http.Error(w, "Did not set image setup name", http.StatusBadRequest)
		log.Errorf("Did not sent image setup name: %v", err)
		return
	}

	err = api.store.CreateImageSetup(username, &image)
	if err != nil {
		http.Error(w, "Failed to create image setup", http.StatusBadRequest)
		log.Errorf("Error creating database entry: %v", err)
		return
	}

	http.Error(w, "Successfully created image setup", http.StatusOK)
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
func (api *API) findImageSetupsByUsername(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	// TODO: Better unique error returns
	imageSetup, err := api.store.FindImageSetupsByUsername(username)
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
func (api *API) getImageSetup(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	uuid_, err := GetTag("uuid", w, r)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("UUID not found in URI: %v", err)
		return
	}

	setup, err := api.store.GetImageSetup(username, uuid_)
	if err != nil {
		http.Error(w, "Failed to find image setup", http.StatusBadRequest)
		log.Errorf("Cannot find image setup: %v", err)
		return
	}
	util.PrettyPrintStruct(setup.Images)
	_ = json.NewEncoder(w).Encode(setup)
}

// addImageToImageSetup add an ImageModel to the associated ImageSetup
// Example request: POST /[name]/image_setup/[uuid]
// Example body: {"Uuid": "3a760707-c160-40fa-81be-430b75131ddc", "Version": 3}
// Example response:
//  {"Name": "Linux Kernel 2",
//   "Images": [{"UUIDImage":"3a760707-c160-40fa-81be-430b75131ddc","VersionNumber":3}],
//   "User":"ValentijnvdBeek",
//   "UUID":"2b59ff94-7fb6-4239-b2e6-82f1e30f4355"}
func (api *API) addImageToImageSetup(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil {
		http.Error(w, "Failed to add image to image setups", http.StatusBadRequest)
		log.Errorf("Username not found in URI: %v", err)
		return
	}

	uuid_, err := GetTag("uuid", w, r)
	if err != nil {
		http.Error(w, "Failed to find image setups", http.StatusBadRequest)
		log.Errorf("UUID not found in URI: %v", err)
		return
	}

	imageMsg := model.ImageSetupMessage{}
	err = json.NewDecoder(r.Body).Decode(&imageMsg)
	if err != nil {
		http.Error(w, "Cannot find image UUID in JSON message.", http.StatusBadRequest)
		log.Errorf("Cannot find image UUID: %v", err)
		return
	}

	image, err := api.store.GetImageByUUID(images.ImageUUID(imageMsg.Uuid))

	if err != nil {
		http.Error(w, "Failed to add image to image setups", http.StatusBadRequest)
		log.Errorf("Cannot find images: %v", err)
		return
	}

	imageSetup, err := api.store.GetImageSetup(username, uuid_)

	if err != nil {
		http.Error(w, "Failed to add image to image setups", http.StatusBadRequest)
		log.Errorf("Cannot find image setup: %v", err)
		return
	}

	version := images.Version{
		Version:        imageMsg.Version,
		ImageModelUUID: image.UUID,
	}

	api.store.AddImageToImageSetup(&imageSetup, image, version)

	_ = json.NewEncoder(w).Encode(imageSetup)
}

// RegisterImageHandlers sets the metadata for each of the routes and registers them to the global handler
func (api *API) RegisterImageHandlers() {
	api.routes = append(api.routes, Route{
		URI:         "/image/{uuid}",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.GetImage,
		Method:      http.MethodGet,
		Description: "Gets information about an image",
	})

	api.routes = append(api.routes, Route{
		URI:         "/image/{uuid}/latest",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.DownloadLatestImage,
		Method:      http.MethodPost,
		Description: "Offers the latest version of the image",
	})

	api.routes = append(api.routes, Route{
		URI:         "/image/{uuid}/docker",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.RunDocker,
		Method:      http.MethodPost,
		Description: "Uploads a new version of the image",
	})

	api.routes = append(api.routes, Route{
		URI:         "/image/{uuid}/{version}",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.DownloadImage,
		Method:      http.MethodGet,
		Description: "Requests a particular version of the image",
	})

	api.routes = append(api.routes, Route{
		URI:         "/image/{uuid}",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.UploadImage,
		Method:      http.MethodPost,
		Description: "Uploads a new version of the image",
	})

	api.routes = append(api.routes, Route{
		URI:         "/user/{name}/image_setup",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.createImageSetup,
		Method:      http.MethodPost,
		Description: "Creates an image setup",
	})

	api.routes = append(api.routes, Route{
		URI:         "/{name}/image_setup",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.findImageSetupsByUsername,
		Method:      http.MethodGet,
		Description: "Find image setups by username",
	})

	api.routes = append(api.routes, Route{
		URI:         "/{name}/image_setup/{uuid}",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.getImageSetup,
		Method:      http.MethodGet,
		Description: "Get a specific image setup",
	})

	api.routes = append(api.routes, Route{
		URI:         "/{name}/image_setup/{uuid}",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.addImageToImageSetup,
		Method:      http.MethodPost,
		Description: "Add image to the setup system",
	})
}

// RunDocker takes a Dockerfile and generates a bootable OS image
// Request request: /image/{uuid}/docker
func (api *API) RunDocker(w http.ResponseWriter, r *http.Request) {
	// Get the reader to the multireader
	mr, err := r.MultipartReader()

	if err != nil {
		http.Error(w, "Cannot parse POST form", http.StatusBadRequest)
		log.Errorf("Cannot parse POST form: %v", err)
		return
	}

	uniqueID, err := GetTag("uuid", w, r)
	if err != nil {
		return
	}

	version, err := createNewVersion(uniqueID, api)
	if err != nil {
		http.Error(w, "cannot fetch the image from the database", http.StatusNotFound)
		log.Errorf("cannot fetch image from database: %v", err)
		return
	}

	// We only use the first part right now, but this might change
	p, err := mr.NextPart()

	// Go error's handling is pain
	if err != nil {
		http.Error(w, "File upload failed", http.StatusInternalServerError)
		log.Errorf("Cannot upload image: %v", err)
		return
	}

	// One liner which closes the file at the end of the call.
	defer func() {
		if err = p.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()

	dir := api.diskpath + "/" + uniqueID

	if err != nil {
		http.Error(w, "Cannot open destination file", http.StatusInternalServerError)
		log.Errorf("Cannot open destination file: %v", err)
		return
	}

	// Delete the directory at the end of the program
	//defer os.RemoveAll(dir)

	// Write the docker file to the directory
	f, err := os.OpenFile(dir+"/Dockerfile", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Cannot write to DockerImage file: %v", err)
		return
	}

	defer f.Close()

	err = fs.CopyStream(p, f)
	if err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Cannot write to DockerImage file: %v", err)
		return
	}

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	// Run a shell script which creates the image
	// It is written in shell for convience, it could be rewritten in Go if someone really cares.
	cmd := exec.Command("utils/docker2image.sh", dir)

	fmt.Println(cmd)
	stderr, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Cannot run the docker2image script: %v", err)
		return
	}

	if err := cmd.Run(); err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Cannot run the docker2image script: %v", err)

		// If we error out here we probably don't care
		slurp, _ := io.ReadAll(stderr)
		log.Errorf("Output: %s", slurp)
		return
	}

	if err := os.Rename(dir+"/"+"image.img",
		fmt.Sprintf(api.diskpath+images.FilePathFmt, uniqueID, version.Version)); err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Failed to move image file: %v", err)
		return
	}

	newPath := fmt.Sprintf(api.diskpath+"/%s/Dockerfile-%d", uniqueID, version.Version)
	if err := os.Rename(dir+"/"+"Dockerfile", newPath); err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Failed to move dockerfile: %v", err)
		return
	}

	http.Error(w, "Successfully uploaded image: "+strconv.FormatUint(version.Version, 10), http.StatusOK)
}
