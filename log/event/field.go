package event

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// Field type is a generic type to build Event Metadata
type Field map[string]interface{}

// ToMap method returns the Field in it's (raw) string-interface{} map format
func (f Field) ToMap() map[string]interface{} {
	return f
}

// ToStructPB method will convert the metadata in the protobuf Event as a pointer to a
// structpb.Struct, returning this and an error if any.
//
// The metadata (a map[string]interface{}) is converted to JSON (bytes), and this data is
// unmarshalled into a *structpb.Struct object.
func (f Field) ToStructPB() (*structpb.Struct, error) {
	b, err := json.Marshal(f.ToMap())
	if err != nil {
		return nil, err
	}

	s := &structpb.Struct{}
	err = protojson.Unmarshal(b, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Encode method is similar to Field.ToStructPB(), but it does not return any errors.
func (f Field) Encode() *structpb.Struct {
	s, _ := f.ToStructPB()
	return s
}
