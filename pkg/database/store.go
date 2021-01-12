package database

import "github.com/baas-project/baas/pkg/model"

type Store interface {

	// GetMachineByMac retrieves a machine based on it's mac address.
	GetMachineByMac(mac string) (*model.Machine, error)

	// GetMachines returns a list of all machines in the database
	GetMachines() ([]model.Machine, error)

	// UpdateMachineByMac changes the value of a machine based.
	// A mac address is used as key. Mac may be the empty string.
	// In that case the mac address of the given machine is used as key.
	//
	// The machine however, must contain a MacAddress to be used as key.
	UpdateMachineByMac(machine model.Machine, mac string) error

}

