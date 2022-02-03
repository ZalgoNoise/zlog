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

	logger := log.New("test-logs", &log.JSONFmt{})

	logger.SetOuts(os.Stdout)
	logger.AddOuts(logFile1, logFile2)
	logger.Infoln("test log")
	logger.Debugf("%v\n", len(data))
	logger.Warnln("big warning")
	logger.Fields(map[string]interface{}{
		"path":  "/src/srv/stack",
		"error": 9,
		"proc": map[string]interface{}{
			"test": true,
		},
	}).Warnln("urgent error")
	// log.Panicln("i am out")

}
