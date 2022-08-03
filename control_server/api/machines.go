// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/baas-project/baas/pkg/model/images"
	machinemodel "github.com/baas-project/baas/pkg/model/machine"
	"github.com/baas-project/baas/pkg/model/user"

	"github.com/baas-project/baas/pkg/util"
	"gorm.io/gorm"

	"github.com/baas-project/baas/pkg/fs"
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

// DeleteMachine Deletes a machine from the database
// Example request: DELETE machine/[mac]
// Example response: Successfully deleted
func (api_ *API) DeleteMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mac, ok := vars["mac"]
	if !ok || mac == "" {
		http.Error(w, "Invalid mac", http.StatusBadRequest)
		log.Error("Invalid mac given")
		return
	}

	machine, err := api_.store.GetMachineByMac(util.MacAddress{Address: mac})
	if err != nil {
		http.Error(w, "Failed to delete machine", http.StatusInternalServerError)
		log.Errorf("Cannot find machine with mac address: %s (%v)", mac, err)
		return
	}

	image, err := api_.store.GetMachineImageByMac(util.MacAddress{Address: mac})

	if err != nil {
		http.Error(w, "Failed to get the next boot setup", http.StatusBadRequest)
		log.Errorf("Failed to get the machine image: %v", err)
		return
	}

	err = api_.store.DeleteMachine(machine)
	if err != nil {
		http.Error(w, "Failed to delete machine", http.StatusInternalServerError)
		log.Errorf("Machine %s deletion failed with error code: %v", mac, err)
		return
	}

	err = os.RemoveAll(fmt.Sprintf(api_.diskpath+"/%s", image.UUID))
	if err != nil {
		http.Error(w, "Failed to delete machine", http.StatusInternalServerError)
		log.Errorf("Machine %s deletion failed with error code: %v", mac, err)
		return
	}

	http.Error(w, "Successfully deleted the machine", http.StatusOK)
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
	var machine machinemodel.MachineModel
	err := json.NewDecoder(r.Body).Decode(&machine)
	util.PrettyPrintStruct(machine)

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

	_ = json.NewEncoder(w).Encode(&machine)
}

// CreateMachine creates the machine in the database and returns a JSON object representing it
func (api_ *API) CreateMachine(w http.ResponseWriter, r *http.Request) {
	var machine machinemodel.MachineModel
	err := json.NewDecoder(r.Body).Decode(&machine)
	if err != nil {
		http.Error(w, "invalid machine given", http.StatusBadRequest)
		log.Errorf("Invalid machine given: %v", err)
		return
	}

	err = api_.store.CreateMachine(&machine)
	if ErrorWrite(w, err, "Cannot create machine") != nil {
		return
	}

	// Generate the UUID and create the entry in the database.
	// We don't actually make an image file yet.
	machineImage, err := images.CreateMachineImageModel(machine.MacAddress)
	machineImage.Name = machine.MacAddress.Address
	machineImage.UUID = images.ImageUUID(uuid.New().String())

	if ErrorWrite(w, err, "Cannot create machine") != nil {
		return
	}

	// api_.store.CreateImage(&machineImage.ImageModel)
	api_.store.CreateMachineImage(machineImage)

	if err != nil {
		http.Error(w, "couldn't create image model", http.StatusInternalServerError)
		log.Errorf("decode create model: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&machine)
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

	mac, ok := vars["mac"]
	if !ok || mac == "" {
		http.Error(w, "Invalid mac address", http.StatusBadRequest)
		log.Error("Invalid mac address given")
		return
	}

	image, err := api_.store.GetMachineImageByMac(util.MacAddress{Address: mac})

	if err != nil {
		http.NotFound(w, r)
		log.Errorf("failed to read disk image (%v)", err)
		return
	}

	f, err := image.OpenImageFile(0)
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
	bootInfo, err := api_.store.GetNextBootSetup(machine.MacAddress.Address)

	if err == gorm.ErrRecordNotFound {
		http.Error(w, "No boot setup found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Error with finding boot setup", http.StatusBadRequest)
		log.Errorf("Database error: %v", err)
		return
	}

	// TODO: Fix foreign key to version
	resp, err := api_.store.GetImageSetup(string(bootInfo.SetupUUID))

	if err != nil {
		http.Error(w, "Failed to get the next boot setup", http.StatusInternalServerError)
		log.Errorf("Failed to get the image setup: %v", err)
		return
	}

	// Circumvents a problem in the foreign key where the version is
	// not properly loaded into struct. This should be fixed.
	for i := range resp.Images {
		version, verr := api_.store.GetVersionByID(resp.Images[i].VersionID)

		if verr != nil {
			http.Error(w, "Failed to get the next boot setup", http.StatusBadRequest)
			log.Errorf("Failed to get the machine image: %v", verr)
			return
		}

		resp.Images[i].Version = *version
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
	var bootSetup images.BootSetup
	err = json.NewDecoder(r.Body).Decode(&bootSetup)

	if err != nil {
		http.Error(w, "Invalid machine given", http.StatusBadRequest)
		log.Errorf("Invalid machine given: %v", err)
		return
	}

	bootSetup.MachineMAC = machine.MacAddress.Address
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
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.GetMachine,
		Method:      http.MethodGet,
		Description: "Gets a machine from the database",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machines",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.GetMachines,
		Method:      http.MethodGet,
		Description: "Gets all the machines from the database",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine",
		Permissions: []user.UserRole{user.Admin},
		UserAllowed: true,
		Handler:     api_.UpdateMachine,
		Method:      http.MethodPut,
		Description: "Updates a machine",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine",
		Permissions: []user.UserRole{user.Admin},
		UserAllowed: true,
		Handler:     api_.CreateMachine,
		Method:      http.MethodPost,
		Description: "Creates a new machine",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}",
		Permissions: []user.UserRole{user.Admin},
		UserAllowed: false,
		Handler:     api_.DeleteMachine,
		Method:      http.MethodDelete,
		Description: "Deletes a machine from the database",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/disk/{uuid}",
		Permissions: []user.UserRole{user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.UploadDiskImage,
		Method:      http.MethodPost,
		Description: "Uploads the image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/image",
		Permissions: []user.UserRole{user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.DownloadDiskImage,
		Method:      http.MethodGet,
		Description: "Downloads the disk image",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/boot",
		Permissions: []user.UserRole{user.Moderator, user.Admin},
		UserAllowed: false,
		Handler:     api_.BootInform,
		Method:      http.MethodGet,
		Description: "Gets the next configuration a machine is going to boot into",
	})

	api_.Routes = append(api_.Routes, Route{
		URI:         "/machine/{mac}/boot",
		Permissions: []user.UserRole{user.User, user.Moderator, user.Admin},
		UserAllowed: true,
		Handler:     api_.SetBootSetup,
		Method:      http.MethodPost,
		Description: "Adds a boot configuration to the queue",
	})
}
