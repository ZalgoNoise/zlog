package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
	"github.com/zalgonoise/zlog/store/fs"
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

	customLog, err := fs.New("/tmp/test-custom.log")
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}

	customLog.MaxSize(5)

	multi := log.MultiLogger(

		log.New(
			log.WithPrefix("alpha-log"),
			log.WithFormat(log.FormatText),
		),
		log.New(
			log.WithPrefix("beta-log"),
			log.WithFormat(log.FormatJSON),
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
			"var":    "event.Level_trace",
			"val":    event.Level_trace,
		},
	).Log(
		event.New().Level(event.Level_trace).Message("this is a custom level log entry").Build(),
	)

	chlogger := log.NewLogCh(multi)
	defer chlogger.Close()

	logCh, _ := chlogger.Channels()

	logCh <- event.New().Prefix("test-chan-log").Level(event.Level_trace).Message("test log message").Metadata(
		map[string]interface{}{
			"type":    "trace",
			"data":    "this is a buffered logger in a goroutine",
			"test_id": 0,
		},
	).Build()

	chlogger.Log(event.New().Prefix("test-chan-log").Level(event.Level_trace).Message("test log message").Metadata(
		map[string]interface{}{
			"type":    "trace",
			"data":    "this is a buffered logger in a goroutine",
			"test_id": 1,
		},
	).Build())

	chlogger.Log(
		event.New().Prefix("test-chan-log").Level(event.Level_trace).Message("test log message").Metadata(
			map[string]interface{}{
				"type":    "trace",
				"data":    "this is a buffered logger in a goroutine",
				"test_id": 2,
			},
		).Build(),
		event.New().Prefix("test-chan-log").Level(event.Level_warn).Message("warn runtime").Metadata(
			map[string]interface{}{
				"type":    "warn",
				"data":    "this is a buffered logger in a goroutine",
				"test_id": 3,
			},
		).Build(),
	)

	// go func() {
	// 	logCh <- event.New().Prefix("test-chan-log").Level(event.Level_panic).Message("break runime").Metadata(
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
		log.WithFormat(text.New().Time(text.LTRubyDate).LevelFirst().DoubleSpace().Color().Upper().Build()),
		log.SkipExit,
	)

	newLogger.Log(event.New().Level(event.Level_panic).Prefix("color").Sub("logger-service").Message("hello universe!").Build())

	n, err := newLogger.Write(event.New().
		Message("gob-encoded tester").
		Level(event.Level_fatal).
		Prefix("binary").
		Metadata(event.Field{
			"test-data":     "with content",
			"another-field": "also with content",
			"obj": event.Field{
				"nested": true,
			},
			"last-one": true,
		}).
		Build().
		Encode(),
	)

	fmt.Println(n, err)

	csvLogger := log.New(
		log.WithPrefix("csv-logger"),
		log.WithOut(os.Stdout),
		log.WithFormat(log.FormatCSV),
		log.SkipExit,
	)

	csvLogger.Log(event.New().Sub("CSV").Message("hello from CSV!").Build())
	csvLogger.Log(event.New().Prefix("csv-test").Sub("CSV").Message("hello from CSV with custom prefix").Build())
	csvLogger.Log(event.New().Level(event.Level_panic).Sub("CSV").Message("hello from CSV with custom level").Build())
	csvLogger.Log(
		event.New().
			Prefix("test-all").
			Sub("CSV").
			Level(event.Level_warn).
			Message("hello from CSV with all of it").
			Metadata(event.Field{
				"content":   true,
				"test-data": "this is test data",
				"inner-field": event.Field{
					"custom":  true,
					"content": "yes",
				},
			}).
			Build(),
	)

	xmlLogger := log.New(
		log.WithPrefix("xml-logger"),
		log.WithOut(os.Stdout),
		log.WithFormat(log.FormatXML),
		log.SkipExit,
	)

	xmlLogger.Log(event.New().Sub("XML").Message("hello from XML!").Build())
	xmlLogger.Log(event.New().Prefix("xml-test").Sub("XML").Message("hello from XML with custom prefix").Build())
	xmlLogger.Log(event.New().Level(event.Level_panic).Sub("XML").Message("hello from XML with custom level").Build())
	xmlLogger.Log(
		event.New().
			Prefix("test-all").
			Sub("XML").
			Level(event.Level_warn).
			Message("hello from XML with all of it").
			Metadata(event.Field{
				"content":   true,
				"test-data": "this is test data",
				"inner-field": event.Field{
					"custom":  true,
					"content": "yes",
				},
			}).
			Build(),
	)

	filteredLogger := log.New(
		log.WithPrefix("filtered-logger"),
		log.WithOut(os.Stdout),
		log.WithFormat(text.New().LevelFirst().Color().Upper().Build()),
		log.WithFilter(event.Level_error),
		log.SkipExit,
	)

	filteredLogger.Log(event.New().Level(event.Level_trace).Prefix("filter").Sub("logger-service").Message("trace").Build())
	filteredLogger.Log(event.New().Level(event.Level_warn).Prefix("filter").Sub("logger-service").Message("warn").Build())
	filteredLogger.Log(event.New().Level(event.Level_error).Prefix("filter").Sub("logger-service").Message("error").Build())
	filteredLogger.Log(event.New().Level(event.Level_fatal).Prefix("filter").Sub("logger-service").Message("fatal").Build())
	filteredLogger.Log(event.New().Level(event.Level_panic).Prefix("filter").Sub("logger-service").Message("panic").Build())

	stackMessage := event.New().
		Level(event.Level_fatal).
		Prefix("with-stack").
		Sub("stacktrace").
		Message("failed to execute with error").
		Metadata(event.Field{
			"critical": true,
		}).
		CallStack(true).
		Build()

	filteredLogger.Log(stackMessage)

	jsonLogger := log.New(
		log.WithFormat(log.FormatJSON),
		log.WithOut(os.Stdout),
		log.SkipExit,
	)

	jsonLogger.Log(stackMessage)
	fmt.Println()
	xmlLogger.Log(stackMessage)
}
