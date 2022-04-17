package json

import (
	"encoding/json"

	"github.com/zalgonoise/zlog/log/event"
)

// FmtJSON struct describes the different manipulations and processing that a JSON LogFormatter
// can apply to a LogMessage
type FmtJSON struct{}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtJSON) Format(log *event.Event) (buf []byte, err error) {
	// remove trailing newline on JSON format
	if log.Msg[len(log.Msg)-1] == 10 {
		log.Msg = log.Msg[:len(log.Msg)-1]
	}

	return json.Marshal(log)
}
