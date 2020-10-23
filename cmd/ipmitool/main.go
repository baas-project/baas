package main

import (
	"baas/pkg/ipmi"
	"context"
	"log"
	"os"
	"time"
)

func main() {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, 10 * time.Second)
	defer cancel()

	conn, err := ipmi.NewConnection(ctx, os.Args[1], "baas", os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(conn.Reboot(ctx))
}
