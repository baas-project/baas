package database

import (
	errors2 "errors"
	"fmt"
	"github.com/baas-project/baas/pkg/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)
import "gorm.io/driver/sqlite"

const InMemoryPath = "file::memory:"

type SqliteStore struct {
	*gorm.DB
}

func (s SqliteStore) CreateImage(username string, image *model.ImageModel) error {
	user, err := s.GetUserByName(username)
	if err != nil {
		return errors.Wrap(err, "get user by name")
	}
	res := s.Model(user).Association("Images").Append(image)

	if res != nil { return res }
	v := model.Version{
		Version: time.Now().Unix(),
		ImageModelID: image.ID,
	}
	image.Versions = append(image.Versions, v)

	s.DB.Create(&v)
	return res
}

func (s SqliteStore) GetImageByUUID(uuid model.ImageUUID) (*model.ImageModel, error) {
	image := model.ImageModel{UUID: uuid}
	res := s.Where("UUID = ?", uuid).
		Preload("Versions").
		First(&image)
	fmt.Print(image.Versions)
	return &image, res.Error
}

func (s SqliteStore) GetImagesByUsername(username string) ([]model.ImageModel, error) {
	var images []model.ImageModel

	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.id = image_models.user_model_id").
		Where("user_models.name = ?", username).
		Find(&images)

	return images, res.Error
}

func (s SqliteStore) GetImagesByNameAndUsername(name string, username string) ([]model.ImageModel, error) {
	var images []model.ImageModel
	res := s.Table("image_models").
		Preload("Versions").
		Joins("join user_models on user_models.id = image_models.user_model_id").
		Where("user_models.name = ? AND image_models.name = ?", username, name).
		Find(&images)
	return images, res.Error
}

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
		Select("*").
		Joins("left join disk_mapping_models on disk_mapping_models.machine_setup_id = machine_models.id").
		Find(&machines)

	return machines, res.Error
}

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

func (s SqliteStore) GetUserByName(name string) (*model.UserModel, error) {
	user := model.UserModel{}
	res := s.Where("name = ?", name).First(&user)
	return &user, errors.Wrap(res.Error, "find user")
}

func (s SqliteStore) GetUserById(id uint) (*model.UserModel, error) {
	user := model.UserModel{}
	res := s.Where("id = ?", id).First(&user)
	return &user, errors.Wrap(res.Error, "find user by id")
}

func (s SqliteStore) GetUsers() (users []model.UserModel, _ error) {
	res := s.Find(&users)
	return users, res.Error
}

func (s SqliteStore) CreateUser(user *model.UserModel) error {
	return s.Save(user).Error
}

func NewSqliteStore(dbpath string) (Store, error) {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
			})

	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}

	err = db.AutoMigrate(
		&model.Version{},
		&model.ImageModel{},
		&model.UserModel{},
		&model.MachineModel{},
		&model.DiskMappingModel{},
		&model.MachineSetup{},
		&model.MacAddress{},
	)

	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	return SqliteStore{
		db,
	}, nil
}
