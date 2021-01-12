package database

import (
	"github.com/baas-project/baas/pkg/model"
)

type InMemoryStore struct {
	machines map[string] model.Machine
}

func (i InMemoryStore) UpdateMachineByMac(machine model.Machine, mac string) error {
	if machine.MacAddress == "" {
		return Error("machine.MacAddress is empty")
	}


	if mac == "" {
		mac = machine.MacAddress
	}

	i.machines[mac] = machine

	return nil
}

func (i InMemoryStore) GetMachineByMac(mac string) (*model.Machine, error) {
	machine, ok := i.machines[mac]
	if !ok {
		return nil, NotFound
	}

	return &machine, nil
}

func NewInMemoryStore() Store {
	return InMemoryStore {
		machines: map[string]model.Machine{},
	}
}
