package bson

import (
	"time"

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

	type Event struct {
		Time   time.Time              `bson:"timestamp,omitempty"`
		Prefix string                 `bson:"service,omitempty"`
		Sub    string                 `bson:"module,omitempty"`
		Level  string                 `bson:"level,omitempty"`
		Msg    string                 `bson:"message,omitempty"`
		Meta   map[string]interface{} `bson:"metadata,omitempty"`
	}

	return bson.Marshal(Event{
		Time:   log.GetTime().AsTime(),
		Prefix: log.GetPrefix(),
		Sub:    log.GetSub(),
		Level:  log.GetLevel().String(),
		Msg:    log.GetMsg(),
		Meta:   log.Meta.AsMap(),
	})
}
