package event

// LogLevel type describes a numeric value for a log level with priority increasing in
// relation to its value
//
// LogLevel also implements the Stringer interface, used to convey this log level in a message
type LogLevel int32

const (
	LLTrace LogLevel = iota // LogLevel Trace
	LLDebug                 // LogLevel Debug
	LLInfo                  // LogLevel Info
	LLWarn                  // LogLevel Warning
	LLError                 // LogLevel Error
	LLFatal                 // LogLevel Fatal
	_                       // [reserved]
	_                       // [reserved]
	_                       // [reserved]
	LLPanic                 // LogLevel Panic
)

// String method is defined for LogLevel objects to implement the Stringer interface
//
// It returns the string to which this log level is mapped to, in `LogTypeVals`
func (ll LogLevel) String() string {
	return LogTypeVals[ll]
}

// Int method returns a LogLevel's value as an integer, to be used for comparison with
// input log level filters
func (ll LogLevel) Int() int {
	return int(ll)
}

var (
	// LogTypeVals is an enum map to convert LogLevels to its string representation
	LogTypeVals = map[LogLevel]string{
		0: "trace",
		1: "debug",
		2: "info",
		3: "warn",
		4: "error",
		5: "fatal",
		9: "panic",
	}

	// LogTypeVals is an enum map to convert LogLevels from its string representation
	// to an int value
	LogTypeKeys = map[string]int{
		"trace": 0,
		"debug": 1,
		"info":  2,
		"warn":  3,
		"error": 4,
		"fatal": 5,
		"panic": 9,
	}
)
