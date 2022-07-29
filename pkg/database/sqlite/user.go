// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import (
	"github.com/baas-project/baas/pkg/model"
	"github.com/pkg/errors"
)

// GetUserByUsername gets the first user with the associated username from the database.
func (s Store) GetUserByUsername(name string) (*model.UserModel, error) {
	user := model.UserModel{}
	res := s.Where("username = ?", name).First(&user)
	return &user, res.Error
}

// GetUserByID gets the user with the specified id from the database.
func (s Store) GetUserByID(id uint) (*model.UserModel, error) {
	user := model.UserModel{}
	res := s.Where("id = ?", id).First(&user)
	return &user, errors.Wrap(res.Error, "find user by id")
}

// GetUsers gets all the users out of the database.
func (s Store) GetUsers() (users []model.UserModel, _ error) {
	res := s.Find(&users)
	return users, res.Error
}

// CreateUser creates a new user
func (s Store) CreateUser(user *model.UserModel) error {
	return s.Save(user).Error
}

// DeleteUser remove a user
func (s Store) RemoveUser(user *model.UserModel) error {
	return s.Delete(user).Error
}

// ModifyUser modifies a user
func (s Store) ModifyUser(user *model.UserModel) error {
	return s.Updates(user).Error
}
