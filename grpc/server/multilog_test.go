package server

import (
	"reflect"
	"testing"
)

func TestMultiLogger(t *testing.T) {
	module := "GRPCLogServer"
	funcname := "MultiLogger()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		l     []LogServer
		wants LogServer
	}

	var loggers = []LogServer{
		NilServer(),
		NilServer(),
		NilServer(),
		NilServer(),
	}

	var tests = []test{
		{
			name:  "empty call",
			l:     []LogServer{},
			wants: nil,
		},
		{
			name:  "nil call",
			l:     nil,
			wants: nil,
		},
		{
			name:  "one logger",
			l:     []LogServer{loggers[0]},
			wants: loggers[0],
		},
		{
			name: "multiple loggers",
			l:    []LogServer{loggers[0], loggers[1], loggers[2]},
			wants: &multiLogger{
				loggers: []LogServer{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "nested multiloggers",
			l:    []LogServer{loggers[0], MultiLogger(loggers[1], loggers[2])},
			wants: &multiLogger{
				loggers: []LogServer{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "add nil logger in the mix",
			l:    []LogServer{loggers[0], loggers[1], nil},
			wants: &multiLogger{
				loggers: []LogServer{loggers[0], loggers[1]},
			},
		},
		{
			name:  "add nil loggers in the mix",
			l:     []LogServer{loggers[0], nil, nil},
			wants: loggers[0],
		},
		{
			name:  "add only loggers",
			l:     []LogServer{nil, nil, nil},
			wants: nil,
		},
	}

	var verify = func(idx int, test test) {
		ml := MultiLogger(test.l...)

		if !reflect.DeepEqual(ml, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				ml,
				test.name,
			)
		}

		// test interface calls
		if ml != nil {
			ml.Serve()
			ml.Stop()
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
