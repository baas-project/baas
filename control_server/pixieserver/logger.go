// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pixieserver

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

var logSync sync.Mutex

func logWithStdLog(subsys, msg string) {
	logSync.Lock()
	defer logSync.Unlock()
	log.Debugf("[%s] %s", subsys, msg)
}

func logWithStdFmt(subsys, msg string) {
	logSync.Lock()
	defer logSync.Unlock()
	fmt.Printf("[%s] %s\n", subsys, msg)
}
