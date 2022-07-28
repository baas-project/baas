// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package images

import (
	"github.com/baas-project/baas/pkg/util"
)

// FilesystemType is the type of filesystem that is used by the image
type FilesystemType string

const (
	// FileSystemTypeFAT32 defines a disk using the universal FAT32 filesystem
	FileSystemTypeFAT32 FilesystemType = "fat32"
	// FileSystemTypeEXT4 defines a disk using the Linux EXT4 filesystem
	FileSystemTypeEXT4 = "ext4"
)

// MachineImageModel is the model defining the image which holds data about the specific machine
type MachineImageModel struct {
	ImageModel
	MachineMAC util.MacAddress `gorm:"foreignKey:Adress;constraint:onUpdate:CASCADE,OnDelete:CASCADE"`
	Filesystem FilesystemType
	Size       uint // filesize in MiB
}

// CreateMachineModel creates a simple machine image for the designated machine
func CreateMachineModel(image ImageModel, mac util.MacAddress) (*MachineImageModel, error) {
	machineImage := MachineImageModel{ImageModel: image,
		MachineMAC: mac,
		Size:       128,
		Filesystem: FileSystemTypeEXT4,
	}

	return &machineImage, nil
}
