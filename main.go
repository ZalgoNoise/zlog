package main

import (
	"fmt"
	"os"
	"time"

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

	customLog, err := log.NewLogfile("/tmp/test-custom.log")
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

	multi.SetPrefix("custom-level").Fields(
		map[string]interface{}{
			"level":  0,
			"string": "trace",
			"var":    "log.LLTrace",
			"val":    log.LLTrace,
		},
	).Log(log.LLTrace, "this is a custom level log entry")

	chl := log.New("test", log.TextFormat)
	logCh := make(chan log.ChLogMessage)
	go func() {
		for {
			msg, ok := <-logCh
			if ok {
				chl.SetPrefix(msg.Prefix).Fields(msg.Metadata).Log(msg.Level, msg.Msg)
			} else {
				fmt.Println("channel closed")
				break
			}
		}
	}()

	logCh <- log.ChLogMessage{
		Prefix: "test-chan-log",
		Level:  log.LLTrace,
		Msg:    "test log message",
		Metadata: map[string]interface{}{
			"type":    "trace",
			"data":    "this is a buffered logger in a goroutine",
			"test_id": 0,
		},
	}

	time.Sleep(10 * time.Second)

	logCh <- log.ChLogMessage{
		Prefix: "test-chan-log",
		Level:  log.LLTrace,
		Msg:    "test log message",
		Metadata: map[string]interface{}{
			"type":    "trace",
			"data":    "this is a buffered logger in a goroutine",
			"test_id": 1,
		},
	}

	logCh <- log.ChLogMessage{
		Prefix: "test-chan-log",
		Level:  log.LLTrace,
		Msg:    "test log message",
		Metadata: map[string]interface{}{
			"type":    "trace",
			"data":    "this is a buffered logger in a goroutine",
			"test_id": 2,
		},
	}

}
