package main

import (
	"baas/pkg/api"
	"fmt"
	"log"
)

func main() {
	c := APIClient{baseURL: fmt.Sprintf("http://control_server:%d", api.Port)}

	log.Println(c.BootInform())
}
