// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import (
	"github.com/baas-project/baas/pkg/model/user"
	"github.com/pkg/errors"
)

// GetUserByUsername gets the first user with the associated username from the database.
func (s Store) GetUserByUsername(name string) (*user.UserModel, error) {
	userModel := user.UserModel{}
	res := s.Where("username = ?", name).First(&userModel)
	return &userModel, res.Error
}

// GetUserByID gets the user with the specified id from the database.
func (s Store) GetUserByID(id uint) (*user.UserModel, error) {
	userModel := user.UserModel{}
	res := s.Where("id = ?", id).First(&userModel)
	return &userModel, errors.Wrap(res.Error, "find user by id")
}

// GetUsers gets all the users out of the database.
func (s Store) GetUsers() (users []user.UserModel, _ error) {
	res := s.Find(&users)
	return users, res.Error
}

// CreateUser creates a new user
func (s Store) CreateUser(user *user.UserModel) error {
	return s.Save(user).Error
}

// RemoveUser deletes a user from the database
func (s Store) RemoveUser(user *user.UserModel) error {
	return s.Delete(user).Error
}

// ModifyUser modifies a user
func (s Store) ModifyUser(user *user.UserModel) error {
	return s.Updates(user).Error
}
