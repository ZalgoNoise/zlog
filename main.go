package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/store"
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

	customLog, err := store.NewLogfile("/tmp/test-custom.log")
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
		log.NewTextFormat().Time(log.LTRubyDate).LevelFirst().DoubleSpace().Color().Upper().Build(),
		log.SkipExitCfg,
	)

	newLogger.Log(log.NewMessage().Level(log.LLPanic).Prefix("color").Sub("logger-service").Message("hello universe!").Build())

	n, err := newLogger.Write(log.NewMessage().
		Message("gob-encoded tester").
		Level(log.LLFatal).
		Prefix("binary").
		Metadata(log.Field{
			"test-data":     "with content",
			"another-field": "also with content",
			"obj": log.Field{
				"nested": true,
			},
			"last-one": true,
		}).
		Build().
		Bytes(),
	)

	fmt.Println(n, err)

	csvLogger := log.New(
		log.WithPrefix("csv-logger"),
		log.WithOut(os.Stdout),
		log.CSVFormat,
		log.SkipExitCfg,
	)

	csvLogger.Log(log.NewMessage().Sub("CSV").Message("hello from CSV!").Build())
	csvLogger.Log(log.NewMessage().Prefix("csv-test").Sub("CSV").Message("hello from CSV with custom prefix").Build())
	csvLogger.Log(log.NewMessage().Level(log.LLPanic).Sub("CSV").Message("hello from CSV with custom level").Build())
	csvLogger.Log(
		log.NewMessage().
			Prefix("test-all").
			Sub("CSV").
			Level(log.LLWarn).
			Message("hello from CSV with all of it").
			Metadata(log.Field{
				"content":   true,
				"test-data": "this is test data",
				"inner-field": log.Field{
					"custom":  true,
					"content": "yes",
				},
			}).
			Build(),
	)

	xmlLogger := log.New(
		log.WithPrefix("xml-logger"),
		log.WithOut(os.Stdout),
		log.XMLFormat,
		log.SkipExitCfg,
	)

	xmlLogger.Log(log.NewMessage().Sub("XML").Message("hello from XML!").Build())
	xmlLogger.Log(log.NewMessage().Prefix("xml-test").Sub("XML").Message("hello from XML with custom prefix").Build())
	xmlLogger.Log(log.NewMessage().Level(log.LLPanic).Sub("XML").Message("hello from XML with custom level").Build())
	xmlLogger.Log(
		log.NewMessage().
			Prefix("test-all").
			Sub("XML").
			Level(log.LLWarn).
			Message("hello from XML with all of it").
			Metadata(log.Field{
				"content":   true,
				"test-data": "this is test data",
				"inner-field": log.Field{
					"custom":  true,
					"content": "yes",
				},
			}).
			Build(),
	)

	filteredLogger := log.New(
		log.WithPrefix("filtered-logger"),
		log.WithOut(os.Stdout),
		log.NewTextFormat().LevelFirst().Color().Upper().Build(),
		log.SkipExitCfg,
		log.WithFilter(log.LLError),
	)

	filteredLogger.Log(log.NewMessage().Level(log.LLTrace).Prefix("filter").Sub("logger-service").Message("trace").Build())
	filteredLogger.Log(log.NewMessage().Level(log.LLWarn).Prefix("filter").Sub("logger-service").Message("warn").Build())
	filteredLogger.Log(log.NewMessage().Level(log.LLError).Prefix("filter").Sub("logger-service").Message("error").Build())
	filteredLogger.Log(log.NewMessage().Level(log.LLFatal).Prefix("filter").Sub("logger-service").Message("fatal").Build())
	filteredLogger.Log(log.NewMessage().Level(log.LLPanic).Prefix("filter").Sub("logger-service").Message("panic").Build())

	stackMessage := log.NewMessage().
		Level(log.LLFatal).
		Prefix("with-stack").
		Sub("stacktrace").
		Message("failed to execute with error").
		Metadata(log.Field{
			"critical": true,
		}).
		CallStack(true).
		Build()

	filteredLogger.Log(stackMessage)

	jsonLogger := log.New(
		log.JSONCfg,
		log.SkipExitCfg,
		log.WithOut(os.Stdout),
	)

	jsonLogger.Log(stackMessage)
	fmt.Println()
	xmlLogger.Log(stackMessage)
}
