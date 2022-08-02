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
		log.Fatalf("unexpected error creating temp file: %v", err)
	}
	defer os.Remove(tempF.Name())

	// use the temp file as a LogFile
	logF, err := fs.New(
		tempF.Name(),
	)
	if err != nil {
		log.Fatalf("unexpected error creating logfile: %v", err)
	}

	// set up a max size in MB, to auto-rotate
	logF.MaxSize(50)

	// create a simple logger using the logfile as output, log messages to it
	var logger = log.New(
		log.WithOut(logF),
	)
	logger.Log(
		event.New().Message("log entry written to file").Build(),
	)

	// print out file's's content
	b, err := os.ReadFile(tempF.Name())
	if err != nil {
		log.Fatalf("unexpected error reading logfile's data: %v", err)
	}

	fmt.Println(string(b))
}
