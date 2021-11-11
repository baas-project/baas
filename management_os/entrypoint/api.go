package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/baas-project/baas/pkg/model"

	"github.com/pkg/errors"

	"github.com/baas-project/baas/pkg/api"
	//	"github.com/baas-project/baas/pkg/util"
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

	resp, err := http.Get(url)
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
func (a *APIClient) DownloadDiskHTTP(uuid model.DiskUUID, version uint) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/image/%s/%d", a.baseURL, uuid, version)
	log.Infof("downloading disk %v over http from %s", uuid, url)

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
func (a *APIClient) UploadDiskHTTP(r io.Reader, uuid model.DiskUUID) error {
	url := fmt.Sprintf("%s/image/%s", a.baseURL, uuid)
	log.Debugf("uploading disk %v over http to %s", uuid, url)

	// Create a pipe so we only pass around the streams, if we try to write the actual file or program will be reaped
	// Dammit Calli
	boundary := "JanRellermeyer"
	fileName := "image.img"
	fileHeader := "Content-Type: application/octet-stream"
	fileFormat := "--%s\r\nContent-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\n%s\r\n\r\n"
	filePart := fmt.Sprintf(fileFormat, boundary, fileName, fileHeader)
	end := fmt.Sprintf("\r\n--%s\r\n", boundary)

	body := io.MultiReader(strings.NewReader(filePart), r, strings.NewReader(end))

	// TODO: Do not work with a max but get the file size
	resp, err := http.Post(url, fmt.Sprintf("multipart/form-data; boundary=%s", boundary),
		body)

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
