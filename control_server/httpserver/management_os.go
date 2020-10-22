package httpserver

import (
	"log"
	"net/http"

	"baas/control_server/machines"
)

// ManagementOsHandler is a struct on which functions are defined that respond to requests
// from the management OS. This struct holds state necessary for the request handlers.
type ManagementOsHandler struct {
	machineStore machines.MachineStore
}

// RespondToTestPostRequest is temporary to demonstrate communication.
func (t *ManagementOsHandler) RespondToTestPostRequest(w http.ResponseWriter, r *http.Request) {
	var contents []byte

	_, err := r.Body.Read(contents)
	if err != nil {
		log.Printf("An error occurred: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("%v", contents)
}
