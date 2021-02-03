package main

import (
	"github.com/baas-project/baas/pkg/compression"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/model"
)

// ReadInDisks reads in all disks in the machine setup and uploads them to the control server.
func ReadInDisks(api *APIClient, mac string, setup model.MachineSetup) error {
	log.Info("Reading and uploading disks")

	for _, disk := range setup.Disks {
		log.Debugf("reading disk: %v", disk.Uuid)

		r, err := ReadDisk(disk.Image)
		if err != nil {
			return errors.Wrapf(err, "read disk")
		}

		com, err := compression.Compress(r, disk.Image.DiskCompressionStrategy)
		if err != nil {
			return errors.Wrapf(err, "compressing disk")
		}

		err = UploadDisk(api, com, mac, disk.Uuid, disk.Image)
		if err != nil {
			return errors.Wrapf(err, "uploading disk")
		}
	}
	return nil
}

// UploadDisk uploads a disk to the control server given a transfer strategy.
func UploadDisk(api *APIClient, reader io.Reader, mac string, uuid model.DiskUUID, image model.DiskImage) error {
	log.Debugf("DiskUUID transfer strategy: %v", image.DiskTransferStrategy)
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.UploadDiskHTTP(reader, mac, uuid)
	default:
		return errors.New("unknown transfer strategy")
	}
}
