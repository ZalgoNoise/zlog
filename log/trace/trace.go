package trace

import (
	"regexp"
	"runtime"
)

const (
	callStackMaxSize int    = 1024
	goRoutineRegexp  string = `^goroutine (\d*) \[(.*)\]:$`

	mKey      string = "method"
	rKey      string = "reference"
	idKey     string = "id"
	statusKey string = "status"
	stackKey  string = "stack"
	goRKey    string = "goroutine-"
)

type stacktrace struct {
	buf    []byte
	split  [][]byte
	stacks []goRoutine
	out    map[string]interface{}
}

type callStack struct {
	method    string
	reference string
}
type goRoutine struct {
	id     string
	status string
	stack  []callStack
}

// New function will return the current callstack in a map[string]interface{} format,
// to be added in an event's metadata
func New(all bool) map[string]interface{} {
	return newCallStack().
		getCallStack(all).
		splitCallStack().
		parseCallStack().
		mapCallStack().
		asMap()
}

func newCallStack() *stacktrace {
	return &stacktrace{}
}

func (s *stacktrace) getCallStack(all bool) *stacktrace {
	buf := make([]byte, callStackMaxSize)
	n := runtime.Stack(buf, all)
	s.buf = buf[0:n]

	return s
}

func (s *stacktrace) splitCallStack() *stacktrace {
	if s.buf == nil || len(s.buf) == 0 {
		s.getCallStack(false)
	}

	var line []byte

	for i := 0; i < len(s.buf); i++ {
		if s.buf[i] != 10 {
			line = append(line, s.buf[i])
		} else {
			s.split = append(s.split, line)
			line = []byte{}
		}
	}

	if len(line) != 0 {
		s.split = append(s.split, line)
	}

	return s
}

func (s *stacktrace) parseCallStack() *stacktrace {
	if s.split == nil || len(s.split) == 0 {
		s.splitCallStack()
	}

	cs := goRoutine{}
	regxGoRoutine := regexp.MustCompile(goRoutineRegexp)

	for i := 0; i < len(s.split); i++ {
		if regxGoRoutine.Match(s.split[i]) {
			if cs.id != "" || cs.status != "" {
				s.stacks = append(s.stacks, cs)
				cs = goRoutine{}
			}
			match := regxGoRoutine.FindStringSubmatch(string(s.split[i]))
			cs.id = match[1]
			cs.status = match[2]
		} else {
			if i+1 < len(s.split) && len(s.split[i+1]) > 0 && s.split[i+1][0] == 9 {

				s := callStack{
					method:    string(s.split[i]),
					reference: string(s.split[i+1][1:]),
				}
				cs.stack = append(cs.stack, s)
			}

		}
	}

	if cs.id != "" || cs.status != "" {
		s.stacks = append(s.stacks, cs)
	}

	return s
}

func (s *stacktrace) mapCallStack() *stacktrace {

	if s.stacks == nil || len(s.stacks) == 0 {
		s.parseCallStack()
	}

	field := map[string]interface{}{}

	for _, s := range s.stacks {
		stackMap := []map[string]interface{}{}
		for _, v := range s.stack {

			f := map[string]interface{}{
				mKey: v.method,
				rKey: v.reference,
			}

			stackMap = append(stackMap, f)
		}

		field[goRKey+s.id] = map[string]interface{}{
			idKey:     s.id,
			statusKey: s.status,
			stackKey:  stackMap,
		}

	}

	s.out = field

	return s
}

func (s *stacktrace) asMap() map[string]interface{} {
	if s.out == nil || len(s.out) <= 0 {
		s.mapCallStack()
	}
	return s.out
}
