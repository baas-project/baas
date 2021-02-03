package main

import (
	"github.com/baas-project/baas/pkg/compression"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/model"
)

// WriteOutDisks Downloads, Decompresses and finally Writes a disk image to disk
func WriteOutDisks(api *APIClient, mac string, setup model.MachineSetup) error {
	log.Info("Downloading and writing disks")

	for _, disk := range setup.Disks {
		log.Debugf("writing disk: %v", disk.Uuid)

		reader, err := DownloadDisk(api, mac, disk.Uuid, disk.Image)
		if err != nil {
			return errors.Wrap(err, "error downloading disk")
		}

		dec, err := compression.Decompress(reader, disk.Image.DiskCompressionStrategy)
		if err != nil {
			return errors.Wrap(err, "error decompressing disk")
		}

		err = WriteDisk(dec, disk.Image)
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
func DownloadDisk(api *APIClient, mac string, uuid model.DiskUUID, image model.DiskImage) (reader io.ReadCloser, _ error) {
	log.Debugf("DiskUUID transfer strategy: %v", image.DiskTransferStrategy)
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.DownloadDiskHTTP(mac, uuid)
	default:
		return nil, errors.New("unknown transfer strategy")
	}
}
