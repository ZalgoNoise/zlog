package client

import "google.golang.org/grpc"

type ConnAddr map[string]*grpc.ClientConn

func (a *ConnAddr) Map() map[string]*grpc.ClientConn {
	return *a
}

func (a *ConnAddr) Keys() []string {
	var keys []string
	for k, _ := range *a {
		keys = append(keys, k)
	}
	return keys
}

func (a ConnAddr) Get(k string) *grpc.ClientConn {
	if a == nil || len(a) == 0 {
		return nil
	}

	return a[k]
}

func (a ConnAddr) Set(k string, conn *grpc.ClientConn) {
	a[k] = conn
}

func (a *ConnAddr) Len() int {
	return len(*a)
}

func (a ConnAddr) Add(addr ...string) {
	if len(addr) == 0 || addr == nil {
		return
	}
	for _, address := range addr {
		if a[address] != nil {
			continue
		} else {
			a[address] = &grpc.ClientConn{}
		}
	}
}

func (a *ConnAddr) Reset() {
	new := ConnAddr(map[string]*grpc.ClientConn{})
	a = &new
}

func (a ConnAddr) Write(p []byte) (n int, err error) {
	a.Add(string(p))
	return a.Len(), nil
}
