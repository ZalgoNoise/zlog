package main

import (
	"fmt"
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

	customLog, err := log.NewLogfile("/tmp/test-custom")
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}

	customLog.MaxSize(5)

	multi := log.MultiLogger(

		log.New("alpha-log", &log.TextFmt{}),
		log.New("beta-log", &log.JSONFmt{}, logFile1, logFile2, customLog),
	)

	multi.Info("test log")
	multi.SetPrefix("multi-logs").Debugf("%v", len(data))
	multi.Warn("multi warning")
	multi.SetPrefix("prod-logs").Fields(map[string]interface{}{
		"path":  "/src/srv/stack",
		"error": 9,
		"proc": map[string]interface{}{
			"test": true,
		},
	}).Warn("urgent error")

}
