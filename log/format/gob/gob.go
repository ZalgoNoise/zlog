package gob

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/zalgonoise/zlog/log/event"
)

// FmtGob struct allows marshalling an event.Event as gob-encoded bytes
type FmtGob struct{}

// Format method will take in a pointer to an event.Event, and return the execution of its
// encode() method; which converts it to gob-encoded bytes, returning a slice of bytes and an
// error
func (f *FmtGob) Format(log *event.Event) ([]byte, error) {
	type Event struct {
		Time   time.Time
		Prefix string
		Sub    string
		Level  string
		Msg    string
		Meta   map[string]interface{}
	}

	buf := &bytes.Buffer{}

	gob.Register(Event{})
	gob.Register(map[string]interface{}{})

	e := Event{
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
