package text

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/log/event"
)

func TestNew(t *testing.T) {
	module := "FmtText"
	funcname := "New()"

	type test struct {
		name        string
		wants       *FmtText
		time        LogTimestamp
		levelFirst  bool
		doubleSpace bool
		color       bool
		upper       bool
		noTimestamp bool
		noHeaders   bool
		noLevel     bool
	}

	var tests = []test{
		{
			name: "check defaults",
			wants: &FmtText{
				timeFormat: LTRFC3339Nano.String(),
			},
		},
		{
			name: "custom everything, no clashes",
			wants: &FmtText{
				timeFormat:  LTRubyDate.String(),
				levelFirst:  true,
				doubleSpace: true,
				colored:     true,
				upper:       true,
			},
			time:        LTRubyDate,
			levelFirst:  true,
			doubleSpace: true,
			color:       true,
			upper:       true,
		},
		{
			name: "nothing but message, with clashes",
			wants: &FmtText{
				timeFormat:  LTRubyDate.String(),
				noTimestamp: true,
				noHeaders:   true,
				noLevel:     true,
				doubleSpace: true,
			},
			time:        LTRubyDate,
			noLevel:     true,
			noHeaders:   true,
			noTimestamp: true,
			levelFirst:  true,
			doubleSpace: true,
			color:       true,
			upper:       true,
		},
	}

	var init = func(test test) *FmtText {
		f := New()
		if test.time != "" {
			f.Time(test.time)
		}
		if test.levelFirst {
			f.LevelFirst()
		}
		if test.doubleSpace {
			f.DoubleSpace()
		}
		if test.color {
			f.Color()
		}
		if test.upper {
			f.Upper()
		}
		if test.noTimestamp {
			f.NoTimestamp()
		}
		if test.noHeaders {
			f.NoHeaders()
		}
		if test.noLevel {
			f.NoLevel()
		}
		return f.Build()
	}

	var verify = func(idx int, test test, f *FmtText) {
		if !reflect.DeepEqual(*f, *test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output formatter mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				*test.wants,
				*f,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		f := init(test)
		verify(idx, test, f)
	}
}

func TestFormat(t *testing.T) {
	module := "FmtText"
	funcname := "Format()"

	type test struct {
		name  string
		e     *event.Event
		f     *FmtText
		regex string
	}

	var tests = []test{
		{
			name:  "basic event",
			e:     event.New().Message("null").Build(),
			f:     New().Build(),
			regex: `\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z\]\s+\[info\]\s+\[log\]\s+null`,
		},
		{
			name:  "complete event",
			e:     event.New().Prefix("test").Sub("testing").Message("null").Metadata(event.Field{"a": true}).Build(),
			f:     New().Build(),
			regex: `\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z\]\s+\[info\]\s+\[test\]\s+\[testing\]\s+null\s+\[ a = true \]`,
		},
		{
			name:  "complete event; double space",
			e:     event.New().Prefix("test").Sub("testing").Message("null").Metadata(event.Field{"a": true}).Build(),
			f:     New().DoubleSpace().Build(),
			regex: `\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z\]\s+\[info\]\s+\[test\]\s+\[testing\]\s+null\s+\[ a = true \]`,
		},
		{
			name:  "complete event; level first, double space",
			e:     event.New().Prefix("test").Sub("testing").Message("null").Metadata(event.Field{"a": true}).Build(),
			f:     New().LevelFirst().DoubleSpace().Build(),
			regex: `\[info\]\s+\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z\]\s+\[test\]\s+\[testing\]\s+null\s+\[ a = true \]`,
		},
		{
			name:  "complete event; color, uppercase",
			e:     event.New().Prefix("test").Sub("testing").Message("null").Metadata(event.Field{"a": true}).Build(),
			f:     New().Color().Upper().Build(),
			regex: `\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z\]\s+\[\\[36mINFO\\[0m\]\s+\[TEST\]\s+\[TESTING\]\s+null\s+\[ a = true \]`,
		},
	}

	var init = func(test test) ([]byte, error) {
		return test.f.Format(test.e)
	}

	var verify = func(idx int, test test) {
		b, err := init(test)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- formatting failed with an error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		r := regexp.MustCompile(test.regex)

		if !r.MatchString(string(b)) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch: matching expression %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.regex,
				string(b),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestFmtTime(t *testing.T) {
	module := "FmtText"
	funcname := "fmtTime()"

	type test struct {
		name  string
		time  LogTimestamp
		regex string
	}

	var tests = []test{
		{
			name:  "RFC3339Nano",
			time:  LTRFC3339Nano,
			regex: `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.+`, // 2022-09-19T18:36:43.454942597Z
		},
		{
			name:  "RFC3339",
			time:  LTRFC3339,
			regex: `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.+`, //2022-09-19T18:36:43Z
		},
		{
			name:  "RFC822Z",
			time:  LTRFC822Z,
			regex: `\d{2}\s+((Jan)|(Feb)|(Mar)|(Apr)|(May)|(Jun)|(Jul)|(Aug)|(Sep)|(Oct)|(Nov)|(Dec))\s+\d{2}\s+\d{2}:\d{2}\s+\+\d{4}`,
		},
		{
			name:  "RubyDate",
			time:  LTRubyDate,
			regex: `((Mon)|(Tue)|(Wed)|(Thu)|(Fri)|(Sat)|(Sun))\s+((Jan)|(Feb)|(Mar)|(Apr)|(May)|(Jun)|(Jul)|(Aug)|(Sep)|(Oct)|(Nov)|(Dec))\s+\d{2}\s+\d{2}:\d{2}:\d{2}\s+\+\d{4}\s+\d{4}`,
		},
		{
			name:  "UNIX_NANO",
			time:  LTUnixNano,
			regex: `\d+`,
		},
		{
			name:  "UNIX_MILLI",
			time:  LTUnixMilli,
			regex: `\d+`,
		},
		{
			name:  "UNIX_MICRO",
			time:  LTUnixMicro,
			regex: `\d+`,
		},
	}

	var init = func(test test) *FmtText {
		return New().Time(test.time).Build()
	}

	var verify = func(idx int, test test) {
		f := init(test)

		out := f.fmtTime(time.Now())

		r := regexp.MustCompile(test.regex)

		if !r.MatchString(out) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch: matching expression %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.regex,
				out,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestFmtMetadata(t *testing.T) {
	module := "FmtText"
	funcname := "FmtMetadata()"

	type test struct {
		name  string
		m     map[string]interface{}
		regex string
	}

	var tests = []test{
		{
			name:  "empty event",
			m:     event.Field{},
			regex: "",
		},
		{
			name: "basic event",
			m: event.Field{
				"a": true,
				"b": false,
			},
			regex: `\[(( a = true )|(;)|( b = false )){3}]`,
		},
		{
			name: "basic event w/ strings",
			m: event.Field{
				"a": "yes",
				"b": "no",
			},
			regex: `\[(( a = "yes" )|(;)|( b = "no" )){3}]`,
		},
		{
			name: "basic event w/ event.Field",
			m: event.Field{
				"a": "yes",
				"b": event.Field{
					"c": true,
					"d": false,
				},
				"e": true,
			},
			regex: `\[(( a = "yes" )|(;)|( b = \[(( c = true )|( d = false )|(;)){3}\] )|( e = true )){5}]`,
		},
		{
			name: "basic event w/ map[string]interface{}",
			m: map[string]interface{}{
				"a": "yes",
				"b": map[string]interface{}{
					"c": true,
					"d": false,
				},
				"e": true,
			},
			regex: `\[(( a = "yes" )|(;)|( b = \[(( c = true )|( d = false )|(;)){3}\] )|( e = true )){5}]`,
		},
		{
			name: "basic event w/ []event.Field",
			m: event.Field{
				"a": "yes",
				"b": []event.Field{
					{"c": true},
					{"d": false},
				},
				"e": true,
			},
			regex: `\[(( a = "yes" )|(;)|( b = \[(( \[ c = true \] )|( \[ d = false \] )|(;)){3}\] )|( e = true )){5}]`,
		},
		{
			name: "basic event w/ []map[string]interface{}",
			m: map[string]interface{}{
				"a": "yes",
				"b": []map[string]interface{}{
					{"c": true},
					{"d": false},
				},
				"e": true,
			},
			regex: `\[(( a = "yes" )|(;)|( b = \[(( \[ c = true \] )|( \[ d = false \] )|(;)){3}\] )|( e = true )){5}]`,
		},
	}

	var init = func(test test) string {
		return New().Build().FmtMetadata(test.m)
	}

	var verify = func(idx int, test test) {
		out := init(test)

		r := regexp.MustCompile(test.regex)

		if !r.MatchString(out) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch: matching expression %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.regex,
				out,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
