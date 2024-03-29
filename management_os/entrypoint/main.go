// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	syslog "log/syslog"
	"os"
	"os/exec"

	"github.com/baas-project/baas/pkg/model/images"

	"net"

	log "github.com/sirupsen/logrus"
	sysruslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/sirupsen/logrus/hooks/writer"

	"github.com/baas-project/baas/pkg/httplog"

	"github.com/baas-project/baas/pkg/api"
)

func getMacAddr() (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var as []string
	for _, ifa := range ifas {
		if ifa.Flags&net.FlagUp == 0 {
			continue
		}
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}

	if len(as) == 0 {
		return "", err
	}

	return as[0], nil
}

var baseurl = fmt.Sprintf("http://control_server:%d", api.Port)

func init() {
	file, err := os.OpenFile("/var/log/baas.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666)

	log.SetFormatter(&BaasFormatter{log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)

	if err != nil {
		log.Warn("Cannot log to file")
	} else {
		log.AddHook(&writer.Hook{
			Writer: file,
			LogLevels: []log.Level{
				log.PanicLevel,
				log.ErrorLevel,
				log.DebugLevel,
				log.FatalLevel,
				log.WarnLevel,
				log.InfoLevel,
			},
		})
	}

	log.AddHook(httplog.NewLogHook(fmt.Sprintf("%s/log", baseurl), "MMOS"))
	hook, err := sysruslog.NewSyslogHook("", "", syslog.LOG_DEBUG, "")
	if err != nil {
		log.Warn("Cannot open syslog")
	} else {
		log.AddHook(hook)
	}
}

func getLastSetup(machine *MachineImage) images.ImageSetup {
	var lastSetup images.ImageSetup
	// Fetch information on the last setup
	if v, _ := machine.Exists("last_setup.json"); v {
		f, err := machine.Open("last_setup.json")
		if err != nil {
			log.Warnf("Cannot read the last setup information: %v", err)
		}

		_ = json.NewDecoder(f).Decode(&lastSetup)
		err = f.Close()
		if err != nil {
			log.Warnf("Cannot close the image: %v", err)
		}
		err = machine.Remove("last_setup.json")
		if err != nil {
			log.Warnf("Cannot close the image: %v", err)
		}
	} else {
		f, err := machine.Create("last_setup.json")
		if err != nil {
			log.Warnf("Cannot create the last setup: %v", err)
		}
		lastSetup = images.ImageSetup{}
		err = f.Close()

		if err != nil {
			log.Warnf("Cannot close the last setup file: %v", err)
		}
	}

	return lastSetup
}

func initializeMachine() *images.ImageSetup {
	var machine MachineImage
	machine.Initialise("/dev/sda1", "/mnt/machine")
	machine.Mount()
	defer machine.Unmount()

	// Get the partition cache
	getPartitions(&machine)
	printPartitions()

	lastSetup := getLastSetup(&machine)
	// Unmount the machine partition so it can be overwritten if needed, the partition should remain the same.
	return &lastSetup
}

func teardownMachine(imageSetup *images.ImageSetup) {
	var machine MachineImage
	machine.Initialise("/dev/sda1", "/mnt/machine")
	machine.Mount()
	defer machine.Unmount()

	// Store the current image setup
	f, err := machine.Open("last_setup.json")

	defer func() {
		err = f.Close()
		if err != nil {
			log.Errorf("Cannot close last setup file: %v", err)
		}
	}()

	if err != nil {
		log.Errorf("Cannot write setup to disk: %v", err)
	}
	err = json.NewEncoder(f).Encode(imageSetup)

	if err != nil {
		log.Errorf("Cannot encode the image setup: %v", err)
	}

	// Write the partition list
	pc, err := machine.Open("partitions_cache.json")

	defer func() {
		err = pc.Close()
		if err != nil {
			log.Errorf("Cannot close partition cache: %v", err)
		}
	}()

	printPartitions()
	writePartitionJSON(pc)
}

func main() {
	conf := getConfig()
	c := NewAPIClient(baseurl)
	mac, err := getMacAddr()

	if err != nil {
		log.Fatal(err)
	}

	lastSetup := initializeMachine()
	if conf.UploadDisk && lastSetup.UUID != "" {
		if err = ReadInDisks(c, lastSetup); err != nil {
			log.Fatalf("Failed to read the disks: %v", err)
		}
	} else {
		log.Info("Uploading disks disabled in configuration file.")
	}

	imageSetup, err := c.BootInform(mac)
	if err != nil {
		log.Fatal(err)
	}

	if err = WriteOutDisks(c, mac, imageSetup); err != nil {
		log.Fatal(err)
	}
	log.Info("reprovisioning done")

	teardownMachine(imageSetup)

	// This presumes that the second option is the hard disk
	if conf.SetNextBoot {
		log.Info("Setting the BootNext parameter")
		cmd := exec.Command("efibootmgr", "-n", "1")
		log.Info(cmd.String())
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	if conf.RebootAfterFinish {
		cmd := exec.Command("systemctl", "reboot")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
