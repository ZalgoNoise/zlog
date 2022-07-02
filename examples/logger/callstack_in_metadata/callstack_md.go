package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

// set up an indented JSON logger; package-level as an example
//
// you'd set this up within your logic however it's convenient
var logger = log.New(log.CfgFormatJSONIndent)

// placeholder operation for visibility in the callstack
func operation(value int) bool {
	return subOperation(value)
}

// placeholder sub-operation for visibility in the callstack
//
// error is printed whenever input is zero
func subOperation(value int) bool {
	if value == 0 {
		logger.Log(
			event.New().
				Level(event.Level_error).
				Message("operation failed").
				Metadata(event.Field{
					"error": "input cannot be zero", // custom metadata
					"input": value,                  // custom metadata
				}).
				CallStack(true). // add (complete) callstack to metadata
				Build(),
		)
		return false
	}
	return true
}

func main() {
	// all goes well until something happens within your application
	for a := 5; a >= 0; a-- {
		if operation(a) {
			continue
		}
		break
	}
}
