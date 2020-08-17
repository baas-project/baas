package machines

import "net"

type Machine struct {
	MacAddress         string
	LastKnownIpAddress net.IPAddr
	Architecture       SystemArchitecture
	Info               *MachineInfo
}

type MachineInfo struct {
	LastBootedManagementOs bool
}
