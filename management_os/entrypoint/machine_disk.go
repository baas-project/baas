// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// MachineImage stores the target directory and device file associated with an image.
type MachineImage struct {
	DeviceFile string
	target     string
}

// Inspiration from the following, it does however count as being trivial
// https://github.com/jsgilmore/mount/blob/master/mounts_linux.go#L209

// ext4MountOptions are the options used to mount the ext4 image representing the
// the machine storage
var ext4MountOptions = "journal_checksum,errors=remount-ro,data=ordered"

// Initialise creates the image using a particular partition and a target directory
func (image *MachineImage) Initialise(file string, target string) {
	image.DeviceFile = file
	image.target = target

	if err := os.Mkdir(target, 0755); err != nil {
		log.Warnf("Failed to create mount target: %v", err)
		return
	}
}

func (image *MachineImage) Mount() {
	var flags uintptr
	flags = syscall.MS_NOATIME | syscall.MS_SILENT | syscall.MS_NODEV
	flags |= syscall.MS_NOEXEC | syscall.MS_NOSUID

	err := syscall.Mount(image.DeviceFile, image.target, "ext4", flags, ext4MountOptions)
	if err != nil {
		log.Warnf("Failed to mount %s to %s: %v", image.DeviceFile, image.target, err)
		return
	}
}

func (image *MachineImage) Unmount() {
	err := syscall.Unmount(image.target, 0)
	if err != nil {
		log.Warnf("Failed to mount %s: %v", image.target, err)
		return
	}
}

// Open opens a file for read or write on disk
func (image *MachineImage) Open(file string) (*os.File, error) {
	return os.OpenFile(image.target+"/"+file, os.O_CREATE|os.O_RDWR,
		0755)
}

// Create file creates a file on disk
func (image *MachineImage) Create(file string) (*os.File, error) {
	return os.Create(image.target + "/" + file)
}

// MkdirAll creates all directories given to it on disk
func (image *MachineImage) MkdirAll(dir string, perm os.FileMode) error {
	return os.MkdirAll(image.target+"/"+dir, perm)
}

// Remove deletes a file on disk
func (image *MachineImage) Remove(name string) error {
	return os.Remove(image.target + "/" + name)
}

// RemoveAll removes all files or directories in it's path
func (image *MachineImage) RemoveAll(path string) error {
	return os.RemoveAll(image.target + "/" + path)
}

// Exists checks if the path exists on disk
func (image *MachineImage) Exists(path string) (bool, error) {
	_, err := os.Stat(image.target + "/" + path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
