package machines

type Machine struct {
	MacAddress string
	Architecture *SystemArchitecture
	Info *MachineInfo
}

type MachineInfo struct {
	LastBootedManagementOs bool
}
