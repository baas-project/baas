package api

import (
	"encoding/json"
	"fmt"
	pkgapi "github.com/baas-project/baas/pkg/api"
	"github.com/baas-project/baas/pkg/fs"
	"github.com/baas-project/baas/pkg/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"syscall"
)

func (api *Api) GetMachine(w http.ResponseWriter, r *http.Request) {
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

func (api *Api) GetMachines(w http.ResponseWriter, r *http.Request) {
	machines, err := api.store.GetMachines()
	if err != nil {
		http.Error(w, "couldn't get machines", http.StatusInternalServerError)
		log.Errorf("get machines: %v", err)
		return
	}

	e := json.NewEncoder(w)
	_ = e.Encode(machines)
}

func (api *Api) UpdateMachine(w http.ResponseWriter, r *http.Request) {
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
func (api *Api) UploadDiskImage(w http.ResponseWriter, r *http.Request) {
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
func (api *Api) DownloadDiskImage(w http.ResponseWriter, r *http.Request) {
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

// BootInform handles all incoming boot inform requests
func (api *Api) BootInform(w http.ResponseWriter, r *http.Request) {
	var bootInform pkgapi.BootInformRequest

	if err := json.NewDecoder(r.Body).Decode(&bootInform); err != nil {
		log.Errorf("Error while parsing json: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Debug("Received BootInform request, serving Reprovisioning information")

	// handle things based on bootinform

	// Request data from database for what to do with this machine
	uuid1 := "alpineresult.iso"
	uuid2 := "alpine.iso"
	location := "/dev/sda"

	// Prepare response
	resp := pkgapi.ReprovisioningInfo{
		Prev: model.MachineSetup{
			Ephemeral: false,
			Disks: []model.DiskMappingModel{
				{
					Uuid: uuid1,
					Image: model.DiskImage{
						DiskType:             model.DiskTypeRaw,
						DiskTransferStrategy: model.DiskTransferStrategyHTTP,
						//DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
						Location: location,
					},
				},
			},
		},
		Next: model.MachineSetup{
			Ephemeral: false,
			Disks: []model.DiskMappingModel{
				{
					Uuid: uuid2,
					Image: model.DiskImage{
						DiskType:             model.DiskTypeRaw,
						DiskTransferStrategy: model.DiskTransferStrategyHTTP,
						//DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
						Location: location,
					},
				},
			},
		},
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Errorf("Error while serialising json: %v", err)
		http.Error(w, "Error while serialising response json", http.StatusInternalServerError)
		return
	}

	r.Header.Set("content-type", "application/json")
}
