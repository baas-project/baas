package main

import (
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/model"
)

// ReadInDisks reads in all disks in the machine setup and uploads them to the control server.
func ReadInDisks(api *APIClient, setup model.MachineSetup) error {
	log.Info("Reading and uploading disks")

	for uuid, disk := range setup.Disks {
		log.Infof("<upload> reading disk: %v", uuid)
		r, err := ReadDisk(disk)
		if err != nil {
			return errors.Wrapf(err, "read disk")
		}

		log.Info("<upload> Compressing the disk")
		com, err := Compress(r, disk)
		if err != nil {
			return errors.Wrapf(err, "compressing disk")
		}

		log.Info("<upload> Uploading disk to server")
		err = UploadDisk(api, com, uuid, disk)
		if err != nil {
			return errors.Wrapf(err, "uploading disk")
		}
		log.Info("<upload> Finished the upload")
	}
	return nil
}

// UploadDisk uploads a disk to the control server given a transfer strategy.
func UploadDisk(api *APIClient, reader io.Reader, uuid model.DiskUUID, image model.DiskImage) error {
	log.Debugf("Disk transfer strategy: %v", image.DiskTransferStrategy)
	switch image.DiskTransferStrategy {
	case model.DiskTransferStrategyHTTP:
		return api.UploadDiskHTTP(reader, uuid)
	default:
		return errors.New("unknown transfer strategy")
	}
}
