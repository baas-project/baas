// Package model provides the common data structures used in the communication between the control server and management os
package model

// DiskType describes the type of a disk image, this can also describe the filesystem contained within
type DiskType int

const (
	// DiskTypeRaw is the most simple DiskType of which nothing extra is known
	DiskTypeRaw DiskType = iota
)

// DiskTransferStrategy describes the strategy used to down- and upload a disk image
type DiskTransferStrategy int

const (
	// DiskTransferStrategyHTTP uses HTTP to transfer the disk image
	DiskTransferStrategyHTTP DiskTransferStrategy = iota
)

// DiskCompressionStrategy the various available disk compression strategies
type DiskCompressionStrategy int

const (
	// DiskCompressionStrategyNone doesn't compress
	DiskCompressionStrategyNone DiskCompressionStrategy = iota
)

// DiskImage describes a single disk image on the machine
type DiskImage struct {
	DiskType
	DiskTransferStrategy
	DiskCompressionStrategy
	// Location is used to determine in combination with the DiskTransferStrategy how to retrieve the image
	Location string
}

// DiskUUID is the linux by-uuid of a disk
type DiskUUID = string


// MachineSetup describes the setup for a machine during a session
type MachineSetup struct {
	// Ephemeral determines if we should save the state after the session ends
	Ephemeral bool
	// Disks is a map from disk uuids to disk images
	Disks map[DiskUUID]DiskImage
}
