package pixieserver

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
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
