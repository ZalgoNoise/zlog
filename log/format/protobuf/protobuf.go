package protobuf

import (
	"github.com/zalgonoise/zlog/log/event"
	"google.golang.org/protobuf/proto"
)

// FmtPB struct is a placeholder LogFormatter to seamlessly convert protobuf messages into
// byte slices for data transmission
type FmtPB struct{}

// Format method will take in a pointer to an event.Event; and returns a slice of bytes
// and an error.
//
// This method will process the input event.Event and marshal it according to this LogFormatter
func (f *FmtPB) Format(log *event.Event) (buf []byte, err error) {
	return proto.Marshal(log)
}
