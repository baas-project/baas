package machines

import "net"

// Machine stores information intrinsic to a machine. Used together with the MachineStore.
type Machine struct {
	MacAddress         string
	LastKnownIPAddress net.IPAddr
	Architecture       SystemArchitecture
	Info               *MachineInfo
}

// MachineInfo stores information about the state of a machine. Used together with the Machine struct.
type MachineInfo struct {
	LastBootedManagementOs bool
}
