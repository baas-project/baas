// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package images

import (
	"fmt"
	"os"

	model "github.com/baas-project/baas/pkg/model/machine"

	"github.com/baas-project/baas/pkg/util"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// FilesystemType is the type of filesystem that is used by the image
type FilesystemType string

// MachineImageModel is the model defining the image which holds data about the specific machine
type MachineImageModel struct {
	ImageModel
	// Store the machine id
	Machine    model.MachineModel `gorm:"foreignKey:MachineMAC;references:Address;not null;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	MachineMAC string             `gorm:"not null;primaryKey"`

	Size uint // filesize in MiB
}

// CreateMachineImageModel creates a simple machine image for the designated machine
func CreateMachineImageModel(mac util.MacAddress) (*MachineImageModel, error) {
	baseImage := ImageModel{
		Name:                    mac.Address,
		DiskCompressionStrategy: DiskCompressionStrategyNone,
		Type:                    "machine",
		UUID:                    ImageUUID(uuid.New().String()),
		Username:                "System",
		Checksum:                "DEADBEEF",
		ImagePath:               os.Getenv("BAAS_DISK_PATH"),
		Filesystem:              FileSystemTypeEXT4,
	}

	machineImage := MachineImageModel{ImageModel: baseImage,
		Size:       128,
		MachineMAC: mac.Address,
	}

	return &machineImage, nil
}

// BeforeCreate creates the machine image model directory and image file.
func (machineImage *MachineImageModel) BeforeCreate(tx *gorm.DB) (ret error) {
	path := os.Getenv("BAAS_DISK_PATH")
	machineImage.ImageModel.ImagePath = path

	// Create the actual image together with the first empty version which a user may or may not use.
	err := os.Mkdir(fmt.Sprintf(path+"/%s", machineImage.UUID), os.ModePerm)
	if err != nil {
		log.Errorf("cannot create image directory: %v", err)
		return
	}

	machineImage.ImageModel.CreateImageFile(machineImage.Size, SizeMegabyte)
	machineImage.ImageModel.GenerateChecksum()
	machineImage.ImageModel.FormatImage()
	return
}

// AfterDelete removes all the data associated with the machine.
func (machineImage *MachineImageModel) AfterDelete(tx *gorm.DB) (ret error) {
	machineImage.ImageModel.AfterDelete(tx)
	return
}
