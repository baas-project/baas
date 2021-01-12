package model

import (
	"gorm.io/gorm"
)

// SystemArchitecture defines constants describing the architecture of machines.
type SystemArchitecture string

const (
	// Arm64 is the 64 bit Arm architecture
	Arm64 SystemArchitecture = "Arm64"
	// X86_64 is the 64 bit x86 architecture
	X86_64 SystemArchitecture = "x86_64" //nolint
	// Unknown is any architecture which baas could not identify.
	Unknown SystemArchitecture = "unknown"
)

// Name gets the name of an architecture as a string. Convenience function,
// but actually does very little as the name is also the value of the constant.
func (id *SystemArchitecture) Name() string {
	return string(*id)
}

// Machine stores information intrinsic to a machine. Used together with the MachineStore.
type Machine struct {
	gorm.Model
	MacAddress   string
	Name         string
	Architecture SystemArchitecture

	// DiskUUIDs are the linux by-uuids this machine has
	DiskUUIDs []DiskUUID

	// Managed indicates that a machine should be managed by BAAS (if false baas will not touch the machine in any way)
	Managed bool

	// ShouldReprovision indicates if at bootinform time this machine should be (re)provisioned to NextSetup
	ShouldReprovision bool
	CurrentSetup MachineSetup
	// NextSetup stores the machine setup of what the machine should become after reprovisioning
	// MUST be non-nil if ShouldReprovision is true else it MAY be nil
	NextSetup	 *MachineSetup
}

// MachineSetup describes the setup for a machine during a session
type MachineSetup struct {
	// Ephemeral determines if we should save the state after the session ends
	Ephemeral bool
	// Disks is a map from uuids to disk images. Each disk image has a unique uuid,
	// not related to the /dev/disk/by-uuid on the booting system.
	Disks map[DiskUUID]DiskImage
}
