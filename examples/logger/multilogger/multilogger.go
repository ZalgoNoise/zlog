package main

import (
	"bytes"
	"fmt"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {
	buf := new(bytes.Buffer)

	stdLogger := log.New() // default logger printing to stdErr
	jsonLogger := log.New( // custom JSON logger, writing to buffer
		log.WithOut(buf),
		log.CfgFormatJSONIndent,
	)

	// join both loggers
	logger := log.MultiLogger(
		stdLogger,
		jsonLogger,
	)

	// print messages to stderr
	logger.Info("some event occurring")
	logger.Warn("a warning pops-up")
	logger.Log(
		event.New().Level(event.Level_error).
			Message("and finally an error").
			Metadata(event.Field{
				"code":      5,
				"some-data": true,
			}).
			Build())

	// print buffer content
	fmt.Print("\n---\n- JSON data:\n---\n", buf.String())
}
