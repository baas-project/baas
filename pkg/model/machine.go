// Package model defines the entities which are used inside the database.
package model

import (
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

	// We want to store the version of the disk
	Version uint64 `gorm:"foreignKey:version"`

	// The image and the disk mapping for this image
	ImageUUID string `gorm:"foreignKey:UUID"`

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

	// Isn't this going to be one in most, if not all, cases?
	MacAddress uint64
}
