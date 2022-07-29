// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package database defines the interface to interact with the database.
package database

import (
	"github.com/baas-project/baas/pkg/images"
	"github.com/baas-project/baas/pkg/model"
	"github.com/baas-project/baas/pkg/util"
)

// Store defines the functions which should be exported by any concrete database implementation
type Store interface {

	// GetMachineByMac retrieves a machine based on its mac address.
	GetMachineByMac(mac util.MacAddress) (*model.MachineModel, error)
	GetMachineImageByMac(mac util.MacAddress) (*images.MachineImageModel, error)

	// GetMachines returns a list of all machines in the database
	GetMachines() ([]model.MachineModel, error)
	CreateMachine(machine *model.MachineModel) error

	// UpdateMachine changes the value of a machine based.
	// The mac address is used as key.
	UpdateMachine(machine *model.MachineModel) error
	AddBootSetupToMachine(bootSetup *model.BootSetup) error
	GetNextBootSetup(machineMAC string) (model.BootSetup, error)
	DeleteMachine(machine *model.MachineModel) error

	GetUserByUsername(name string) (*model.UserModel, error)
	GetUserByID(id uint) (*model.UserModel, error)
	GetUsers() ([]model.UserModel, error)
	CreateUser(user *model.UserModel) error

	GetImageByUUID(uuid images.ImageUUID) (*images.ImageModel, error)
	GetImagesByUsername(username string) ([]images.ImageModel, error)
	GetImagesByNameAndUsername(name string, username string) ([]images.ImageModel, error)
	CreateImage(image *images.ImageModel)
	DeleteImage(image *images.ImageModel) error
	UpdateImage(image *images.ImageModel) error
	CreateNewImageVersion(version images.Version)

	// You could use weird Go polymorphisms here, but I guess I will just copy and paste code
	CreateMachineImage(image *images.MachineImageModel)
	CreateImageSetup(username string, image *images.ImageSetup) error
	AddImageToImageSetup(setup *images.ImageSetup, image *images.ImageModel, version images.Version, update bool)
	FindImageSetupsByUsername(username string) (*[]images.ImageSetup, error)
	GetImageSetup(imageSetup string) (images.ImageSetup, error)
	GetImageSetups(username string) (*[]images.ImageSetup, error)

	ModifyImageSetup(imageSetup *images.ImageSetup) error
	DeleteImageSetup(imageSetup *images.ImageSetup) error
	RemoveImageFromImageSetup(setup *images.ImageSetup, image *images.ImageModel, version images.Version, update bool) error
}
