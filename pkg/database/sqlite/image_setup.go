// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import (
	"github.com/baas-project/baas/pkg/model/images"
	"github.com/pkg/errors"
)

// CreateImageSetup creates a collection of images in history.
func (s Store) CreateImageSetup(username string, image *images.ImageSetup) error {
	_, err := s.GetUserByUsername(username)
	if err != nil {
		return errors.Wrap(err, "get user by name")
	}

	// res := s.Model(user).Association("Image_Setups").Append(image)
	return s.Create(&image).Error
}

// FindImageSetupsByUsername finds all ImageSetups associated with a particular user.
func (s Store) FindImageSetupsByUsername(username string) (*[]images.ImageSetup, error) {
	var userImageSetup []images.ImageSetup
	res := s.Table("image_setups").
		Joins("join user_models on image_setups.user = user_models.username").
		Where("user_models.username = ?", username).
		Find(&userImageSetup)
	return &userImageSetup, res.Error
}

// AddImageToImageSetup adds an image pegged to a particular version to the setup.
func (s Store) AddImageToImageSetup(setup *images.ImageSetup, image *images.ImageModel, version images.Version,
	update bool) {
	setup.AddImage(image, version, update)
	s.DB.Updates(setup)
}

// GetImageSetup an image setup associated with a particular UUID.
func (s Store) GetImageSetup(uuid string) (images.ImageSetup, error) {
	var imageSetup images.ImageSetup
	res := s.Table("image_setups").
		Preload("Images").
		Preload("Images.Image").
		Where("image_setups.uuid = ?", uuid).
		First(&imageSetup)
	return imageSetup, res.Error
}

// GetImageSetups finds the image setups associated with a user
func (s Store) GetImageSetups(username string) (*[]images.ImageSetup, error) {
	var imageSetups []images.ImageSetup
	res := s.Table("image_setups").
		Preload("Images").
		Preload("Images.Image").
		Where("image_setups.Username = ?", username).
		Find(&imageSetups)

	return &imageSetups, res.Error
}

// DeleteImageSetup deletes an image setup
func (s Store) DeleteImageSetup(imageSetup *images.ImageSetup) error {
	return s.Delete(imageSetup).Unscoped().Error
}

// ModifyImageSetup changes the metadata of an image setup
func (s Store) ModifyImageSetup(imageSetup *images.ImageSetup) error {
	return s.Updates(imageSetup).Error
}

// RemoveImageFromImageSetup removes a particular iamge from the image setup
func (s Store) RemoveImageFromImageSetup(setup *images.ImageSetup, targetImage *images.ImageModel, version images.Version, update bool) error {
	var found images.ImageFrozen
	for _, image := range setup.Images {
		if image.UUIDImage == targetImage.UUID {
			found = image
		}
	}

	return s.Delete(&found).Error
}
