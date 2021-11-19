package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"gorm.io/gorm"
)

var conf = &oauth2.Config{
	ClientID:     "2162911b22578f57f3e0",
	ClientSecret: "198f08805ae79b48d8afd943780a335514fdf640",
	Scopes:       []string{"user"},
	Endpoint:     github.Endpoint,
}

// returnUserByOAuth gets or creates the associated user from the database.
func (api *API) returnUserByOAuth(username string, email string, realName string) (*model.UserModel, error) {
	user, err := api.store.GetUserByUsername(username)

	// Create the user if we cannot find it in the database.
	if err == gorm.ErrRecordNotFound {
		user = &model.UserModel{
			Username: username,
			Name:     realName,
			Email:    email,
			Role:     model.User,
		}

		api.store.CreateUser(user)
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

// LoginGithub defines the entrypoint to start the OAuth flow
func (api *API) LoginGithub(w http.ResponseWriter, _ *http.Request) {
	// Redirect the user
	uri := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	w.WriteHeader(200)
	w.Write([]byte("Visit the following URI: a " + uri))

}

// LoginGithubCallback gets the token and creates the user model for the Github User
func (api *API) LoginGithubCallback(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session, _ := api.session.Get(r, "session-name")

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

	user, err := api.returnUserByOAuth(loginInfo.Login, loginInfo.Email, loginInfo.Email)

	if err != nil {
		http.Error(w, "Cannot find the user in the database", http.StatusBadRequest)
		return
	}

	uuid, err := uuid.NewUUID()

	if err != nil {
		http.Error(w, "Cannot generate UUID", http.StatusBadRequest)
		return
	}

	// Set the session ID and username
	session.Values["Session"] = uuid.String()
	session.Values["Username"] = user.Username
	fmt.Println(session.Values)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(session.ID)
	response := fmt.Sprintf("Session cookie: %s\n", session.ID)
	w.WriteHeader(200)

	w.Write([]byte(response))

}
