package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
)

type MachineImage struct {
	DeviceFile string
	target     string
}

// Inspiration from the following, it does however count as being trivial
// https://github.com/jsgilmore/mount/blob/master/mounts_linux.go#L209

// ext4MountOptions are the options used to mount the ext4 image representing the
// the machine storage
var ext4MountOptions = "journal_checksum,errors=remount-ro,data=ordered"

func (image *MachineImage) initialize(file string, target string) {
	image.DeviceFile = file
	image.target = target

	if err := os.Mkdir(target, 0755); err != nil {
		log.Warnf("Failed to create mount target: %v", err)
		return
	}

	var flags uintptr
	flags = syscall.MS_NOATIME | syscall.MS_SILENT | syscall.MS_NODEV
	flags |= syscall.MS_NOEXEC | syscall.MS_NOSUID

	err := syscall.Mount(file, target, "ext4", flags, ext4MountOptions)
	if err != nil {
		log.Warnf("Failed to mount %s to %s: %v", file, target, err)
		return
	}
}

func (image *MachineImage) Open(file string) (*os.File, error) {
	return os.OpenFile(image.target+"/"+file, os.O_CREATE|os.O_RDWR,
		0755)
}

func (image *MachineImage) Create(file string) (*os.File, error) {
	return os.Create(image.target + "/" + file)
}

func (image *MachineImage) MkdirAll(dir string, perm os.FileMode) error {
	return os.MkdirAll(image.target+"/"+dir, perm)
}

func (image *MachineImage) Remove(name string) error {
	return os.Remove(image.target + "/" + name)
}

func (image *MachineImage) RemoveAll(path string) error {
	return os.RemoveAll(image.target + "/" + path)
}

func (image *MachineImage) Exists(path string) (bool, error) {
	_, err := os.Stat(image.target + "/" + path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
