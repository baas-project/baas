package main

import (
	gzip "github.com/klauspost/pgzip"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/model"
)

func setupDisk(api *APIClient, uuid model.DiskUUID, disk model.DiskImage) error {
	log.Debugf("writing disk: %v", uuid)

	reader, err := DownloadDisk(api, uuid, disk)
	if err != nil {
		return errors.Wrap(err, "error downloading disk")
	}

	// Kind of a dirty hack which I am not super proud of. However, GZip's reader has an extra close method that
	// we need to deal with. This can only be done after writing everything, which means we need to keep the type
	// somehow. Casting it upwards is not allowed, hence this is the only solution I could find. Maybe there
	// is a neater way out there. Feel free to change this.
	var dec io.Reader
	if disk.DiskCompressionStrategy == model.DiskCompressionStrategyGZip {
		r, err := gzip.NewReader(reader)

		if err != nil {
			return errors.Wrap(err, "Opening GZip stream")
		}

		defer func () {
			err = r.Close()
			if err != nil {
				log.Warnf("Cannot close GZip stream: '%s'", err)
			}
		}()

		// Cast down to common Reader
		dec = r
	} else {
		dec, err = Decompress(reader, disk)
		if err != nil {
			return errors.Wrap(err, "error decompressing disk")
		}
	}

	err = WriteDisk(dec, disk)
	if err != nil {
		return errors.Wrap(err, "error writing disk")
	}

	err = reader.Close()
	if err != nil {
		return errors.Wrap(err, "couldn't close download body")
	}
	return nil
}
// WriteOutDisks Downloads, Decompresses and finally Writes a disk image to disk
func WriteOutDisks(api *APIClient, setup model.MachineSetup) error {
	log.Info("Downloading and writing disks")

	for uuid, disk := range setup.Disks {
		// Yes, you could inline this function but this crews with the defers mechanism that Go has.
		// By using a separate method call we ensure that the file are closed whenever they are no longer
		// needed rather than waiting for the entire cycle.
		err := setupDisk(api, uuid, disk)

		if err != nil {
			return err
		}
	}

	return nil
}

// DownloadDisk downloads a disk from the network using the image's DiskTransferStrategy
func DownloadDisk(api *APIClient, uuid model.DiskUUID, image model.DiskImage) (reader io.ReadCloser, _ error) {
	log.Debugf("Disk transfer strategy: %v", image.DiskTransferStrategy)
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.DownloadDiskHTTP(uuid)
	default:
		return nil, errors.New("unknown transfer strategy")
	}
}
