package database

import (
	"github.com/baas-project/baas/pkg/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"testing"
	"time"
)

type User struct {
	gorm.Model
	CreditCards []CreditCard
}

type CreditCard struct {
	gorm.Model
	Number string
	UserID uint
}

func TestNewSqliteStore(t *testing.T) {

	db, err := gorm.Open(sqlite.Open(InMemoryPath), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&model.Version{}, &model.ImageModel{})
	assert.NoError(t, err)

	im := model.ImageModel{
		Name: "aaa",
		Versions: []model.Version{
			{
				Version: time.Now(),
			}, {
				Version: time.Now(),
			}, {
				Version: time.Now(),
			},
		},
		UUID:     "yeet",
		DiskUUID: "yote",
	}

	err = db.Create(&im).Error
	assert.NoError(t, err)

	imr := model.ImageModel{}
	db.Preload(clause.Associations).First(&imr)

	assert.Equal(t, len(imr.Versions), 3)

}
