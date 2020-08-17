// Package machines defines a series of functions and types to identify
// and work with machines.
//
// Each machine will uniquely identify itself my it's mac address.
// When a machine requests to boot, and we know some information about
// the machine, we can send it the correct disk images. The first time
// a machine boots (and we don't have any information about it), the
// management os is booted. (TODO: what happens on arm machines where the
// TODO: management kernel needs to be different?).
// It will establish the capabilities of the system (like how many disks
// it has and how large, etc. This is saved in the store. The next time
// the machine boots, this can be used to verify image sent to it. If a
package machines

type MachineStore interface {
	// GetMachine Gets the machine identified by this mac address.
	// Returns a new Machine struct with the requested mac address in it
	// when the machine was not found.
	//
	// The error return type is for data stores which may error when
	// looking up values.
	GetMachine(macAddress string) (Machine, error)

	// UpdateMachine updates the machine identified by it's mac address.
	UpdateMachine(machine Machine) error
}

type InMemoryMachineStore struct {
	machines map[string]Machine
}

func (i InMemoryMachineStore) GetMachine(macAddress string) (Machine, error) {
	machine, ok := i.machines[macAddress]
	if !ok {
		return Machine{
			MacAddress: macAddress,
		}, nil
	}

	return machine, nil
}

func (i InMemoryMachineStore) UpdateMachine(machine Machine) error {
	macAddress := machine.MacAddress
	i.machines[macAddress] = machine

	return nil
}
