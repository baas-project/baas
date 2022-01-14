// Package model defines the entities which are used inside the database.
package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/baas-project/baas/pkg/images"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

// SystemArchitecture defines constants describing the architecture of machines.
type SystemArchitecture string

const (
	// Arm64 is the 64-bit Arm architecture
	Arm64 SystemArchitecture = "Arm64"
	// X86_64 is the 64-bit x86 architecture
	X86_64 SystemArchitecture = "x86_64" //nolint
	// Unknown is any architecture which baas could not identify.
	Unknown SystemArchitecture = "unknown"
)

// Name gets the name of an architecture as a string. Convenience function,
// but actually does very little as the name is also the value of the constant.
func (id *SystemArchitecture) Name() string {
	return string(*id)
}

// BootSetup stores what the next boot for the machine should look like.
// It functions somewhat like a queue where it removes the first value from the database.
type BootSetup struct {
	gorm.Model `json:"-"`

	// Store the machine id
	MachineModelID uint `gorm:"foreignKey:ID"`

	// Store the setup that should be loaded onto the machine
	Setup images.ImageSetup `gorm:"foreignKey:ID"`

	// Should the image changes be uploaded to the server?
	Update bool
}

// MacAddress is a structure containing the unique Mac Address
type MacAddress struct {
	Address string
}

func (mac MacAddress) GormDataType() string {
	return "INTEGER"
}

func (mac MacAddress) GormValue(_ context.Context, _ *gorm.DB) clause.Expr {
	hex, err := strconv.ParseUint(strings.ReplaceAll(mac.Address, ":", ""), 16, 64)
	if err != nil {

	}
	return clause.Expr{
		SQL:  "?",
		Vars: []interface{}{fmt.Sprintf("%d", hex)},
	}
}

func (mac *MacAddress) Scan(v interface{}) error {
	bs, ok := v.(int64)
	if !ok {
		return errors.New("cannot parse mac address")
	}

	builder := strings.Builder{}
	num := fmt.Sprintf("%x", bs)
	for i, v := range []byte(num) {
		if i != 0 && i%2 == 0 {
			builder.WriteByte(':')
		}
		builder.WriteByte(v)
	}
	mac.Address = builder.String()
	return nil

}

// MachineModel stores information intrinsic to a machine. Used together with the MachineStore.
type MachineModel struct {
	gorm.Model `json:"-"`

	// General Info
	Name         string
	Architecture SystemArchitecture

	// Managed indicates that a machine should be managed by BAAS (if false baas will not touch the machine in any way)
	Managed bool

	// MacAddress is the mac address associated with this machine
	MacAddress MacAddress
}
