package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigFile(t *testing.T) {
	config := getConfig()
	assert.True(t, config.UploadDisk)
	assert.True(t, config.SetNextBoot)
	assert.False(t, config.RebootAfterFinish)
}
