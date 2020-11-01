package main

import (
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/model"
)

// WriteOutDisks Downloads, Decompresses and finally Writes a disk image to disk
func WriteOutDisks(api *APIClient, setup model.MachineSetup) error {
	log.Info("Downloading and writing disks")

	for uuid, disk := range setup.Disks {
		log.Debugf("writing disk: %v", uuid)

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
	log.Debugf("Disk transfer strategy: %v", image.DiskTransferStrategy)
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.DownloadDiskHTTP(uuid)
	default:
		return nil, errors.New("unknown transfer strategy")
	}
}
