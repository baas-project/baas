// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/codingsince1985/checksum"
	"io"
	"os"
	"syscall"

	"github.com/baas-project/baas/pkg/images"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/baas-project/baas/pkg/fs"
)

// ReadDisk reads a disk from a file and returns a stream
func ReadDisk(image *images.ImageModel) (io.ReadCloser, error) {
	partition := getPartition(image.UUID)
	file, err := os.OpenFile(partition.DeviceFile, syscall.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening path %s", partition.DeviceFile)
	}

	return file, nil
}

// WriteDisk Writes an image to disk using an io reader and disk image definition
func WriteDisk(reader io.Reader, image *images.ImageModel) error {
	partition := getPartition(image.UUID)
	logrus.Debug("Writing to disk")
	if partition != nil {
		printPartition(*partition)
	}
	checksum, err := checksum.CRC32(partition.DeviceFile)
	if err != nil {
		logrus.Errorf("Cannot get checksum: %v", err)
	}

	if image.Checksum != "" && image.Checksum == checksum {
		return nil
	}

	file, err := os.OpenFile(partition.DeviceFile, syscall.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "error opening path %s", partition.DeviceFile)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Errorf("error closing: %s %s", partition.DeviceFile, err.Error())
		}
	}()

	return errors.Wrap(fs.CopyStream(reader, file), "Error compressing")
}
