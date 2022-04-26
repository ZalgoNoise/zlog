package text

import (
	"reflect"
	"testing"
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
