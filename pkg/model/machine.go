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

// DiskModel is the layout of a Disk of a MachineModel
type DiskModel struct {
	gorm.Model `json:"-"`

	UUID           DiskUUID
	MachineModelID uint
}

// MacAddress stores an MAC address associated to a machine.
type MacAddress struct {
	gorm.Model `json:"-"`

	Mac            string
	MachineModelID uint
}

// BootSetup stores what the next boot for the machine should look like.
// It functions somewhat like a queue where it removes the first value from the database.
type BootSetup struct {
	gorm.Model `json:"-"`

	// Store the machine id
	MachineModelID uint `gorm:"foreignKey:ID"`

	// We want to store the version of the disk
	Version uint `gorm:"foreignKey:version"`

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

	// DiskUUIDs are the linux by-uuids this machine has
	DiskUUIDs []DiskModel `gorm:"foreignKey:ID"`

	// MAC addresses associated to the machine
	// Isn't this going to be one in most, if not all, cases?
	MacAddresses []MacAddress
}

// DiskMappingModel stores the images and the target device for an DiskImage.
type DiskMappingModel struct {
	gorm.Model `json:"-"`

	MachineSetupID uint

	// Why can image only be flashed onto one device file
	UUID    DiskUUID
	Image   DiskImage `gorm:"embedded"`
	Version uint
}

// MachineSetup describes the setup for a machine during a session
type MachineSetup struct {
	gorm.Model `json:"-"`

	// Ephemeral determines if we should save the state after the session ends
	Ephemeral bool
	// Disks is a map from uuids to disk images. Each disk image has a unique uuid,
	// not related to the /dev/disk/by-uuid on the booting system.
	Disks []DiskMappingModel
}
