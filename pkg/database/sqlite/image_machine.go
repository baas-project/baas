// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import (
	"github.com/baas-project/baas/pkg/model/images"
	"github.com/baas-project/baas/pkg/util"
)

// CreateMachineImage creates the image entity in the database and adds the first version to it.
func (s Store) CreateMachineImage(image *images.MachineImageModel) {
	s.DB.Create(image)
}

// GetMachineImageByMac fetches the image with the versions using mac address of their machine as a key
func (s Store) GetMachineImageByMac(mac util.MacAddress) (*images.MachineImageModel, error) {
	image := images.MachineImageModel{}
	res := s.Where("machine_mac = ?", mac).
		Preload("Versions").
		First(&image)
	return &image, res.Error
}
