package main

import (
	"io"

	"github.com/baas-project/baas/pkg/compression"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/model"
)

// ReadInDisks reads in all disks in the machine setup and uploads them to the control server.
func ReadInDisks(api *APIClient, setup model.MachineSetup) error {
	log.Info("Reading and uploading disks")

	for _, disk := range setup.Disks {
		log.Debugf("reading disk: %v", disk.UUID)

		r, err := ReadDisk(disk.Image)
		if err != nil {
			return errors.Wrapf(err, "read disk")
		}

		log.Debug("Compressing disk")
		com, err := compression.Compress(r, disk.Image.DiskCompressionStrategy)
		if err != nil {
			return errors.Wrapf(err, "compressing disk")
		}

		log.Debug("Uploading image")
		err = UploadDisk(api, com, disk.UUID, disk.Image)
		if err != nil {
			return errors.Wrapf(err, "uploading disk")
		}
	}
	return nil
}

// UploadDisk uploads a disk to the control server given a transfer strategy.
func UploadDisk(api *APIClient, reader io.Reader, uuid model.DiskUUID, image model.DiskImage) error {
	log.Debugf("DiskUUID transfer strategy: %v", image.DiskTransferStrategy)
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.UploadDiskHTTP(reader, uuid)
	default:
		return errors.New("unknown transfer strategy")
	}
}
