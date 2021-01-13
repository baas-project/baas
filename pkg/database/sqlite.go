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
	machine := model.MachineModel{
		MacAddress: mac,
	}

	res := s.Find(&machine).Limit(1)
	return &machine, errors.Wrap(res.Error, "find machine")
}

func (s SqliteStore) GetMachines() (machines []model.MachineModel, _ error) {
	res := s.Find(&machines)
	return machines, res.Error
}

func (s SqliteStore) UpdateMachine(machine *model.MachineModel) error {
	result := model.MachineModel{}

	resp := s.Where("mac_address = ?", machine.MacAddress).First(&result)
	if resp.Error != nil {
		if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			// insert
			return s.Create(machine).Error
		} else {
			return resp.Error
		}
	}

	machine.ID = result.ID

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
	)

	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	return SqliteStore{
		db,
	}, nil
}
