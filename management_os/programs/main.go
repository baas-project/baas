package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// TestRequest is just temporary until there's actually something to send.
type TestRequest struct {
	Test string `json:"test"`
}

func main() {
	var buf bytes.Buffer

	req := TestRequest{
		Test: "test",
	}

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		panic(err)
	}

	r, err := http.Post("http://control_server:4848/mmos/test", "application/json", &buf)
	if err != nil {
		panic(err)
	}

	err = r.Body.Close()
	if err != nil {
		panic(err)
	}
}
