package images

import (
	"gorm.io/gorm"
)

// ImageFrozen defines a collection of Images where are pegged to a specific version.
type ImageFrozen struct {
	gorm.Model     `json:"-"`
	Image          *ImageModel `json:"-" gorm:"foreignKey:UUIDImage;references:UUID"`
	UUIDImage      ImageUUID
	Version        Version `json:"-" gorm:"foreignKey:VersionNumber;references:Version"`
	VersionNumber  uint64
	ImageSetupUUID ImageUUID `json:"-"`
}

// ImageSetup defines a collection of Images
type ImageSetup struct {
	gorm.Model `json:"-"`
	Name       string        `gorm:"not null"`
	Images     []ImageFrozen `gorm:"foreignKey:ImageSetupUUID;references:UUID"`
	User       string        `gorm:"foreignKey:Username"`
	UUID       ImageUUID     `gorm:"uniqueIndex;primaryKey;unique"`
}

// CreateImageSetup creates an ImageSetup of a specified name.
func CreateImageSetup(name string) ImageSetup {
	return ImageSetup{
		Name:   name,
		Images: []ImageFrozen{},
	}
}

// AddImage takes an Image and a Version to adds both to Image list in ImageSetup
func (setup *ImageSetup) AddImage(image *ImageModel, version Version) {
	setup.Images = append(setup.Images, ImageFrozen{
		Image:   image,
		Version: version,
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
			return image, &frozenImage.Version
		}
	}
	return nil, nil
}
