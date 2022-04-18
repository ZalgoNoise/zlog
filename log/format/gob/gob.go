package gob

import (
	"bytes"
	"encoding/gob"

	"github.com/zalgonoise/zlog/log/event"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FmtGob struct allows marshalling a LogMessage as gob-encoded bytes
type FmtGob struct{}

// Format method will take in a pointer to a LogMessage, and return the execution of its
// encode() method; which converts it to gob-encoded bytes, returning a slice of bytes and an
// error
func (f *FmtGob) Format(log *event.Event) ([]byte, error) {
	buf := &bytes.Buffer{}
	gob.Register(&timestamppb.Timestamp{})
	gob.Register(event.Level(0))
	gob.Register(event.Field{})
	gob.Register(map[string]interface{}{})

	enc := gob.NewEncoder(buf)

	err := enc.Encode(log)

	return buf.Bytes(), err

}
