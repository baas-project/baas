package pixieserver

import (
	"fmt"
	"log"
	"sync"
)

var logSync sync.Mutex

func logWithStdLog(subsys, msg string) {
	logSync.Lock()
	defer logSync.Unlock()
	log.Printf("[%s] %s", subsys, msg)
}

func logWithStdFmt(subsys, msg string) {
	logSync.Lock()
	defer logSync.Unlock()
	fmt.Printf("[%s] %s\n", subsys, msg)
}
