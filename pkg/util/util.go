// Package util defines various functions and types which are useful throughout the program
package util

import (
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"
)

// ProgressReporter is a struct which contains a reader which records the progress made
// Idea and basic implementation taken from:
// https://stackoverflow.com/questions/26050380/go-tracking-post-request-progress
type ProgressReporter struct {
	R     io.Reader
	Max   uint
	sent  int
	atEOF bool
}

func (pr *ProgressReporter) Read(p []byte) (int, error) {
	n, err := pr.R.Read(p)
	pr.sent += n

	if err == io.EOF {
		pr.atEOF = true
	}

	if n%1_048_567 == 0 {
		pr.report()
	}
	return n, err
}

func (pr *ProgressReporter) report() {
	log.Infof("sent %d of %d bytes\n", pr.sent, pr.Max)
	if pr.atEOF {
		log.Info("DONE")
	}
}

// PrettyPrintStruct prints a nice looking version of a struct
// TODO: Give this a better place in the code than randomly inside API
func PrettyPrintStruct(a interface{}) {
	// If I had a nickel for every time that the best way in a language to pretty print a datastructure is to cast it into a JSON
	// structure and printing that, I would have two nickels. That is not a lot, but it is funny that it happened twice.
	s, _ := json.MarshalIndent(a, "", "\t")
	log.Info(string(s))
}
