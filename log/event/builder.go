package event

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/zalgonoise/zlog/log/trace"
	pb "github.com/zalgonoise/zlog/proto/message"
)

// EventBuilder struct describes the elements in a Events's builder, which will
// be the target of different changes until its `Build()` method is called -- returning
// then a pointer to a Events object
type EventBuilder struct {
	time     time.Time
	prefix   string
	sub      string
	level    string
	msg      string
	metadata map[string]interface{}
}

// New function is the initializer of a EventBuilder. From this call, further
// EventBuilder methods can be chained since they all return pointers to the same object
func New() *EventBuilder {
	return &EventBuilder{}
}

// Prefix method will set the prefix element in the EventBuilder with string p, and
// return the builder
func (b *EventBuilder) Prefix(p string) *EventBuilder {
	b.prefix = p
	return b
}

// Sub method will set the sub-prefix element in the EventBuilder with string s, and
// return the builder
func (b *EventBuilder) Sub(s string) *EventBuilder {
	b.sub = s
	return b
}

// Message method will set the message element in the EventBuilder with string m, and
// return the builder
func (b *EventBuilder) Message(m string) *EventBuilder {
	b.msg = m
	return b
}

// Level method will set the level element in the EventBuilder with LogLevel l, and
// return the builder
func (b *EventBuilder) Level(l LogLevel) *EventBuilder {
	b.level = l.String()
	return b
}

// Metadata method will set (or add) the metadata element in the EventBuilder
// with map m, and return the builder
func (b *EventBuilder) Metadata(m map[string]interface{}) *EventBuilder {
	if m == nil {
		return b
	}

	if b.metadata == nil || len(b.metadata) == 0 {
		b.metadata = m
	} else {
		for k, v := range m {
			b.metadata[k] = v
		}
	}
	return b
}

// CallStack method will grab the current call stack, and add it as a "callstack" object
// in the EventBuilder's metadata.
func (b *EventBuilder) CallStack(all bool) *EventBuilder {
	if b.metadata == nil {
		b.metadata = map[string]interface{}{}
	}
	b.metadata["callstack"] = trace.New(all)

	return b
}

// FromProto method will decode a protobuf MessageRequest, returning a pointer to
// a EventBuilder.
//
// Considering the amount of optional elements, all elements are verified (besides the
// message elements) and defaults are applied when unset.
func (b *EventBuilder) FromProto(in *pb.MessageRequest) *EventBuilder {
	if in.Time == nil {
		b.time = time.Now()
	} else {
		b.time = in.Time.AsTime()
	}

	if in.Level == nil {
		b.level = LLInfo.String()
	} else {
		b.level = LogLevel(int(*in.Level)).String()
	}

	if in.Prefix == nil {
		b.prefix = "log"
	} else {
		b.prefix = *in.Prefix
	}

	if in.Sub == nil {
		b.sub = ""
	} else {
		b.sub = *in.Sub
	}

	if in.Meta == nil {
		b.metadata = map[string]interface{}{}
	} else {
		b.metadata = in.Meta.AsMap()
	}

	b.msg = in.Msg

	return b
}

func (b *EventBuilder) FromGob(p []byte) (*Event, error) {
	msg := &Event{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(msg)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Build method will create a new timestamp, review all elements in the `EventBuilder`,
// apply any defaults to non-defined elements, and return a pointer to a Event
func (b *EventBuilder) Build() *Event {
	if b.time.IsZero() {
		b.time = time.Now()
	}

	if b.prefix == "" {
		b.prefix = "log"
	}

	if b.level == "" {
		b.level = LLInfo.String()
	}

	if b.metadata == nil {
		b.metadata = map[string]interface{}{}
	}

	return &Event{
		Time:     b.time,
		Prefix:   b.prefix,
		Sub:      b.sub,
		Level:    b.level,
		Msg:      b.msg,
		Metadata: b.metadata,
	}
}
