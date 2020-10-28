package httplog

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type HttpLogHook struct {
	url string
	origin string
}

type LogMessage struct {
	Level string
	Message string
	Origin string
}

func (h HttpLogHook) Levels() []log.Level {
	return log.AllLevels
}

func (h HttpLogHook) Fire(entry *log.Entry) error {
	res, err := json.Marshal(LogMessage {
		Level: entry.Level.String(),
		Message: entry.Message,
		Origin: h.origin,
	})

	if err != nil {
		return errors.Wrap(err, "marshall log entry")
	}

	resp, err := http.Post(h.url, "application/json", bytes.NewReader(res))
	if err != nil {
		return errors.Wrap(err, "sending log")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("sending log failed: %s with status %d", body, resp.StatusCode)
	}

	err = resp.Body.Close()
	if err != nil {
		return errors.Wrap(err, "close body")
	}

	return nil
}


func NewLogHook(url, origin string) *HttpLogHook {
	return &HttpLogHook{
		url,
		origin,
	}
}
