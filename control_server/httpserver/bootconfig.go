package httpserver

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"baas/control_server/machines"
)

// BootConfigHandler serves boot configurations (from the ServeBootConfigurations function)
// to pixiecore, so it knows what os to boot.
type BootConfigHandler struct {
	// A different configuration needs to be served depending on the system architecture.
	// The machine store stores this information.
	MachineStore machines.MachineStore
}

type bootConfigResponse struct {
	// Kernel to boot.
	Kernel string `json:"kernel"`

	// Initramfs to boot.
	Initramfs []string `json:"initrd"`

	// Message to print before booting.
	Message string `json:"message"`

	// Kernel command line parameters.
	Cmdline string `json:"cmdline"`
}

func getBootConfig(arch machines.SystemArchitecture) *bootConfigResponse {
	switch arch {
	case machines.X8664:
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

// ServeBootConfigurations actually responds to requests from pixiecore.
func (p BootConfigHandler) ServeBootConfigurations(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	mac := vars["mac"]

	addr, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		log.Printf("An error occurred: %v", err)
		writer.WriteHeader(500)
		return
	}

	log.Printf("Serving boot config for %v at ip: %v", mac, addr)

	m, err := p.MachineStore.GetMachine(mac)
	if err != nil {
		log.Printf("An error occurred: %v", err)
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
