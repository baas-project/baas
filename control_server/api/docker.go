// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"github.com/baas-project/baas/pkg/fs"
	"github.com/baas-project/baas/pkg/images"
	"github.com/baas-project/baas/pkg/model"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

// RunDocker takes a Dockerfile and generates a bootable OS image
// Request request: /image/{uuid}/docker
func (api_ *API) RunDocker(w http.ResponseWriter, r *http.Request) {
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

	version, err := CreateNewVersion(uniqueID, api_.store)
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

	dir := api_.diskpath + "/" + uniqueID

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
		fmt.Sprintf(api_.diskpath+images.FilePathFmt, uniqueID, version.Version)); err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Failed to move image file: %v", err)
		return
	}

	newPath := fmt.Sprintf(api_.diskpath+"/%s/Dockerfile-%d", uniqueID, version.Version)
	if err := os.Rename(dir+"/"+"Dockerfile", newPath); err != nil {
		http.Error(w, "Cannot compile docker image", http.StatusInternalServerError)
		log.Errorf("Failed to move dockerfile: %v", err)
		return
	}

	http.Error(w, "Successfully uploaded image: "+strconv.FormatUint(version.Version, 10), http.StatusOK)
}

func (api_ *API) RegisterImageDockerHandlers() {
	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}/docker",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api_.RunDocker,
		Method:      http.MethodPost,
		Description: "Uploads a new version of the image",
	})
}
