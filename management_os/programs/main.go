package main

import (
	"log"
)

func main() {
	if err := CopyFile("/tmp/test.img", "/tmp/test2.img"); err != nil {
		log.Fatal(err)
	}
}
