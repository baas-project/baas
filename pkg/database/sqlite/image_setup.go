package sqlite

import (
	"github.com/baas-project/baas/pkg/images"
	"github.com/pkg/errors"
)

// CreateImageSetup creates a collection of images in history.
func (s Store) CreateImageSetup(username string, image *images.ImageSetup) error {
	_, err := s.GetUserByUsername(username)
	if err != nil {
		return errors.Wrap(err, "get user by name")
	}

	//res := s.Model(user).Association("Image_Setups").Append(image)
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
