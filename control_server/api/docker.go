// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/baas-project/baas/pkg/model/images"
	"github.com/baas-project/baas/pkg/model/user"

	"github.com/baas-project/baas/pkg/fs"
	log "github.com/sirupsen/logrus"
)

func extractPart(r *http.Request) (*multipart.Part, error) {
	// Get the reader and writer of the multireader
	mr, err := r.MultipartReader()
	if err != nil {
		log.Errorf("cannot parse POST form: %v", err)
		return nil, err
	}

	p, err := mr.NextPart()

	if err != nil {
		log.Errorf("cannot fetch image from database: %v", err)
		return nil, err
	}

	return p, nil
}

func runDockerCommand(dir string) error {
	cmd := exec.Command("utils/docker2image.sh", dir)

	stderr, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("Failed to open a standard output pipe: %v", err)
		return err
	}

	if err := cmd.Run(); err != nil {
		log.Errorf("Cannot run the docker2image script: %v", err)

		// If we error out here we probably don't care
		slurp, _ := io.ReadAll(stderr)
		log.Errorf("Output: %s", slurp)
		return err
	}

	return nil
}

func renameFiles(dir string, diskpath string, filepathfmt string, uniqueID string, version uint64) error {
	if err := os.Rename(dir+"/"+"image.img",
		fmt.Sprintf(diskpath+filepathfmt, uniqueID, version)); err != nil {
		log.Errorf("Failed to move image file: %v", err)
		return err
	}

	newPath := fmt.Sprintf(diskpath+"/%s/Dockerfile-%d", uniqueID, version)
	if err := os.Rename(dir+"/"+"Dockerfile", newPath); err != nil {
		log.Errorf("Failed to move dockerfile: %v", err)
		return err
	}

	return nil
}

// RunDocker takes a Dockerfile and generates a bootable OS image
// Request request: /image/{uuid}/docker
func (api_ *API) RunDocker(w http.ResponseWriter, r *http.Request) {
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

	p, err := extractPart(r)
	if err != nil {
		http.Error(w, "cannot parse POST form", http.StatusBadRequest)
		return
	}

	// One liner which closes the file at the end of the call.
	defer func() {
		if err = p.Close(); err != nil {
			log.Errorf("Cannot close upload file: %v", err)
		}
	}()

	dir := api_.diskpath + "/" + uniqueID

	// Delete the directory at the end of the program
	// defer os.RemoveAll(dir)

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

	// Run a shell script which creates the image
	// It is written in shell for convience, it could be rewritten in Go if someone really cares.
	if runDockerCommand(dir) != nil {
		http.Error(w, "cannot compile docker image", http.StatusInternalServerError)
		return
	}

	if renameFiles(dir, api_.diskpath, images.FilePathFmt, uniqueID, version.Version) != nil {
		http.Error(w, "cannot compile docker image", http.StatusInternalServerError)
		return
	}

	http.Error(w, "Successfully uploaded image: "+strconv.FormatUint(version.Version, 10), http.StatusOK)
}

// RegisterImageDockerHandlers sets the metadata for each of the routes and registers them to the global handler
func (api_ *API) RegisterImageDockerHandlers() {
	api_.Routes = append(api_.Routes, Route{
		URI:         "/image/{uuid}/docker",
		Permissions: []user.UserRole{user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.RunDocker,
		Method:      http.MethodPost,
		Description: "Uploads a new version of the image",
	})
}
