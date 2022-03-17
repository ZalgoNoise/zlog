package client

type ConnAddr []string

func (a *ConnAddr) Strings() []string {
	return *a
}

func (a *ConnAddr) Len() int {
	return len(*a)
}

func (a *ConnAddr) Add(addr ...string) {
	if len(addr) == 0 || addr == nil {
		return
	}
	*a = append(*a, addr...)
}

func (a *ConnAddr) Reset() {
	new := ConnAddr([]string{})
	a = &new
}

func (a ConnAddr) Write(p []byte) (n int, err error) {
	a = append(a, string(p))
	return len(a), nil
}
