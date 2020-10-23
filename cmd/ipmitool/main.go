package main

import (
	"baas/pkg/ipmi"
	"context"
	"encoding/json"
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

	res, err := conn.GetBootDev(ctx)
	if err != nil {
		log.Fatal(err)
	}

	jsonb, _ := json.Marshal(&res)

	log.Println(string(jsonb))
}
