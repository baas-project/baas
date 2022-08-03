// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"github.com/baas-project/baas/pkg/model/images"
	"github.com/baas-project/baas/pkg/model/user"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/baas-project/baas/pkg/database/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestApi_CreateImage(t *testing.T) {
	store, err := sqlite.NewSqliteStore(sqlite.InMemoryPath)
	assert.NoError(t, err)

	userVar := user.UserModel{
		Username: "test",
		Role:     "userVar",
	}

	err = store.CreateUser(&userVar)
	assert.NoError(t, err)

	image := images.ImageModel{
		Name:     "yeet",
		Username: "test",
	}

	var mi bytes.Buffer
	err = json.NewEncoder(&mi).Encode(image)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "/tmp")
	request := httptest.NewRequest(http.MethodPost, "/user/test/image", &mi)
	request.Header.Add("type", "system")
	request.Header.Add("origin", "http://localhost:9090")

	handler.ServeHTTP(resp, request)
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
	os.RemoveAll("/tmp/" + string(decoded.UUID))
}

func TestApi_GetImage(t *testing.T) {
	store, err := sqlite.NewSqliteStore(sqlite.InMemoryPath)
	assert.NoError(t, err)

	userVar := user.UserModel{
		Username: "test",
		Name:     "Test System",
		Role:     "User",
	}

	err = store.CreateUser(&userVar)
	assert.NoError(t, err)

	image := images.ImageModel{
		Name:     "abc",
		UUID:     "def",
		Username: "test",
	}

	store.CreateImage(&image)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "/tmp")
	request := httptest.NewRequest(http.MethodGet, "/image/def", nil)
	request.Header.Add("type", "system")
	request.Header.Add("origin", "http://localhost:9090")

	handler.ServeHTTP(resp, request)
	assert.Equal(t, resp.Code, http.StatusOK)

	decoded := images.ImageModel{}
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	assert.NoError(t, err)

	assert.Equal(t, image.UUID, decoded.UUID)
	assert.Equal(t, image.Name, decoded.Name)
}
