package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	"baas/control_server/machines"
	"baas/pkg/api"
	"baas/pkg/model"
)

// ManagementOsHandler is a struct on which functions are defined that respond to requests
// from the management OS. This struct holds state necessary for the request handlers.
type ManagementOsHandler struct {
	machineStore machines.MachineStore
}

// RespondToTestPostRequest is temporary to demonstrate communication.
func (m *ManagementOsHandler) RespondToTestPostRequest(w http.ResponseWriter, r *http.Request) {
	var contents []byte

	_, err := r.Body.Read(contents)
	if err != nil {
		log.Printf("An error occurred: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("%v", contents)
}

// BootInform handles all incoming boot inform requests
func (m *ManagementOsHandler) BootInform(w http.ResponseWriter, r *http.Request) {
	var bootInform api.BootInformRequest

	if err := json.NewDecoder(r.Body).Decode(&bootInform); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// handle things based on bootinform

	// Request data from database for what to do with this machine
	uuid := "uuid"
	location := "location"

	// Prepare response
	resp := api.ReprovisioningInfo{
		Prev: model.MachineSetup{
			Ephemeral: true,
		},
		Next: model.MachineSetup{
			Ephemeral: true,
			Disks: map[model.DiskUUID]model.DiskImage{
				uuid: {
					DiskType:             model.DiskTypeRaw,
					DiskTransferStrategy: model.DiskTransferStrategyHTTP,
					Location:             location,
				},
			},
		},
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, "Error while serialising response json", http.StatusInternalServerError)
		return
	}

	r.Header.Set("content-type", "application/json")
}
