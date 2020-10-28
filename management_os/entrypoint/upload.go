package main

import (
	"io"
	"os"
	"syscall"

	"github.com/pkg/errors"

	"baas/pkg/model"
)

// ReadInDisks reads in all disks in the machine setup and uploads them to the control server.
func ReadInDisks(api *APIClient, setup model.MachineSetup) error {
	for uuid, disk := range setup.Disks {
		r, err := ReadDisk(disk)
		if err != nil {
			return errors.Wrapf(err, "read disk")
		}

		com, err := Compress(r, disk)
		if err != nil {
			return errors.Wrapf(err, "compressing disk")
		}

		err = UploadDisk(api, com, uuid, disk)
		if err != nil {
			return errors.Wrapf(err, "uploading disk")
		}
	}
	return nil
}

// UploadDisk uploads a disk to the control server given a transfer strategy.
func UploadDisk(api *APIClient, reader io.ReadCloser, uuid model.DiskUUID, image model.DiskImage) error {
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.UploadDiskHTTP(reader, uuid)
	default:
		return errors.New("unknown transfer strategy")
	}
}

// Compress is a decorator to compress a disk image stream
func Compress(reader io.ReadCloser, image model.DiskImage) (io.ReadCloser, error) {
	switch image.DiskCompressionStrategy {
	case model.DiskCompressionStrategyNone:
		return reader, nil
	default:
		return nil, errors.New("unknown decompression strategy")
	}
}

// ReadDisk reads a disk from a file and returns a stream
func ReadDisk(image model.DiskImage) (io.ReadCloser, error) {
	file, err := os.OpenFile(image.Location, syscall.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening path %s", image.Location)
	}

	return file, nil
}
