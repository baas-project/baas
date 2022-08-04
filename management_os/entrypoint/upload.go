// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"

	"github.com/baas-project/baas/pkg/compression"
	"github.com/baas-project/baas/pkg/model/images"
	"github.com/baas-project/baas/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ReadInDisks reads in all disks in the machine setup and uploads them to the control server.
func ReadInDisks(api *APIClient, setup *images.ImageSetup) error {
	log.Info("Reading and uploading disks")

	for _, image := range setup.Images {
		log.Debugf("reading disk: %v", image.Image.UUID)
		util.PrettyPrintStruct(image)
		if !image.Update {
			log.Debugf("Image %s update is not set", image.UUIDImage)
			continue
		}

		r, err := ReadDisk(&image.Image)
		if err != nil {
			return errors.Wrapf(err, "read disk")
		}

		log.Debug("Compressing disk")
		com, err := compression.Compress(r, image.Image.DiskCompressionStrategy)
		if err != nil {
			return errors.Wrapf(err, "compressing disk")
		}

		log.Debug("Uploading image")
		err = UploadDisk(api, com, &image.Image)
		if err != nil {
			return errors.Wrapf(err, "uploading disk")
		}
	}
	return nil
}

// UploadDisk uploads a disk to the control server given a transfer strategy.
func UploadDisk(api *APIClient, reader io.Reader, uuid *images.ImageModel) error {
	return api.UploadDiskHTTP(reader, string(uuid.UUID))
}
