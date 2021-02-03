package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"

	"github.com/baas-project/baas/pkg/httplog"

	"github.com/baas-project/baas/pkg/api"
)

var baseurl = fmt.Sprintf("http://control_server:%d", api.Port)

func init() {
	log.AddHook(httplog.NewLogHook(fmt.Sprintf("%s/log", baseurl), "MMOS"))
}

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

func main() {
	c := NewAPIClient(baseurl)

	mac, err := getMacAddr()
	if err != nil {
		log.Fatal(err)
	}

	prov, err := c.BootInform(mac)
	if err != nil {
		log.Fatal(err)
	}

	if !prov.Prev.Ephemeral {
		if err := ReadInDisks(c, mac, prov.Prev); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Info("Not downloading any disk because previous session was ephemeral")
	}

	if err := WriteOutDisks(c, mac, prov.Next); err != nil {
		log.Fatal(err)
	}

	log.Info("reprovisioning done")
}
