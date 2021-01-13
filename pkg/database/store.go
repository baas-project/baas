package database

import "github.com/baas-project/baas/pkg/model"

type Store interface {

	// GetMachineByMac retrieves a machine based on it's mac address.
	GetMachineByMac(mac string) (*model.MachineModel, error)

	// GetMachines returns a list of all machines in the database
	GetMachines() ([]model.MachineModel, error)

	// UpdateMachine changes the value of a machine based.
	// The mac address is used as key.
	UpdateMachine(machine *model.MachineModel) error

	//
	GetUserByName(name string) (*model.UserModel, error)

	//
	GetUsers() ([]model.UserModel, error)

	//
	CreateUser(user *model.UserModel) error

	GetImageByUUID(uuid model.ImageUUID) (*model.ImageModel, error)

	CreateImage(username string, image model.ImageModel) error
}
