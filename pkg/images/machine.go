// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package images

import (
	"github.com/baas-project/baas/pkg/util"
)

type FilesystemType string

const (
	FileSystemTypeFAT32 FilesystemType = "fat32"
	FilesystemEXT4                     = "ext4"
)

type MachineImageModel struct {
	ImageModel
	MachineMAC util.MacAddress
	Filesystem FilesystemType
	Size       uint // filesize in MiB
}

func CreateMachineModel(image ImageModel, mac util.MacAddress) (*MachineImageModel, error) {
	machineImage := MachineImageModel{ImageModel: image,
		MachineMAC: mac,
		Size:       128,
		Filesystem: FilesystemEXT4,
	}

	return &machineImage, nil
}
