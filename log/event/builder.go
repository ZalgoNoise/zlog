package event

import (
	"time"

	"github.com/zalgonoise/zlog/log/trace"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EventBuilder struct describes the elements in a Events's builder, which will
// be the target of different changes until its `Build()` method is called -- returning
// then a pointer to a Events object
type EventBuilder struct {
	time     *time.Time
	prefix   *string
	sub      *string
	level    *Level
	msg      string
	metadata *map[string]interface{}
}

// New function is the initializer of a EventBuilder. From this call, further
// EventBuilder methods can be chained since they all return pointers to the same
//
// This builder returns an EventBuilder struct with initialized pointers to its elements,
// which is exactly the same implementation as the protobuf message, without
// protobuf-specific data types
func New() *EventBuilder {
	var (
		time     time.Time
		prefix   string = Default_Event_Prefix
		sub      string
		level    Level = Default_Event_Level
		metadata map[string]interface{}
	)

	return &EventBuilder{
		time:     &time,
		prefix:   &prefix,
		sub:      &sub,
		level:    &level,
		metadata: &metadata,
	}
}

// Prefix method will set the prefix element in the EventBuilder with string p, and
// return the builder
func (b *EventBuilder) Prefix(p string) *EventBuilder {
	*b.prefix = p
	return b
}

// Sub method will set the sub-prefix element in the EventBuilder with string s, and
// return the builder
func (b *EventBuilder) Sub(s string) *EventBuilder {
	*b.sub = s
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
func (b *EventBuilder) Level(l Level) *EventBuilder {
	*b.level = l
	return b
}

// Metadata method will set (or add) the metadata element in the EventBuilder
// with map m, and return the builder
func (b *EventBuilder) Metadata(m map[string]interface{}) *EventBuilder {
	if m == nil {
		return b
	}

	if b.metadata == nil || len(*b.metadata) == 0 {
		b.metadata = &m
	} else {
		mcopy := *b.metadata
		for k, v := range m {
			mcopy[k] = v
		}
		b.metadata = &mcopy
	}
	return b
}

// CallStack method will grab the current call stack, and add it as a "callstack" object
// in the EventBuilder's metadata.
func (b *EventBuilder) CallStack(all bool) *EventBuilder {
	if b.metadata == nil {
		b.metadata = &map[string]interface{}{}
	}
	mcopy := *b.metadata
	mcopy["callstack"] = trace.New(all)
	b.metadata = &mcopy

	return b
}

// Build method will create a new timestamp, review all elements in the `EventBuilder`,
// apply any defaults to non-defined elements, and return a pointer to a Event
func (b *EventBuilder) Build() *Event {
	var timestamp *timestamppb.Timestamp
	var loglevel *Level
	var meta *structpb.Struct

	if t := *b.time; t.IsZero() {
		timestamp = timestamppb.Now()
	} else {
		timestamp = timestamppb.New(*b.time)
	}

	if b.prefix == nil {
		*b.prefix = Default_Event_Prefix
	}

	if b.level == nil {
		*loglevel = Default_Event_Level
	} else {
		loglevel = b.level
	}

	if b.metadata == nil {
		meta = &structpb.Struct{}
	} else {
		f := Field(*b.metadata)
		meta = f.Encode()
	}

	return &Event{
		Time:   timestamp,
		Prefix: b.prefix,
		Sub:    b.sub,
		Level:  loglevel,
		Msg:    &b.msg,
		Meta:   meta,
	}
}
