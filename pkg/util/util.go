// Package util defines various functions and types which are useful throughout the program
package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"strconv"
	"strings"

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
func PrettyPrintStruct(a interface{}) {
	// If I had a nickel for every time that the best way in a language to pretty print a datastructure is to cast it into a JSON
	// structure and printing that, I would have two nickels. That is not a lot, but it is funny that it happened twice.
	s, _ := json.MarshalIndent(a, "", "\t")
	log.Info(string(s))
}

// MacAddress is a structure containing the unique Mac Address
type MacAddress struct {
	Address string
}

func (mac MacAddress) GormDataType() string {
	return "INTEGER"
}

func (mac MacAddress) GormValue(_ context.Context, _ *gorm.DB) clause.Expr {
	hex, err := strconv.ParseUint(strings.ReplaceAll(mac.Address, ":", ""), 16, 64)
	if err != nil {

	}
	return clause.Expr{
		SQL:  "?",
		Vars: []interface{}{fmt.Sprintf("%d", hex)},
	}
}

func (mac *MacAddress) Scan(v interface{}) error {
	bs, ok := v.(int64)
	if !ok {
		return errors.New("cannot parse mac address")
	}

	builder := strings.Builder{}
	num := fmt.Sprintf("%x", bs)
	for i, v := range []byte(num) {
		if i != 0 && i%2 == 0 {
			builder.WriteByte(':')
		}
		builder.WriteByte(v)
	}
	mac.Address = builder.String()
	return nil

}
