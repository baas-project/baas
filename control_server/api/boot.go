// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/baas-project/baas/pkg/model/machine"

	"github.com/baas-project/baas/pkg/util"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

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

func getBootConfig(arch machine.SystemArchitecture) *bootConfigResponse {
	var bootConfig bootConfigResponse
	arch = machine.SystemArchitecture(strings.ToLower(string(arch)))

	switch arch {
	case machine.X86_64:
		bootConfig = bootConfigResponse{
			Kernel: "http://localhost:4848/static/vmlinuz",
			Initramfs: []string{
				"http://localhost:4848/static/initramfs",
			},
			Message: "Booting into X86 management kernel.",
			Cmdline: "root=sr0",
		}
	case machine.Arm64:
		log.Warn("Received request to boot an ARM64 machine, which has not been implemented yet.")
	default:
		log.Warn("Architecture is not supported")
	}

	return &bootConfig
}

// ServeBootConfigurations actually responds to requests from pixiecore.
func (api_ *API) ServeBootConfigurations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mac := vars["mac"]

	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Errorf("Error while trying to get remote ip address: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Serving boot config for %v at ip: %v", mac, addr)

	m, err := api_.store.GetMachineByMac(util.MacAddress{Address: mac})
	if err != nil {
		log.Errorf("Couldn't find machine in store: %v", err)
		http.Error(w, "Cannot serve the boot configuration", http.StatusNotFound)
		return
	}

	resp := getBootConfig(m.Architecture)
	if resp == nil {
		log.Error("Couldn't find appropriate bootconfig for this machine")
		http.Error(w, "Cannot serve the boot configuration", http.StatusNotFound)
		return
	}

	log.Debugf("Sending boot config %v", resp)

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Errorf("Couldn't write bootconfig to network: %v", err)
		http.Error(w, "Cannot serve the boot configuration", http.StatusInternalServerError)
	}
}
