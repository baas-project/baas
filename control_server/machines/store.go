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

import (
	"fmt"
	"sync"
)

// MachineStore provides functions to operate on machine stores.
// Machine stores store machine information (mac address, architecture etc)
// which can be used to provide the right machines with the right data to boot
// into the management OS, and to provide the management OS with disk images and other data.
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

// InMemoryMachineStore stores machine information (mac address, architecture etc) in-memory.
type InMemoryMachineStore struct {
	lock     sync.Mutex
	machines map[string]Machine
}

// InMemoryStore creates a new in-memory machine store.
func InMemoryStore() *InMemoryMachineStore {
	return &InMemoryMachineStore{
		machines: make(map[string]Machine),
	}
}

// GetMachine retrieves a machine in the machine store for the InMemoryMachineStore.
func (i *InMemoryMachineStore) GetMachine(macAddress string) (Machine, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	machine, ok := i.machines[macAddress]
	if !ok {
		return Machine{}, fmt.Errorf("machine with mac address %v not found", macAddress)
	}

	return machine, nil
}

// UpdateMachine updates the machine in the machine store for the InMemoryMachineStore.
func (i *InMemoryMachineStore) UpdateMachine(machine Machine) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	macAddress := machine.MacAddress

	i.machines[macAddress] = machine

	return nil
}
