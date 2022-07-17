// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFile(t *testing.T) {
	config := getConfig()
	assert.True(t, config.UploadDisk)
	assert.True(t, config.SetNextBoot)
	assert.False(t, config.RebootAfterFinish)
}
