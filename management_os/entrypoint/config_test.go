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
