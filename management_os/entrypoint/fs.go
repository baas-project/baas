package main

import (
	"io"
	"os"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"baas/pkg/fs"
	"baas/pkg/model"
)

// ReadDisk reads a disk from a file and returns a stream
func ReadDisk(image model.DiskImage) (io.ReadCloser, error) {
	file, err := os.OpenFile(image.Location, syscall.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening path %s", image.Location)
	}

	return file, nil
}

// WriteDisk Writes an image to disk using an io reader and disk image definition
func WriteDisk(reader io.Reader, image model.DiskImage) error {
	file, err := os.OpenFile(image.Location, syscall.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "error opening path %s", image.Location)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Errorf("error closing: %s %s", image.Location, err.Error())
		}
	}()

	return errors.Wrap(fs.CopyStream(reader, file), "error copying stream")
}
