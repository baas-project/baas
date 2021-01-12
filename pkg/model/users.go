package model

import "gorm.io/gorm"

// User (noun) one who uses, not necessarily a single person
type User struct {
	gorm.Model

	// Name is a human-readable identifier for a user (or entity) of the system
	Name string

	// Images is a list of ImageModel of this user
	Images []ImageModel
}
