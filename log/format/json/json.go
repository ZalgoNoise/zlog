package json

import (
	"encoding/json"
	"time"

	"github.com/zalgonoise/zlog/log/event"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FmtJSON struct describes the different manipulations and processing that a JSON LogFormatter
// can apply to an event.Event
type FmtJSON struct{}

type entry struct {
	Time   time.Time              `json:"timestamp,omitempty"`
	Prefix string                 `json:"service,omitempty"`
	Sub    string                 `json:"module,omitempty"`
	Level  string                 `json:"level,omitempty"`
	Msg    string                 `json:"message,omitempty"`
	Meta   map[string]interface{} `json:"metadata,omitempty"`
}

func Decode(b []byte) (*event.Event, error) {
	e := new(entry)

	err := json.Unmarshal(b, e)

	if err != nil {
		return nil, err
	}

	log := event.New().
		Prefix(e.Prefix).
		Sub(e.Sub).
		Level(event.Level(event.Level_value[e.Level])).
		Message(e.Msg).
		Metadata(e.Meta).
		Build()

	log.Time = timestamppb.New(e.Time)

	return log, nil
}

// Format method will take in a pointer to an event.Event; and returns a buffer and an error.
//
// This method will process the input event.Event and marshal it according to this LogFormatter
func (f *FmtJSON) Format(log *event.Event) (buf []byte, err error) {
	// remove trailing newline on JSON format
	if log.GetMsg()[len(log.GetMsg())-1] == 10 {
		*log.Msg = log.GetMsg()[:len(log.GetMsg())-1]
	}

	return json.Marshal(entry{
		Time:   log.GetTime().AsTime(),
		Prefix: log.GetPrefix(),
		Sub:    log.GetSub(),
		Level:  log.GetLevel().String(),
		Msg:    log.GetMsg(),
		Meta:   log.Meta.AsMap(),
	})
}
