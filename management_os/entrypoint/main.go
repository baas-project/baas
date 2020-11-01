package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/baas-project/baas/pkg/httplog"

	"github.com/baas-project/baas/pkg/api"
)

var baseurl = fmt.Sprintf("http://control_server:%d", api.Port)

func init() {
	log.AddHook(httplog.NewLogHook(fmt.Sprintf("%s/log", baseurl), "MMOS"))
}

func main() {
	c := NewAPIClient(baseurl)

	prov, err := c.BootInform()
	if err != nil {
		log.Fatal(err)
	}

	if !prov.Prev.Ephemeral {
		if err := ReadInDisks(c, prov.Prev); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Info("Not downloading any disk because previous session was ephemeral")
	}

	if err := WriteOutDisks(c, prov.Next); err != nil {
		log.Fatal(err)
	}

	log.Info("reprovisioning done")
}
