package httpserver

import (
	"baas/control_server/machines"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
)

type BootConfigHandler struct {
	MachineStore machines.MachineStore
}


type bootConfigResponse struct {
	// Kernel to boot.
	Kernel    string   `json:"kernel"`

	// Initramfs to boot.
	Initramfs []string `json:"initrd"`

	// Message to print before booting.
	Message   string   `json:"message"`

	// Kernel command line parameters.
	Cmdline   string   `json:"cmdline"`
}

func getBootConfig(arch machines.SystemArchitecture) *bootConfigResponse {
	switch arch {
	case machines.X86_64:
		return &bootConfigResponse{
			Kernel: "http://localhost:4848/static/vmlinuz",
			Initramfs: []string{
				"http://localhost:4848/static/initramfs",
			},
			Message: "Booting into X86 management kernel.",
			Cmdline: "",
		}
	case machines.Arm64:
		fallthrough
	case machines.Unknown:
		fallthrough
	default:
		return nil
	}
}

func (p BootConfigHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)
	mac := vars["mac"]

	addr, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		log.Printf("An error occured: %v", err)
		writer.WriteHeader(500)
		return
	}

	log.Printf("Serving boot config for %v at ip: %v", mac, addr)

	m, err := p.MachineStore.GetMachine(mac)
	if err != nil {
		log.Printf("An error occured: %v", err)
		writer.WriteHeader(500)
		return
	}

	resp := getBootConfig(m.Architecture)
	if resp == nil {
		log.Printf("Couldn't find appropriate bootconfig for this machine")
		writer.WriteHeader(404)
		return
	}

	log.Printf("Sending boot config %v", resp)

	if err := json.NewEncoder(writer).Encode(&resp); err != nil {
		panic(err)
	}
}
