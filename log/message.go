package log

import (
	"errors"
	"io"
)

type Logger interface {
	io.ReadWriter
}

var logTypeKeys = map[string]int{
	"trace": 0,
	"debug": 1,
	"info":  2,
	"warn":  3,
	"error": 4,
	"fatal": 5,
	"panic": 9,
}

var logTypeVals = map[int]string{
	0: "trace",
	1: "debug",
	2: "info",
	3: "warn",
	4: "error",
	5: "fatal",
	9: "panic",
}

//TODO:
// - add timestamp
// - add data joining method (l.Write should build a message with its metadata)
//
type LogMessage struct {
	logType string
	logMsg  []byte
}

func (l *LogMessage) Read(p []byte) (n int, err error) {
	if len(l.logMsg) > 0 {
		p = l.logMsg
		return len(p), nil
	}
	if len(p) == 0 {
		return 0, nil
	}

	return -1, errors.New("invalid data")
}

func (l *LogMessage) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		l.logMsg = p
		return len(l.logMsg), nil
	}

	if len(p) == 0 {
		return 0, nil
	}

	return -1, errors.New("invalid data")
}
