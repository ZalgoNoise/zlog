package csv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FmtCSV struct describes the different manipulations and processing that a CSV LogFormatter
// can apply to an event.Event
type FmtCSV struct {
	UnixTime bool
	JsonMeta bool
}

type entry [6]string

// FmtCSVBuilder struct allows creating custom CSV Formatters. Its default values will leave
// its supported options set as false, so it's not required to always use this struct.
//
// Its options will allow:
// - setting Unix micros as timestamp (in string format)
// - setting a JSON metadata formatter instead of text-based
type FmtCSVBuilder struct {
	unixTime bool
	jsonMeta bool
}

// New function will create a new instance of a FmtCSVBuilder
func New() *FmtCSVBuilder {
	return &FmtCSVBuilder{}
}

// Unix method will set the FmtCSV's timestamp as Unix micros
func (b *FmtCSVBuilder) Unix() *FmtCSVBuilder {
	b.unixTime = true
	return b
}

// JSON method will set the FmtCSV's metadata as JSON format
func (b *FmtCSVBuilder) JSON() *FmtCSVBuilder {
	b.jsonMeta = true
	return b
}

// Build method will create a (custom) FmtCSV object based on the builder's configuration,
// and return a pointer to it
func (b *FmtCSVBuilder) Build() *FmtCSV {
	return &FmtCSV{
		UnixTime: b.unixTime,
		JsonMeta: b.jsonMeta,
	}
}

// Format method will take in a pointer to an event.Event; and returns a buffer and an error.
//
// This method will process the input event.Event and marshal it according to this LogFormatter
func (f *FmtCSV) Format(log *event.Event) (buf []byte, err error) {
	b := bytes.NewBuffer(buf)
	w := csv.NewWriter(b)

	// prepare time value
	var t string

	if f.UnixTime {
		// Unix micros in string format
		t = strconv.FormatInt(log.Time.AsTime().UnixNano(), 10)
	} else {
		// RFC 3339 timestamp in string format
		t = log.Time.AsTime().Format(text.LTRFC3339Nano.String())
	}

	// prepare metadata value
	var m string
	if f.JsonMeta {
		// marshal as JSON
		b, err := json.Marshal(log.Meta.AsMap())
		if err != nil {
			return nil, err
		}
		m = string(b)
	} else {
		// use FmtText to marshal the metadata
		txt := &text.FmtText{}
		m = txt.FmtMetadata(log.Meta.AsMap())
	}

	// default format for:
	// "timestamp","level","prefix","sub","message","metadata"
	record := entry{
		t,
		log.GetLevel().String(),
		log.GetPrefix(),
		log.GetSub(),
		log.GetMsg(),
		m,
	}

	if err = w.Write(record[:]); err != nil {
		return nil, err
	}

	w.Flush()

	if err = w.Error(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil

}

func Decode(buf []byte) (e *event.Event, err error) {
	b := bytes.NewBuffer(buf)
	r := csv.NewReader(b)

	record, err := r.Read()

	if err != nil {
		return nil, err
	}

	meta := new(map[string]interface{})
	err = json.Unmarshal([]byte(record[5]), meta)

	if err != nil {
		return nil, err
	}

	timestamp, err := convTime(record[0])

	e = event.New().
		Level(event.Level(event.Level_value[record[1]])).
		Prefix(record[2]).
		Sub(record[3]).
		Message(record[4]).
		Metadata(*meta).
		Build()

	e.Time = timestamppb.New(timestamp)

	return e, err
}

func convTime(in string) (out time.Time, err error) {
	rfcTime, rErr := convRFC3339(in)

	if rErr != nil {
		err = fmt.Errorf("RFC3339 timestamp conversion failed: %w", rErr)

		unixTime, uErr := convUnix(in)

		if uErr != nil {
			err = fmt.Errorf("unix timestamp conversion failed: %v; %w", uErr, err)
			return
		}

		out = unixTime
		err = nil

		return

	}

	out = rfcTime
	err = nil

	return
}

func convRFC3339(t string) (time.Time, error) {
	return time.Parse(text.LTRFC3339Nano.String(), t)
}

func convUnix(t string) (time.Time, error) {
	var out time.Time

	usec := t[:10]
	unano := t[len(t)-9:]

	unixtime, err := strconv.ParseInt(usec, 10, 64)

	if err != nil {
		return out, err
	}

	unixnano, err := strconv.ParseInt(unano, 10, 64)
	if err != nil {
		return out, err
	}
	return time.Unix(unixtime, unixnano), nil
}
