package event

import (
	"google.golang.org/protobuf/proto"
)

// Marshal method will convert the protobuf Event into a slice of bytes, also returning
// potential errors in the convertion
func (e *Event) Marshal() ([]byte, error) {
	return proto.Marshal(e)
}

// Unmarshal method will convert the input slice of bytes as a protobuf Event, stored in the
// method receiver. Returns an error if any.
func (e *Event) Unmarshal(b []byte) error {
	return proto.Unmarshal(b, e)
}

// Encode method is similar to Event.Marshal(), but it does not return any errors
func (e *Event) Encode() []byte {
	b, _ := e.Marshal()
	return b
}

// Decode method is similar to Event.Unmarshal, but it does not return any errors
func (e *Event) Decode(b []byte) {
	_ = proto.Unmarshal(b, e) // deliberately ignore error in this method call
}

// Decode function will take in a slice of bytes and convert it to an Event protobuf,
// returning this and an error if any.
func Decode(b []byte) (*Event, error) {
	var e = &Event{}

	err := proto.Unmarshal(b, e)

	if err != nil {
		return nil, err
	}
	return e, nil
}

// Encode function will take in a protobuf Event, and convert it to a slice of bytes,
// returning this and an error if any.
func Encode(e *Event) (b []byte, err error) {
	return proto.Marshal(e)
}
