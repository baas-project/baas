package httplog

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func CreateLogHandler(logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var l LogMessage

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




