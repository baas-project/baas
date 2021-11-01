package api

import (
	"encoding/json"
	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// GetTag helper function which gets the name out of the request
// Returns the name in the URI
func GetTag(tag string, w http.ResponseWriter, r *http.Request) (string, error) {
	vars := mux.Vars(r)
	res, ok := vars[tag]

	if !ok || res == "" {
		http.Error(w, tag + " not found", http.StatusBadRequest)
		log.Errorf(tag + " not provided")
		return "", errors.New(tag + " not found")
	}

	return res, nil
}

// GetName is a shorthand for GetTag(name, r, w)
func GetName(w http.ResponseWriter, r *http.Request) (string, error) {
	return GetTag("name", w, r)
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
func (api *Api) CreateImage(w http.ResponseWriter, r *http.Request) {
	name, err := GetName(w, r)
	if err != nil { return }

	image := model.ImageModel{}
	err = json.NewDecoder(r.Body).Decode(&image)

	// Input validation
	if image.Name == "" {
		http.Error(w, "Name is not allowed to be empty", http.StatusBadRequest)
		return
	} else if len(image.Versions) != 0 {
		http.Error(w, "There shouldn't be a version", http.StatusBadRequest)
		return
	} else if image.DiskUUID == "" {
		http.Error(w, "DiskUUID is not allowed to be empty", http.StatusBadRequest)
		return
	}

	// Create the first version of the image. It may not make the most sense, though.
	image.Versions = append(image.Versions, model.Version{
		Version: time.Now(),
	})

	if err != nil {
		http.Error(w, "couldn't decode image model", http.StatusBadRequest)
		log.Errorf("decode image model: %v", err)
		return
	}

	// Generate the UUID and create the entry in the database.
	// We don't actually make an image file yet.
	image.UUID = model.ImageUUID(uuid.New().String())
	err = api.store.CreateImage(name, image)
	if err != nil {
		http.Error(w, "couldn't create image model", http.StatusInternalServerError)
		log.Errorf("decode create model: %v", err)
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
//    "Versions": null,
//    "UUID": "a9c11954-6161-410b-b238-c03df5c529e9",
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
func (api *Api) GetImagesByUser(w http.ResponseWriter, r *http.Request) {
	name, err := GetName(w, r)
	if err != nil { return }

	images, err := api.store.GetImagesByUsername(name)

	if err != nil {
		http.Error(w, "couldn't get images", http.StatusInternalServerError)
		log.Errorf("get images by users: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(images)
}

func (api *Api) GetImageByName(w http.ResponseWriter, r *http.Request) {
	username, err := GetName(w, r)
	if err != nil { return }

	imageName, err := GetTag("image_name", w, r)
	if err != nil { return }

	// TODO: Change to images
	// TODO: Fetch by name and user id
	// TODO: Security needs to be done using auth system instead, no role checking in this route code.
	image, err := api.store.GetImageByName(imageName)
	if err != nil {
		http.Error(w, "couldn't get image", http.StatusInternalServerError)
		log.Errorf("get image by name: %v", err)
		return
	}

	givenUser, err := api.store.GetUserByName(username)
	if err != nil {
		http.Error(w, "couldn't find user", http.StatusInternalServerError)
		log.Errorf("get user by id: %v", err)
		return
	}

	// In the image setup there is no easily usable info to check the user permissions or the user who requested the
	// image. Therefore we need to check the permissions ourselves. You might be able to do this a bit faster with a
	// clever-ish SQL query. Please keep in mind this should only go for /user/ images and these restrictions do not
	// apply to system images. TODO: Ensure that works when system images are implemented
	if givenUser.Role != "admin" {
		imageUser, err := api.store.GetUserById(image.UserModelID)

		if err != nil {
			http.Error(w, "couldn't find user", http.StatusInternalServerError)
			log.Errorf("get user by id: %v", err)
			return
		}

		if imageUser.Name != username {
			http.Error(w, "wrong permissions and/or user", http.StatusForbidden)
			log.Errorf("Wrong permission image access %v", err)
			return
		}
	}
	_ = json.NewEncoder(w).Encode(image)
}

func (api *Api) GetImage(w http.ResponseWriter, r *http.Request) {
	uniqueId, err := GetTag("uuid", w, r)
	if err != nil { return }

	res, err := api.store.GetImageByUUID(model.ImageUUID(uniqueId))
	if err != nil {
		http.Error(w, "couldn't get image", http.StatusInternalServerError)
		log.Errorf("get image: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(res)
}

func (api *Api) DownloadImage(w http.ResponseWriter, r *http.Request) {

}

func (api *Api) UploadImage(w http.ResponseWriter, r *http.Request) {

}
