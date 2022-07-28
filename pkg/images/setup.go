// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package images

import (
	"gorm.io/gorm"
)

// ImageFrozen defines a collection of Images where are pegged to a specific version.
type ImageFrozen struct {
	gorm.Model `json:"-"`
	Image      ImageModel `gorm:"foreignKey:UUIDImage;references:UUID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	UUIDImage  ImageUUID  `gorm:"not null;"`
	Version    Version    `json:"-" gorm:"foreignKey:VersionID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	VersionID  uint64     `gorm:"not null;"`
	// ImageSetup     ImageSetup `json:"-" gorm:"foreignKey:UUID;referencesImageSetupUUID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ImageSetupUUID ImageUUID `json:"-"`
	Update         bool      `gorm:"not null;default:false"`
}

// ImageSetup defines a collection of Images
type ImageSetup struct {
	Name     string        `gorm:"not null"`
	Images   []ImageFrozen `gorm:"foreignKey:ImageSetupUUID;references:UUID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	Username string        `gorm:"foreignKey:Username;not null;"`
	UUID     ImageUUID     `gorm:"uniqueIndex;primaryKey;unique;not null;"`
}

// CreateImageSetup creates an ImageSetup of a specified name.
func CreateImageSetup(name string) ImageSetup {
	return ImageSetup{
		Name:   name,
		Images: []ImageFrozen{},
	}
}

// AddImage takes an Image and a Version to adds both to Image list in ImageSetup
func (setup *ImageSetup) AddImage(image *ImageModel, version Version, update bool) {
	setup.Images = append(setup.Images, ImageFrozen{
		Image:   *image,
		Version: version,
		Update:  update,
	})
}

// AddFrozenImages adds all the given ImageFrozen to the ImageSetup
func (setup *ImageSetup) AddFrozenImages(images ...ImageFrozen) {
	setup.Images = append(setup.Images, images...)
}

// GetImageFromSetup queries the Image list to find aspecified ImageModel
func (setup ImageSetup) GetImageFromSetup(name string) (*ImageModel, *Version) {
	for _, frozenImage := range setup.Images {
		image := frozenImage.Image
		if image.Name == name {
			return &image, &frozenImage.Version
		}
	}
	return nil, nil
}
