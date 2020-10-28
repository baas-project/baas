package main

import (
	"baas/pkg/httplog"
	"fmt"
	log "github.com/sirupsen/logrus"

	"baas/pkg/api"
)

var baseurl = fmt.Sprintf("http://control_server:%d", api.Port)

func init() {
	log.AddHook(httplog.NewLogHook(baseurl  + "/log", "MMOS"))
}


func main() {
	c := APIClient{baseURL: baseurl}

	prov, err := c.BootInform()
	if err != nil {
		log.Fatal(err)
	}

	if !prov.Prev.Ephemeral {
		if err := ReadInDisks(&c, prov.Prev); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Info("Not downloading any disk because previous session was ephemeral")
	}

	if err := WriteOutDisks(&c, prov.Next); err != nil {
		log.Fatal(err)
	}

	log.Info("reprovisioning done")
}
