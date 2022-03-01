package main

import (
	"bytes"
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

	customLog, err := log.NewLogfile("/tmp/test-custom.log")
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}

	customLog.MaxSize(5)

	multi := log.MultiLogger(

		log.New(
			log.WithPrefix("alpha-log"),
			log.TextCfg,
		),
		log.New(
			log.WithPrefix("beta-log"),
			log.JSONCfg,
			log.WithOut(
				logFile1, logFile2, customLog,
			)),
	)

	multi.Info("test log")
	multi.Prefix("multi-logs").Debugf("%v", len(data))
	multi.Warn("multi warning")
	multi.Prefix("prod-logs").Fields(map[string]interface{}{
		"path":  "/src/srv/stack",
		"error": 9,
		"proc": map[string]interface{}{
			"test": true,
		},
	}).Warn("urgent error")

	multi.Prefix("custom-level").Fields(
		map[string]interface{}{
			"level":  0,
			"string": "trace",
			"var":    "log.LLTrace",
			"val":    log.LLTrace,
		},
	).Log(
		log.NewMessage().Level(log.LLTrace).Message("this is a custom level log entry").Build(),
	)

	chlogger := log.NewLogCh(multi)
	defer chlogger.Close()

	logCh, _ := chlogger.Channels()

	logCh <- log.NewMessage().Prefix("test-chan-log").Level(log.LLTrace).Message("test log message").Metadata(
		map[string]interface{}{
			"type":    "trace",
			"data":    "this is a buffered logger in a goroutine",
			"test_id": 0,
		},
	).Build()

	chlogger.Log(log.NewMessage().Prefix("test-chan-log").Level(log.LLTrace).Message("test log message").Metadata(
		map[string]interface{}{
			"type":    "trace",
			"data":    "this is a buffered logger in a goroutine",
			"test_id": 1,
		},
	).Build())

	chlogger.Log(
		log.NewMessage().Prefix("test-chan-log").Level(log.LLTrace).Message("test log message").Metadata(
			map[string]interface{}{
				"type":    "trace",
				"data":    "this is a buffered logger in a goroutine",
				"test_id": 2,
			},
		).Build(),
		log.NewMessage().Prefix("test-chan-log").Level(log.LLWarn).Message("warn runtime").Metadata(
			map[string]interface{}{
				"type":    "warn",
				"data":    "this is a buffered logger in a goroutine",
				"test_id": 3,
			},
		).Build(),
	)

	// go func() {
	// 	logCh <- log.NewMessage().Prefix("test-chan-log").Level(log.LLPanic).Message("break runime").Metadata(
	// 		map[string]interface{}{
	// 			"type":    "panic",
	// 			"data":    "this is a goroutine panic into a logger in a goroutine",
	// 			"test_id": 4,
	// 		},
	// 	).Build()
	// }()

	var newBuf = &bytes.Buffer{}

	newLogger := log.New(
		log.WithPrefix("multi-conf"),
		log.WithOut(os.Stdout, newBuf),
		log.TextCfg,
	)

	newLogger.Log(log.NewMessage().Message("hello universe!").Build())

}
