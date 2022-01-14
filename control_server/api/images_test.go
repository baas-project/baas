package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baas-project/baas/pkg/database"
	"github.com/baas-project/baas/pkg/images"
	"github.com/baas-project/baas/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestApi_CreateImage(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	user := model.UserModel{
		Name:   "test",
		Images: nil,
	}

	err = store.CreateUser(&user)
	assert.NoError(t, err)

	image := images.ImageModel{
		Name: "yeet",
	}

	var mi bytes.Buffer
	err = json.NewEncoder(&mi).Encode(image)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/user/"+user.Name+"/image", &mi))

	assert.Equal(t, http.StatusCreated, resp.Code)

	decoded := images.ImageModel{}
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	assert.NoError(t, err)

	assert.NotEmpty(t, decoded.UUID)
	assert.Equal(t, image.Name, decoded.Name)

	res, err := store.GetImageByUUID(decoded.UUID)
	assert.NoError(t, err)

	assert.Equal(t, decoded.UUID, res.UUID)
	assert.Equal(t, image.Name, res.Name)
}

func TestApi_GetImage(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	user := model.UserModel{
		Name:   "test",
		Images: nil,
	}

	err = store.CreateUser(&user)
	assert.NoError(t, err)

	image := images.ImageModel{
		Name: "abc",
		UUID: "def",
	}

	err = store.CreateImage(user.Name, &image)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/image/"+string(image.UUID), nil))

	assert.Equal(t, resp.Code, http.StatusOK)

	decoded := images.ImageModel{}
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	assert.NoError(t, err)

	assert.Equal(t, image.UUID, decoded.UUID)
	assert.Equal(t, image.Name, decoded.Name)
}
