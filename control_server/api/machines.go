// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"github.com/baas-project/baas/pkg/images"
	"github.com/baas-project/baas/pkg/util"
	"github.com/codingsince1985/checksum"
	"gorm.io/gorm"

	"github.com/baas-project/baas/pkg/fs"
	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// GetMachine GETs any machine in the database based on its MAC address
// Example message: machine/00:11:22:33:44:55:66
// Example response: {"name": "Machine 1",
//                    "Architecture": "x86_64",
//                    "MacAddresses": [{"Mac": "00:11:22:33:44:55:66}]}
func (api_ *API) GetMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mac, ok := vars["mac"]
	if !ok || mac == "" {
		http.Error(w, "invalid mac address", http.StatusBadRequest)
		log.Error("Invalid mac address given")
		return
	}

	machine, err := api_.store.GetMachineByMac(util.MacAddress{Address: mac})
	if err != nil {
		http.Error(w, "couldn't get machine", http.StatusInternalServerError)
		log.Errorf("get machine by mac: %v", err)
		return
	}

	e := json.NewEncoder(w)
	_ = e.Encode(machine)
}

// GetMachines fetches all the machines from the database using a GET request
// Example request: machines
// Example response: {"name": "Machine 1",
//                    "Architecture": "x86_64",
//                    "MacAddresses": [{"Mac": "00:11:22:33:44:55:66}]}
func (api_ *API) GetMachines(w http.ResponseWriter, _ *http.Request) {
	machines, err := api_.store.GetMachines()
	if err != nil {
		http.Error(w, "couldn't get machines", http.StatusInternalServerError)
		log.Errorf("get machines: %v", err)
		return
	}

	e := json.NewEncoder(w)
	_ = e.Encode(machines)
}

// UpdateMachine updates (or adds) the machine to the database.
//
// Example of a JSON message:
//     {
//        "name": "Hello World",
//        "Architecture": "x86_64",
//        "Managed": true,
//        "DiskUUIDs": null,
//        "MacAddresses": [{
//            "Mac": "52:54:00:d9:71:15",
//            "MachineModelID": 12
//        }]
//     }
//
func (api_ *API) UpdateMachine(w http.ResponseWriter, r *http.Request) {
	var machine model.MachineModel
	err := json.NewDecoder(r.Body).Decode(&machine)

	if err != nil {
		http.Error(w, "invalid machine given", http.StatusBadRequest)
		log.Errorf("Invalid machine given: %v", err)
		return
	}

	err = api_.store.UpdateMachine(&machine)
	if err != nil {
		http.Error(w, "couldn't update machine", http.StatusInternalServerError)
		log.Errorf("get update machine: %v", err)
		return
	}
}

func (api_ *API) CreateMachine(w http.ResponseWriter, r *http.Request) {
	var machine model.MachineModel
	err := json.NewDecoder(r.Body).Decode(&machine)
	if err != nil {
		http.Error(w, "invalid machine given", http.StatusBadRequest)
		log.Errorf("Invalid machine given: %v", err)
		return
	}

	api_.store.CreateMachine(&machine)
	// Generate the UUID and create the entry in the database.
	// We don't actually make an image file yet.
	machineImage, err := images.CreateMachineModel(images.ImageModel{}, machine.MacAddress)
	if err != nil {

	}

	// Fill the machine image with default values, in particular ensure that it
	// is not compressed
	machineImage.UUID = images.ImageUUID(uuid.New().String())
	machineImage.Type = "machine"
	machineImage.DiskCompressionStrategy = images.DiskCompressionStrategyNone
	machineImage.Name = machine.MacAddress.Address

	// Create the actual image together with the first empty version which a user may or may not use.
	err = os.Mkdir(fmt.Sprintf(api_.diskpath+"/%s", machineImage.UUID), os.ModePerm)
	if err != nil {
		http.Error(w, "could not create image", http.StatusInternalServerError)
		log.Errorf("cannot create image directory: %v", err)
		return
	}

	err = machineImage.CreateImageFile(machineImage.Size, api_.diskpath, images.SizeMegabyte)

	if err != nil {
		http.Error(w, "Cannot create the image file", http.StatusInternalServerError)
		log.Errorf("image creation failed: %v", err)
		return
	}

	f, err := OpenImageFile(string(machineImage.UUID), "0", api_)

	if err != nil {
		http.Error(w, "couldn't decode image model", http.StatusBadRequest)
		log.Errorf("failed to open the image file: %v", err)
		return
	}

	chk, err := checksum.CRCReader(f)
	if err != nil {
		http.Error(w, "couldn't decode image model", http.StatusBadRequest)
		log.Errorf("Can't generate the checksum: %v", err)
		return
	}

	machineImage.Checksum = chk

	api_.store.CreateImage(&machineImage.ImageModel)
	api_.store.CreateMachineImage(machineImage)

	if err != nil {
		http.Error(w, "couldn't create image model", http.StatusInternalServerError)
		log.Errorf("decode create model: %v", err)
		return
	}

	// Create an EXT4 partition scheme on disk
	path := fmt.Sprintf(api_.diskpath+"/%s/0.img", machineImage.UUID)
	cmd := exec.Command("mkfs.ext4", path)
	err = cmd.Run()
	if err != nil {
		http.Error(w, "Cannot create the image file", http.StatusInternalServerError)
		log.Fatalf("Creating ext4 partition failed: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&machineImage)

}

// UploadDiskImage allows the management os to upload disk images
func (api_ *API) UploadDiskImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["uuid"]
	if !ok || id == "" {
		http.Error(w, "Invalid uuid", http.StatusBadRequest)
		log.Error("Invalid uuid given")
		return
	}

	mac, ok := vars["mac"]
	if !ok || mac == "" {
		http.Error(w, "Invalid mac address", http.StatusBadRequest)
		log.Error("Invalid mac address given")
		return
	}

	path := fmt.Sprintf("%s/%s", api_.diskpath, id)
	temppath := fmt.Sprintf("%s.%s.tmp", path, uuid.New().String())

	f, err := os.OpenFile(temppath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		http.NotFound(w, r)
		log.Errorf("failed to open/create disk image (%v)", err)
		return
	}

	err = fs.CopyStream(r.Body, f)
	if err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		log.Errorf("failed to write file (%v)", err)
		return
	}

	err = os.Rename(temppath, path)
	if err != nil {
		http.Error(w, "failed to move file", http.StatusInternalServerError)
		log.Errorf("failed to move file (%v)", err)
		return
	}
}

// DownloadDiskImage provides disk images for the management os to download
func (api_ *API) DownloadDiskImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["uuid"]
	if !ok || id == "" {
		http.Error(w, "Invalid uuid", http.StatusBadRequest)
		log.Error("Invalid uuid given")
		return
	}

	mac, ok := vars["mac"]
	if !ok || mac == "" {
		http.Error(w, "Invalid mac address", http.StatusBadRequest)
		log.Error("Invalid mac address given")
		return
	}

	f, err := os.OpenFile(fmt.Sprintf("%s/%s", api_.diskpath, id), syscall.O_RDONLY, os.ModePerm)
	if err != nil {
		http.NotFound(w, r)
		log.Errorf("failed to read disk image (%v)", err)
		return
	}

	r.Header.Set("Content-Type", "application/octet-stream")

	err = fs.CopyStream(f, w)
	if err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		log.Errorf("failed to write file (%v)", err)
		return
	}
}

// BootInform handles all incoming boot inform requests
func (api_ *API) BootInform(w http.ResponseWriter, r *http.Request) {
	// First we fetch the id associated of the
	vars := mux.Vars(r)
	mac, ok := vars["mac"]

	if !ok || mac == "" {
		http.Error(w, "mac address is not found", http.StatusBadRequest)
		log.Errorf("mac not provided")
		return
	}

	machine, err := api_.store.GetMachineByMac(util.MacAddress{Address: mac})

	if err != nil {
		http.Error(w, "Cannot find the machine in the database", http.StatusBadRequest)
		log.Errorf("Machine not found")
		return
	}

	log.Debug("Received BootInform request, serving Reprovisioning information")

	// Get the next boot configuration based on a FIFO queue.
	bootInfo, err := api_.store.GetNextBootSetup(machine.ID)

	if err == gorm.ErrRecordNotFound {
		http.Error(w, "No boot setup found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Error with finding boot setup", http.StatusBadRequest)
		log.Errorf("Database error: %v", err)
		return
	}

	resp, err := api_.store.GetImageSetup(string(bootInfo.SetupUUID))

	if err != nil {
		http.Error(w, "Failed to get the next boot setup", http.StatusInternalServerError)
		log.Errorf("Failed to get the image setup: %v", err)
		return
	}

	image, err := api_.store.GetMachineImageByMac(util.MacAddress{Address: mac})

	if err != nil {
		http.Error(w, "Failed to get the next boot setup", http.StatusBadRequest)
		log.Errorf("Failed to get the machine image: %v", err)
		return
	}

	// Add the machine image to the list
	resp.Images = append(resp.Images, images.ImageFrozen{
		Image: image.ImageModel,
		Version: images.Version{
			Version: 0,
		},
	})

	if err != nil {
		http.Error(w, "Failed to get the next boot setup", http.StatusBadRequest)
		log.Errorf("Failed to fetch image setup: %v", err)
		return
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Errorf("Error while serialising json: %v", err)
		http.Error(w, "Error while serialising response json", http.StatusInternalServerError)
		return
	}

	r.Header.Set("content-type", "application/json")

}

// SetBootSetup adds an image to the schedule to be flashed onto the machine
// Example request: POST machine/52:54:00:d9:71:93/boot
// Example body: {"Version": 1636116090, "ImageUUID": "74368cec-7903-4233-87b7-564195619dce", "update": true}
// Example response: {
//   "MachineModelID": 1,
//   "Version": 1636116090,
//   "ImageUUID": "74368cec-7903-4233-87b7-564195619dce",
//   "Update": true}
func (api_ *API) SetBootSetup(w http.ResponseWriter, r *http.Request) {
	// First we fetch the id associated of the
	vars := mux.Vars(r)
	mac, ok := vars["mac"]

	if !ok || mac == "" {
		http.Error(w, "mac address is not found", http.StatusBadRequest)
		log.Errorf("mac not provided")
		return
	}

	machine, err := api_.store.GetMachineByMac(util.MacAddress{Address: mac})

	if err != nil {
		http.Error(w, "Cannot find the machine in the database", http.StatusBadRequest)
		log.Errorf("Machine not found")
		return
	}

	// Fetch the data from the body
	var bootSetup model.BootSetup
	err = json.NewDecoder(r.Body).Decode(&bootSetup)

	if err != nil {
		http.Error(w, "Invalid machine given", http.StatusBadRequest)
		log.Errorf("Invalid machine given: %v", err)
		return
	}

	util.PrettyPrintStruct(bootSetup)

	bootSetup.MachineModelID = machine.ID
	err = api_.store.AddBootSetupToMachine(&bootSetup)

	if err != nil {
		http.Error(w, "cannot add the bootsetup to the machine", http.StatusBadRequest)
		log.Errorf("Cannot add boot info: %v", err)
		return
	}

	e := json.NewEncoder(w)
	_ = e.Encode(bootSetup)
}

// RegisterMachineHandlers sets the metadata for each of the routes and registers them to the global handler
func (api_ *API) RegisterMachineHandlers() {
	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api_.GetMachine,
		Method:      http.MethodGet,
		Description: "Gets a machine from the database",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machines",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api_.GetMachines,
		Method:      http.MethodGet,
		Description: "Gets all the machines from the database",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine",
		Permissions: []model.UserRole{model.Admin},
		UserAllowed: false,
		Handler:     api_.UpdateMachine,
		Method:      http.MethodPut,
		Description: "Updates a machine",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine",
		Permissions: []model.UserRole{model.Admin},
		UserAllowed: false,
		Handler:     api_.CreateMachine,
		Method:      http.MethodPost,
		Description: "Creates a new machine",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/disk/{uuid}",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api_.UploadDiskImage,
		Method:      http.MethodPost,
		Description: "Uploads the image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/disk/{uuid}",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: true,
		Handler:     api_.DownloadDiskImage,
		Method:      http.MethodGet,
		Description: "Downloads the disk image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/boot",
		Permissions: []model.UserRole{model.Moderator, model.Admin},
		UserAllowed: false,
		Handler:     api_.BootInform,
		Method:      http.MethodGet,
		Description: "Gets the next configuration a machine is going to boot into",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/boot",
		Permissions: []model.UserRole{model.User, model.Moderator, model.Admin},
		UserAllowed: false,
		Handler:     api_.SetBootSetup,
		Method:      http.MethodPost,
		Description: "Adds a boot configuration to the queue",
	})
}
