package log

import (
	"encoding/json"
	"errors"
	"io"
	"time"
)

type Logger interface {
	io.Writer
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

var formatTypeKeys = map[string]int{
	"text": 0,
	"json": 1,
}

var formatTypeVals = map[int]string{
	0: "text",
	1: "json",
}

//TODO:
// - add timestamp
// - add data joining method (l.Write should build a message with its metadata)
//
type LogMessage struct {
	Service   string                 `json:"service,omitempty"`
	Module    string                 `json:"module,omitempty"`
	LogType   string                 `json:"type,omitempty"`
	LogMsg    []byte                 `json:"message,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	format    int
	output    []byte
}

// func (l *LogMessage) Read(p []byte) (n int, err error) {
// 	if len(l.logMsg) > 0 {
// 		p = []byte(l.logMsg)
// 		return len(p), nil
// 	}
// 	if len(p) == 0 {
// 		return 0, nil
// 	}

// 	return -1, errors.New("invalid data")
// }

func (l *LogMessage) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		l.SetTime()

		if l.LogType == "" {
			l.SetLevel(2)
		}

		l.LogMsg = p

		l.build()

		return len(p), nil
	}

	if len(p) == 0 {
		return 0, nil
	}

	return -1, errors.New("invalid data")
}

func (l *LogMessage) Fields(fields map[string]interface{}) {
	l.Metadata = fields
}

func (l *LogMessage) SetLevel(level int) {
	l.LogType = logTypeVals[level]
}

func (l *LogMessage) SetFormat(fmt string) {
	l.format = formatTypeKeys[fmt]
}

func (l *LogMessage) SetTime() {
	l.Timestamp = time.Now().String()
}

func (l *LogMessage) build() {
	if l.format == 0 {
		l.output = []byte(
			"[" + l.Service + "]" +
				"[" + l.Module + "]" +
				"\t[" + l.LogType + "] " +
				string(l.LogMsg) + "\t" +
				"[" + l.Timestamp + "]\n",
		)
	} else if l.format == 1 {
		buf, err := json.Marshal(l)
		if err != nil {
			l.output = []byte{}
		}
		l.output = buf
	}
}
