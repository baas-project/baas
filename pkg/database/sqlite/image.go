// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import "github.com/baas-project/baas/pkg/images"

// CreateImage creates the image entity in the database and adds the first version to it.
func (s Store) CreateImage(image *images.ImageModel) {
	var versions []images.Version
	image.Versions = append(versions, images.Version{Version: 0, ImageModelUUID: image.UUID})
	s.DB.Create(image)
}

// GetImageByUUID fetches the image with the versions using their UUID as a key
func (s Store) GetImageByUUID(uuid images.ImageUUID) (*images.ImageModel, error) {
	image := images.ImageModel{UUID: uuid}
	res := s.Where("UUID = ?", uuid).
		Preload("Versions").
		First(&image)

	return &image, res.Error
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

// GetImagesByNameAndUsername gets all the images associated with a user which have the same human-readable name.
// This theoretically possible, but it is unsure whether this actually holds in any real-world scenario.
func (s Store) GetImagesByNameAndUsername(name string, username string) ([]images.ImageModel, error) {
	var userImages []images.ImageModel
	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.id = image_models.user_model_id").
		Where("user_models.name = ? AND image_models.name = ?", username, name).
		Find(&userImages)
	return userImages, res.Error
}
