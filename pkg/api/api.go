// Package api defines structures which are transferred over the network
package api

import "baas/pkg/model"

const Port int = 4848

// BootInformRequest is the data which the machine (client) sends to the control server on initial boot
type BootInformRequest struct {
}

// ReprovisioningInfo is to inform the management OS of the previous and next machine state and session
type ReprovisioningInfo struct {
	Prev model.MachineSetup
	Next model.MachineSetup
}
