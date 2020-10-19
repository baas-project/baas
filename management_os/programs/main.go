package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type TestRequest struct {
	test string `json:"test"`
}

func main() {
	var buf bytes.Buffer

	req := TestRequest{
		test: "test",
	}

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		panic(err)
	}

	_, err := http.Post("http://control_server:4848/mmos/test", "application/json", &buf);
	if err != nil {
		panic(err)
	}
}
