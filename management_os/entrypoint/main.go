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

	log.Printf("Starting upload")

	if !prov.Prev.Ephemeral {
		if err := ReadInDisks(&c, prov.Prev); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Starting Download")

	if err := WriteOutDisks(&c, prov.Next); err != nil {
		log.Fatal(err)
	}

	log.Printf("Done!")
}
