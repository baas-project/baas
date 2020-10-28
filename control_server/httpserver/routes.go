package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/gorilla/mux"

	"baas/pkg/fs"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"

	"baas/control_server/machines"
	"baas/pkg/api"
	"baas/pkg/model"
)

// Routes is a struct on which functions are defined that respond to requests
// from the management OS. This struct holds state necessary for the request handlers.
type Routes struct {
	machineStore machines.MachineStore
	diskpath     string
}

// NewRoutes creates a new Routes struct.
func NewRoutes(store machines.MachineStore, diskpath string) *Routes {
	return &Routes{
		machineStore: store,
		diskpath:     diskpath,
	}
}

// BootInform handles all incoming boot inform requests
func (routes *Routes) BootInform(w http.ResponseWriter, r *http.Request) {
	var bootInform api.BootInformRequest

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
	resp := api.ReprovisioningInfo{
		Prev: model.MachineSetup{
			Ephemeral: false,
			Disks: map[model.DiskUUID]model.DiskImage{
				uuid1: {
					DiskType:             model.DiskTypeRaw,
					DiskTransferStrategy: model.DiskTransferStrategyHTTP,
					//DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
					Location: location,
				},
			},
		},
		Next: model.MachineSetup{
			Ephemeral: false,
			Disks: map[model.DiskUUID]model.DiskImage{
				uuid2: {
					DiskType:             model.DiskTypeRaw,
					DiskTransferStrategy: model.DiskTransferStrategyHTTP,
					//DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
					Location: location,
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

// UploadDiskImage allows the management os to upload disk images
func (routes *Routes) UploadDiskImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["uuid"]
	if !ok || id == "" {
		http.Error(w, "Invalid uuid", http.StatusBadRequest)
		log.Error("Invalid uuid given")
		return
	}

	path := fmt.Sprintf("%s/%s", routes.diskpath, id)
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
func (routes *Routes) DownloadDiskImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["uuid"]
	if !ok || id == "" {
		http.Error(w, "Invalid uuid", http.StatusBadRequest)
		log.Error("Invalid uuid given")
		return
	}

	f, err := os.OpenFile(fmt.Sprintf("%s/%s", routes.diskpath, id), syscall.O_RDONLY, os.ModePerm)
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
