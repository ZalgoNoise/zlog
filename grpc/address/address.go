package address

import "google.golang.org/grpc"

type ConnAddr map[string]*grpc.ClientConn

func (a *ConnAddr) Map() map[string]*grpc.ClientConn {
	return *a
}

func (a *ConnAddr) Keys() []string {
	var keys []string
	for k := range *a {
		keys = append(keys, k)
	}
	return keys
}

func (a *ConnAddr) Get(k string) *grpc.ClientConn {
	if a == nil || len(*a) == 0 {
		return nil
	}
	v := *a
	return v[k]
}

func (a *ConnAddr) Set(k string, conn *grpc.ClientConn) {
	v := *a
	v[k] = conn

	a = &v
}

func (a *ConnAddr) Len() int {
	return len(*a)
}

func (a *ConnAddr) Add(addr ...string) {
	if len(addr) == 0 || addr == nil {
		return
	}

	v := *a

	for _, address := range addr {
		if v[address] != nil {
			continue
		} else {
			v[address] = &grpc.ClientConn{}
		}
	}
	a = &v
}

func (a *ConnAddr) Reset() {
	new := ConnAddr(map[string]*grpc.ClientConn{})
	a = &new
}

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

func (a *ConnAddr) Write(p []byte) (n int, err error) {
	a.Add(string(p))
	return a.Len(), nil
}

func New(addr ...string) *ConnAddr {
	if len(addr) == 0 {
		return nil
	}

	var a = &ConnAddr{}
	a.Add(addr...)

	return a

}
