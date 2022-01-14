package images

import "gorm.io/gorm"

type ImageFrozen struct {
	image   ImageModel
	Version Version
}

// ImageSetup defines a collection of images
type ImageSetup struct {
	gorm.Model `json:"-"`
	Name       string
	images     []ImageFrozen
	User       string `gorm:"foreignKey:Username"`
}

func CreateImageSetup(name string) ImageSetup {
	return ImageSetup{
		Name:   name,
		images: []ImageFrozen{},
	}
}

func (setup ImageSetup) AddImage(image ImageModel, version Version) {
	setup.images = append(setup.images, ImageFrozen{
		image:   image,
		Version: version,
	})
}

func (setup ImageSetup) AddFrozenImages(images ...ImageFrozen) {
	setup.images = append(setup.images, images...)
}

func (setup ImageSetup) GetImageFromSetup(name string) (*ImageModel, *Version) {
	for _, frozenImage := range setup.images {
		image := frozenImage.image
		if image.Name == name {
			return &image, &frozenImage.Version
		}
	}
	return nil, nil
}
