package store

import "testing"

func TestNilWritterWrite(t *testing.T) {
	module := "nilWriter"
	funcname := "Write()"

	type test struct {
		name string
		buf  []byte
		n    int
		err  error
	}

	var tests = []test{
		{
			name: "test empty buffer",
			buf:  []byte{},
			n:    0,
			err:  nil,
		},
		{
			name: "test buffer with content",
			buf:  []byte("testing"),
			n:    0,
			err:  nil,
		},
	}

	var verify = func(idx int, test test, n int, err error) {
		if err != test.err {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: wanted %v ; got %v",
				idx,
				module,
				funcname,
				test.err,
				err,
			)
			return
		}

		if n != test.n {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected number of bytes written: wanted %v ; got %v",
				idx,
				module,
				funcname,
				test.n,
				n,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s]",
			idx,
			module,
			funcname,
		)

	}

	for idx, test := range tests {
		var nilw = new(nilWritter)

		n, err := nilw.Write(test.buf)

		verify(idx, test, n, err)
	}
}
