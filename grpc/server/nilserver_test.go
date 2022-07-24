package server

import "testing"

func TestNilServer(t *testing.T) {
	module := "GRPCLogServer"
	funcname := "MultiLogger()"

	_ = module
	_ = funcname

	type test struct {
		name string
	}

	var tests = []test{
		{
			name: "creating a nil LogServer",
		},
	}

	var verify = func(idx int, test test) {
		s := NilServer()

		if _, ok := s.(*nilLogServer); !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] call did not output an obj of type *nilLogServer -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
		}

		// test interface calls
		if s != nil {
			s.Serve()
			s.Stop()
			s.Channels()
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
