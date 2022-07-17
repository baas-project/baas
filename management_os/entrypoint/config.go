// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"os"
	"sync"

	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
)

// Config is the structure of the YAML file in code
type Config struct {
	UploadDisk        bool
	RebootAfterFinish bool
	SetNextBoot       bool
}

var conf *Config

func getConfig() *Config {
	var once sync.Once
	if conf == nil {
		once.Do(func() {
			file, err := os.OpenFile("/etc/baas.toml", os.O_RDONLY, 0644)

			if err != nil {
				log.Errorf("Cannot open the configuration file: '%s'", err)
			}

			content, err := io.ReadAll(file)

			if err != nil {
				log.Errorf("Cannot read configuration file: '%s'", err)
			}

			var config Config
			log.Info("Creating the configuration file object")

			err = toml.Unmarshal(content, &config)
			if err != nil {
				log.Errorf("Cannot load configuration file: '%s'", err.Error())
			}
			conf = &config
		})
	}

	return conf
}
