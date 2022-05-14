package address

import (
	"reflect"
	"testing"
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
		var addr *ConnAddr = New(test.addr...)

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

		if !reflect.DeepEqual(keys, test.keys) {
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

// func TestNew(t *testing.T) {
// 	module := "ConnAddr"
// 	funcname := "New()"

// 	_ = module
// 	_ = funcname

// 	type test struct {
// 		name string
// 	}

// 	var tests = []test{}

// 	var verify = func(idx int, test test) {}

// 	for idx, test := range tests {
// 		verify(idx, test)
// 	}

// }
