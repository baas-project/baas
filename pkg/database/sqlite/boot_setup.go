package sqlite

import "github.com/baas-project/baas/pkg/model"

// AddBootSetupToMachine adds a configuration for booting to the specified machine
func (s Store) AddBootSetupToMachine(bootSetup *model.BootSetup) error {
	return s.Save(bootSetup).Error
}

// GetNextBootSetup fetches the first machine from the database.
func (s Store) GetNextBootSetup(machineID uint) (model.BootSetup, error) {
	var bootSetup model.BootSetup
	res := s.Table("boot_setups").
		Where("machine_model_id = ?", machineID).
		First(&bootSetup).
		Delete(&bootSetup)
	return bootSetup, res.Error
}
