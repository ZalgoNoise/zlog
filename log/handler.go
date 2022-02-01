package log

import (
	"fmt"
	"io"
	"os"
)

type logOutput struct {
	output io.Writer
	errors io.Writer
}

// basic output types (implementing io.Reader and io.Writer interfaces)
// - os.StdOut / os.StdErr
// - to file (TODO)
// - to HTTP req (TODO)
var logOutputKeys = map[int]string{
	0: "console",
	// 1: "file",
	// 2: "http",
}

var consoleOut = logOutput{
	output: os.Stdout,
	errors: os.Stderr,
}

type Log struct {
	outputs    []logOutput
	logMessage LogMessage
}

func (l *Log) Log(level int, msg string) {
	// log level needs to be set to logMessage and set within its Write method
	//     snippet below is overwritten by its following buffer
	//
	// _, _ = l.logMessage.Write([]byte(logTypeVals[level]))

	n, err := l.logMessage.Write([]byte(msg))

	if n == 0 {
		return
	}

	if err != nil {
		_, _ = l.logMessage.Write([]byte("logger error while writing input message to buffer"))
	}

	// TODO:
	// append contents to file, don't overwrite
	for idx, _ := range l.outputs {
		var out io.Writer

		if level > 3 {
			out = l.outputs[idx].errors
		} else {
			out = l.outputs[idx].output
		}

		_, err := out.Write(l.logMessage.output)

		if err != nil {
			fmt.Printf("[LOGGER][ERR] Unable to write data to output stream")
		}
	}

}

//TODO: add a setter method for the outputs; needs to be io.ReadWriteCloser

func New() *Log {
	return &Log{
		outputs: []logOutput{consoleOut},
	}
}
