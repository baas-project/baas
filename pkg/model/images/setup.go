// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package images

import (
	"github.com/baas-project/baas/pkg/model/machine"
	"gorm.io/gorm"
)

// ImageFrozen defines a collection of Images where are pegged to a specific version.
type ImageFrozen struct {
	gorm.Model `json:"-"`
	Image      ImageModel `gorm:"foreignKey:UUIDImage;references:UUID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	UUIDImage  ImageUUID  `gorm:"not null;" json:"-"`
	Version    Version    `gorm:"foreignKey:VersionID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	VersionID  uint64     `gorm:"not null" json:"-"`

	// ImageSetup     ImageSetup `json:"-" gorm:"foreignKey:UUID;referencesImageSetupUUID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ImageSetupUUID ImageUUID `json:"-"`
	Update         bool      `gorm:"not null;default:false"`
}

// ImageSetup defines a collection of Images
type ImageSetup struct {
	gorm.Model `json:"-"`
	Name       string        `gorm:"not null"`
	Images     []ImageFrozen `gorm:"foreignKey:ImageSetupUUID;references:UUID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	Username   string        `gorm:"foreignKey:Username;not null;"`
	UUID       ImageUUID     `gorm:"uniqueIndex;primaryKey;unique;not null;"`
}

// BootSetup stores what the next boot for the machine should look like.
// It functions somewhat like a queue where it removes the first value from the database.
type BootSetup struct {
	gorm.Model `json:"-"`

	// Store the machine id
	Machine    machine.MachineModel `gorm:"foreignKey:MachineMAC;references:Address;not null;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	MachineMAC string               `gorm:"not null;primaryKey"`

	// Store the setup that should be loaded onto the machine
	Setup     ImageSetup `gorm:"foreignKey:SetupUUID;references:UUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SetupUUID ImageUUID  `gorm:"not null;primaryKey"`

	// Should the image changes be uploaded to the server?
	Update bool `gorm:"not null;"`
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
func (setup *ImageSetup) GetImageFromSetup(name string) (*ImageModel, *Version) {
	for _, frozenImage := range setup.Images {
		image := frozenImage.Image
		if image.Name == name {
			return &image, &frozenImage.Version
		}
	}
	return nil, nil
}
