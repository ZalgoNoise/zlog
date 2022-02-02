package main

import (
	"os"

	"github.com/ZalgoNoise/zlog/log"
)

func main() {
	logFile1, err := os.Create("/tmp/test-log-1.log")
	if err != nil {
		panic(err)
	}
	logFile2, err := os.Create("/tmp/test-log-2.log")
	if err != nil {
		panic(err)
	}
	data := []int{
		2, 3, 5,
	}

	log := log.New("test-logs", 1)

	log.SetOuts(os.Stdout)
	log.AddOuts(logFile1, logFile2)

	log.Infoln("test log")
	log.Debugf("%v\n", len(data))
	log.Warnln("big warning")
	// log.Panicln("i am out")
}
