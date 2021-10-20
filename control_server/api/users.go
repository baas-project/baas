package api

import (
	"encoding/json"
	"github.com/baas-project/baas/pkg/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (api *Api) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := api.store.GetUsers()

	if err != nil {
		http.Error(w, "couldn't get users", http.StatusInternalServerError)
		log.Errorf("get users: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(users)
}

func (api *Api) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.UserModel
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "invalid user given", http.StatusBadRequest)
		log.Errorf("Invalid user given: %v", err)
		return
	}

	if user.Name == "" {
		http.Error(w, "No username given", http.StatusBadRequest)
		return
	} else if user.Email == "" {
		http.Error(w, "No email given", http.StatusBadRequest)
		return
	} else if user.Role == "" {
		http.Error(w, "No role given", http.StatusBadRequest)
		return
	}

	err = api.store.CreateUser(&user)
	if err != nil {
		http.Error(w, "couldn't create user", http.StatusInternalServerError)
		log.Errorf("create user: %v", err)
		return
	}
}

func (api *Api) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok || name == "" {
		http.Error(w, "name not found", http.StatusBadRequest)
		log.Errorf("name not provided in get user")
		return
	}

	users, err := api.store.GetUserByName(name)

	if err != nil {
		http.Error(w, "couldn't get users", http.StatusInternalServerError)
		log.Errorf("get users: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(users)
}
