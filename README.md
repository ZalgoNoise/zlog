# zlog
A lightweight Golang library to handle logging 

_________________________

<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/LoggerGopher-small.png" />
</p>


## Index

1. [Overview](#overview)
1. [Features](#features)
	1. [Simple API](#simple-api)
	1. [Highly configurable](#highly-configurable)
	1. [Feature-rich events](#feature-rich-events)
		1. [Data structure](#data-structure)
		1. [Event builder](#event-builder)
		1. [Structured metadata](#structured-metadata)
	1. [Different formatters](#different-formatters)
		1. [Text](#text)
		1. [JSON](#json)
		1. [BSON](#bson)
		1. [CSV](#csv)
		1. [XML](#xml)
		1. [Protobuf](#protobuf)
		1. [Gob](#gob)
1. [Installation](#installation)
1. [Usage](#usage)
1. [Integration](#integration)
1. [Benchmarks](#benchmarks)
1. [Contributing](#contributing)

_________________


### Overview 


This project started (like many others) as a means for me to learn and understand how logging works (in Go and in general), among other interesting Go design patterns. 

Basing myself off of the standard library `log` package, the goal was to create a new, _minimalist_ logger while introducing great features found in open-source projects like [logrus](https://github.com/sirupsen/logrus).

Very quickly it became apparent that this project had little or no minimalism as it grew, as I intended to add new features as I learned new technologies and techniques.

That being the case, the goal morphed from simplicity to feature-rich **and** developer-friendly at the same time -- using abstractions and wrappers to allow more complex configuration or behavior if the dev wants, while (trying to) keep it idiomatic when using simple or default configurations. 
 
_________________


### Features

This library provides a feature-rich structured logger, ready to write to many types of outputs (standard out / error, to buffers, to files and databases) and over-the-wire (via gRPC).

#### Simple API

<!-- 
	Add an image of a `log.Infof()` call
-->

The [Logger interface](log/logger.go#L95) in this library provides a set complete set of idiomatic methods which allow to either control the logger:


```go
type Logger interface {
	io.Writer
	Printer

	SetOuts(outs ...io.Writer) Logger
	AddOuts(outs ...io.Writer) Logger
	Prefix(prefix string) Logger
	Sub(sub string) Logger
	Fields(fields map[string]interface{}) Logger
	IsSkipExit() bool
}
```

...or to use its [Printer interface](log/print.go#L18) and print messages in the `fmt.Print()` / `fmt.Println()` / `fmt.Printf()` way:


```go
type Printer interface {
	Output(m *event.Event) (n int, err error)
	Log(m ...*event.Event)

	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicln(v ...interface{})
	Panicf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalln(v ...interface{})
	Fatalf(format string, v ...interface{})

	Error(v ...interface{})
	Errorln(v ...interface{})
	Errorf(format string, v ...interface{})

	Warn(v ...interface{})
	Warnln(v ...interface{})
	Warnf(format string, v ...interface{})

	Info(v ...interface{})
	Infoln(v ...interface{})
	Infof(format string, v ...interface{})

	Debug(v ...interface{})
	Debugln(v ...interface{})
	Debugf(format string, v ...interface{})

	Trace(v ...interface{})
	Traceln(v ...interface{})
	Tracef(format string, v ...interface{})
}
```


#### Highly configurable 


<!-- 
	Add an image of `log.New()` with added configuration
-->

Creating a new logger with, for example, [`log.New()`](log/logger.go#L134) takes any number of configurations (including none, for the default configuration). This allows added modularity to the way your logger should behave.

Method / Variable | Description
:--:|:--:
[`NilLogger()`](log/conf.go#L154) | create a __nil-logger__ (that doesn't write anything to anywhere)
[`WithPrefix(string)`](log/conf.go#L165) | set a __default prefix__
[`WithSub(string)`](log/conf.#L172) | set a __default sub-prefix__
[`WithOut(...io.Writer)`](log/format.go#L179) | set (a) __default writer(s)__
[`WithFormat(LogFormatter)`](log/format.go#L33) | set the formatter for the log event output content
[`SkipExit` config](log/conf.go#L78) | set the __skip-exit option__ (to skip `os.Exit(1)` and `panic()` calls)
[`WithFilter(event.Level)`](log/conf.go#L203) | set a __log-level filter__
[`WithDatabase(...io.WriteCloser)`](log/conf.go#L211) | set a __database writer__ (if using a database)

Beyond the functions and preset configurations above, the package also exposes the following preset for the [default config](log/conf.go#L56):

```go
var DefaultConfig LoggerConfig = &multiconf{
  confs: []LoggerConfig{
    WithFormat(TextColorLevelFirst),
    WithOut(),
    WithPrefix(event.Default_Event_Prefix),
  },
}
```

...and the following [(initialized) presets](log/conf.go#L77) for several useful "defaults":

```go
var (
	DefaultCfg    = LoggerConfigs[0]  // default LoggerConfig
	SkipExit      = LoggerConfigs[1]  // skip-exits LoggerConfig
	StdOut        = LoggerConfigs[7]  // os.Stderr LoggerConfig
	PrefixDefault = LoggerConfigs[8]  // default-prefix LoggerConfig
	FilterInfo    = LoggerConfigs[9]  // Info-filtered LoggerConfig
	FilterWarn    = LoggerConfigs[10] // Warn-filtered LoggerConfig
	FilterError   = LoggerConfigs[11] // Error-filtered LoggerConfig
	NilConfig     = LoggerConfigs[12] // empty / nil LoggerConfig
	EmptyConfig   = LoggerConfigs[12] // empty / nil LoggerConfig
)
```

#### Feature-rich events

<!--
	Add an image of `event.New()....Build()`
-->

##### Data structure

The events are defined in a protocol buffer format, in [`proto/event.proto`](proto/event.proto#L20); to give it a seamless integration as a gRPC logger's request message:

```protobuf
message Event {
    optional google.protobuf.Timestamp time = 1;
    optional string prefix = 2 [ default = "log" ];
    optional string sub = 3;
    optional Level level = 4 [ default = info ];
    required string msg = 5;
    optional google.protobuf.Struct meta = 6;
}
```


##### Event builder


An event is created with a builder pattern, by defining a set of elements before _spitting out_ the resulting object. 

The event builder will allow chaining methods after [`event.New()`](log/event/builder.go#L29) until the [`Build()`](log/event/builder.go#L107) method is called. Below is a list of all available methods to the [`event.EventBuilder`](log/event/builder.go#L14):

Method signature | Description
:--:|:--:
[`Prefix(p string) *EventBuilder`](log/event/builder.go#L47) | set the prefix element
[`Sub(s string) *EventBuilder`](log/event/builder.go#L54) | set the sub-prefix element
[`Message(m string) *EventBuilder`](log/event/builder.go#L61) | set the message body element
[`Level(l Level) *EventBuilder`](log/event/builder.go#L68) | set the level element
[`Metadata(m map[string]interface{}) *EventBuilder`](log/event/builder.go#L75) | set (or add to) the metadata element
[`CallStack(all bool) *EventBuilder`](log/event/builder.go#L94) | grab the current call stack, and add it as a "callstack" object in the event's metadata
[`Build() *Event`](log/event/builder.go#L107) | build an event with configured elements, defaults applied where needed, and by adding a timestamp

##### Structured metadata

Metadata is added to the as a `map[string]interface{}` which is compatible with JSON output (for the most part, for most the common data types). This allows a list of key-value pairs where the key is always a string (an identifier) and the value is the data itself, regardless of the type.

The event package also exposes a unique type ([`event.Field`](log/event/field.go#L11)):

```go
// Field type is a generic type to build Event Metadata
type Field map[string]interface{}
```

The [`event.Field`](log/event/field.go#L11) type exposes three methods to allow fast / easy conversion to [`structpb.Struct`](https://pkg.go.dev/google.golang.org/protobuf/types/known/structpb#Struct) pointers; needed for the protobuf encoders:

```go
// AsMap method returns the Field in it's (raw) string-interface{} map format
func (f Field) AsMap() map[string]interface{} {}

// ToStructPB method will convert the metadata in the protobuf Event as a pointer to a
// structpb.Struct, returning this and an error if any.
//
// The metadata (a map[string]interface{}) is converted to JSON (bytes), and this data is
// unmarshalled into a *structpb.Struct object.
func (f Field) ToStructPB() (*structpb.Struct, error) {}

// Encode method is similar to Field.ToStructPB(), but it does not return any errors.
func (f Field) Encode() *structpb.Struct {}
```


#### Different formatters

The logger can output events in several different formats, listed below:

##### Text

<!-- 
	Add an image of text formatter's output in a terminal
-->

The text formatter allows an array of options, with the text formatter sub-package exposing a builder to create a text formatter. Below is the list of methods you can expect when calling [`text.New()....Build()`](log/format/text/text.go#L87):

Method | Description
:--:|:--:
[`Time(LogTimestamp)](log/format/text/text.go#L93)` | define the timestamp format, based on the exposed list of timestamps, from the table below
[`LevelFirst()`](log/format/text/text.go#L100) | place the log level as the first element in the line
[`DoubleSpace()`](log/format/text/text.go#L107) | place double-tab-spaces between elements (`\t\t`)
[`Color()`](log/format/text/text.go#L114) | add color to log levels (it is skipped on Windows CLI, as it doesn't support it)
[`Upper()`](log/format/text/text.go#L121) | make log level, prefix and sub-prefix uppercase
[`NoTimestamp()`](log/format/text/text.go#L128) | skip adding the timestamp element
[`NoHeaders()`](log/format/text/text.go#L135) | skip adding the prefix and sub-prefix elements
[`NoLevel()`](log/format/text/text.go#L142) | skip adding the log level element

Regarding the timestamp constraints, please note the available timestamps for the text formatter:

Constant | Description
:--:|:--:
[`LTRFC3339Nano`](log/format/text/text.go#L48) | Follows the standard in `time.RFC3339Nano`
[`LTRFC3339`](log/format/text/text.go#L49) | Follows the standard in `time.RFC3339`
[`LTRFC822Z`](log/format/text/text.go#L50) | Follows the standard in `time.RFC822Z`
[`LTRubyDate`](log/format/text/text.go#L51) | Follows the standard in `time.RubyDate`
[`LTUnixNano`](log/format/text/text.go#L52) | Displays a unix timestamp, in nanos
[`LTUnixMilli`](log/format/text/text.go#L53) | Displays a unix timestamp, in millis
[`LTUnixMicro`](log/format/text/text.go#L54) | Displays a unix timestamp, in micros

The library also exposes a few initialized preset configurations using text formatters, as in the list below. While these are [`LoggerConfig`](log/conf.go#L21) presets, they're a wrapper for [the same formatter](log/format.go#L18), which is also available by not including the `Cfg` prefix:

```go
var (
	CfgFormatText                  = WithFormat(text.New().Build())  // default
	CfgTextLongDate                = WithFormat(text.New().Time(text.LTRFC3339).Build())  // with a RFC3339 date format
	CfgTextShortDate               = WithFormat(text.New().Time(text.LTRFC822Z).Build())  // with a RFC822Z date format
	CfgTextRubyDate                = WithFormat(text.New().Time(text.LTRubyDate).Build())  // with a RubyDate date format
	CfgTextDoubleSpace             = WithFormat(text.New().DoubleSpace().Build()) // with double spaces
	CfgTextLevelFirstSpaced        = WithFormat(text.New().DoubleSpace().LevelFirst().Build()) // with level-first and double spaces
	CfgTextLevelFirst              = WithFormat(text.New().LevelFirst().Build()) // with level-first
	CfgTextColorDoubleSpace        = WithFormat(text.New().DoubleSpace().Color().Build()) // with color and double spaces
	CfgTextColorLevelFirstSpaced   = WithFormat(text.New().DoubleSpace().LevelFirst().Color().Build()) // with color, level-first and double spaces
	CfgTextColorLevelFirst         = WithFormat(text.New().LevelFirst().Color().Build()) // with color and level-first
	CfgTextColor                   = WithFormat(text.New().Color().Build()) // with color
	CfgTextOnly                    = WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()) // with only the text content
	CfgTextNoHeaders               = WithFormat(text.New().NoHeaders().Build()) // without headers
	CfgTextNoTimestamp             = WithFormat(text.New().NoTimestamp().Build()) // without timestamp
	CfgTextColorNoTimestamp        = WithFormat(text.New().NoTimestamp().Color().Build()) // without timestamp
	CfgTextColorUpperNoTimestamp   = WithFormat(text.New().NoTimestamp().Color().Upper().Build()) // without timestamp and uppercase headers
)

```


##### JSON


<!-- 
	Add an image of JSON formatter's output in a terminal
-->

The JSON formatter allow generating JSON-formatted events in different ways. These formatters are already initialized as [`LoggerConfig`](log/format.go#L69) and [`LogFormatter`](log/format.go#L39) objects.

[This formatter](log/format/json/json.go#L13) allows creating JSON events separated by newlines or not, and also to optionally add indentation:

```go
type FmtJSON struct {
	SkipNewline bool
	Indent      bool
}
```

Also note how the [`LoggerConfig`](log/format.go#L69) presets are exposed. While these are a wrapper for the same formatter, they are also available as [`LogFormatter`](log/format.go#L39) by not including the `Cfg` prefix:

```go
var (
	CfgFormatJSON                  = WithFormat(&json.FmtJSON{})  // default
	CfgFormatJSONSkipNewline       = WithFormat(&json.FmtJSON{SkipNewline: true}) // with a skip-newline config
	CfgFormatJSONIndentSkipNewline = WithFormat(&json.FmtJSON{SkipNewline: true, Indent: true}) // with a skip-newline and indentation config
	CfgFormatJSONIndent            = WithFormat(&json.FmtJSON{Indent: true}) // with an indentation config
)
```



##### BSON


<!-- 
	Add an image of BSON formatter's output in a terminal
-->


##### CSV


<!-- 
	Add an image of BSON formatter's output in a terminal
-->


##### XML


<!-- 
	Add an image of BSON formatter's output in a terminal
-->


##### Protobuf


<!-- 
	Add an image of BSON formatter's output in a terminal
-->


##### Gob


<!-- 
	Add an image of BSON formatter's output in a terminal
-->
________________


### Installation

_________________

### Usage

_______________

### Integration

_______________


### Benchmarks

______________

### Contributing

_____________


_WIP: this repository is in a beta stage and is not yet usable for production_
