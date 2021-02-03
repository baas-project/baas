package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/baas-project/baas/pkg/model"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/api"
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
func (a *APIClient) BootInform(mac string) (*api.ReprovisioningInfo, error) {
	url := fmt.Sprintf("%s/machine/%s/boot", a.baseURL, mac)
	log.Debugf("Sending boot inform request to %s", url)

	b, err := json.Marshal(&api.BootInformRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't marshal boot inform json")
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
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
		return nil, errors.Errorf("inform request failed (%s) to %s", strings.TrimSpace(string(msg)), url)
	}

	var info api.ReprovisioningInfo

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, errors.Wrap(err, "couldn't deserialize inform request response")
	}

	return &info, nil
}

// DownloadDiskHTTP Downloads a disk image from the control_server over HTTP
func (a *APIClient) DownloadDiskHTTP(mac string, uuid model.DiskUUID) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/machine/%s/disk/%s", a.baseURL, mac, uuid)
	log.Debugf("downloading disk %v over http from %s", uuid, url)

	//nolint we are returning a readcloser so the body will be closed later
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "error dl disk")
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)

		return nil, errors.Errorf("http error while downloading disk (%s)", strings.TrimSpace(string(b)))
	}

	log.Debugf("done downloading disk %v over http", uuid)

	return resp.Body, nil
}

// UploadDiskHTTP uploads a disk image given the http strategy
func (a *APIClient) UploadDiskHTTP(r io.Reader, mac string, uuid model.DiskUUID) error {
	url := fmt.Sprintf("%s/machine/%s/disk/%s", a.baseURL, mac, uuid)
	log.Debugf("uploading disk %v over http to %s", uuid, url)

	// Post closes r if able to, so no manual close is necessary
	resp, err := http.Post(url, "application/octet-stream", r)
	if err != nil {
		return errors.Wrap(err, "upload disk")
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)

		return errors.Errorf("upload disk http (%s) to %s", strings.TrimSpace(string(b)), url)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Failed to close reader (%v)", err)
		}
	}()

	log.Debugf("done uploading disk %v over http", uuid)

	return nil
}
