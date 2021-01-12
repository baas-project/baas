package database

import (
	"github.com/baas-project/baas/pkg/model"
)

type InMemoryStore struct {
	machines map[string] model.Machine
	users map[string]model.User
}

func (i InMemoryStore) GetMachines() ([]model.Machine, error) {
	var res []model.Machine
	for _, machine := range i.machines {
		res = append(res, machine)
	}

	return res, nil
}

func (i InMemoryStore) GetUserByName(name string) (*model.User, error) {
	user, ok := i.users[name]
	if !ok {
		return nil, NotFound
	}

	return &user, nil
}

func (i InMemoryStore) UpdateUser(user *model.User) error {
	i.users[user.Name] = *user
	return nil
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
