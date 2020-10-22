package main

import (
	"fmt"
	"log"

	"baas/pkg/api"
)

func main() {
	c := APIClient{baseURL: fmt.Sprintf("http://control_server:%d", api.Port)}

	prov, err := c.BootInform()
	if err != nil {
		log.Fatal(err)
	}

	if !prov.Prev.Ephemeral {
		log.Println("idk lp0 on fire i'm not a unix admin")
	}

	log.Fatal(WriteOutDisks(&c, prov.Next))
}
