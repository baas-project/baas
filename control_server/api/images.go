package api

import (
	"encoding/json"
	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (api *Api) CreateImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok || name == "" {
		http.Error(w, "name not found", http.StatusBadRequest)
		log.Errorf("name not provided in get user")
		return
	}

	image := model.ImageModel{}
	err := json.NewDecoder(r.Body).Decode(&image)

	if err != nil {
		http.Error(w, "couldn't decode image model", http.StatusBadRequest)
		log.Errorf("decode image model: %v", err)
		return
	}

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

func (api *Api) GetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid, ok := vars["uuid"]
	if !ok || uuid == "" {
		http.Error(w, "uuid not found", http.StatusBadRequest)
		log.Errorf("uuid not provided in get image")
		return
	}

	res, err := api.store.GetImageByUUID(model.ImageUUID(uuid))
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
