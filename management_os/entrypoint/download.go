package main

import (
	"io"
	"log"
	"os"
	"syscall"

	"baas/pkg/fs"

	"github.com/pkg/errors"

	"baas/pkg/model"
)

// WriteOutDisks Downloads, Decompresses and finally Writes a disk image to disk
func WriteOutDisks(api *APIClient, setup model.MachineSetup) error {
	for uuid, disk := range setup.Disks {
		reader, err := DownloadDisk(api, uuid, disk)
		if err != nil {
			return errors.Wrap(err, "error downloading disk")
		}

		dec, err := Decompress(reader, disk)
		if err != nil {
			return errors.Wrap(err, "error decompressing disk")
		}

		err = WriteDisk(dec, disk)
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
		return api.DownloadDiskHTTP(uuid)
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

// WriteDisk Writes an image to disk using an io reader and disk image definition
func WriteDisk(reader io.Reader, image model.DiskImage) error {
	file, err := os.OpenFile(image.Location, syscall.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "error opening path %s", image.Location)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Printf("error closing: %s %s", image.Location, err.Error())
		}
	}()

	return errors.Wrap(fs.CopyStream(reader, file), "error copying stream")
}
