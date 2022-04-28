package gob

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/zalgonoise/zlog/log/event"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FmtGob struct allows marshalling an event.Event as gob-encoded bytes
type FmtGob struct{}

type entry struct {
	Time   time.Time
	Prefix string
	Sub    string
	Level  string
	Msg    string
	Meta   map[string]interface{}
}

func Decode(b []byte) (*event.Event, error) {
	e := new(entry)

	buf := bytes.NewBuffer(b)

	dec := gob.NewDecoder(buf)

	err := dec.Decode(e)

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

// Format method will take in a pointer to an event.Event, and return the execution of its
// encode() method; which converts it to gob-encoded bytes, returning a slice of bytes and an
// error
func (f *FmtGob) Format(log *event.Event) ([]byte, error) {

	buf := &bytes.Buffer{}

	gob.Register(entry{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register([]map[string]interface{}{})

	e := entry{
		Time:   log.GetTime().AsTime(),
		Prefix: log.GetPrefix(),
		Sub:    log.GetSub(),
		Level:  log.GetLevel().String(),
		Msg:    log.GetMsg(),
		Meta:   log.Meta.AsMap(),
	}

	enc := gob.NewEncoder(buf)

	err := enc.Encode(e)

	return buf.Bytes(), err

}
