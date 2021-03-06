// Package api defines structures which are transferred over the network
package api

import "github.com/baas-project/baas/pkg/model"

// Port is the port on which the control server listens
const Port int = 4848

// BootInformRequest is the data which the machine (client) sends to the control server on initial boot
type BootInformRequest struct {
}

// ReprovisioningInfo is to inform the management OS of the previous and next machine state and session
type ReprovisioningInfo struct {
	Prev model.MachineSetup
	Next model.MachineSetup
}
