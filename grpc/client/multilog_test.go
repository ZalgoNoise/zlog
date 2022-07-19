package client

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestMultiLogger(t *testing.T) {
	module := "GRPCLogger"
	funcname := "MultiLogger()"

	type test struct {
		name  string
		l     []GRPCLogger
		wants GRPCLogger
	}

	var loggers = []GRPCLogger{
		NilClient(),
		NilClient(),
		NilClient(),
		NilClient(),
	}

	var tests = []test{
		{
			name:  "empty call",
			l:     []GRPCLogger{},
			wants: nil,
		},
		{
			name:  "nil call",
			l:     nil,
			wants: nil,
		},
		{
			name:  "one logger",
			l:     []GRPCLogger{loggers[0]},
			wants: loggers[0],
		},
		{
			name: "multiple loggers",
			l:    []GRPCLogger{loggers[0], loggers[1], loggers[2]},
			wants: &multiLogger{
				loggers: []GRPCLogger{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "nested multiloggers",
			l:    []GRPCLogger{loggers[0], MultiLogger(loggers[1], loggers[2])},
			wants: &multiLogger{
				loggers: []GRPCLogger{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "add nil logger in the mix",
			l:    []GRPCLogger{loggers[0], loggers[1], nil},
			wants: &multiLogger{
				loggers: []GRPCLogger{loggers[0], loggers[1]},
			},
		},
		{
			name:  "add nil loggers in the mix",
			l:     []GRPCLogger{loggers[0], nil, nil},
			wants: loggers[0],
		},
		{
			name:  "add only loggers",
			l:     []GRPCLogger{nil, nil, nil},
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

		// test interface calls --
		// TODO (zalgonoise): add individual tests for these
		if ml != nil {
			// ChanneledLogger impl
			ml.Close()
			ml.Channels()

			// io.Writer impl
			ml.Write(event.New().Message("null").Build().Encode())

			// log.Logger impl
			ml.SetOuts(&bytes.Buffer{})
			ml.AddOuts(&bytes.Buffer{})
			ml.Prefix("null")
			ml.Sub("null")
			ml.Fields(map[string]interface{}{"ok": true})
			ml.IsSkipExit()

			// log.Printer impl
			ml.Output(event.New().Message("null").Build())
			ml.Log(event.New().Message("null").Build())
			ml.Print("null")
			ml.Println("null")
			ml.Printf("null")
			ml.Panic("null")
			ml.Panicln("null")
			ml.Panicf("null")
			ml.Fatal("null")
			ml.Fatalln("null")
			ml.Fatalf("null")
			ml.Error("null")
			ml.Errorln("null")
			ml.Errorf("null")
			ml.Warn("null")
			ml.Warnln("null")
			ml.Warnf("null")
			ml.Info("null")
			ml.Infoln("null")
			ml.Infof("null")
			ml.Debug("null")
			ml.Debugln("null")
			ml.Debugf("null")
			ml.Trace("null")
			ml.Traceln("null")
			ml.Tracef("null")
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
