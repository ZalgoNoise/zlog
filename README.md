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
	1. [Data Stores](#data-stores)
		1. [Writer interface](#writer-interface)
		1. [Logfile](#logfile)
		1. [Databases](#databases)
			1. [SQLite](#sqlite)
			1. [MySQL](#mysql)
			1. [PostgreSQL](#postgresql)
			1. [MongoDB](#mongodb)
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

##### Log levels

Log levels are defined as a protobuf enum, as [`Level` enum](proto/event.proto#L9):

```protobuf
enum Level {
    trace = 0;
    debug = 1;
    info = 2;
    warn = 3;
    error = 4;
    fatal = 5;
    reserved 6 to 8;
    panic = 9;
}
```

The [generated code](log/event/event.pb.go) creates a type and two maps which set these levels:

```go
type Level int32

const (
	Level_trace Level = 0
	Level_debug Level = 1
	Level_info  Level = 2
	Level_warn  Level = 3
	Level_error Level = 4
	Level_fatal Level = 5
	Level_panic Level = 9
)

// Enum value maps for Level.
var (
	Level_name = map[int32]string{
		0: "trace",
		1: "debug",
		2: "info",
		3: "warn",
		4: "error",
		5: "fatal",
		9: "panic",
	}
	Level_value = map[string]int32{
		"trace": 0,
		"debug": 1,
		"info":  2,
		"warn":  3,
		"error": 4,
		"fatal": 5,
		"panic": 9,
	}
)
```

This way, the enum is ready for changes in a consistent and seamless way. Here is how you'd define a log level as you create a new event:

```go
// import "github.com/zalgonoise/zlog/log/event"

e := event.New().
		   Message("Critical warning!"). // add a message body to event
		   Level(event.Level_warn).		 // set log level
		   Build()						 // build it
```

The [`Level` type](log/event/event.pb.go#L25) also has an exposed (custom) method, [`Int() int32`](log/event/level.go#L6), which acts as a quick converter from the map value to an `int32` value.

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

#### Data stores

##### Writer interface


<!-- 
	Add an image of Logger.Write() being called
-->


Not only [`Logger` interface](log/logger.go#L95) uses the [`io.Writer` interface](https://pkg.go.dev/io#Writer) to write to its outputs with its [`Output()` method](log/print.go#L88), it also implements it in its own [`Write()` method](log/logger.go#L307) so it can be used directly as one. This gives the logger more flexibility as it can be vastly integrated with other modules.


The the input slice of bytes is decoded, in case the input is an encoded [`event.Event`](log/event/event.pb.go#L96). If the conversion is successful, the input event is logged as-is.

If it is not an [`event.Event`](log/event/event.pb.go#L96) (there will be an error from the [`Decode()` method](log/event/event.go#L32)), then a new message is created where:
- Log level is set to the default value (info)
- Prefix, sub-prefix and metadata are added from the logger's configuration, or defaults if they aren't set.
- input byte stream is converted to a string, and that will be log event message body:

```go
func (l *logger) Write(p []byte) (n int, err error) {
	// decode bytes
	m, err := event.Decode(p)

	// default to printing message as if it was a byte slice payload for the log event body
	if err != nil {
		return l.Output(event.New().
			Level(event.Default_Event_Level).
			Prefix(l.prefix).
			Sub(l.sub).
			Message(string(p)).
			Metadata(l.meta).
			Build())
	}

	// print message
	return l.Output(m)
}
```

##### Logfile 


<!-- 
	Add an image of a directory containing some logfiles
-->

This library also provides a simple [`Logfile`](store/fs/logfile.go#L18) (an actual file in the disk where log entries are written to) configuration with appealing features for simple applications.

```go
type Logfile struct {
	path   string
	file   *os.File
	size   int64
	rotate int
}
```

The [`Logfile`](store/fs/logfile.go#L18) exposes a few methods that could be helpful to keep the events organized:

Method | Description
:--:|:--:
[`MaxSize(mb int) *Logfile`](store/fs/logfile.go#L59) | sets the rotation indicator for the Logfile, or, the target size when should the logfile be rotated (in MBs)
[`Size() (int64, error)`](store/fs/logfile.go#L139) | a wrapper for an [`os.File.Stat()`](https://pkg.go.dev/os#File.Stat) followed by [`fs.FileInfo.Size()`](https://pkg.go.dev/io/fs#FileInfo)
[`IsTooHeavy() bool`](store/fs/logfile.go#L151) | verify the file's size and rotate it if exceeding the set maximum weight (in the Logfile's rotate element)
[`Write(b []byte) (n int, err error)`](store/fs/logfile.go#L174) | implement the io.Writer interface, for Logfile to be compatible with Logger as an output to be used


##### Databases

It's perfectly possible to write log events to a database instead of the terminal, a buffer, or a file. It makes it more reliable for a larger scale operation or for the long-run.

This library leverages an ORM to handle interactions with most of the databases, for the sake of simplicity and streamlined testing -- these should focus on using a database as a writer, and not re-testing the database connections, configurations, etc. This is why an ORM is being used. This library uses [GORM](https://gorm.io/) for this purpose.

Databases are not configured to loggers as an [`io.Writer` interface](https://pkg.go.dev/io#Writer) using the [`WithOut()` method](log/conf.go#L179), but with their dedicated [`WithDatabase()` method](log/conf.go#L211). This takes an [`io.WriterCloser` interface](https://pkg.go.dev/io#WriteCloser).


To create this [`io.WriterCloser`](https://pkg.go.dev/io#WriteCloser), either the database package's appropriate `New()` method can be used; or by using its package function for the same purpose, `WithXxx()`.

Note the available database writers, and their features:

##### SQLite


<!-- 
	Add an image of a SQLite logger config and calls (?)
-->



Symbol | Type | Description
:--:|:--:|:--:
[`New(path string) (sqldb io.WriteCloser, err error)`](store/db/sqlite/sqlite.go#L22) | function | takes in a path to a .db file; and create a new instance of a SQLite3 object; returning an io.WriteCloser and an error.
[`*SQLite.Create(msg ...*event.Event) error`](store/db/sqlite/sqlite.go#L39) | method | will register any number of event.Event in the SQLite database, returning an error (exposed method, but it's mostly used internally )
[`*SQLite.Write(p []byte) (n int, err error)`](store/db/sqlite/sqlite.go#L71) | method | implements the io.Writer interface, for SQLite DBs to be used with Logger, as its writer.
[`*SQLite.Close() error`](store/db/sqlite/sqlite.go#L97) | method | method added for compatibility with DBs that require it
[`WithSQLite(path string) log.LoggerConfig`](store/db/sqlite/sqlite.go#L113) | function | takes in a path to a .db file, and a table name; and returns a LoggerConfig so that this type of writer is defined in a Logger

##### MySQL


<!-- 
	Add an image of a MySQL logger config and calls (?)
-->


> Using this package will require the following environment variables to be set:

Variable | Type | Description
:--:|:--:|:--:
`MYSQL_USER` | string | username for the MySQL database connection
`MYSQL_PASSWORD` | string | password for the MySQL database connection

Symbol | Type | Description
:--:|:--:|:--:
[`New(address, database string) (sqldb io.WriteCloser, err error)`](store/db/mysql/mysql.go#L32) | function | takes in a mysql DB address and database name; and create a new instance of a MySQL object; returning an io.WriteCloser and an error.
[`*MySQL.Create(msg ...*event.Event) error`](store/db/mysql/mysql.go#L50) | method | will register any number of event.Event in the MySQL database, returning an error (exposed method, but it's mostly used internally )
[`*MySQL.Write(p []byte) (n int, err error)`](store/db/mysql/mysql.go#L82) | method | implements the io.Writer interface, for MySQL DBs to be used with Logger, as its writer.
[`*MySQL.Close() error`](store/db/mysql/mysql.go#L112) | method | method added for compatibility with DBs that require it
[`WithMySQL(addr, database string) log.LoggerConfig`](store/db/mysql/mysql.go#L166) | function | takes in an address to a MySQL server, and a database name; and returns a LoggerConfig so that this type of writer is defined in a Logger


##### PostgreSQL


<!-- 
	Add an image of a Postgres logger config and calls (?)
-->


> Using this package will require the following environment variables to be set:

Variable | Type | Description
:--:|:--:|:--:
`POSTGRES_USER` | string | username for the Postgres database connection
`POSTGRES_PASSWORD` | string | password for the Postgres database connection

Symbol | Type | Description
:--:|:--:|:--:
[`New(address, port, database string) (sqldb io.WriteCloser, err error)`](store/db/postgres/postgres.go#L33) | function | takes in a postgres DB address, port and database name; and create a new instance of a Postgres object; returning an io.WriteCloser and an error.
[`*Postgres.Create(msg ...*event.Event) error`](store/db/postgres/postgres.go#L52) | method | will register any number of event.Event in the Postgres database, returning an error (exposed method, but it's mostly used internally )
[`*Postgres.Write(p []byte) (n int, err error)`](store/db/postgres/postgres.go#L84) | method | implements the io.Writer interface, for Postgres DBs to be used with Logger, as its writer.
[`*Postgres.Close() error`](store/db/postgres/postgres.go#L118) | method | method added for compatibility with DBs that require it
[`WithPostgres(addr, port, database string) log.LoggerConfig`](store/db/postgres/postgres.go#L172) | function | takes in an address and port to a Postgres server, and a database name; and returns a LoggerConfig so that this type of writer is defined in a Logger



##### MongoDB


<!-- 
	Add an image of a Mongo logger config and calls (?)
-->


> Using this package will require the following environment variables to be set:

Variable | Type | Description
:--:|:--:|:--:
`MONGO_USER` | string | username for the Mongo database connection
`MONGO_PASSWORD` | string | password for the Mongo database connection

Symbol | Type | Description
:--:|:--:|:--:
[`New(address, database, collection string) (io.WriteCloser, error)`](store/db/mongo/mongo.go#L36) | function | takes in a mysql DB address and database name; and create a new instance of a Mongo object; returning an io.WriteCloser and an error.
[`*Mongo.Create(msg ...*event.Event) error`](store/db/mongo/mongo.go#L85) | method | will register any number of event.Event in the Mongo database, returning an error (exposed method, but it's mostly used internally )
[`*Mongo.Write(p []byte) (n int, err error)`](store/db/mongo/mongo.go#L132) | method | implements the io.Writer interface, for Mongo DBs to be used with Logger, as its writer.
[`*Mongo.Close() error`](store/db/mongo/mongo.go#L79) | method | used to terminate the live connection to the MongoDB instance
[`WithPostgres(addr, port, database string) log.LoggerConfig`](store/db/mongo/mongo.go#L163) | function | takes in the address to the mongo server, and a database and collection name; and returns a LoggerConfig so that this type of writer is defined in a Logger


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
