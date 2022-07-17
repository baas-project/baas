// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package model defines the entities which are used inside the database.
package model

import (
	"github.com/baas-project/baas/pkg/images"
	"github.com/baas-project/baas/pkg/util"
	"gorm.io/gorm"
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

// BootSetup stores what the next boot for the machine should look like.
// It functions somewhat like a queue where it removes the first value from the database.
type BootSetup struct {
	gorm.Model `json:"-"`

	// Store the machine id
	MachineModelID uint `gorm:"foreignKey:ID"`

	// Store the setup that should be loaded onto the machine
	SetupUUID images.ImageUUID

	// Should the image changes be uploaded to the server?
	Update bool
}

// MachineModel stores information intrinsic to a machine. Used together with the MachineStore.
type MachineModel struct {
	gorm.Model `json:"-"`

	// General Info
	Name         string
	Architecture SystemArchitecture

	// Managed indicates that a machine should be managed by BAAS (if false baas will not touch the machine in any way)
	Managed bool

	// MacAddress is the mac address associated with this machine
	MacAddress util.MacAddress
}
