package bson

import (
	"github.com/zalgonoise/zlog/log/event"
	"go.mongodb.org/mongo-driver/bson"
)

// FmtBSON struct describes the different manipulations and processing that a BSON LogFormatter
// can apply to a LogMessage
type FmtBSON struct{}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtBSON) Format(log *event.Event) (buf []byte, err error) {
	// remove trailing newline on JSON format
	if log.GetMsg()[len(log.GetMsg())-1] == 10 {
		*log.Msg = log.GetMsg()[:len(log.GetMsg())-1]
	}

	return bson.Marshal(log)
}
