// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"

	"github.com/baas-project/baas/pkg/images"

	"github.com/baas-project/baas/pkg/compression"
	gzip "github.com/klauspost/pgzip"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

func setupDisk(api *APIClient, mac string, image *images.ImageModel, version uint64) error {
	log.Debugf("writing disk: %v", mac)

	reader, err := DownloadDisk(api, image, version)
	if err != nil {
		return errors.Wrap(err, "error downloading disk")
	}

	// Kind of a dirty hack which I am not super proud of. However, GZip's reader has an extra close method that
	// we need to deal with. This can only be done after writing everything, which means we need to keep the type
	// somehow. Casting it upwards is not allowed, hence this is the only solution I could find. Maybe there
	// is a neater way out there. Feel free to change this.
	var dec io.Reader
	if image.DiskCompressionStrategy == images.DiskCompressionStrategyGZip {
		r, err2 := gzip.NewReader(reader)

		if err2 != nil {
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
		dec, err = compression.Decompress(reader, image.DiskCompressionStrategy)
		if err != nil {
			return errors.Wrap(err, "error decompressing disk")
		}
	}

	err = WriteDisk(dec, image)
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
func WriteOutDisks(api *APIClient, mac string, setup *images.ImageSetup) error {
	log.Info("Downloading and writing disks")

	for _, image := range setup.Images {
		log.Warnf("Image UUID: %s", image.Image.UUID)
		// Yes, you could inline this function but this screws with the defers mechanism that Go has.
		// By using a separate method call we ensure that the file are closed whenever they are no longer
		// needed rather than waiting for the entire cycle.
		err := setupDisk(api, mac, &image.Image, image.VersionNumber)

		if err != nil {
			return errors.Wrap(err, "couldn't close download body")
		}
	}

	return nil
}

// DownloadDisk downloads a disk from the network using the image's DiskTransferStrategy
func DownloadDisk(api *APIClient, image *images.ImageModel, version uint64) (reader io.ReadCloser, _ error) {
	log.Debugf("Downloading image: %s", image.UUID)
	return api.DownloadDiskHTTP(image.UUID, version)
}
