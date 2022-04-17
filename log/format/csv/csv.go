package csv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strconv"

	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
)

// FmtCSV struct describes the different manipulations and processing that a CSV LogFormatter
// can apply to a LogMessage
type FmtCSV struct {
	UnixTime bool
	JsonMeta bool
}

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

// NewCSVFormat function will create a new instance of a FmtCSVBuilder
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

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtCSV) Format(log *event.Event) (buf []byte, err error) {
	b := bytes.NewBuffer(buf)
	w := csv.NewWriter(b)

	// prepare time value
	var t string

	if f.UnixTime {
		// Unix micros in string format
		t = strconv.FormatInt(log.Time.Unix(), 10)
	} else {
		// RFC 3339 timestamp in string format
		t = log.Time.Format(text.LTRFC3339Nano.String())
	}

	// prepare metadata value
	var m string
	if f.JsonMeta {
		// marshal as JSON
		b, err := json.Marshal(log.Metadata)
		if err != nil {
			return nil, err
		}
		m = string(b)
	} else {
		// use FmtText to marshal the metadata
		txt := &text.FmtText{}
		m = txt.FmtMetadata(log.Metadata)
	}

	// default format for:
	// "timestamp","level","prefix","sub","message","metadata"
	record := []string{
		t,
		log.Level,
		log.Prefix,
		log.Sub,
		log.Msg,
		m,
	}

	if err = w.Write(record); err != nil {
		return nil, err
	}

	w.Flush()

	if err = w.Error(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil

}
