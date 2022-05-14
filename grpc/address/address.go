package address

import "google.golang.org/grpc"

// ConnAddr is a map of addresses which refer to pointers to grpc.ClientConn
//
// This map is used to build dynamic connections for gRPC Loggers.
type ConnAddr map[string]*grpc.ClientConn

// New function will take in any number of addr and initialize a ConnAddr object with
// those addresses. Then, returns a pointer to this ConnAddr object.
func New(addr ...string) *ConnAddr {
	if addr == nil || len(addr) == 0 {
		return nil
	}

	var input []string

	for _, a := range addr {
		if a == "" {
			continue
		}

		input = append(input, a)
	}

	if len(input) == 0 {
		return nil
	}

	var a = &ConnAddr{}
	a.Add(input...)

	return a

}

// Map method will return a ConnAddr object in a map[string]*grpc.ClientConn format
func (a *ConnAddr) AsMap() map[string]*grpc.ClientConn {
	return *a
}

// Add method will allocate the input strings as entries in the map, with initialized
// pointers to grpc.ClientConn
func (a *ConnAddr) Add(addr ...string) {
	if addr == nil || len(addr) == 0 {
		return
	}

	v := *a

	for _, address := range addr {
		if address == "" {
			continue
		}
		if v[address] != nil {
			continue
		} else {
			v[address] = &grpc.ClientConn{}
		}
	}
	a = &v
}

// Keys method will return a ConnAddr object's keys (its addresses) in a slice of strings
func (a *ConnAddr) Keys() []string {
	var keys []string
	for k := range *a {
		keys = append(keys, k)
	}
	return keys
}

// Get method will return the pointer to a grpc.ClientConn, as referenced in the input
// address k
func (a *ConnAddr) Get(k string) *grpc.ClientConn {
	if a == nil || len(*a) == 0 {
		return nil
	}
	v := *a
	return v[k]
}

// Set method will allocate the input connection to the input string, within the ConnAddr
// map (overwritting it if existing)
func (a *ConnAddr) Set(k string, conn *grpc.ClientConn) {
	v := *a
	v[k] = conn

	a = &v
}

// Len method will return the size of the ConnAddr map
func (a *ConnAddr) Len() int {
	return len(*a)
}

// Reset method will overwrite the existing ConnAddr map with a new, empty one.
func (a *ConnAddr) Reset() {
	new := ConnAddr(map[string]*grpc.ClientConn{})
	a = &new
}

// Unset method will remove the input addr strings from the ConnAddr map, if existing
func (a *ConnAddr) Unset(addr ...string) {
	if len(addr) == 0 || addr == nil {
		return
	}

	v := *a

	for _, address := range addr {
		delete(v, address)
	}

	a = &v
}

// Write method is an implementation of io.Writer, so that the ConnAddr map can be used
// in a gRPC Logger's SetOuts() and AddOuts() methods. These need to conform with the
// Logger interface that implements these methods with a variatic number of io.Writer.
//
// For the same layer of compatibility to be possible in a gRPC Logger (who will write
// its log entries in a remote server), it uses these methods to implement its way of
// altering the existing connections, instead of dismissing this part of the implementation
// all together.
//
// ...That being said -- this is not any io.Writer.
func (a *ConnAddr) Write(p []byte) (n int, err error) {
	a.Add(string(p))
	return a.Len(), nil
}
