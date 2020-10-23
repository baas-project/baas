package httpserver

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
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
	case machines.X86_64:
		return &bootConfigResponse{
			Kernel: "http://localhost:4848/static/vmlinuz",
			Initramfs: []string{
				"http://localhost:4848/static/initramfs",
			},
			Message: "Booting into X86 management kernel.",
			Cmdline: "root=sr0",
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
func (p *BootConfigHandler) ServeBootConfigurations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mac := vars["mac"]

	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Errorf("Error while trying to get remote ip address: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Serving boot config for %v at ip: %v", mac, addr)

	m, err := p.MachineStore.GetMachine(mac)
	if err != nil {
		log.Errorf("Couldn't find machine in store: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := getBootConfig(m.Architecture)
	if resp == nil {
		log.Errorf("Couldn't find appropriate bootconfig for this machine")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Debugf("Sending boot config %v", resp)

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Errorf("Couldn't write bootconfig to network")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
