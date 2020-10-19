package httpserver

import (
	"baas/control_server/machines"
	"log"
	"net/http"
)

type TestHandler struct {
	machineStore machines.MachineStore
}

func (t TestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
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