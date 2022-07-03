package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/fs"
)

func main() {
	// create a new temporary file in /tmp
	tempF, err := ioutil.TempFile("/tmp", "zlog_test_fs-")

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error creating temp file: %v", err)
	}

	// cleanup -- remove temp file
	defer os.Remove(tempF.Name())

	// use the temp file as a LogFile
	logF, err := fs.New(
		tempF.Name(),
	)

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error creating temp file: %v", err)
	}

	// set up a max size in MB, to auto-rotate
	logF.MaxSize(50)

	// create a simple logger using the logfile as output
	var logger = log.New(
		log.WithOut(logF),
	)

	// log messages with the Logger as usual
	logger.Log(
		event.New().Message("log entry written to file").Build(),
	)

	// or use the logger as a Writer, to that logfile
	n, err := logger.Write(
		event.New().Message("another log entry written to file").Build().Encode(), // event.Encode()  method to convert msg to []byte
	)

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error: %v", err)
	}

	if n == 0 {
		// zlog's standard logger warn
		log.Warn("zero bytes written")
	}

	// print out file's's content
	b, err := os.ReadFile(tempF.Name())

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error: %v", err)
	}

	fmt.Println(string(b))

}
