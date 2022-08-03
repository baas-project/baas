// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package machine declares the entities related to the machines stored in the database
package machine

import (
	"github.com/baas-project/baas/pkg/util"
)

// SystemArchitecture defines constants describing the architecture of machines.
type SystemArchitecture string

const (
	// Arm64 is the 64-bit Arm architecture
	Arm64 SystemArchitecture = "Arm64"
	// X86_64 is the 64-bit x86 architecture
	X86_64 SystemArchitecture = "x86_64" //nolint
	// Unknown is any architecture which baas could not identify.
	Unknown SystemArchitecture = "unknown"
)

// Name gets the name of an architecture as a string. Convenience function,
// but actually does very little as the name is also the value of the constant.
func (id *SystemArchitecture) Name() string {
	return string(*id)
}

// MachineModel stores information intrinsic to a machine. Used together with the MachineStore.
// nolint: golint
type MachineModel struct {
	// General Info
	Name         string `gorm:"unique"`
	Architecture SystemArchitecture

	// Managed indicates that a machine should be managed by BAAS (if false baas will not touch the machine in any way)
	Managed bool

	// MacAddress is the mac address associated with this machine
	MacAddress util.MacAddress `gorm:"embedded;unique;primaryKey"`
	ImageUUID  string
}
