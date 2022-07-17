// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package model

import (
	"gorm.io/gorm"
)

// UserRole is an enum which stores the roles a user can have.
type UserRole string

const (
	// User can just use images and change their own image
	User UserRole = "user"
	// Moderator can change or upload system images
	Moderator = "moderator"
	// Admin can do anything on the system
	Admin = "admin"
)

// UserModel (noun) one who uses, not necessarily a single person
type UserModel struct {
	gorm.Model `json:"-"`

	// Name is a human-readable identifier for a user (or entity) of the system
	Username string `gorm:"unique;not null;primaryKey"`
	Name     string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
	Role     UserRole
}
