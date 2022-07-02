package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
)

func main() {

	// setup a simple text logger, with custom formatting
	custTextLogger := log.New(
		log.WithFormat(
			text.New().
				Color().
				DoubleSpace().
				LevelFirst().
				Upper().
				Time(text.LTRubyDate).
				Build(),
		),
	)

	// setup a simple JSON logger
	jsonLogger := log.New(log.CfgFormatJSON)

	// setup a simple XML logger
	xmlLogger := log.New(log.CfgFormatXML)

	// setup a simple CSV logger
	csvLogger := log.New(log.CfgFormatCSV)

	// setup a simple BSON logger
	bsonLogger := log.New(log.CfgFormatBSON)

	// setup a simple protobuf logger
	pbLogger := log.New(log.CfgFormatProtobuf)

	// setup a simple Gob logger
	gobLogger := log.New(log.CfgFormatGob)

	// join all loggers
	multiLogger := log.MultiLogger(
		custTextLogger,
		jsonLogger,
		xmlLogger,
		csvLogger,
		bsonLogger,
		pbLogger,
		gobLogger,
	)

	// example message to print
	var msg = event.New().Message("message from a formatted logger").Build()

	// print the message to standard out, with different formats
	multiLogger.Log(msg)
}
