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
		MacAddress:        "abc",
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

	m, err := store.GetMachineByMac(machine.MacAddress)

	m.Model = gorm.Model{}

	assert.NoError(t, err)
	assert.EqualValues(t, m, &machine)
}

func TestApi_UpdateMachineExists(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	machine := model.MachineModel{
		MacAddress:        "abc",
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

	m, err := store.GetMachineByMac(machine.MacAddress)
	m.Model = gorm.Model{}

	assert.NoError(t, err)
	assert.EqualValues(t, m, &machine)

	machine.Name = "xxx"

	mj = bytes.Buffer{}
	err = json.NewEncoder(&mj).Encode(machine)
	assert.NoError(t, err)

	resp = httptest.NewRecorder()
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodPut, "/machine", &mj))

	assert.Equal(t, resp.Code, http.StatusOK)

	m, err = store.GetMachineByMac(machine.MacAddress)

	m.Model = gorm.Model{}

	assert.NoError(t, err)
	assert.EqualValues(t, m, &machine)
}


func TestApi_GetMachine(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	machine := model.MachineModel{
		MacAddress:        "abc",
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
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/machine/" + machine.MacAddress, nil))

	assert.NoError(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)

	var dm model.MachineModel
	err = json.NewDecoder(resp.Body).Decode(&dm)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.EqualValues(t, dm, machine)
}

func TestApi_GetMachines(t *testing.T) {
	store, err := database.NewSqliteStore(database.InMemoryPath)
	assert.NoError(t, err)

	machine1 := model.MachineModel{
		MacAddress:        "abc",
		Name:              "bca",
		Architecture:      model.X86_64,
		DiskUUIDs:         nil,
		Managed:           false,
		ShouldReprovision: false,
		CurrentSetup:      model.MachineSetup{},
		NextSetup:         nil,
	}

	machine2 := model.MachineModel{
		MacAddress:        "cba",
		Name:              "bca",
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

	assert.NoError(t, err)
	assert.Contains(t, dm, machine1)
	assert.Contains(t, dm, machine2)
}
