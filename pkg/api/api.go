// Package api provides the common data structures used in the communication between the control server and management os
package api

// DiskType describes the type of a disk image, this can also describe the filesystem contained within
type DiskType int

const (
	// DiskTypeRaw is the most simple DiskType of which nothing extra is known
	DiskTypeRaw DiskType = iota
)

// DiskTransferStrategy describes the strategy used to down- and upload a disk image
type DiskTransferStrategy int

const (
	// DiskDownloadStrategyHTTP uses HTTP to transfer the disk image
	DiskDownloadStrategyHTTP DiskTransferStrategy = iota
)

// DiskImage describes a single disk image on the machine
type DiskImage struct {
	DiskType
	DiskTransferStrategy
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

// ReprovisioningInfo is to inform the management OS of the previous and next machine state and session
type ReprovisioningInfo struct {
	Prev MachineSetup
	Next MachineSetup
}
