// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import (
	"testing"

	"github.com/baas-project/baas/pkg/images"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	err = db.AutoMigrate(&images.Version{}, &images.ImageModel{})
	assert.NoError(t, err)

	im := images.ImageModel{
		Name:     "aaa",
		Versions: []images.Version{},
		UUID:     "yeet",
	}

	err = db.Create(&im).Error
	assert.NoError(t, err)

	imr := images.ImageModel{}
	db.Preload(clause.Associations).First(&imr)

	assert.Equal(t, imr.Name, "aaa")
	assert.Equal(t, string(imr.UUID), "yeet")
	assert.Equal(t, len(imr.Versions), 0)
}
