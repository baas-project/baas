// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httplog provides a log for logrus, which enables log lines to be sent
// using http requests.
package httplog

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// CreateLogHandler creates a new log handler. It takes a logger as parameter,
// and returns a HandlerFunc which can be used with the go http framework.
func CreateLogHandler(logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var l logMessage

		err := json.NewDecoder(r.Body).Decode(&l)
		if err != nil {
			log.Errorf("failed to read log message (%v)", err)
		}

		level, err := log.ParseLevel(l.Level)
		if err != nil {
			log.Errorf("failed to parse log level (%v)", err)
		}

		switch level {
		case log.PanicLevel, log.FatalLevel, log.ErrorLevel:
			logger.Errorf("[%s] %s", l.Origin, l.Message)
		case log.WarnLevel:
			logger.Warnf("[%s] %s", l.Origin, l.Message)
		case log.DebugLevel:
			logger.Debugf("[%s] %s", l.Origin, l.Message)
		case log.TraceLevel:
			logger.Tracef("[%s] %s", l.Origin, l.Message)
		default:
			logger.Infof("[%s] %s", l.Origin, l.Message)
		}
	}
}
