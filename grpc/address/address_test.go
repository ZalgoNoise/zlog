package address

import (
	"testing"

	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	module := "ConnAddr"
	funcname := "New() / AsMap()"

	_ = module
	_ = funcname

	type test struct {
		name string
		addr []string
		len  int
	}

	var tests = []test{
		{
			name: "one address",
			addr: []string{"example.com"},
			len:  1,
		},
		{
			name: "multiple addresses",
			addr: []string{"example.com", "example.net"},
			len:  2,
		},
		{
			name: "empty string",
			addr: []string{""},
			len:  0,
		},
		{
			name: "multiple addresses with emtpy string",
			addr: []string{"example.com", "", "example.net"},
			len:  2,
		},
		{
			name: "nil input",
			addr: nil,
			len:  0,
		},
		{
			name: "multiple addresses, all empty",
			addr: []string{"", "", ""},
			len:  0,
		},
	}

	var verify = func(idx int, test test) {
		var addr *ConnAddr

		if test.addr == nil {
			addr = New()
		} else {
			addr = New(test.addr...)
		}

		if addr == nil && test.len != 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil map: wanted %v items ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.len,
				nil,
				test.name,
			)
			return
		} else if addr == nil && test.len == 0 {
			return
		}

		if len(addr.AsMap()) != test.len {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.len,
				len(addr.AsMap()),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestAdd(t *testing.T) {
	module := "ConnAddr"
	funcname := "Add()"

	_ = module
	_ = funcname

	type test struct {
		name string
		c    *ConnAddr
		addr []string
		len  int
	}

	var tests = []test{
		{
			name: "one address",
			c:    New("domain.com"),
			addr: []string{"example.com"},
			len:  2,
		},
		{
			name: "same address",
			c:    New("domain.com"),
			addr: []string{"domain.com"},
			len:  1,
		},
		{
			name: "multiple addresses",
			c:    New("domain.com"),
			addr: []string{"example.com", "example.net"},
			len:  3,
		},
		{
			name: "empty string",
			c:    New("domain.com"),
			addr: []string{""},
			len:  1,
		},
		{
			name: "multiple addresses with emtpy string",
			c:    New("domain.com"),
			addr: []string{"example.com", "", "example.net"},
			len:  3,
		},
		{
			name: "nil input",
			c:    New("domain.com"),
			addr: nil,
			len:  1,
		},
		{
			name: "multiple addresses, all empty",
			c:    New("domain.com"),
			addr: []string{"", "", ""},
			len:  1,
		},
	}

	var verify = func(idx int, test test) {
		test.c.Add(test.addr...)

		if len(test.c.AsMap()) != test.len {
			if len(test.c.AsMap()) != test.len {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					test.len,
					len(test.c.AsMap()),
					test.name,
				)
				return
			}
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestKeys(t *testing.T) {
	module := "ConnAddr"
	funcname := "Keys()"

	_ = module
	_ = funcname

	type test struct {
		name string
		c    *ConnAddr
		keys []string
	}

	var tests = []test{
		{
			name: "one address",
			c:    New("domain.com"),
			keys: []string{"domain.com"},
		},
		{
			name: "multiple addresses",
			c:    New("domain.com", "example.com", "example.net"),
			keys: []string{"domain.com", "example.com", "example.net"},
		},
	}

	var verify = func(idx int, test test) {
		keys := test.c.Keys()

		if len(keys) != len(test.keys) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				len(test.keys),
				len(keys),
				test.name,
			)
			return
		}

		var spree []bool
		for a := 0; a < len(test.keys); a++ {
			for b := 0; b < len(keys); b++ {
				if test.keys[a] == keys[b] {
					spree = append(spree, true)
				}
			}
		}

		// if !reflect.DeepEqual(keys, test.keys) {
		if len(spree) != len(test.keys) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.keys,
				keys,
				test.name,
			)
			return
		}

		// for _, k := range keys {
		// 	conn := test.c.Get(k)
		// 	if conn == nil {
		// 		t.Errorf(
		// 			"#%v -- FAILED -- [%s] [%s] output error in key %s: wanted a pointer to grpc.ClientConn ; got %v -- action: %s",
		// 			idx,
		// 			module,
		// 			funcname,
		// 			k,
		// 			nil,
		// 			test.name,
		// 		)
		// 		return
		// 	}
		// }
	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestGet(t *testing.T) {
	module := "ConnAddr"
	funcname := "Get()"

	_ = module
	_ = funcname

	type test struct {
		name string
		c    *ConnAddr
		keys []string
		ok   bool
	}

	var tests = []test{
		{
			name: "one address",
			c:    New("domain.com"),
			keys: []string{"domain.com"},
			ok:   true,
		},
		{
			name: "multiple addresses",
			c:    New("domain.com", "example.com", "example.net"),
			keys: []string{"domain.com", "example.com", "example.net"},
			ok:   true,
		},
		{
			name: "non-existing address",
			c:    New("domain.com"),
			keys: []string{"example.com"},
		},
		{
			name: "Get() from empty map",
			c:    &ConnAddr{},
			keys: []string{"example.com"},
		},
	}

	var verify = func(idx int, test test) {
		for _, k := range test.keys {
			conn := test.c.Get(k)
			if conn == nil && test.ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] output error in key %s: wanted a pointer to grpc.ClientConn ; got %v -- action: %s",
					idx,
					module,
					funcname,
					k,
					nil,
					test.name,
				)
				return
			}
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestSet(t *testing.T) {
	module := "ConnAddr"
	funcname := "Set() / Len()"

	_ = module
	_ = funcname

	type test struct {
		name string
		c    *ConnAddr
		keys []string
		len  int
	}

	var tests = []test{
		{
			name: "one address",
			c:    New("test.com"),
			keys: []string{"domain.com"},
			len:  2,
		},
		{
			name: "multiple addresses",
			c:    New("test.com"),
			keys: []string{"domain.com", "example.com", "example.net"},
			len:  4,
		},
		{
			name: "non-existing address",
			c:    New("test.com"),
			keys: []string{"example.com"},
			len:  2,
		},
		{
			name: "Get() from empty map",
			c:    &ConnAddr{},
			keys: []string{"example.com"},
			len:  1,
		},
	}

	var verify = func(idx int, test test) {
		for _, k := range test.keys {
			test.c.Set(k, &grpc.ClientConn{})
		}

		if test.c.Len() != test.len {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.len,
				test.c.Len(),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestReset(t *testing.T) {
	module := "ConnAddr"
	funcname := "Reset()"

	_ = module
	_ = funcname

	type test struct {
		name string
		c    *ConnAddr
	}

	var tests = []test{
		{
			name: "one address",
			c:    New("test.com"),
		},
		{
			name: "multiple addresses",
			c:    New("domain.com", "example.com", "example.net"),
		},
		{
			name: "empty ConnAddr pointer",
			c:    &ConnAddr{},
		},
	}

	var verify = func(idx int, test test) {
		test.c.Reset()

		if test.c.Len() != 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				0,
				test.c.Len(),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestUnset(t *testing.T) {
	module := "ConnAddr"
	funcname := "Unset()"

	_ = module
	_ = funcname

	type test struct {
		name string
		c    *ConnAddr
		keys []string
		len  int
	}

	var tests = []test{
		{
			name: "no addresses",
			c:    New("test.com"),
			keys: []string{},
			len:  1,
		},
		{
			name: "one address",
			c:    New("test.com"),
			keys: []string{"test.com"},
			len:  0,
		},
		{
			name: "multiple addresses",
			c:    New("domain.com", "example.com", "example.net"),
			keys: []string{"domain.com", "example.com"},
			len:  1,
		},
		{
			name: "absent address",
			c:    New("test.com"),
			keys: []string{"example.com"},
			len:  1,
		},
		{
			name: "empty ConnAddr pointer",
			c:    &ConnAddr{},
			keys: []string{"example.com"},
			len:  0,
		},
	}

	var verify = func(idx int, test test) {
		test.c.Unset(test.keys...)

		if test.c.Len() != test.len {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				0,
				test.c.Len(),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWrite(t *testing.T) {
	module := "ConnAddr"
	funcname := "Write()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		c     *ConnAddr
		input []byte
		len   int
	}

	var tests = []test{
		{
			name:  "no addresses",
			c:     New("test.com"),
			input: []byte("example.com"),
			len:   2,
		},
		{
			name:  "multiple addresses",
			c:     New("domain.com", "example.com", "example.net"),
			input: []byte("test.com"),
			len:   4,
		},
		{
			name:  "multiple addresses",
			c:     New("domain.com", "example.com", "example.net"),
			input: []byte{},
			len:   3,
		},
	}

	var verify = func(idx int, test test) {
		n, err := test.c.Write(test.input)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if n != test.len {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				0,
				test.c.Len(),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
