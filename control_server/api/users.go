package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baas-project/baas/pkg/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// GetUsers fetches all the users from the database
// Example request: users
// Response: [{"Name": "Valentijn", "Email": "v.d.vandebeek@student.tudelft.nl",
//             "Role": "admin", "Image": null}
func (api *API) GetUsers(w http.ResponseWriter, _ *http.Request) {
	users, err := api.store.GetUsers()

	if err != nil {
		http.Error(w, "couldn't get users", http.StatusInternalServerError)
		log.Errorf("get users: %v", err)
		return
	}

	_ = json.NewEncoder(w).Encode(users)
}

// CreateUser creates a new user in the database
// Example request: user, {"name": "William Narchi",
//                         "email", "w.narchi1@student.tudelft.nl",
//                         "role": "user"}
// Response: Either an error message or success.
func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.UserModel
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "invalid user given", http.StatusBadRequest)
		log.Errorf("Invalid user given: %v", err)
		return
	}

	if user.Username == "" {
		http.Error(w, "No username given", http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "No name given", http.StatusBadRequest)
		return
	}

	if user.Email == "" {
		http.Error(w, "No email given", http.StatusBadRequest)
		return
	}

	if user.Role == "" {
		http.Error(w, "No role given", http.StatusBadRequest)
		return
	}

	err = api.store.CreateUser(&user)
	if err != nil {
		http.Error(w, "couldn't create user", http.StatusInternalServerError)
		log.Errorf("create user: %v", err)
		return
	}
	_, err = fmt.Fprintf(w, "Successfully created user\n")
	if err != nil {
		log.Error("Error writing over http")
		return
	}
}

// GetLoggedInUser gets the currently logged-in user and returns it.
// Example request: user/me
func (api *API) GetLoggedInUser(w http.ResponseWriter, r *http.Request) {
	session, _ := api.session.Get(r, "session-name")
	username, ok := session.Values["Username"].(string)

	if !ok {
		http.Error(w, "Cannot find username", http.StatusBadRequest)
		return
	}

	user, err := api.store.GetUserByUsername(username)

	if err != nil {
		http.Error(w, "Cannot find user: "+username, http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(user)
}

// GetUser fetches a user based on their name and returns it
// Example request: user/Jan
// Response: {"Name": "Jan",
//            "Email": "v.d.vandebeek@student.tudelft.nl",
//            "role": "admin"}
func (api *API) GetUser(w http.ResponseWriter, r *http.Request) {
	session, _ := api.session.Get(r, "session-name")

	username, ok := session.Values["Username"].(string)
	if !ok {
		http.Error(w, "Username not found", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok || name == "" {
		http.Error(w, "name not found", http.StatusBadRequest)
		log.Errorf("name not provided in get user")
		return
	}

	user, err := api.store.GetUserByUsername(name)

	// Annoyingly enough we can't be more specific due to error wrapping... I swear, this language.
	if err != nil {
		http.Error(w, "couldn't get users", http.StatusInternalServerError)
		log.Errorf("get users: %v", err)
		return
	}

	// Check if the user is allowed to access the profile.
	if user.Role != model.Admin && user.Username != username {
		http.Error(w, "Cannot access this user", http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(w).Encode(user)
}

// RegisterUserHandlers sets the metadata for each of the routes and registers them to the global handler
func (api *API) RegisterUserHandlers() {
	api.routes = append(api.routes, Route{
		URI:         "/users",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: false,
		Handler:     api.GetUsers,
		Method:      http.MethodGet,
		Description: "Gets all the users from the database",
	})

	api.routes = append(api.routes, Route{
		URI:         "/user",
		Permissions: []model.UserRole{model.Admin},
		UserAllowed: false,
		Handler:     api.CreateUser,
		Method:      http.MethodPost,
		Description: "Adds a new user to the database",
	})

	api.routes = append(api.routes, Route{
		URI:         "/user/me",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.GetLoggedInUser,
		Method:      http.MethodGet,
		Description: "Gets the user who is currently logged in",
	})

	api.routes = append(api.routes, Route{
		URI:         "/user/{name}",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.GetUser,
		Method:      http.MethodGet,
		Description: "Gets information about a particular user",
	})

	api.routes = append(api.routes, Route{
		URI:         "/user/{name}/image",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.CreateImage,
		Method:      http.MethodPost,
		Description: "Creates a new image",
	})

	api.routes = append(api.routes, Route{
		URI:         "/user/{name}/images",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.GetImagesByUser,
		Method:      http.MethodGet,
		Description: "Gets all the images owned by a particular user",
	})

	api.routes = append(api.routes, Route{
		URI:         "/user/{name}/images/{image_name}",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api.GetImagesByName,
		Method:      http.MethodGet,
		Description: "Finds all the images by this user with a particular name",
	})
}
