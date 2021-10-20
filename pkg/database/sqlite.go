package database

import (
	"github.com/baas-project/baas/pkg/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)
import "gorm.io/driver/sqlite"

const InMemoryPath = "file::memory:"

type SqliteStore struct {
	*gorm.DB
}

func (s SqliteStore) CreateImage(username string, image model.ImageModel) error {
	user, err := s.GetUserByName(username)
	if err != nil {
		return errors.Wrap(err, "get user by name")
	}

	return s.Model(user).Association("Images").Append(&image)
}

func (s SqliteStore) GetImageByUUID(uuid model.ImageUUID) (*model.ImageModel, error) {
	res := model.ImageModel{UUID: uuid}

	return &res, s.Find(&res).Error
}

func (s SqliteStore) GetMachineByMac(mac string) (*model.MachineModel, error) {

	machine := model.MachineModel{}
	res := s.Table("machine_models").
		Preload("MacAddresses").
		Select("*").
		Joins("join mac_addresses on mac_addresses.machine_model_id = machine_models.id").
		Where("mac_addresses.mac = ?", mac).
		Limit(1).
		Find(&machine)

	return &machine, errors.Wrap(res.Error, "find machine")
}

func (s SqliteStore) GetMachines() (machines []model.MachineModel, _ error) {
	res := s.Preload("MacAddresses").Find(&machines)
	return machines, res.Error
}

func (s SqliteStore) UpdateMachine(machine *model.MachineModel) error {
	if len(machine.MacAddresses) == 0 {
		return errors.New("no mac address in original machine")
	}

	old, err := s.GetMachineByMac(machine.MacAddresses[0].Mac)

	if err != nil {
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
	user := model.UserModel{
		Name: name,
	}
	res := s.Find(&user).Limit(1)
	return &user, errors.Wrap(res.Error, "find user")
}

func (s SqliteStore) GetUsers() (users []model.UserModel, _ error) {
	res := s.Find(&users)
	return users, res.Error
}

func (s SqliteStore) CreateUser(user *model.UserModel) error {
	return s.Save(user).Error
}

func NewSqliteStore(dbpath string) (Store, error) {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
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
