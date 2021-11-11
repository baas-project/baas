package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"syscall"

	pkgapi "github.com/baas-project/baas/pkg/api"
	"github.com/baas-project/baas/pkg/fs"
	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetMachine GETs any machine in the database based on its MAC address
// Example message: machine/00:11:22:33:44:55:66
// Example response: {"name": "Machine 1",
//                    "Architecture": "x86_64",
//                    "MacAddresses": [{"Mac": "00:11:22:33:44:55:66}]}
func (api *API) GetMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mac, ok := vars["mac"]
	if !ok || mac == "" {
		http.Error(w, "invalid mac address", http.StatusBadRequest)
		log.Error("Invalid mac address given")
		return
	}

	machine, err := api.store.GetMachineByMac(mac)
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
func (api *API) GetMachines(w http.ResponseWriter, _ *http.Request) {
	machines, err := api.store.GetMachines()
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
//        "ShouldReprovision": true,
//        "CurrentSetup": null,
//        "NextSetup": null,
//        "DiskUUIDs": null,
//        "MacAddresses": [{
//            "Mac": "52:54:00:d9:71:15",
//            "MachineModelID": 12
//        }]
//     }
//
func (api *API) UpdateMachine(w http.ResponseWriter, r *http.Request) {
	var machine model.MachineModel
	err := json.NewDecoder(r.Body).Decode(&machine)
	if err != nil {
		http.Error(w, "invalid machine given", http.StatusBadRequest)
		log.Errorf("Invalid machine given: %v", err)
		return
	}

	err = api.store.UpdateMachine(&machine)
	if err != nil {
		http.Error(w, "couldn't update machine", http.StatusInternalServerError)
		log.Errorf("get update machine: %v", err)
		return
	}
}

// UploadDiskImage allows the management os to upload disk images
func (api *API) UploadDiskImage(w http.ResponseWriter, r *http.Request) {
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

	path := fmt.Sprintf("%s/%s", api.diskpath, id)
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
func (api *API) DownloadDiskImage(w http.ResponseWriter, r *http.Request) {
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

	f, err := os.OpenFile(fmt.Sprintf("%s/%s", api.diskpath, id), syscall.O_RDONLY, os.ModePerm)
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

// generateMachineSetup generates the MachineSetup model
func generateMachineSetup(setup model.BootSetup) model.MachineSetup {
	return model.MachineSetup{
		Ephemeral: false,
		Disks: []model.DiskMappingModel{
			{
				UUID:    setup.ImageUUID,
				Version: setup.Version,
				Image: model.DiskImage{
					DiskType:             model.DiskTypeRaw,
					DiskTransferStrategy: model.DiskTransferStrategyHTTP,
					Location:             "/dev/sda",
				},
			},
		},
	}
}

// BootInform handles all incoming boot inform requests
func (api *API) BootInform(w http.ResponseWriter, r *http.Request) {
	// First we fetch the id associated of the
	vars := mux.Vars(r)
	mac, ok := vars["mac"]

	if !ok || mac == "" {
		http.Error(w, "mac address is not found", http.StatusBadRequest)
		log.Errorf("mac not provided")
		return
	}

	machine, err := api.store.GetMachineByMac(mac)

	if err != nil {
		http.Error(w, "Cannot find the machine in the database", http.StatusBadRequest)
		log.Errorf("Machine not found")
		return
	}

	log.Debug("Received BootInform request, serving Reprovisioning information")

	// Get the next boot configuration based on a FIFO queue.
	bootInfo, err := api.store.GetNextBootSetup(machine.ID)

	if err == gorm.ErrRecordNotFound {
		http.Error(w, "No boot setup found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Error with finding boot setup", http.StatusBadRequest)
		log.Errorf("Database error: %v", err)
		return
	}

	// Use the same table to get the last deleted setup (which is the one running now)
	lastSetup, err := api.store.GetLastDeletedBootSetup(machine.ID)

	if err != gorm.ErrRecordNotFound && err != nil {
		http.Error(w, "Error with fetching the boot history", http.StatusBadRequest)
		return
	}

	var prev model.MachineSetup
	if err != gorm.ErrRecordNotFound && lastSetup.Update {
		prev = generateMachineSetup(lastSetup)
	}

	resp := pkgapi.ReprovisioningInfo{
		Prev: prev,
		Next: generateMachineSetup(bootInfo),
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
func (api *API) SetBootSetup(w http.ResponseWriter, r *http.Request) {
	// First we fetch the id associated of the
	vars := mux.Vars(r)
	mac, ok := vars["mac"]

	if !ok || mac == "" {
		http.Error(w, "mac address is not found", http.StatusBadRequest)
		log.Errorf("mac not provided")
		return
	}

	machine, err := api.store.GetMachineByMac(mac)

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

	// pkgapi.PrettyPrintStruct(bootSetup)

	bootSetup.MachineModelID = machine.ID
	err = api.store.AddBootSetupToMachine(&bootSetup)

	if err != nil {
		http.Error(w, "cannot add the bootsetup to the machine", http.StatusBadRequest)
		log.Errorf("Cannot add boot info: %v", err)
		return
	}

	e := json.NewEncoder(w)
	_ = e.Encode(bootSetup)
}
