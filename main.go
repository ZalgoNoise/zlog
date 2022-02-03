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

	logger.SetOuts(os.Stdout).AddOuts(logFile1, logFile2)

	logger.Info("test log")
	logger.SetPrefix("debug-logs").Debugf("%v", len(data))
	logger.Warn("big warning")
	logger.SetPrefix("prod-logs").Fields(map[string]interface{}{
		"path":  "/src/srv/stack",
		"error": 9,
		"proc": map[string]interface{}{
			"test": true,
		},
	}).Warn("urgent error")
	// log.Panicln("i am out")

}
