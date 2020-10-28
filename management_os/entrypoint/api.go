package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"baas/pkg/model"

	"github.com/pkg/errors"

	"baas/pkg/api"
)

// APIClient is the client for all communication with the server
type APIClient struct {
	baseURL string
}

// NewAPIClient creates a new APIClient struct
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
	}
}

// BootInform informs the server that we have booted
func (a *APIClient) BootInform() (*api.ReprovisioningInfo, error) {
	log.Debug("Sending boot inform request")

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
			log.Errorf("Failed to close body (%v)", err)
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
func (a *APIClient) DownloadDiskHTTP(uuid model.DiskUUID) (io.ReadCloser, error) {
	log.Debugf("downloading disk %v over http", uuid)

	//nolint we are returning a readcloser so the body will be closed later
	resp, err := http.Get(fmt.Sprintf("%s/mmos/disk/%s", a.baseURL, uuid))
	if err != nil {
		return nil, errors.Wrap(err, "error dl disk")
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)

		return nil, errors.Errorf("http error while downloading disk (%s)", string(b))
	}

	log.Debugf("done downloading disk %v over http", uuid)

	return resp.Body, nil
}

// UploadDiskHTTP uploads a disk image given the http strategy
func (a *APIClient) UploadDiskHTTP(r io.ReadCloser, uuid model.DiskUUID) error {
	log.Debugf("uploading disk %v over http", uuid)

	// Post closes r if able to, so no manual close is necessary
	resp, err := http.Post(fmt.Sprintf("%s/mmos/disk/%s", a.baseURL, uuid), "application/octet-stream", r)
	if err != nil {
		return errors.Wrap(err, "upload disk")
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)

		return errors.Errorf("upload disk http (%s)", string(b))
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("Failed to close reader (%v)", err)
		}
	}()

	log.Debugf("done uploading disk %v over http", uuid)

	return nil
}
