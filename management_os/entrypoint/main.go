package main

import (
	"fmt"
	"os"
	"os/exec"

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

func main() {
	conf := getConfig()
	c := NewAPIClient(baseurl)

	mac, err := getMacAddr()
	if err != nil {
		log.Fatal(err)
	}

	prov, err := c.BootInform(mac)
	if err != nil {
		log.Fatal(err)
	}

	if !conf.UploadDisk {
		log.Info("Uploading disks disabled in configuration file.")
	}

	if !prov.Prev.Ephemeral {
		if err = ReadInDisks(c, mac, prov.Prev); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Info("Not downloading any disk because previous session was ephemeral")
	}

	if err = WriteOutDisks(c, mac, prov.Next); err != nil {
		log.Fatal(err)
	}

	log.Info("reprovisioning done")
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
