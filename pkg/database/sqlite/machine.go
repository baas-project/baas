package sqlite

import (
	errors2 "errors"

	"github.com/baas-project/baas/pkg/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GetMachineByMac gets any machine with the associated MAC addresses from the database
func (s Store) GetMachineByMac(mac model.MacAddress) (*model.MachineModel, error) {
	machine := model.MachineModel{}
	res := s.Table("machine_models").
		Where("mac_address = ?", mac).
		First(&machine)

	return &machine, res.Error
}

// GetMachines returns the values in the machine_models database.
// TODO: Fetch foreign relations.
func (s Store) GetMachines() (machines []model.MachineModel, _ error) {
	res := s.Find(&machines)
	return machines, res.Error
}

// UpdateMachine updates the information about the machine or creates a machine where one does not yet exist.
func (s Store) UpdateMachine(machine *model.MachineModel) error {
	m, err := s.GetMachineByMac(machine.MacAddress)

	if errors2.Is(err, gorm.ErrRecordNotFound) {
		return s.Save(machine).Error
	} else if err != nil {
		return errors.Wrap(err, "get machine")
	}

	m.Architecture = machine.Architecture
	m.Managed = machine.Managed
	m.Name = machine.Name
	s.Save(&m)
	return nil
}
