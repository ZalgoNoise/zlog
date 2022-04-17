package log

// import (
// 	"bytes"
// 	"encoding/gob"
// 	"encoding/json"
// 	"time"

// 	"github.com/zalgonoise/zlog/log/format/xml"
// 	"github.com/zalgonoise/zlog/log/trace"
// 	pb "github.com/zalgonoise/zlog/proto/message"
// 	"google.golang.org/protobuf/encoding/protojson"
// 	"google.golang.org/protobuf/types/known/structpb"
// 	"google.golang.org/protobuf/types/known/timestamppb"
// )

// // LogLevel type describes a numeric value for a log level with priority increasing in
// // relation to its value
// //
// // LogLevel also implements the Stringer interface, used to convey this log level in a message
// type LogLevel int32

// const (
// 	LLTrace LogLevel = iota // LogLevel Trace
// 	LLDebug                 // LogLevel Debug
// 	LLInfo                  // LogLevel Info
// 	LLWarn                  // LogLevel Warning
// 	LLError                 // LogLevel Error
// 	LLFatal                 // LogLevel Fatal
// 	_                       // [reserved]
// 	_                       // [reserved]
// 	_                       // [reserved]
// 	LLPanic                 // LogLevel Panic
// )

// // String method is defined for LogLevel objects to implement the Stringer interface
// //
// // It returns the string to which this log level is mapped to, in `LogTypeVals`
// func (ll LogLevel) String() string {
// 	return LogTypeVals[ll]
// }

// // Int method returns a LogLevel's value as an integer, to be used for comparison with
// // input log level filters
// func (ll LogLevel) Int() int {
// 	return int(ll)
// }

// var (
// 	// LogTypeVals is an enum map to convert LogLevels to its string representation
// 	LogTypeVals = map[LogLevel]string{
// 		0: "trace",
// 		1: "debug",
// 		2: "info",
// 		3: "warn",
// 		4: "error",
// 		5: "fatal",
// 		9: "panic",
// 	}

// 	// LogTypeVals is an enum map to convert LogLevels from its string representation
// 	// to an int value
// 	LogTypeKeys = map[string]int{
// 		"trace": 0,
// 		"debug": 1,
// 		"info":  2,
// 		"warn":  3,
// 		"error": 4,
// 		"fatal": 5,
// 		"panic": 9,
// 	}
// )

// // Field type is a generic type to build LogMessage Metadata
// type Field map[string]interface{}

// // ToMap method returns the Field in it's (raw) string-interface{} map format
// func (f Field) ToMap() map[string]interface{} {
// 	return f
// }

// // type field struct {
// // 	Key string      `xml:"key,omitempty"`
// // 	Val interface{} `xml:"value,omitempty"`
// // }

// // func mappify(data map[string]interface{}) []field {
// // 	var fields []field

// // 	for k, v := range data {
// // 		switch value := v.(type) {
// // 		case []map[string]interface{}:
// // 			f := []field{}

// // 			for _, im := range value {
// // 				ifield := field{}
// // 				for ik, iv := range im {
// // 					ifield.Key = ik
// // 					ifield.Val = iv
// // 				}

// // 				f = append(f, ifield)
// // 			}

// // 			fields = append(fields, field{
// // 				Key: k,
// // 				Val: f,
// // 			})
// // 		case []Field:
// // 			f := []field{}

// // 			for _, im := range value {
// // 				ifield := field{}
// // 				for ik, iv := range im.ToMap() {
// // 					ifield.Key = ik
// // 					ifield.Val = iv
// // 				}

// // 				f = append(f, ifield)
// // 			}

// // 			fields = append(fields, field{
// // 				Key: k,
// // 				Val: f,
// // 			})
// // 		case map[string]interface{}:
// // 			fields = append(fields, field{
// // 				Key: k,
// // 				Val: mappify(value),
// // 			})
// // 		case Field:
// // 			fields = append(fields, field{
// // 				Key: k,
// // 				Val: mappify(value.ToMap()),
// // 			})
// // 		default:
// // 			fields = append(fields, field{
// // 				Key: k,
// // 				Val: value,
// // 			})
// // 		}
// // 	}

// // 	return fields
// // }

// // ToXML method returns the Field in a list of key-value objects,
// // compatible with XML marshalling of data objects
// func (f Field) ToXML() []xml.Field {
// 	return xml.Mappify(f.ToMap())
// }

// // LogMessage struct describes a Log Message's elements, already in a format that can be
// // parsed by a valid formatter.
// type LogMessage struct {
// 	Time     time.Time              `json:"timestamp,omitempty" xml:"timestamp,omitempty" bson:"timestamp,omitempty"`
// 	Prefix   string                 `json:"service,omitempty" xml:"service,omitempty" bson:"service,omitempty"`
// 	Sub      string                 `json:"module,omitempty" xml:"module,omitempty" bson:"module,omitempty"`
// 	Level    string                 `json:"level,omitempty" xml:"level,omitempty" bson:"level,omitempty"`
// 	Msg      string                 `json:"message,omitempty" xml:"message,omitempty" bson:"message,omitempty"`
// 	Metadata map[string]interface{} `json:"metadata,omitempty" xml:"metadata,omitempty" bson:"metadata,omitempty"`
// }

// func (m *LogMessage) encode() ([]byte, error) {
// 	buf := &bytes.Buffer{}
// 	gob.Register(Field{})
// 	gob.Register(map[string]interface{}{})

// 	enc := gob.NewEncoder(buf)

// 	err := enc.Encode(m)

// 	return buf.Bytes(), err
// }

// // Bytes method will return a LogMessage as a gob-encoded slice of bytes. It is compatible with
// // a Logger's io.Writer implementation, as its Write() method will decode this type of data
// func (m *LogMessage) Bytes() []byte {
// 	// skip error checking
// 	buf, _ := m.encode()
// 	return buf
// }

// // Proto method will convert this LogMessage into a protobuf MessageRequest,
// // while skipping the (potential) returning error.
// func (m *LogMessage) Proto() *pb.MessageRequest {
// 	msg, _ := m.ToProto()
// 	return msg
// }

// // ToProto method will conver this LogMessage into a protobuf MessageRequest,
// // returning a pointer to one and an error.
// //
// // This is possible only by encoding the message's metadata into JSON bytes,
// // to then encode as a struct protobuf with protojson.
// //
// // In gRPC, metadata will make the messages heavier; as there is no way to send
// // simply any arbitrary JSON data via gRPC without encoding the keys in, too. It's
// // either this, or the .proto file would need to describe the metadata format
// func (m *LogMessage) ToProto() (*pb.MessageRequest, error) {
// 	b, err := json.Marshal(m.Metadata)
// 	if err != nil {
// 		return nil, err
// 	}
// 	s, err := encodeProto(b)
// 	if err != nil {
// 		return nil, err
// 	}

// 	level := pb.Level(LogTypeKeys[m.Level])

// 	return &pb.MessageRequest{
// 		Time:   timestamppb.New(m.Time),
// 		Prefix: &m.Prefix,
// 		Sub:    &m.Sub,
// 		Level:  &level,
// 		Msg:    m.Msg,
// 		Meta:   s,
// 	}, nil
// }

// func encodeProto(in []byte) (*structpb.Struct, error) {
// 	s := &structpb.Struct{}
// 	err := protojson.Unmarshal(in, s)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return s, nil
// }

// // MessageBuilder struct describes the elements in a LogMessage's builder, which will
// // be the target of different changes until its `Build()` method is called -- returning
// // then a pointer to a LogMessage object
// type MessageBuilder struct {
// 	time     time.Time
// 	prefix   string
// 	sub      string
// 	level    string
// 	msg      string
// 	metadata map[string]interface{}
// }

// // NewMessage function is the initializer of a MessageBuilder. From this call, further
// // MessageBuilder methods can be chained since they all return pointers to the same object
// func NewMessage() *MessageBuilder {
// 	return &MessageBuilder{}
// }

// // Prefix method will set the prefix element in the MessageBuilder with string p, and
// // return the builder
// func (b *MessageBuilder) Prefix(p string) *MessageBuilder {
// 	b.prefix = p
// 	return b
// }

// // Sub method will set the sub-prefix element in the MessageBuilder with string s, and
// // return the builder
// func (b *MessageBuilder) Sub(s string) *MessageBuilder {
// 	b.sub = s
// 	return b
// }

// // Message method will set the message element in the MessageBuilder with string m, and
// // return the builder
// func (b *MessageBuilder) Message(m string) *MessageBuilder {
// 	b.msg = m
// 	return b
// }

// // Level method will set the level element in the MessageBuilder with LogLevel l, and
// // return the builder
// func (b *MessageBuilder) Level(l LogLevel) *MessageBuilder {
// 	b.level = l.String()
// 	return b
// }

// // Metadata method will set (or add) the metadata element in the MessageBuilder
// // with map m, and return the builder
// func (b *MessageBuilder) Metadata(m map[string]interface{}) *MessageBuilder {
// 	if m == nil {
// 		return b
// 	}

// 	if b.metadata == nil || len(b.metadata) == 0 {
// 		b.metadata = m
// 	} else {
// 		for k, v := range m {
// 			b.metadata[k] = v
// 		}
// 	}
// 	return b
// }

// // CallStack method will grab the current call stack, and add it as a "callstack" object
// // in the MessageBuilder's metadata.
// func (b *MessageBuilder) CallStack(all bool) *MessageBuilder {
// 	if b.metadata == nil {
// 		b.metadata = map[string]interface{}{}
// 	}
// 	b.metadata["callstack"] = trace.New(all)

// 	return b
// }

// // FromProto method will decode a protobuf MessageRequest, returning a pointer to
// // a MessageBuilder.
// //
// // Considering the amount of optional elements, all elements are verified (besides the
// // message elements) and defaults are applied when unset.
// func (b *MessageBuilder) FromProto(in *pb.MessageRequest) *MessageBuilder {
// 	if in.Time == nil {
// 		b.time = time.Now()
// 	} else {
// 		b.time = in.Time.AsTime()
// 	}

// 	if in.Level == nil {
// 		b.level = LLInfo.String()
// 	} else {
// 		b.level = LogLevel(int(*in.Level)).String()
// 	}

// 	if in.Prefix == nil {
// 		b.prefix = "log"
// 	} else {
// 		b.prefix = *in.Prefix
// 	}

// 	if in.Sub == nil {
// 		b.sub = ""
// 	} else {
// 		b.sub = *in.Sub
// 	}

// 	if in.Meta == nil {
// 		b.metadata = map[string]interface{}{}
// 	} else {
// 		b.metadata = in.Meta.AsMap()
// 	}

// 	b.msg = in.Msg

// 	return b
// }

// func (b *MessageBuilder) FromGob(p []byte) (*LogMessage, error) {
// 	msg := &LogMessage{}

// 	buf := bytes.NewBuffer(p)
// 	dec := gob.NewDecoder(buf)
// 	err := dec.Decode(msg)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return msg, nil
// }

// // Build method will create a new timestamp, review all elements in the `MessageBuilder`,
// // apply any defaults to non-defined elements, and return a pointer to a LogMessage
// func (b *MessageBuilder) Build() *LogMessage {
// 	if b.time.IsZero() {
// 		b.time = time.Now()
// 	}

// 	if b.prefix == "" {
// 		b.prefix = "log"
// 	}

// 	if b.level == "" {
// 		b.level = LLInfo.String()
// 	}

// 	if b.metadata == nil {
// 		b.metadata = map[string]interface{}{}
// 	}

// 	return &LogMessage{
// 		Time:     b.time,
// 		Prefix:   b.prefix,
// 		Sub:      b.sub,
// 		Level:    b.level,
// 		Msg:      b.msg,
// 		Metadata: b.metadata,
// 	}
// }
