package api

import (
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


}

//func (api *Api) GetImage(w http.ResponseWriter, r *http.Request) {
//
//}
