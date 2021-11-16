// Adapted: https://stackoverflow.com/questions/48971780/how-to-change-the-format-of-log-output-in-logrus
package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// BaasFormatter is a custom format which is used in the project
type BaasFormatter struct {
	log.TextFormatter
}

// ANSI Colours
const (
	Black    int = 30
	Red      int = 31
	Green    int = 32
	Yellow   int = 33
	Blue     int = 34
	Magenta  int = 35
	Cyan     int = 36
	White    int = 37
	BBlack   int = 90
	BRed     int = 91
	BGreen   int = 92
	BYellow  int = 93
	GBlue    int = 94
	GMagenta int = 95
	GCyan    int = 97
	GWhite   int = 98
)

// Format formats an entry in a custom format
func (f *BaasFormatter) Format(entry *log.Entry) ([]byte, error) {
	var levelColor int

	// These are ANSI colours
	switch entry.Level {
	case log.DebugLevel, log.TraceLevel:
		levelColor = BBlack // gray
	case log.WarnLevel:
		levelColor = Yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		levelColor = Red
	default:
		levelColor = Blue
	}

	return []byte(fmt.Sprintf("[%s] - \x1b[%dm%s\x1b[0m - %s\n",
		entry.Time.Format(f.TimestampFormat), levelColor,
		strings.ToUpper(entry.Level.String()), entry.Message)), nil
}
