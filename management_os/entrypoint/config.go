package main

import (
	"io"
	"os"
	"sync"

	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	UploadDisk bool
	RebootAfterFinish bool
	SetNextBoot bool
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

