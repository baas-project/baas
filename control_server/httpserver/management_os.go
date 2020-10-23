package httpserver

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
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

// BootInform handles all incoming boot inform requests
func (m *ManagementOsHandler) BootInform(w http.ResponseWriter, r *http.Request) {
	var bootInform api.BootInformRequest

	if err := json.NewDecoder(r.Body).Decode(&bootInform); err != nil {
		log.Errorf("Error while parsing json: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Debug("Received BootInform request, serving Reprovisioning information")

	// handle things based on bootinform

	// Request data from database for what to do with this machine
	uuid := "uuid"
	location := "/dev/sda"

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
		log.Errorf("Error while serializing json: %v", err)
		http.Error(w, "Error while serialising response json", http.StatusInternalServerError)
		return
	}

	r.Header.Set("content-type", "application/json")
}
