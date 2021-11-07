// Package database defines the interface to interact with the database.
package database

import "github.com/baas-project/baas/pkg/model"

// Store defines the functions which should be exported by any concrete database implementation
type Store interface {

	// GetMachineByMac retrieves a machine based on its mac address.
	GetMachineByMac(mac string) (*model.MachineModel, error)

	// GetMachines returns a list of all machines in the database
	GetMachines() ([]model.MachineModel, error)

	// UpdateMachine changes the value of a machine based.
	// The mac address is used as key.
	UpdateMachine(machine *model.MachineModel) error
	AddBootSetupToMachine(bootSetup *model.BootSetup) error
	GetNextBootSetup(machineID uint) (model.BootSetup, error)

	GetUserByName(name string) (*model.UserModel, error)
	GetUserByID(id uint) (*model.UserModel, error)
	GetUsers() ([]model.UserModel, error)
	CreateUser(user *model.UserModel) error

	GetImageByUUID(uuid model.ImageUUID) (*model.ImageModel, error)
	GetImagesByUsername(username string) ([]model.ImageModel, error)
	GetImagesByNameAndUsername(name string, username string) ([]model.ImageModel, error)
	CreateImage(username string, image *model.ImageModel) error
	CreateNewImageVersion(version model.Version)
}
