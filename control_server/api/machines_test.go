package api

import (
	"bytes"
	"encoding/json"
	"github.com/baas-project/baas/pkg/database"
	"github.com/baas-project/baas/pkg/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApi_UpdateMachine(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	machine := model.MachineModel{
		MacAddresses: []model.MacAddress{
			{Mac: "abc"},
		},
		Name:              "bca",
		Architecture:      model.X86_64,
		DiskUUIDs:         nil,
		Managed:           false,
		ShouldReprovision: false,
		CurrentSetup:      model.MachineSetup{},
		NextSetup:         nil,
	}

	var mj bytes.Buffer
	err = json.NewEncoder(&mj).Encode(machine)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodPut, "/machine", &mj))

	assert.NoError(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)

	m, err := store.GetMachineByMac(machine.MacAddresses[0].Mac)
	assert.NoError(t, err)

	assert.Equal(t, m.Name, machine.Name)
	assert.Equal(t, m.Architecture, machine.Architecture)
	assert.Equal(t, m.MacAddresses[0].Mac, machine.MacAddresses[0].Mac)
}

func TestApi_UpdateMachineExists(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	machine := model.MachineModel{
		MacAddresses: []model.MacAddress{
			{Mac: "abc"},
		},
		Name:              "bca",
		Architecture:      model.X86_64,
		DiskUUIDs:         nil,
		Managed:           false,
		ShouldReprovision: false,
		CurrentSetup:      model.MachineSetup{},
		NextSetup:         nil,
	}

	var mj bytes.Buffer
	err = json.NewEncoder(&mj).Encode(machine)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodPut, "/machine", &mj))

	assert.Equal(t, resp.Code, http.StatusOK)

	m, err := store.GetMachineByMac(machine.MacAddresses[0].Mac)
	m.Model = gorm.Model{}

	assert.NoError(t, err)
	assert.Equal(t, m.Name, machine.Name)
	assert.Equal(t, m.Architecture, machine.Architecture)
	assert.Equal(t, m.MacAddresses[0].Mac, machine.MacAddresses[0].Mac)

	machine.Name = "xxx"

	mj = bytes.Buffer{}
	err = json.NewEncoder(&mj).Encode(machine)
	assert.NoError(t, err)

	resp = httptest.NewRecorder()
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodPut, "/machine", &mj))

	assert.Equal(t, resp.Code, http.StatusOK)

	m, err = store.GetMachineByMac(machine.MacAddresses[0].Mac)

	assert.NoError(t, err)
	assert.Equal(t, m.Name, machine.Name)
	assert.Equal(t, m.Architecture, machine.Architecture)
	assert.Equal(t, m.MacAddresses[0].Mac, machine.MacAddresses[0].Mac)
}

func TestApi_GetMachine(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	machine := model.MachineModel{
		MacAddresses: []model.MacAddress{
			{Mac: "abc"},
		},
		Name:              "bca",
		Architecture:      model.X86_64,
		DiskUUIDs:         nil,
		Managed:           false,
		ShouldReprovision: false,
		CurrentSetup:      model.MachineSetup{},
		NextSetup:         nil,
	}

	err = store.UpdateMachine(&machine)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()

	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/machine/"+machine.MacAddresses[0].Mac, nil))

	assert.NoError(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)

	var dm model.MachineModel
	err = json.NewDecoder(resp.Body).Decode(&dm)
	assert.NoError(t, err)

	assert.NoError(t, err)

	assert.Equal(t, dm.Name, machine.Name)
	assert.Equal(t, dm.Architecture, machine.Architecture)
	assert.Equal(t, dm.MacAddresses[0].Mac, machine.MacAddresses[0].Mac)
}

func TestApi_GetMachines(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	machine1 := model.MachineModel{
		MacAddresses: []model.MacAddress{
			{Mac: "abc"},
		},
		Name:              "bca",
		Architecture:      model.X86_64,
		DiskUUIDs:         nil,
		Managed:           false,
		ShouldReprovision: false,
		CurrentSetup:      model.MachineSetup{},
		NextSetup:         nil,
	}

	machine2 := model.MachineModel{
		MacAddresses: []model.MacAddress{
			{Mac: "cba"},
		},
		Name:              "bcd",
		Architecture:      model.X86_64,
		DiskUUIDs:         nil,
		Managed:           false,
		ShouldReprovision: false,
		CurrentSetup:      model.MachineSetup{},
		NextSetup:         nil,
	}

	err = store.UpdateMachine(&machine1)
	assert.NoError(t, err)
	err = store.UpdateMachine(&machine2)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()

	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/machines", nil))

	assert.NoError(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)

	var dm []model.MachineModel
	err = json.NewDecoder(resp.Body).Decode(&dm)
	assert.NoError(t, err)

	assert.Len(t, dm, 2)

	dm1 := dm[0]
	dm2 := dm[1]

	assert.NotEqual(t, dm1.Name, dm2.Name)
	if dm1.Name == machine2.Name {
		dm1, dm2 = dm2, dm1
	}

	assert.NoError(t, err)
	assert.Equal(t, dm1.Name, machine1.Name)
	assert.Equal(t, dm1.Architecture, machine1.Architecture)
	assert.Equal(t, dm1.MacAddresses[0].Mac, machine1.MacAddresses[0].Mac)

	assert.Equal(t, dm2.Name, machine2.Name)
	assert.Equal(t, dm2.Architecture, machine2.Architecture)
	assert.Equal(t, dm2.MacAddresses[0].Mac, machine2.MacAddresses[0].Mac)
}
