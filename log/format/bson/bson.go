package bson

import (
	"time"

	"github.com/zalgonoise/zlog/log/event"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FmtBSON struct describes the different manipulations and processing that a BSON LogFormatter
// can apply to an event.Event
type FmtBSON struct{}

type entry struct {
	Time   time.Time              `bson:"timestamp,omitempty"`
	Prefix string                 `bson:"service,omitempty"`
	Sub    string                 `bson:"module,omitempty"`
	Level  string                 `bson:"level,omitempty"`
	Msg    string                 `bson:"message,omitempty"`
	Meta   map[string]interface{} `bson:"metadata,omitempty"`
}

// Format method will take in a pointer to an event.Event; and returns a buffer and an error.
//
// This method will process the input event.Event and marshal it according to this LogFormatter
func (f *FmtBSON) Format(log *event.Event) (buf []byte, err error) {
	// remove trailing newline on JSON format
	if log.GetMsg()[len(log.GetMsg())-1] == 10 {
		*log.Msg = log.GetMsg()[:len(log.GetMsg())-1]
	}

	return bson.Marshal(entry{
		Time:   log.GetTime().AsTime(),
		Prefix: log.GetPrefix(),
		Sub:    log.GetSub(),
		Level:  log.GetLevel().String(),
		Msg:    log.GetMsg(),
		Meta:   log.Meta.AsMap(),
	})
}

func Decode(b []byte) (*event.Event, error) {
	ent := new(entry)

	err := bson.Unmarshal(b, ent)
	if err != nil {
		return nil, err
	}

	e := event.New().
		Level(event.Level(event.Level_value[ent.Level])).
		Prefix(ent.Prefix).
		Sub(ent.Sub).
		Message(ent.Msg).
		Metadata(ent.Meta).
		Build()

	e.Time = timestamppb.New(ent.Time)
	return e, nil

}
