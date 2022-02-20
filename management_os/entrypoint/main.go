package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/baas-project/baas/pkg/images"

	"net"

	log "github.com/sirupsen/logrus"
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

func main() {
	conf := getConfig()
	var machine MachineImage
	machine.Initialise("/dev/sda1", "/mnt/machine")
	machine.Mount()
	c := NewAPIClient(baseurl)

	// Get the partition cache
	getPartitions(&machine)
	printPartitions()
	mac, err := getMacAddr()
	if err != nil {
		log.Fatal(err)
	}

	lastSetup := getLastSetup(&machine)
	// Unmount the machine partition so it can be overwritten if needed, the partition should remain the same.
	machine.Unmount()

	imageSetup, err := c.BootInform(mac)
	if err != nil {
		log.Fatal(err)
	}

	if conf.UploadDisk && lastSetup.UUID != "" {
		if err = ReadInDisks(c, lastSetup); err != nil {
			log.Fatalf("Failed to read the disks: %v", err)
		}
	} else {
		log.Info("Uploading disks disabled in configuration file.")
	}

	if err = WriteOutDisks(c, mac, imageSetup); err != nil {
		log.Fatal(err)
	}
	log.Info("reprovisioning done")

	// Reopen the machine file target
	machine.Mount()

	// Store the current image setup
	f, err := machine.Open("last_setup.json")
	if err != nil {
		log.Fatalf("Cannot write setup to disk: %v", err)
	}
	err = json.NewEncoder(f).Encode(imageSetup)

	if err != nil {
		log.Fatalf("Cannot encode the image setup: %v", err)
	}

	// Write the partition list
	f, err = machine.Open("partitions_cache.json")
	if err != nil {
		log.Fatalf("Cannot open the partition cache: %v", err)
		return
	}

	printPartitions()
	writePartitionJSON(f)
	err = f.Close()

	if err != nil {
		log.Warnf("Cannot close the partition cache file: %v", err)
	}

	machine.Unmount()

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
