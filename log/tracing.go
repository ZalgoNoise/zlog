package log

import (
	"fmt"
	"regexp"
	"runtime"
)

type stacktrace struct {
	buf    []byte
	split  [][]byte
	stacks []goRoutine
	out    Field
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

func newCallStack() *stacktrace {
	return &stacktrace{}
}

func (s *stacktrace) getCallStack(all bool) *stacktrace {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, all)
	s.buf = buf[0:n]

	fmt.Printf("\n\n%s\n\n", string(s.buf))

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
	regxGoRoutine := regexp.MustCompile(`^goroutine (\d*) \[(.*)\]:$`)

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
			if i+1 < len(s.split) && s.split[i+1][0] == 9 {

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

	field := Field{}

	for _, s := range s.stacks {
		stackMap := []Field{}
		for _, v := range s.stack {

			f := Field{
				"method":    v.method,
				"reference": v.reference,
			}

			stackMap = append(stackMap, f)
		}

		field["goroutine-"+s.id] = Field{
			"id":     s.id,
			"status": s.status,
			"stack":  stackMap,
		}

	}

	s.out = field

	return s
}

func (s *stacktrace) toMap() map[string]interface{} {
	if s.out == nil {
		s.mapCallStack()
	}
	return s.out
}

func (s *stacktrace) toField() Field {
	if s.out == nil {
		s.mapCallStack()
	}
	return s.out
}
