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

	logger := log.New("test-logs", &log.TextFmt{})

	logger.SetOuts(os.Stdout)
	logger.AddOuts(logFile1, logFile2)
	logger.Infoln("test log")
	logger.Debugf("%v\n", len(data))
	logger.Warnln("big warning")
	// log.Panicln("i am out")

}
