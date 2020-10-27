package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"baas/pkg/model"

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
	if err != nil {
		return nil, errors.Wrap(err, "failed sending inform request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Print(err)
		}
	}()

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

// DownloadDiskHTTP Downloads a disk image from the control_server over HTTP
func (a *APIClient) DownloadDiskHTTP(_ model.DiskUUID, _ model.DiskImage) (io.ReadCloser, error) {
	//nolint we are returning a readcloser so the body will be closed later
	resp, err := http.Get(a.baseURL + "/static/alpine-virt-3.12.1-x86_64.iso")
	if err != nil {
		return nil, errors.Wrap(err, "error dl disk")
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)

		return nil, errors.Errorf("http error while downloading disk (%s)", string(b))
	}

	return resp.Body, nil
}
