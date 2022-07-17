// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyFileSmall(t *testing.T) {
	fromFileName := os.TempDir() + "/baas-TestCopyFileSmallFrom"
	toFileName := os.TempDir() + "/baas-TestCopyFileSmallTo"

	if _, err := os.Stat(fromFileName); !os.IsNotExist(err) {
		assert.Fail(t, "file exists")
	}
	if _, err := os.Stat(toFileName); !os.IsNotExist(err) {
		assert.Fail(t, "file exists")
	}

	from, err := os.Create(fromFileName)
	assert.NoError(t, err)

	content := []byte("Hello World!")
	n, err := from.Write(content)
	assert.NoError(t, err)
	assert.Equal(t, len(content), n)

	// SUT
	err = CopyFile(fromFileName, toFileName)
	assert.NoError(t, err)

	result, err := ioutil.ReadFile(toFileName)
	assert.NoError(t, err)

	assert.Equal(t, content, result)

	err = os.Remove(toFileName)
	assert.NoError(t, err)
	err = os.Remove(fromFileName)
	assert.NoError(t, err)
}

func TestCopyFileLarge(t *testing.T) {
	fromFileName := os.TempDir() + "/baas-TestCopyFileLargeFrom"
	toFileName := os.TempDir() + "/baas-TestCopyFileLargeTo"

	if _, err := os.Stat(fromFileName); !os.IsNotExist(err) {
		assert.Fail(t, "file exists")
	}
	if _, err := os.Stat(toFileName); !os.IsNotExist(err) {
		assert.Fail(t, "file exists")
	}

	from, err := os.Create(fromFileName)
	assert.NoError(t, err)

	content := []byte(strings.Repeat("A", int(blocksize+1)))
	assert.True(t, len(content) > int(blocksize))
	n, err := from.Write(content)
	assert.NoError(t, err)
	assert.Equal(t, len(content), n)

	// SUT
	err = CopyFile(fromFileName, toFileName)
	assert.NoError(t, err)

	result, err := ioutil.ReadFile(toFileName)
	assert.NoError(t, err)

	assert.Equal(t, content, result)

	err = os.Remove(toFileName)
	assert.NoError(t, err)
	err = os.Remove(fromFileName)
	assert.NoError(t, err)
}

func TestCopyFile512(t *testing.T) {
	fromFileName := os.TempDir() + "/baas-TestCopyFileLargeFrom"
	toFileName := os.TempDir() + "/baas-TestCopyFileLargeTo"

	if _, err := os.Stat(fromFileName); !os.IsNotExist(err) {
		assert.Fail(t, "file exists")
	}
	if _, err := os.Stat(toFileName); !os.IsNotExist(err) {
		assert.Fail(t, "file exists")
	}

	from, err := os.Create(fromFileName)
	assert.NoError(t, err)

	content := []byte(strings.Repeat("A", int(blocksize)))
	assert.True(t, len(content) == int(blocksize))
	n, err := from.Write(content)
	assert.NoError(t, err)
	assert.Equal(t, len(content), n)

	// SUT
	err = CopyFile(fromFileName, toFileName)
	assert.NoError(t, err)

	result, err := ioutil.ReadFile(toFileName)
	assert.NoError(t, err)

	assert.Equal(t, content, result)

	err = os.Remove(toFileName)
	assert.NoError(t, err)
	err = os.Remove(fromFileName)
	assert.NoError(t, err)
}
