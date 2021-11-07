package database

import (
	errors2 "errors"
	"time"

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
func (s SqliteStore) CreateImage(username string, image *model.ImageModel) error {
	user, err := s.GetUserByName(username)
	if err != nil {
		return errors.Wrap(err, "get user by name")
	}
	res := s.Model(user).Association("Images").Append(image)

	if res != nil {
		return res
	}
	v := model.Version{
		Version:      time.Now().Unix(),
		ImageModelID: image.ID,
	}
	image.Versions = append(image.Versions, v)

	s.DB.Create(&v)
	return res
}

// GetImageByUUID fetches the image with the versions using their UUID as a key
func (s SqliteStore) GetImageByUUID(uuid model.ImageUUID) (*model.ImageModel, error) {
	image := model.ImageModel{UUID: uuid}
	res := s.Where("UUID = ?", uuid).
		Preload("Versions").
		First(&image)

	return &image, res.Error
}

// GetImagesByUsername fetches all the images associated to a user.
func (s SqliteStore) GetImagesByUsername(username string) ([]model.ImageModel, error) {
	var images []model.ImageModel

	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.id = image_models.user_model_id").
		Where("user_models.name = ?", username).
		Find(&images)

	return images, res.Error
}

// GetImagesByNameAndUsername gets all the images associated with a user which have the same human-readable name.
// This theoretically possible, but it is unsure whether this actually holds in any real-world scenario.
func (s SqliteStore) GetImagesByNameAndUsername(name string, username string) ([]model.ImageModel, error) {
	var images []model.ImageModel
	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.id = image_models.user_model_id").
		Where("user_models.name = ? AND image_models.name = ?", username, name).
		Find(&images)
	return images, res.Error
}

// GetMachineByMac gets any machine with the associated MAC addresses from the database
func (s SqliteStore) GetMachineByMac(mac string) (*model.MachineModel, error) {
	machine := model.MachineModel{}
	res := s.Table("machine_models").
		Preload("MacAddresses").
		Select("*").
		Joins("join mac_addresses on mac_addresses.machine_model_id = machine_models.id").
		Where("mac_addresses.mac = ?", mac).
		First(&machine)

	return &machine, res.Error
}

// GetMachines returns the values in the machine_models database.
// TODO: Fetch foreign relations.
func (s SqliteStore) GetMachines() (machines []model.MachineModel, _ error) {
	res := s.
		Preload("MacAddresses").
		Find(&machines)

	return machines, res.Error
}

// UpdateMachine updates the information about the machine or creates a machine where one does not yet exist.
func (s SqliteStore) UpdateMachine(machine *model.MachineModel) error {
	if len(machine.MacAddresses) == 0 {
		return errors.New("no mac address in original machine")
	}

	old, err := s.GetMachineByMac(machine.MacAddresses[0].Mac)

	if errors2.Is(err, gorm.ErrRecordNotFound) {

	} else if err != nil {
		return errors.Wrap(err, "get machine")
	}

	// Create a new array containing the old MacAddresses
	var macAddresses []model.MacAddress
	copy(macAddresses, old.MacAddresses)

	// O(nm) operation to add those MacAddresses which filters out the MAC addresses already registered for this
	// machine. This is fairly slow, but this is a fairly rare operation and the nm is bounded by the amount of network
	// cards associated with any given machine. In all likelihood this will be somewhere around 1 <= n <= 5.
	for _, mac := range machine.MacAddresses {
		found := false

		for _, oldMac := range old.MacAddresses {
			if mac == oldMac {
				break
			}
			found = true
		}

		if !found {
			macAddresses = append(macAddresses, mac)
		}
	}

	machine.MacAddresses = macAddresses

	machine.ID = old.ID

	return s.Save(machine).Error
}

// AddBootSetupToMachine adds a configuration for booting to the specified machine
func (s SqliteStore) AddBootSetupToMachine(bootSetup *model.BootSetup) error {
	return s.Save(bootSetup).Error
}

// GetNextBootSetup fetches the first machine from the database.
func (s SqliteStore) GetNextBootSetup(machineID uint) (model.BootSetup, error) {
	var bootSetup model.BootSetup
	res := s.Table("boot_setups").Where("machine_model_id = ?", machineID).First(&bootSetup)
	return bootSetup, res.Error
}

// GetUserByName gets the first user with the associated username from the database.
func (s SqliteStore) GetUserByName(name string) (*model.UserModel, error) {
	user := model.UserModel{}
	res := s.Where("name = ?", name).First(&user)
	return &user, errors.Wrap(res.Error, "find user")
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
func (s SqliteStore) CreateNewImageVersion(version model.Version) {
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
		&model.DiskMappingModel{},
		&model.ImageModel{},
		&model.MacAddress{},
		&model.MachineModel{},
		&model.MachineSetup{},
		&model.UserModel{},
		&model.Version{},
	)

	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	return SqliteStore{
		db,
	}, nil
}
