package main

import (
	"bytes"
	"fmt"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {
	// use a bytes Buffer as a writer
	var buf = &bytes.Buffer{}

	// create a simple logger using the buffer as output
	var logger = log.New(
		log.WithOut(buf),
	)

	// log messages with the Logger as usual
	logger.Log(
		event.New().Message("buffered log entry").Build(),
	)

	// or use the logger as a Writer, to that buffer
	n, err := logger.Write(
		event.New().Message("another buffered log entry").Build().Encode(), // event.Encode()  method to convert msg to []byte
	)

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error: %v", err)
	}

	if n == 0 {
		// zlog's standard logger warn
		log.Warn("zero bytes written")
	}

	// print out buffer's content
	fmt.Println(buf.String())

}
