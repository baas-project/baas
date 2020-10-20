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
func (t ManagementOsHandler) RespondToTestPostRequest(writer http.ResponseWriter, request *http.Request) {
	var contents []byte

	_, err := request.Body.Read(contents)
	if err != nil {
		log.Printf("An error occurred: %v", err)
		writer.WriteHeader(500)
		return
	}

	log.Printf("%v", contents)

	writer.WriteHeader(200)
}
