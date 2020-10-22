package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"baas/pkg/api"
)

// APIClient is the client for all communication with the server
type APIClient struct {
	baseURL string
}

// BootInform informs the server that we have booted
func (a *APIClient) BootInform() (*api.ReprovisioningInfo, error) {
	b, err := json.Marshal(&api.BootInformRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't marshal boot inform json")
	}

	resp, err := http.Post(a.baseURL+"/mmos/inform", "application/json", bytes.NewReader(b))
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Print(err)
		}
	}()
	if err != nil {
		return nil, errors.Wrap(err, "failed sending inform request")
	}

	if resp.StatusCode != http.StatusOK {
		msg, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("inform request failed (%s)", string(msg))
	}

	var info api.ReprovisioningInfo

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, errors.Wrap(err, "couldn't deserialize inform request response")
	}

	return &info, nil
}
