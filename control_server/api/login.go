// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"gorm.io/gorm"
)

var conf = &oauth2.Config{
	ClientID:     "2162911b22578f57f3e0",
	ClientSecret: os.Getenv("GITHUB_SECRET"),
	Scopes:       []string{"user"},
	Endpoint:     github.Endpoint,
}

// returnUserByOAuth gets or creates the associated user from the database.
func (api_ *API) returnUserByOAuth(username string, email string, realName string) (*model.UserModel, error) {
	user, err := api_.store.GetUserByUsername(username)

	// Create the user if we cannot find it in the database.
	if err == gorm.ErrRecordNotFound {
		user = &model.UserModel{
			Username: username,
			Name:     realName,
			Email:    email,
			Role:     model.User,
		}

		api_.store.CreateUser(user)
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

// LoginGithub defines the entrypoint to start the OAuth flow
func (api_ *API) LoginGithub(w http.ResponseWriter, _ *http.Request) {
	// Redirect the user
	uri := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	w.WriteHeader(200)
	w.Write([]byte("Visit the following URI: " + uri + "\n"))
}

// LoginGithubCallback gets the token and creates the user model for the GitHub User
func (api_ *API) LoginGithubCallback(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session, _ := api_.session.Get(r, "session-name")

	// Fetch the single-use code from the URI
	ctx := context.Background()
	code := r.URL.Query()["code"][0]

	// Get the OAuth token
	tok, err := conf.Exchange(ctx, code)

	if err != nil {
		http.Error(w, "Invalid OAuth token: "+code, http.StatusBadRequest)
		return
	}

	// Create a client which sends requests using the token.
	client := conf.Client(ctx, tok)

	// Fetch the user information
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Request to GitHub API failed", http.StatusBadRequest)
		return
	}

	var loginInfo model.GitHubLogin
	if err = json.NewDecoder(resp.Body).Decode(&loginInfo); err != nil {
		http.Error(w, "Cannot parse GitHub data", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	user, err := api_.returnUserByOAuth(loginInfo.Login, loginInfo.Email, loginInfo.Email)

	if err != nil {
		http.Error(w, "Cannot find the user in the database", http.StatusBadRequest)
		return
	}

	uuID, err := uuid.NewUUID()

	if err != nil {
		http.Error(w, "Cannot generate UUID", http.StatusBadRequest)
		return
	}

	// Set the session ID and username
	session.Values["Session"] = uuID.String()
	session.Values["Username"] = user.Username
	session.Values["Role"] = string(user.Role)

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the session cookie
	response := "Please check the cookies to get your session ID!"
	w.WriteHeader(200)
	_, err = w.Write([]byte(response))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
