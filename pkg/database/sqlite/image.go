// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import (
	"github.com/baas-project/baas/pkg/model/images"
	"gorm.io/gorm"
)

// CreateImage creates the image entity in the database and adds the first version to it.
func (s Store) CreateImage(image *images.ImageModel) {
	image.Versions = append(image.Versions, images.Version{Version: 0, ImageModelUUID: image.UUID})
	s.DB.Create(image)
}

// GetImageByUUID fetches the image with the versions using their UUID as a key
func (s Store) GetImageByUUID(uuid images.ImageUUID) (*images.ImageModel, error) {
	image := images.ImageModel{UUID: uuid}
	err := s.Where("UUID = ?", uuid).
		Preload("Versions").
		First(&image).Error

	if err == gorm.ErrRecordNotFound {
		var machine *images.MachineImageModel
		machine, err = s.GetMachineImageByUUID(uuid)
		image = machine.ImageModel
	}

	return &image, err
}

// GetImagesByUsername fetches all the images associated to a user.
func (s Store) GetImagesByUsername(username string) ([]images.ImageModel, error) {
	var userImages []images.ImageModel

	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.username = image_models.username").
		Where("user_models.username = ?", username).
		Find(&userImages)

	return userImages, res.Error
}

// CreateNewImageVersion creates a new version in the database
func (s Store) CreateNewImageVersion(version images.Version) {
	s.Create(&version)
}

// GetVersionByID gets the version associated with a specific ID
func (s Store) GetVersionByID(versionID uint64) (*images.Version, error) {
	var version images.Version
	err := s.Table("versions").Where("id = ?", versionID).First(&version).Error
	return &version, err
}

// GetImagesByNameAndUsername gets all the images associated with a user which have the same human-readable name.
// This theoretically possible, but it is unsure whether this actually holds in any real-world scenario.
func (s Store) GetImagesByNameAndUsername(name string, username string) ([]images.ImageModel, error) {
	var userImages []images.ImageModel
	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.username = image_models.username").
		Where("image_models.name = ?", name).
		Find(&userImages)
	return userImages, res.Error
}

// DeleteImage removes an image from the database
func (s Store) DeleteImage(image *images.ImageModel) error {
	return s.Unscoped().Delete(image).Error
}

// UpdateImage updates an image in the database
func (s Store) UpdateImage(image *images.ImageModel) error {
	return s.Updates(image).Error
}
