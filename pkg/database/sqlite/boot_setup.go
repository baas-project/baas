// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import "github.com/baas-project/baas/pkg/model"

// AddBootSetupToMachine adds a configuration for booting to the specified machine
func (s Store) AddBootSetupToMachine(bootSetup *model.BootSetup) error {
	return s.Save(bootSetup).Error
}

// GetNextBootSetup fetches the first machine from the database.
func (s Store) GetNextBootSetup(machineMAC string) (model.BootSetup, error) {
	var bootSetup model.BootSetup
	res := s.Table("boot_setups").
		Where("machine_mac = ?", machineMAC).
		First(&bootSetup).
		Delete(&bootSetup)
	return bootSetup, res.Error
}
