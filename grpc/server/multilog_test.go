package server

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

type testLogServer struct{}

func (testLogServer) Serve() {}
func (testLogServer) Stop()  {}
func (testLogServer) Channels() (logCh, logSvCh chan *event.Event, errCh chan error) {
	logCh = make(chan *event.Event, 0)
	logSvCh = make(chan *event.Event, 0)
	errCh = make(chan error, 0)

	go func() {
		for {
			select {
			case _ = <-logCh:
				continue // test goes on, first
			case _ = <-logSvCh:
				return // then it stops, on the second (svLogger) call
			}
		}
	}()
	return
}

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
		&testLogServer{},
		&testLogServer{},
		&testLogServer{},
		&testLogServer{},
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
			log, svLog, _ := ml.Channels() // skip error coverage for now

			log <- event.New().Message("null").Build()
			svLog <- event.New().Message("null").Build()
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
