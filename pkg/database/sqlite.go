package database

import (
	errors2 "errors"
	"github.com/baas-project/baas/pkg/images"

	"github.com/baas-project/baas/pkg/model"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InMemoryPath is the path inside the memory pointing to the database
const InMemoryPath = "file::memory:"

// SqliteStore is the database structure
type SqliteStore struct {
	*gorm.DB
}

// CreateImage creates the image entity in the database and adds the first version to it.
func (s SqliteStore) CreateImage(username string, image *images.ImageModel) error {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return errors.Wrap(err, "get user by name")
	}
	res := s.Model(user).Association("Images").Append(image)

	if res != nil {
		return res
	}
	v := images.Version{
		ImageModelID: image.ID,
	}
	image.Versions = append(image.Versions, v)

	s.DB.Create(&v)
	return res
}

// GetImageByUUID fetches the image with the versions using their UUID as a key
func (s SqliteStore) GetImageByUUID(uuid images.ImageUUID) (*images.ImageModel, error) {
	image := images.ImageModel{UUID: uuid}
	res := s.Where("UUID = ?", uuid).
		Preload("Versions").
		First(&image)

	return &image, res.Error
}

// GetImagesByUsername fetches all the images associated to a user.
func (s SqliteStore) GetImagesByUsername(username string) ([]images.ImageModel, error) {
	var userImages []images.ImageModel

	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.id = image_models.user_model_id").
		Where("user_models.name = ?", username).
		Find(&userImages)

	return userImages, res.Error
}

// CreateImageSetup creates a collection of images in history.
func (s SqliteStore) CreateImageSetup(username string, image *images.ImageSetup) error {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return errors.Wrap(err, "get user by name")
	}
	res := s.Model(user).Association("Images").Append(image)
	return res
}

// GetImagesByNameAndUsername gets all the images associated with a user which have the same human-readable name.
// This theoretically possible, but it is unsure whether this actually holds in any real-world scenario.
func (s SqliteStore) GetImagesByNameAndUsername(name string, username string) ([]images.ImageModel, error) {
	var userImages []images.ImageModel
	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.id = image_models.user_model_id").
		Where("user_models.name = ? AND image_models.name = ?", username, name).
		Find(&userImages)
	return userImages, res.Error
}

// GetMachineByMac gets any machine with the associated MAC addresses from the database
func (s SqliteStore) GetMachineByMac(mac model.MacAddress) (*model.MachineModel, error) {
	machine := model.MachineModel{}
	res := s.Table("machine_models").
		Where("mac_address = ?", mac).
		First(&machine)

	return &machine, res.Error
}

// GetMachines returns the values in the machine_models database.
// TODO: Fetch foreign relations.
func (s SqliteStore) GetMachines() (machines []model.MachineModel, _ error) {
	res := s.Find(&machines)
	return machines, res.Error
}

// UpdateMachine updates the information about the machine or creates a machine where one does not yet exist.
func (s SqliteStore) UpdateMachine(machine *model.MachineModel) error {
	m, err := s.GetMachineByMac(machine.MacAddress)

	if errors2.Is(err, gorm.ErrRecordNotFound) {
		return s.Save(machine).Error
	} else if err != nil {
		return errors.Wrap(err, "get machine")
	}

	m.Architecture = machine.Architecture
	m.Managed = machine.Managed
	m.Name = machine.Name
	s.Save(&m)
	return nil
}

// AddBootSetupToMachine adds a configuration for booting to the specified machine
func (s SqliteStore) AddBootSetupToMachine(bootSetup *model.BootSetup) error {
	return s.Save(bootSetup).Error
}

// GetNextBootSetup fetches the first machine from the database.
func (s SqliteStore) GetNextBootSetup(machineID uint) (model.BootSetup, error) {
	var bootSetup model.BootSetup
	res := s.Table("boot_setups").
		Where("machine_model_id = ?", machineID).
		First(&bootSetup).
		Delete(&bootSetup)
	return bootSetup, res.Error
}

// GetLastDeletedBootSetup fetches the previously flashed image from the database which should tell us whether to update the image or not.
func (s SqliteStore) GetLastDeletedBootSetup(machineID uint) (model.BootSetup, error) {
	var bootSetup model.BootSetup
	res := s.Table("boot_setups").
		Unscoped().
		Where("machine_model_id = ? and DELETED_AT IS NOT NULL", machineID).
		Last(&bootSetup)
	return bootSetup, res.Error
}

// GetUserByUsername gets the first user with the associated username from the database.
func (s SqliteStore) GetUserByUsername(name string) (*model.UserModel, error) {
	user := model.UserModel{}
	res := s.Where("username = ?", name).First(&user)
	return &user, res.Error
}

// GetUserByID gets the user with the specified id from the database.
func (s SqliteStore) GetUserByID(id uint) (*model.UserModel, error) {
	user := model.UserModel{}
	res := s.Where("id = ?", id).First(&user)
	return &user, errors.Wrap(res.Error, "find user by id")
}

// GetUsers gets all the users out of the database.
func (s SqliteStore) GetUsers() (users []model.UserModel, _ error) {
	res := s.Find(&users)
	return users, res.Error
}

// CreateUser creates a new user
func (s SqliteStore) CreateUser(user *model.UserModel) error {
	return s.Save(user).Error
}

// CreateNewImageVersion creates a new version in the database
func (s SqliteStore) CreateNewImageVersion(version images.Version) {
	s.Create(&version)
}

// NewSqliteStore creates the database storage using the given string as the database file.
func NewSqliteStore(dbpath string) (Store, error) {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})

	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}

	err = db.AutoMigrate(
		&model.BootSetup{},
		&images.ImageSetup{},
		&images.ImageModel{},
		&model.MachineModel{},
		&model.UserModel{},
		&images.Version{},
	)

	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	return SqliteStore{
		db,
	}, nil
}
