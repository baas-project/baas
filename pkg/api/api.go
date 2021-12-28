// Package api defines structures which are transferred over the network
package api

// Port is the port on which the control server listens
const Port int = 4848

// BootInformRequest is the data which the machine (client) sends to the control server on initial boot
type BootInformRequest struct {
}
