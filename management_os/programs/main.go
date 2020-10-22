package main

import (
	"fmt"
	"log"

	"baas/pkg/api"
)

func main() {
	c := APIClient{baseURL: fmt.Sprintf("http://control_server:%d", api.Port)}

	log.Println(c.BootInform())
}
