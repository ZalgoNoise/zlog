package event

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"

	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Event struct describes a Log Event's elements, already in a format that can be
// parsed by a valid formatter.
type Event struct {
	Time     time.Time              `json:"timestamp,omitempty" xml:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Prefix   string                 `json:"service,omitempty" xml:"service,omitempty" bson:"service,omitempty"`
	Sub      string                 `json:"module,omitempty" xml:"module,omitempty" bson:"module,omitempty"`
	Level    string                 `json:"level,omitempty" xml:"level,omitempty" bson:"level,omitempty"`
	Msg      string                 `json:"message,omitempty" xml:"message,omitempty" bson:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" xml:"metadata,omitempty" bson:"metadata,omitempty"`
}

func (e *Event) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	gob.Register(Field{})
	gob.Register(map[string]interface{}{})

	enc := gob.NewEncoder(buf)

	err := enc.Encode(e)

	return buf.Bytes(), err
}

// Bytes method will return an Event as a gob-encoded slice of bytes. It is compatible with
// a Logger's io.Writer implementation, as its Write() method will decode this type of data
func (e *Event) Bytes() []byte {
	// skip error checking
	buf, _ := e.Encode()
	return buf
}

// Proto method will convert this Event into a protobuf MessageRequest,
// while skipping the (potential) returning error.
func (m *Event) Proto() *pb.MessageRequest {
	msg, _ := m.ToProto()
	return msg
}

// ToProto method will conver this Event into a protobuf MessageRequest,
// returning a pointer to one and an error.
//
// This is possible only by encoding the message's metadata into JSON bytes,
// to then encode as a struct protobuf with protojson.
//
// In gRPC, metadata will make the messages heavier; as there is no way to send
// simply any arbitrary JSON data via gRPC without encoding the keys in, too. It's
// either this, or the .proto file would need to describe the metadata format
func (m *Event) ToProto() (*pb.MessageRequest, error) {
	b, err := json.Marshal(m.Metadata)
	if err != nil {
		return nil, err
	}
	s, err := EncodeProto(b)
	if err != nil {
		return nil, err
	}

	level := pb.Level(LogTypeKeys[m.Level])

	return &pb.MessageRequest{
		Time:   timestamppb.New(m.Time),
		Prefix: &m.Prefix,
		Sub:    &m.Sub,
		Level:  &level,
		Msg:    m.Msg,
		Meta:   s,
	}, nil
}

func EncodeProto(in []byte) (*structpb.Struct, error) {
	s := &structpb.Struct{}
	err := protojson.Unmarshal(in, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Field type is a generic type to build Event Metadata
type Field map[string]interface{}

// ToMap method returns the Field in it's (raw) string-interface{} map format
func (f Field) ToMap() map[string]interface{} {
	return f
}

// // ToXML method returns the Field in a list of key-value objects,
// // compatible with XML marshalling of data objects
// func (f Field) ToXML() []xml.Field {
// 	return xml.Mappify(f.ToMap())
// }
