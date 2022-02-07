package sqlite

import (
	"github.com/baas-project/baas/pkg/database"
	"github.com/baas-project/baas/pkg/images"
	"github.com/baas-project/baas/pkg/model"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InMemoryPath is the path inside the memory pointing to the database
const InMemoryPath = "file::memory:"

// Store is the database structure
type Store struct {
	*gorm.DB
}

// NewSqliteStore creates the database storage using the given string as the database file.
func NewSqliteStore(dbpath string) (database.Store, error) {
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
		&images.MachineImageModel{},
		&model.MachineModel{},
		&model.UserModel{},
		&images.Version{},
		&images.ImageFrozen{},
	)

	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	return Store{
		db,
	}, nil
}
