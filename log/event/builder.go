package event

import (
	"time"

	"github.com/zalgonoise/zlog/log/trace"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EventBuilder struct describes the elements in an Events's builder, which will
// be the target of different changes until its `Build()` method is called -- returning
// then a pointer to an Events object
type EventBuilder struct {
	BTime     *time.Time
	BPrefix   *string
	BSub      *string
	BLevel    *Level
	BMsg      string
	BMetadata *map[string]interface{}
}

// New function is the initializer of an EventBuilder. From this call, further
// EventBuilder methods can be chained since they all return pointers to the same
//
// This builder returns an EventBuilder struct with initialized pointers to its elements,
// which is exactly the same implementation as the protobuf message, without
// protobuf-specific data types
func New() *EventBuilder {
	var (
		prefix   string = Default_Event_Prefix
		sub      string
		level    Level = Default_Event_Level
		metadata map[string]interface{}
	)

	return &EventBuilder{
		BPrefix:   &prefix,
		BSub:      &sub,
		BLevel:    &level,
		BMetadata: &metadata,
	}
}

// Prefix method will set the prefix element in the EventBuilder with string p, and
// return the builder
func (b *EventBuilder) Prefix(p string) *EventBuilder {
	*b.BPrefix = p
	return b
}

// Sub method will set the sub-prefix element in the EventBuilder with string s, and
// return the builder
func (b *EventBuilder) Sub(s string) *EventBuilder {
	*b.BSub = s
	return b
}

// Message method will set the message element in the EventBuilder with string m, and
// return the builder
func (b *EventBuilder) Message(m string) *EventBuilder {
	b.BMsg = m
	return b
}

// Level method will set the level element in the EventBuilder with LogLevel l, and
// return the builder
func (b *EventBuilder) Level(l Level) *EventBuilder {
	*b.BLevel = l
	return b
}

// Metadata method will set (or add) the metadata element in the EventBuilder
// with map m, and return the builder
func (b *EventBuilder) Metadata(m map[string]interface{}) *EventBuilder {
	if m == nil {
		return b
	}

	if b.BMetadata == nil || len(*b.BMetadata) == 0 {
		b.BMetadata = &m
	} else {
		mcopy := *b.BMetadata
		for k, v := range m {
			mcopy[k] = v
		}
		b.BMetadata = &mcopy
	}
	return b
}

// CallStack method will grab the current call stack, and add it as a "callstack" object
// in the EventBuilder's metadata.
func (b *EventBuilder) CallStack(all bool) *EventBuilder {
	if *b.BMetadata == nil {
		*b.BMetadata = map[string]interface{}{}
	}
	mcopy := *b.BMetadata
	mcopy["callstack"] = trace.New(all)
	*b.BMetadata = mcopy

	return b
}

// Build method will create a new timestamp, review all elements in the `EventBuilder`,
// apply any defaults to non-defined elements, and return a pointer to an Event
func (b *EventBuilder) Build() *Event {
	var timestamp *timestamppb.Timestamp = timestamppb.Now()
	var meta *structpb.Struct

	if b.BLevel == nil {
		b.BLevel = new(Level)
		*b.BLevel = Default_Event_Level
	}

	if b.BPrefix == nil {
		b.BPrefix = new(string)
		*b.BPrefix = Default_Event_Prefix
	}

	if b.BSub == nil {
		b.BSub = new(string)
		*b.BSub = ""
	}

	if b.BMetadata == nil {
		meta = new(structpb.Struct)
	} else {
		f := Field(*b.BMetadata)
		meta = f.Encode()
	}

	return &Event{
		Time:   timestamp,
		Prefix: b.BPrefix,
		Sub:    b.BSub,
		Level:  b.BLevel,
		Msg:    &b.BMsg,
		Meta:   meta,
	}
}
