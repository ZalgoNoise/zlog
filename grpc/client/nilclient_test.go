package client

import (
	"bytes"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestNilClient(t *testing.T) {
	module := "GRPCLogger"
	funcname := "MultiLogger()"

	_ = module
	_ = funcname

	type test struct {
		name string
	}

	var tests = []test{
		{
			name: "creating a nil GRPCLogger",
		},
	}

	var verify = func(idx int, test test) {
		c := NilClient()

		if _, ok := c.(*nilLogClient); !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] call did not output an obj of type *nilLogServer -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
		}

		// test interface calls
		if c != nil {
			// ChanneledLogger impl
			c.Close()
			c.Channels()

			// io.Writer impl
			c.Write(event.New().Message("null").Build().Encode())

			// log.Logger impl
			c.SetOuts(&bytes.Buffer{})
			c.AddOuts(&bytes.Buffer{})
			c.Prefix("null")
			c.Sub("null")
			c.Fields(map[string]interface{}{"ok": true})
			c.IsSkipExit()

			// log.Printer impl
			c.Output(event.New().Message("null").Build())
			c.Log(event.New().Message("null").Build())
			c.Print("null")
			c.Println("null")
			c.Printf("null")
			c.Panic("null")
			c.Panicln("null")
			c.Panicf("null")
			c.Fatal("null")
			c.Fatalln("null")
			c.Fatalf("null")
			c.Error("null")
			c.Errorln("null")
			c.Errorf("null")
			c.Warn("null")
			c.Warnln("null")
			c.Warnf("null")
			c.Info("null")
			c.Infoln("null")
			c.Infof("null")
			c.Debug("null")
			c.Debugln("null")
			c.Debugf("null")
			c.Trace("null")
			c.Traceln("null")
			c.Tracef("null")
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}

}
