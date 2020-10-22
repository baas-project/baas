package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"baas/pkg/model"
)

// WriteOutDisks Downloads, Decompresses and finally Writes a disk image to disk
func WriteOutDisks(api *APIClient, setup model.MachineSetup) error {
	for disk := range setup.Disks {
		reader, err := DownloadDisk(api, disk, setup.Disks[disk])
		if err != nil {
			return errors.Wrap(err, "error downloading disk")
		}

		dec, err := Decompress(reader, setup.Disks[disk])
		if err != nil {
			return errors.Wrap(err, "error decompressing disk")
		}

		err = WriteDisk(dec, setup.Disks[disk])
		if err != nil {
			return errors.Wrap(err, "error writing disk")
		}

		err = reader.Close()
		if err != nil {
			return errors.Wrap(err, "couldn't close dowload body")
		}
	}

	return nil
}

// DownloadDisk downloads a disk from the network using the image's DiskTransferStrategy
func DownloadDisk(api *APIClient, uuid model.DiskUUID, image model.DiskImage) (reader io.ReadCloser, _ error) {
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.DownloadDiskHTTP(uuid, image)
	default:
		return nil, errors.New("unknown transfer strategy")
	}
}

// Decompress is a decorator to decompress a disk image stream
func Decompress(reader io.Reader, image model.DiskImage) (io.Reader, error) {
	switch image.DiskCompressionStrategy {
	case model.DiskCompressionStrategyNone:
		return reader, nil
	default:
		return nil, errors.New("unknown decompression strategy")
	}
}

const diskfmt string = "/dev/disk/by-uuid/%s"

// WriteDisk Writes an image to disk using an io reader and disk image definition
func WriteDisk(reader io.Reader, image model.DiskImage) error {
	path := func() string {
		if strings.HasPrefix(image.Location, "/") {
			return image.Location
		}
		return fmt.Sprintf(diskfmt, image.Location)
	}()

	file, err := os.OpenFile(path, syscall.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "error opening path %s", path)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Printf("error closing: %s %s", path, err.Error())
		}
	}()

	return errors.Wrap(CopyStream(reader, file), "error copying stream")
}
