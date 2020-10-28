package httplog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Hook is the actual struct which can be given to logrus.
type Hook struct {
	url    string
	origin string
}

type logMessage struct {
	Level   string
	Message string
	Origin  string
}

// Levels implementation for the http log hook.
func (h Hook) Levels() []log.Level {
	return log.AllLevels
}

// Fire implementation for the http log hook.
func (h Hook) Fire(entry *log.Entry) error {
	res, err := json.Marshal(logMessage{
		Level:   entry.Level.String(),
		Message: entry.Message,
		Origin:  h.origin,
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

// NewLogHook Creates a new http log hook, to be given to logrus.
func NewLogHook(url, origin string) *Hook {
	return &Hook{
		url,
		origin,
	}
}
