package db

import (
	"errors"
	"io"
	"reflect"
	"testing"
)

type writeCloseBuffer struct {
	buf    []byte
	closed bool
	err    error
	short  bool
}

func (wcb *writeCloseBuffer) Write(p []byte) (n int, err error) {
	n = 0

	if wcb.err != nil {
		err = wcb.Error()
		defer wcb.setError(nil)

		return
	}

	wcb.buf = append(wcb.buf, p...)

	if wcb.short {
		n = len(p) / 2
		return
	}

	n = len(p)

	return
}

func (wcb *writeCloseBuffer) Close() error {
	if wcb.err != nil {
		err := wcb.Error()
		defer wcb.setError(nil)

		return err
	}

	wcb.closed = true
	return nil
}

func (wcb *writeCloseBuffer) setError(err error) { wcb.err = err }
func (wcb *writeCloseBuffer) setShort()          { wcb.short = true }
func (wcb *writeCloseBuffer) Error() error       { return wcb.err }
func (wcb *writeCloseBuffer) Reset() {
	wcb.buf = []byte{}
	wcb.closed = false
	wcb.err = nil
	wcb.short = false
}

func TestMultiWriteCloser(t *testing.T) {
	module := "MultiWriteCloser"
	funcname := "MultiWriteCloser()"

	type test struct {
		name  string
		input []io.WriteCloser
		wants io.WriteCloser
	}

	var wcBuffers = []*writeCloseBuffer{
		{}, {}, {}, {},
	}

	var tests = []test{
		{
			name: "single io.WriteCloser",
			input: []io.WriteCloser{
				wcBuffers[0],
			},
			wants: wcBuffers[0],
		},
		{
			name: "multi io.WriteCloser",
			input: []io.WriteCloser{
				wcBuffers[0],
				wcBuffers[1],
				wcBuffers[2],
			},
			wants: &multiWriteCloser{
				writers: []io.WriteCloser{
					wcBuffers[0],
					wcBuffers[1],
					wcBuffers[2],
				},
			},
		},
		{
			name: "nested multiWriteClosers (x1)",
			input: []io.WriteCloser{
				&multiWriteCloser{
					writers: []io.WriteCloser{
						wcBuffers[0],
						wcBuffers[1],
					},
				},
				wcBuffers[2],
				wcBuffers[3],
			},
			wants: &multiWriteCloser{
				writers: []io.WriteCloser{
					wcBuffers[0],
					wcBuffers[1],
					wcBuffers[2],
					wcBuffers[3],
				},
			},
		},
		{
			name: "nested multiWriteClosers (x2)",
			input: []io.WriteCloser{
				&multiWriteCloser{
					writers: []io.WriteCloser{
						wcBuffers[0],
						wcBuffers[1],
					},
				},
				&multiWriteCloser{
					writers: []io.WriteCloser{
						wcBuffers[2],
						wcBuffers[3],
					},
				},
			},
			wants: &multiWriteCloser{
				writers: []io.WriteCloser{
					wcBuffers[0],
					wcBuffers[1],
					wcBuffers[2],
					wcBuffers[3],
				},
			},
		},
		{
			name:  "zero io.WriteClosers",
			input: []io.WriteCloser{},
			wants: nil,
		},
		{
			name:  "nil io.WriteClosers",
			input: nil,
			wants: nil,
		},
	}

	var verify = func(idx int, test test, out io.WriteCloser) {
		if !reflect.DeepEqual(out, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] expected results mismatch: wanted %v ; got %v",
				idx,
				module,
				funcname,
				test.wants,
				out,
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
		out := MultiWriteCloser(test.input...)

		verify(idx, test, out)

	}
}

func TestMultiWriteCloserWrite(t *testing.T) {
	module := "MultiWriteCloser"
	funcname := "Write()"

	type test struct {
		name  string
		w     []*writeCloseBuffer
		short []bool
		err   []error
		input []byte
	}

	var wcBuffers = []*writeCloseBuffer{
		{}, {}, {},
	}

	var tests = []test{
		{
			name:  "simple write, single buffer",
			w:     []*writeCloseBuffer{wcBuffers[0]},
			short: []bool{false},
			err:   []error{nil},
			input: []byte("testing"),
		},
		{
			name:  "simple write, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{false, false, false},
			err:   []error{nil, nil, nil},
			input: []byte("testing"),
		},
		{
			name:  "single error write, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{false, false, false},
			err:   []error{nil, nil, errors.New("write error")},
			input: []byte("testing"),
		},
		{
			name:  "spread errors, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{false, false, false},
			err:   []error{errors.New("unable to find file"), nil, errors.New("write error")},
			input: []byte("testing"),
		},
		{
			name:  "short-write errors, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{true, true, false},
			err:   []error{ErrShortWrite, ErrShortWrite, nil},
			input: []byte("testing"),
		},
	}

	var init = func(test test) []io.WriteCloser {
		var wc []io.WriteCloser

		for idx, b := range test.w {
			b.Reset()

			if test.err[idx] != nil && test.err[idx] != ErrShortWrite {
				b.setError(test.err[idx])
			}

			if test.short[idx] {
				b.setShort()
			}
			wc = append(wc, b)
		}
		return wc
	}

	var cleanup = func(test test) {
		for _, b := range test.w {
			b.Reset()
		}
	}

	var verify = func(idx int, test test, w io.WriteCloser) {
		n, err := w.Write(test.input)

		if err != nil {
			if len(test.err) == 0 {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected error: %v",
					idx,
					module,
					funcname,
					err,
				)
				return
			}

			var errs = []error{err}
			var last error = err

			for {
				var inner error

				inner = errors.Unwrap(last)

				if inner == nil {
					break
				}
				last = inner
				errs = append(errs, inner)
			}

			var testErrs []error
			for _, e := range test.err {
				if e != nil {
					testErrs = append(testErrs, e)
				}
			}

			if len(errs) != len(testErrs) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] error length mismatch: wanted %v errors ; got %v -- n: %v -- %v",
					idx,
					module,
					funcname,
					len(testErrs),
					len(errs),
					n,
					errs,
				)
				return
			}

			for _, e := range errs {
				var ok bool
				for _, te := range testErrs {
					if errors.Is(e, te) {
						ok = true
					}
				}
				if !ok {
					t.Errorf(
						"#%v -- FAILED -- [%s] [%s] error mismatch: error %v does not match expected: %v",
						idx,
						module,
						funcname,
						e,
						testErrs,
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [%s] [%s]",
				idx,
				module,
				funcname,
			)
		}

	}

	for idx, test := range tests {
		wc := init(test)

		w := MultiWriteCloser(wc...)

		verify(idx, test, w)

		cleanup(test)
	}
}

func TestMultiWriteCloserClose(t *testing.T) {
	module := "MultiWriteCloser"
	funcname := "Close()"

	type test struct {
		name  string
		w     []*writeCloseBuffer
		short []bool
		err   []error
	}

	var wcBuffers = []*writeCloseBuffer{
		{}, {}, {},
	}

	var tests = []test{
		{
			name:  "simple write, single buffer",
			w:     []*writeCloseBuffer{wcBuffers[0]},
			short: []bool{false},
			err:   []error{nil},
		},
		{
			name:  "simple write, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{false, false, false},
			err:   []error{nil, nil, nil},
		},
		{
			name:  "single error write, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{false, false, false},
			err:   []error{nil, nil, errors.New("write error")},
		},
		{
			name:  "spread errors, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{false, false, false},
			err:   []error{errors.New("unable to find file"), nil, errors.New("write error")},
		},
		{
			name:  "short-write errors, multi buffer",
			w:     []*writeCloseBuffer{wcBuffers[0], wcBuffers[1], wcBuffers[2]},
			short: []bool{true, true, false},
			err:   []error{ErrShortWrite, ErrShortWrite, nil},
		},
	}

	var init = func(test test) []io.WriteCloser {
		var wc []io.WriteCloser

		for idx, b := range test.w {
			b.Reset()

			if test.err[idx] != nil && test.err[idx] != ErrShortWrite {
				b.setError(test.err[idx])
			}

			if test.short[idx] {
				b.setShort()
			}
			wc = append(wc, b)
		}
		return wc
	}

	var cleanup = func(test test) {
		for _, b := range test.w {
			b.Reset()
		}
	}

	var verify = func(idx int, test test, w io.WriteCloser) {
		err := w.Close()

		if err != nil {
			if len(test.err) == 0 {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected error: %v",
					idx,
					module,
					funcname,
					err,
				)
				return
			}

			var errs = []error{err}
			var last error = err

			for {
				var inner error

				inner = errors.Unwrap(last)

				if inner == nil {
					break
				}
				last = inner
				errs = append(errs, inner)
			}

			var testErrs []error
			for _, e := range test.err {
				if e != nil {
					testErrs = append(testErrs, e)
				}
			}

			if len(errs) != len(testErrs) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] error length mismatch: wanted %v errors ; got %v -- %v",
					idx,
					module,
					funcname,
					len(testErrs),
					len(errs),
					errs,
				)
				return
			}

			for _, e := range errs {
				var ok bool
				for _, te := range testErrs {
					if errors.Is(e, te) {
						ok = true
					}
				}
				if !ok {
					t.Errorf(
						"#%v -- FAILED -- [%s] [%s] error mismatch: error %v does not match expected: %v",
						idx,
						module,
						funcname,
						e,
						testErrs,
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [%s] [%s]",
				idx,
				module,
				funcname,
			)
		}

	}

	for idx, test := range tests {
		wc := init(test)

		w := MultiWriteCloser(wc...)

		verify(idx, test, w)

		cleanup(test)
	}
}
