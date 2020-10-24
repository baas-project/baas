package main

import (
	"baas/pkg/ipmi"
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		log.Fatal("Needs at least 2 args")
	}

	ctx, cancel := context.WithTimeout(ctx, 10 * time.Second)
	defer cancel()

	conn, err := ipmi.NewConnection(ctx, os.Args[1], "root", os.Args[2])
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
