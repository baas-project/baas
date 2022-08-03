// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	machinemodel "github.com/baas-project/baas/pkg/model/machine"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baas-project/baas/pkg/database/sqlite"
	"github.com/baas-project/baas/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestApi_UpdateMachine(t *testing.T) {
	store, err := sqlite.NewSqliteStore(sqlite.InMemoryPath)
	assert.NoError(t, err)

	machine := machinemodel.MachineModel{
		MacAddress:   util.MacAddress{Address: "abc"},
		Name:         "bca",
		Architecture: machinemodel.X86_64,
		Managed:      false,
	}

	var mj bytes.Buffer
	err = json.NewEncoder(&mj).Encode(machine)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "")
	req := httptest.NewRequest(http.MethodPut, "/machine", &mj)
	req.Header.Add("type", "system")
	req.Header.Add("origin", "http://localhost:9090")
	handler.ServeHTTP(resp, req)

	assert.NoError(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)

	m, err := store.GetMachineByMac(machine.MacAddress)
	assert.NoError(t, err)

	assert.Equal(t, m.Name, machine.Name)
	assert.Equal(t, m.Architecture, machine.Architecture)
	assert.Equal(t, m.MacAddress, machine.MacAddress)
}

func TestApi_UpdateMachineExists(t *testing.T) {
	store, err := sqlite.NewSqliteStore(sqlite.InMemoryPath)
	assert.NoError(t, err)

	machine := machinemodel.MachineModel{
		MacAddress:   util.MacAddress{Address: "abc"},
		Name:         "bca",
		Architecture: machinemodel.X86_64,
		Managed:      false,
	}

	var mj bytes.Buffer
	err = json.NewEncoder(&mj).Encode(machine)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	handler := getHandler(store, "", "")
	req := httptest.NewRequest(http.MethodPut, "/machine", &mj)
	req.Header.Add("type", "system")
	req.Header.Add("origin", "http://localhost:9090")
	handler.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusOK)

	m, err := store.GetMachineByMac(machine.MacAddress)

	assert.NoError(t, err)
	assert.Equal(t, m.Name, machine.Name)
	assert.Equal(t, m.Architecture, machine.Architecture)
	assert.Equal(t, m.MacAddress, machine.MacAddress)

	machine.Name = "xxx"

	mj = bytes.Buffer{}
	err = json.NewEncoder(&mj).Encode(machine)
	assert.NoError(t, err)

	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/machine", &mj)
	req.Header.Add("type", "system")
	req.Header.Add("origin", "http://localhost:9090")
	handler.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusOK)

	m, err = store.GetMachineByMac(machine.MacAddress)

	assert.NoError(t, err)
	assert.Equal(t, m.Name, machine.Name)
	assert.Equal(t, m.Architecture, machine.Architecture)
	assert.Equal(t, m.MacAddress, machine.MacAddress)
}

func TestApi_GetMachine(t *testing.T) {
	store, err := sqlite.NewSqliteStore(sqlite.InMemoryPath)
	assert.NoError(t, err)

	machine := machinemodel.MachineModel{
		MacAddress:   util.MacAddress{Address: "abc"},
		Name:         "bca",
		Architecture: machinemodel.X86_64,
		Managed:      false,
	}

	err = store.UpdateMachine(&machine)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/machine/"+machine.MacAddress.Address, nil)
	req.Header.Add("type", "system")
	req.Header.Add("origin", "http://localhost:9090")

	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, req)

	assert.NoError(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)

	var dm machinemodel.MachineModel
	err = json.NewDecoder(resp.Body).Decode(&dm)
	assert.NoError(t, err)

	assert.NoError(t, err)

	assert.Equal(t, dm.Name, machine.Name)
	assert.Equal(t, dm.Architecture, machine.Architecture)
	assert.Equal(t, dm.MacAddress, machine.MacAddress)
}

func TestApi_GetMachines(t *testing.T) {
	store, err := sqlite.NewSqliteStore(sqlite.InMemoryPath)
	assert.NoError(t, err)

	machine1 := machinemodel.MachineModel{
		MacAddress:   util.MacAddress{Address: "abc"},
		Name:         "bca",
		Architecture: machinemodel.X86_64,
		Managed:      false,
	}

	machine2 := machinemodel.MachineModel{
		MacAddress:   util.MacAddress{Address: "cba"},
		Name:         "bcd",
		Architecture: machinemodel.X86_64,
		Managed:      false,
	}

	err = store.UpdateMachine(&machine1)
	assert.NoError(t, err)
	err = store.UpdateMachine(&machine2)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	req.Header.Add("type", "system")
	req.Header.Add("origin", "http://localhost:9090")

	handler := getHandler(store, "", "")
	handler.ServeHTTP(resp, req)

	assert.NoError(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)

	var dm []machinemodel.MachineModel
	err = json.NewDecoder(resp.Body).Decode(&dm)
	assert.NoError(t, err)

	fmt.Println(dm[0].MacAddress)
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
	assert.Equal(t, dm1.MacAddress, machine1.MacAddress)

	assert.Equal(t, dm2.Name, machine2.Name)
	assert.Equal(t, dm2.Architecture, machine2.Architecture)
	assert.Equal(t, dm2.MacAddress, machine2.MacAddress)
}
