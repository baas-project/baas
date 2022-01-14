package main

import (
	"github.com/baas-project/baas/pkg/images"
	"io"

	"github.com/baas-project/baas/pkg/compression"
	gzip "github.com/klauspost/pgzip"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/model"
)

func setupDisk(api *APIClient, mac string, uuid string, disk images.DiskImage, version uint) error {
	log.Debugf("writing disk: %v", mac)

	reader, err := DownloadDisk(api, uuid, disk, version)
	if err != nil {
		return errors.Wrap(err, "error downloading disk")
	}

	// Kind of a dirty hack which I am not super proud of. However, GZip's reader has an extra close method that
	// we need to deal with. This can only be done after writing everything, which means we need to keep the type
	// somehow. Casting it upwards is not allowed, hence this is the only solution I could find. Maybe there
	// is a neater way out there. Feel free to change this.
	var dec io.Reader
	if disk.DiskCompressionStrategy == images.DiskCompressionStrategyGZip {
		r, err := gzip.NewReader(reader)

		if err != nil {
			return errors.Wrap(err, "Opening GZip stream")
		}

		defer func() {
			err = r.Close()
			if err != nil {
				log.Warnf("Cannot close GZip stream: '%s'", err)
			}
		}()

		// Cast down to common Reader
		dec = r // nolint: ineffassign
	} else {
		dec, err = compression.Decompress(reader, disk.DiskCompressionStrategy)
		if err != nil {
			return errors.Wrap(err, "error decompressing disk")
		}

		err = WriteDisk(dec, disk)
		if err != nil {
			return errors.Wrap(err, "error writing disk")
		}
	}

	err = reader.Close()

	if err != nil {
		return errors.Wrap(err, "couldn't close download body")
	}

	return nil
}

// WriteOutDisks Downloads, Decompresses and finally Writes a disk image to disk
func WriteOutDisks(api *APIClient, mac string, setup model.BootSetup) error {
	log.Info("Downloading and writing disks")
	/*
		for _, disk := range setup.Disks {
			// Yes, you could inline this function but this screws with the defers mechanism that Go has.
			// By using a separate method call we ensure that the file are closed whenever they are no longer
			// needed rather than waiting for the entire cycle.
			err := setupDisk(api, mac, disk.UUID, disk.Image, disk.Version)

			if err != nil {
				return errors.Wrap(err, "couldn't close download body")
			}
		}
	*/

	return nil
}

// DownloadDisk downloads a disk from the network using the image's DiskTransferStrategy
func DownloadDisk(api *APIClient, uuid model.BootInfo, image images.DiskImage, version uint) (reader io.ReadCloser, _ error) {
	return api.DownloadDiskHTTP(uuid, version)
}
