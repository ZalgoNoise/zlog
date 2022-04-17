package gob

import (
	"github.com/zalgonoise/zlog/log/event"
)

// FmtGob struct allows marshalling a LogMessage as gob-encoded bytes
type FmtGob struct{}

// Format method will take in a pointer to a LogMessage, and return the execution of its
// encode() method; which converts it to gob-encoded bytes, returning a slice of bytes and an
// error
func (f *FmtGob) Format(log *event.Event) ([]byte, error) {
	return log.Encode()
}
