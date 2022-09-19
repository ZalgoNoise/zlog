# zlog
A lightweight Golang library to handle logging 

_________________________

<p align="center">
  <img src="https://github.com/zalgonoise/zlog/raw/media/img/LoggerGopher-small.png" />
</p>


## Index

1. [Overview](#overview)
1. [Installation](#installation)
1. [Usage](#usage)
	1. [Loggers usage](#loggers-usage)
		1. [Simple Logger](#simple-logger---example)
		1. [Custom Logger](#custom-logger---example)
		1. [Multilogger](#multilogger---example)
		1. [Output formats](#output-formats---example)
		1. [Modular events](#modular-events---example)
		1. [Channeled Logger](#channeled-logger---example)
		1. [Callstack in Metadata](#callstack-in-metadata---example)
	1. [Storing log events](#storing-log-events)
		1. [Logger as a Writer](#logger-as-a-writer---example)
		1. [Writing events to a file](#writing-events-to-a-file---example)
		1. [Writing events to a database](#writing-events-to-a-database)
	1. [Logging events remotely](#logging-events-remotely)
		1. [gRPC server / client](#unary-grpc-server--client---example)
	1. [Building the library](#building-the-library)
		1. [Targets](#targets)
		1. [Target types](#target-types)
1. [Features](#features)
	1. [Simple API](#simple-api)
	1. [Highly configurable](#highly-configurable)
	1. [Feature-rich events](#feature-rich-events)
		1. [Data structure](#data-structure)
		1. [Event builder](#event-builder)
		1. [Log levels](#log-levels)
		1. [Structured metadata](#structured-metadata)
		1. [Callstack in metadata](#callstack-in-metadata)
	1. [Multi-everything](#multi-everything)
	1. [Different formatters](#different-formatters)
		1. [Text](#text)
			1. [Log Timestamps](#log-timestamps)
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
	1. [gRPC](#grpc)
		1. [gRPC Log Service](#grpc-log-service)
		1. [gRPC Log Server](#grpc-log-server)
			1. [Log Server Configs](#log-server-configs)
		1. [gRPC Log Client](#grpc-log-client)
			1. [Log Client Configs](#log-client-configs)
			1. [Log Client Backoff](#log-client-backoff)
		1. [Connection Addresses](#connection-addresses)
1. [Integration](#integration)
	1. [Protobuf code generation](#protobuf-code-generation)
	1. [Adding your own configuration settings](#adding-your-own-configuration-settings)
	1. [Adding methods to a Builder pattern](#adding-methods-to-a-builder-pattern)
	1. [Adding interceptors to gRPC server / client](#adding-interceptors-to-grpc-server--client)
1. [Benchmarks](#benchmarks)
1. [Contributing](#contributing)

_________________


### Overview 


This project started (like many others) as a means for me to learn and understand how logging works (in Go and in general), among other interesting Go design patterns. 

Basing myself off of the standard library `log` package, the goal was to create a new, _minimalist_ logger while introducing great features found in open-source projects like [logrus](https://github.com/sirupsen/logrus).

Very quickly it became apparent that this project had little or no minimalism as it grew, as I intended to add new features as I learned new technologies and techniques.

That being the case, the goal morphed from simplicity to feature-rich **and** developer-friendly at the same time -- using abstractions and wrappers to allow more complex configuration or behavior if the dev wants, while (trying to) keep it idiomatic when using simple or default configurations. 

________________


### Installation

To use the library in a project you're working on, ensure that you've initialized your `go.mod` file by running:

```shell
go mod init ${package_name} # like github.com/user/repo
go mod tidy
```

After doing so, you can `go get` this library:

```shell
go get github.com/zalgonoise/zlog
```

<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/go_get_zlog.png" />
</p>


From this point onward, you can import the library in your code and use it as needed.


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/go_mod_zlog.png" />
</p>



> There are plans to add a CLI version too, to serve as a gRPC Log Server binary or a one-shot logger binary. The corresponding `go install` instructions will be added by then.

_________________

### Usage

This section covers basic usage and typical use-cases of different modules in this library. There are several [individual examples](./examples/) to provide direct context, as well as the [Features](#features) section, which goes in-depth on each module and its functionality. In here you will find reference to certain actions, a snippet from the respective example and a brief explanation of what's happening.

#### Loggers usage

##### Simple Logger - [_example_](./examples/logger/simple_logger/simple_logger.go)

<details>

_Snippet_

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
)

func main() {
	log.Print("this is the simplest approach to entering a log message")
	log.Tracef("and can include formatting: %v %v %s", 3.5, true, "string")
	log.Errorln("which is similar to fmt.Print() method calls")

	log.Panicf("example of a logger panic event: %v", true)
}
```

_Output_

```
[info]  [2022-07-26T17:05:46.208657519Z]        [log]   this is the simplest approach to entering a log message
[trace] [2022-07-26T17:05:46.208750114Z]        [log]   and can include formatting: 3.5 true string
[error] [2022-07-26T17:05:46.208759031Z]        [log]   which is similar to fmt.Print() method calls

[panic] [2022-07-26T17:05:46.208766425Z]        [log]   example of a logger panic event: true
panic: example of a logger panic event: true

goroutine 1 [running]:
github.com/zalgonoise/zlog/log.(*logger).Panicf(0xc0001dafc0, {0x952348, 0x23}, {0xc0001af410, 0x1, 0x1})
        /go/src/github.com/zalgonoise/zlog/log/print.go:226 +0x357
github.com/zalgonoise/zlog/log.Panicf(...)
        /go/src/github.com/zalgonoise/zlog/log/print.go:784
main.main()
        /go/src/github.com/zalgonoise/zlog/examples/logger/simple_logger/simple_logger.go:19 +0x183
exit status 2

```

</details>

The simplest approach to using the logger library is to call its built-in methods, as if they were `fmt.Print()`-like calls. The logger exposes methods for registering messages in [different log levels](#log-levels), defined in its [`Printer` interface](./log/print.go#L18).

Note that there are calls which are configured to halt the application's runtime, like `Fatal()` and `Panic()`. These exit calls can be skipped in the [logger's configuration](#highly-configurable).

More information on the [`Printer` interface](./log/print.go#L18) and the Logger's methods in the [_Simple API_ section](#simple-api).


##### Custom Logger - [_example_](./examples/logger/custom_logger/custom_logger.go)


<details>

_Snippet_

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/zalgonoise/zlog/log"
)

func main() {
	logger := log.New(
		log.WithPrefix("svc"),
		log.WithSub("mod"),
	)

	buf := new(bytes.Buffer)

	logger.SetOuts(buf)
	logger.Prefix("service")
	logger.Sub("module")

	logger.Info("message written to a new buffer")

	fmt.Println(buf.String())
}
```

_Output_

```
[info]  [2022-07-26T17:04:25.371617213Z]        [service]       [module]        message written to a new buffer
```

</details>


The logger is customized on creation, and any number of configuration can be passed to it. This makes it flexible for simple configurations (where only defaults are applied), and makes it granular enough for the complex ones.

Furthermore, it will also expose [certain methods](#simple-api) to allow changes to the logger's configuration during runtime (with `Prefix()`, `Sub()` and `Metadata()`, as well as `AddOuts()` and `SetOuts()` methods). Besides these and for more information on the available configuration functions for loggers, check out the [_Highly configurable_ section](#highly-configurable).




_________________



##### MultiLogger - [_example_](./examples/logger/multilogger/multilogger.go)


<details>

_Snippet_

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {
	buf := new(bytes.Buffer)

	stdLogger := log.New() // default logger printing to stdErr
	jsonLogger := log.New( // custom JSON logger, writing to buffer
		log.WithOut(buf),
		log.CfgFormatJSONIndent,
	)

	// join both loggers
	logger := log.MultiLogger(
		stdLogger,
		jsonLogger,
	)

	// print messages to stderr
	logger.Info("some event occurring")
	logger.Warn("a warning pops-up")
	logger.Log(
		event.New().Level(event.Level_error).
			Message("and finally an error").
			Metadata(event.Field{
				"code":      5,
				"some-data": true,
			}).
			Build())

	// print buffer content
	fmt.Print("\n---\n- JSON data:\n---\n", buf.String())
}
```

_Output_

```
[info]  [2022-07-28T17:12:44.966084127Z]        [log]   some event occurring
[warn]  [2022-07-28T17:12:44.966220938Z]        [log]   a warning pops-up
[error] [2022-07-28T17:12:44.966246265Z]        [log]   and finally an error    [ code = 5 ; some-data = true ] 

---
- JSON data:
---
{
  "timestamp": "2022-07-28T17:12:44.966187177Z",
  "service": "log",
  "level": "info",
  "message": "some event occurring"
}
{
  "timestamp": "2022-07-28T17:12:44.966234771Z",
  "service": "log",
  "level": "warn",
  "message": "a warning pops-up"
}
{
  "timestamp": "2022-07-28T17:12:44.966246265Z",
  "service": "log",
  "level": "error",
  "message": "and finally an error",
  "metadata": {
    "code": 5,
    "some-data": true
  }
}
```

</details>

The logger on [line 21](./examples/logger/multilogger/multilogger.go#L21) is merging any loggers provided as input. In this example, the caller can leverage this functionality to write the same events to different outputs (with different formats), or with certain log level filters (`writerA` will register all events, while `writerB` will register events that are `error` and above).

This approach can be taken with all kinds of loggers, provided that they they implement the same methods as [`Logger` interface](./log/logger.go#L95). By all kinds of loggers, I mean those within this library, such as having a standard-error logger, as well as a gRPC Log Client configured as one, with [`MultiLogger()`](./log/multilog.go#L11). More information on [_Multi-everything_, in its own section](#multi-everything)



_________________



##### Output formats - [_example_](./examples/logger/formatted_logger/formatted_logger.go)


<details>

_Snippet_

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
)

func main() {

	// setup a simple text logger, with custom formatting
	custTextLogger := log.New(
		log.WithFormat(
			text.New().
				Color().
				DoubleSpace().
				LevelFirst().
				Upper().
				Time(text.LTRubyDate).
				Build(),
		),
	)

	// setup a simple JSON logger
	jsonLogger := log.New(log.CfgFormatJSON)

	// setup a simple XML logger
	xmlLogger := log.New(log.CfgFormatXML)

	// (...)

	// join all loggers
	multiLogger := log.MultiLogger(
		custTextLogger,
		jsonLogger,
		xmlLogger,
		// (...)
	)

	// example message to print
	var msg = event.New().Message("message from a formatted logger").Build()

	// print the message to standard out, with different formats
	multiLogger.Log(msg)
}

```

_Output_

```
[INFO]          [Sat Jul 30 13:17:31 +0000 2022]                [LOG]           message from a formatted logger
{"timestamp":"2022-07-30T13:17:31.744955941Z","service":"log","level":"info","message":"message from a formatted logger"}
<entry><timestamp>2022-07-30T13:17:31.744955941Z</timestamp><service>log</service><level>info</level><message>message from a formatted logger</message></entry>
```

</details>

When setting up the [`Logger` interface](./log/logger.go#L95), different formatters can be passed as well. There are common formats already implemented (like JSON, XML, CSV, BSON), as well as a modular text formatter. 

New formatters can also be added seamlessly by complying with their corresponding interfaces. More information on all formatters in the [_Different Formatters_ section](#different-formatters).



_________________



##### Modular events - [_example_](./examples/logger/modular_events/modular_events.go)


<details>

_Snippet_

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {

	// Events can be created and customized with a builder pattern,
	// where each element is defined with a chained method until the
	// Build() method is called.
	//
	// This last method will apply the timestamp to the event and any
	// defaults for missing (required) fields.
	log.Log(
		event.New().
			Prefix("module").
			Sub("service").
			Level(event.Level_warn).
			Metadata(event.Field{
				"data": true,
			}).
			Build(),
		event.New().
			Prefix("mod").
			Sub("svc").
			Level(event.Level_debug).
			Metadata(event.Field{
				"debug": "something something",
			}).
			Build(),
	)
}
```

_Output_

```
[warn]  [2022-07-30T13:21:14.023201168Z]        [module]        [service]               [ data = true ] 
[debug] [2022-07-30T13:21:14.023467597Z]        [mod]   [svc]           [ debug = "something something" ]
```

</details>

The events are created under-the-hood when using methods from the [Printer interface](./log/print.go#L18), but they can also be created using the exposed events builder. This allows using a clean approach when using the logger (using its [`Log()` method](#simple-api)), while keeping the events as detailed as you need. More information on events in the [_Feature-rich Events_ section](#feature-rich-events).


_________________



##### Channeled Logger - [_example_](./examples/logger/channeled_logger/channeled_logger.go)


<details>

_Snippet_

```go
package main

import (
	"time"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/logch"
)

func main() {

	// create a new, basic logger directly as a channeled logger
	chLogger := logch.New(log.New())

	// send messages using its Log() method directly; like the simple one:
	chLogger.Log(
		event.New().Message("one").Build(),
		event.New().Message("two").Build(),
		event.New().Message("three").Build(),
	)

	// or, call its Channels() method to work with the channels directly:
	msgCh, done := chLogger.Channels()

	// send the messages in a separate goroutine, then close the logger
	go func() {
		msgCh <- event.New().Message("four").Build()
		msgCh <- event.New().Message("five").Build()
		msgCh <- event.New().Message("six").Build()

		// give it a millisecond to allow the last message to be printed
		time.Sleep(time.Millisecond)

		// send done signal to stop the process
		done <- struct{}{}
	}()

	// keep-alive until the done signal is received
	for {
		select {
		case <-done:
			return
		}
	}

}
```

_Output_

```
[info]  [2022-07-31T12:15:40.702944256Z]        [log]   one
[info]  [2022-07-31T12:15:40.703050024Z]        [log]   two
[info]  [2022-07-31T12:15:40.703054422Z]        [log]   three
[info]  [2022-07-31T12:15:40.703102802Z]        [log]   four
[info]  [2022-07-31T12:15:40.703156352Z]        [log]   five
[info]  [2022-07-31T12:15:40.703169196Z]        [log]   six
```

</details>

Since loggers are usually kept running in the background (as your app handles events and writes to its logger), you are perfectly able to launch this logger as a goroutine. To simplify the process, an interface is added: ([`ChanneledLogger`](./log/logch/logch.go#L12)). The gist of it is being able to directly launch a logger in a goroutine, with useful methods to interact with it (`Log()`, `Close()` and `Channels()`). More information on this logic in the [_Highly Configurable_ section](#highly-configurable).


_________________



##### Callstack in Metadata - [_example_](./examples/logger/callstack_in_metadata/callstack_md.go)


<details>

_Snippet_

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

// set up an indented JSON logger; package-level as an example
//
// you'd set this up within your logic however it's convenient
var logger = log.New(log.CfgFormatJSONIndent)

// placeholder operation for visibility in the callstack
func operation(value int) bool {
	return subOperation(value)
}

// placeholder sub-operation for visibility in the callstack
//
// error is printed whenever input is zero
func subOperation(value int) bool {
	if value == 0 {
		logger.Log(
			event.New().
				Level(event.Level_error).
				Message("operation failed").
				Metadata(event.Field{
					"error": "input cannot be zero", // custom metadata
					"input": value,                  // custom metadata
				}).
				CallStack(true). // add (complete) callstack to metadata
				Build(),
		)
		return false
	}
	return true
}

func main() {
	// all goes well until something happens within your application
	for a := 5; a >= 0; a-- {
		if operation(a) {
			continue
		}
		break
	}
}
```

_Output_

```json
{
  "timestamp": "2022-08-01T16:06:46.162355505Z",
  "service": "log",
  "level": "error",
  "message": "operation failed",
  "metadata": {
    "callstack": {
      "goroutine-1": {
        "id": "1",
        "stack": [
          {
            "method": "github.com/zalgonoise/zlog/log/trace.(*stacktrace).getCallStack(...)",
            "reference": "/go/src/github.com/zalgonoise/zlog/log/trace/trace.go:54"
          },
          {
            "method": "github.com/zalgonoise/zlog/log/trace.New(0x30?)",
            "reference": "/go/src/github.com/zalgonoise/zlog/log/trace/trace.go:41 +0x7f"
          },
          {
            "method": "github.com/zalgonoise/zlog/log/event.(*EventBuilder).CallStack(0xc000077fc0, 0x30?)",
            "reference": "/go/src/github.com/zalgonoise/zlog/log/event/builder.go:99 +0x6b"
          },
          {
            "method": "main.subOperation(0x0)",
            "reference": "/go/src/github.com/zalgonoise/zlog/examples/logger/callstack_in_metadata/callstack_md.go:31 +0x34b"
          },
          {
            "method": "main.operation(...)",
            "reference": "/go/src/github.com/zalgonoise/zlog/examples/logger/callstack_in_metadata/callstack_md.go:15"
          },
          {
            "method": "main.main()",
            "reference": "/go/src/github.com/zalgonoise/zlog/examples/logger/callstack_in_metadata/callstack_md.go:42 +0x33"
          }
        ],
        "status": "running"
      }
    },
    "error": "input cannot be zero",
    "input": 0
  }
}
```

</details>

When creating an event, you're able to chain the `Callstack(all bool)` method to it, before building the event. The `bool` value it takes represents whether you want a full or trimmed callstack. 

More information on event building in [the _Feature-rich Events_ section](#feature-rich-events), and the [_Callstack in metadata_ section in particular](#callstack-in-metadata).


_________________


#### Storing log events

##### Logger as a Writer - [_example_](./examples/logger/log_as_writer/log_as_writer.go)


<details>

_Snippet_

```go
package main

import (
	"fmt"
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {

	logger := log.New()

	n, err := logger.Write([]byte("Hello, world!"))
	if err != nil {
		fmt.Println("errored: ", err)
		os.Exit(1)
	}

	fmt.Printf("\n---\nn: %v, err: %v\n---\n", n, err)

	n, err = logger.Write(event.New().Message("Hi, world!").Build().Encode())
	if err != nil {
		fmt.Println("errored: ", err)
		os.Exit(1)
	}

	fmt.Printf("\n---\nn: %v, err: %v\n---\n", n, err)
}
```

_Output_

```
[info]  [2022-07-30T12:07:44.451547181Z]        [log]   Hello, world!

---
n: 69, err: <nil>
---
[info]  [2022-07-30T12:07:44.45166375Z] [log]   Hi, world!

---
n: 65, err: <nil>
---
```

</details>

Since the [`Logger` interface](./log/logger.go#L95) also implements the [`io.Writer` interface](https://pkg.go.dev/io#Writer), it can be used in a broader form. The example above shows how simply passing a (string) message as a slice of bytes replicates a `log.Info()` call, and passing an encoded event will actually read its parameters (level, prefix, etc) and register an event accordingly. More information in the [_Writer Interface_ section](#writer-interface).


_________________



##### Writing events to a file - [_example_](./examples/datastore/file/logfile.go)

<details>

_Snippet_

```go
package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/fs"
)

func main() {
	// create a new temporary file in /tmp
	tempF, err := ioutil.TempFile("/tmp", "zlog_test_fs-")

	if err != nil {
		log.Fatalf("unexpected error creating temp file: %v", err)
	}
	defer os.Remove(tempF.Name())

	// use the temp file as a LogFile
	logF, err := fs.New(
		tempF.Name(),
	)
	if err != nil {
		log.Fatalf("unexpected error creating logfile: %v", err)
	}

	// set up a max size in MB, to auto-rotate
	logF.MaxSize(50)

	// create a simple logger using the logfile as output, log messages to it
	var logger = log.New(
		log.WithOut(logF),
	)
	logger.Log(
		event.New().Message("log entry written to file").Build(),
	)

	// print out file's's content
	b, err := os.ReadFile(tempF.Name())
	if err != nil {
		log.Fatalf("unexpected error reading logfile's data: %v", err)
	}

	fmt.Println(string(b))
}
```

_Output_

```
[info]  [2022-08-02T16:28:07.523156566Z]        [log]   log entry written to file
```

</details>

This logger also exposes a package to simplify writing log events to a file. [This package](./store/fs/logfile.go) exposes methods that may be helpful, like max-size limits and auto-rotate features. 

More information on this topic in [the _Logfile_ section](#logfile).


_________________



##### Writing events to a database

<details>


_SQLite Snippet_ - [_example_](./examples/datastore/db/sqlite/sqlite.go)


<details>

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/sqlite"

	"os"
)

const (
	dbPathEnv string = "SQLITE_PATH"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func setupMethodOne(dbPath string) log.Logger {
	// create a new DB writer
	db, err := sqlite.New(dbPath)

	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	// create a new logger with the general DB config
	logger := log.New(
		log.WithDatabase(db),
	)

	return logger
}

func setupMethodTwo(dbPath string) log.Logger {
	// create one with the package function
	return log.New(
		sqlite.WithSQLite(dbPath),
	)
}

func main() {
	// load sqlite db path from environment variable
	sqlitePath, ok := getEnv(dbPathEnv)
	if !ok {
		log.Fatalf("SQLite database path not provided, from env variable %s", dbPathEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne := setupMethodOne(sqlitePath)

	// write a message to the DB
	loggerOne.Log(
		event.New().Message("log entry #1 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #1").Build(),
	)

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(sqlitePath)

	// write a message to the DB
	loggerTwo.Log(
		event.New().Message("log entry #1 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #2").Build(),
	)
}

```

</details>



_MySQL Snippet_ - [_example_](./examples/datastore/db/mysql/mysql.go)



<details>

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/mysql"

	"os"
)

const (
	_         string = "MYSQL_USER"     // account for these env variables
	_         string = "MYSQL_PASSWORD" // account for these env variables
	dbAddrEnv string = "MYSQL_HOST"
	dbPortEnv string = "MYSQL_PORT"
	dbNameEnv string = "MYSQL_DATABASE"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func setupMethodOne(dbAddr, dbPort, dbName string) log.Logger {
	// create a new DB writer
	db, err := mysql.New(dbAddr+":"+dbPort, dbName)

	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	// create a new logger with the general DB config
	logger := log.New(
		log.WithDatabase(db),
	)

	return logger
}

func setupMethodTwo(dbAddr, dbPort, dbName string) log.Logger {
	// create one with the package function
	return log.New(
		mysql.WithMySQL(dbAddr+":"+dbPort, dbName),
	)
}

func main() {
	// load mysql db details from environment variables
	mysqlAddr, ok := getEnv(dbAddrEnv)
	if !ok {
		log.Fatalf("MySQL database address not provided, from env variable %s", dbAddrEnv)
	}

	mysqlPort, ok := getEnv(dbPortEnv)
	if !ok {
		log.Fatalf("MySQL database port not provided, from env variable %s", dbPortEnv)
	}

	mysqlDB, ok := getEnv(dbNameEnv)
	if !ok {
		log.Fatalf("MySQL database name not provided, from env variable %s", dbNameEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne := setupMethodOne(mysqlAddr, mysqlPort, mysqlDB)

	// write a message to the DB
	loggerOne.Log(
		event.New().Message("log entry #1 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #1").Build(),
	)

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(mysqlAddr, mysqlPort, mysqlDB)

	// write a message to the DB
	loggerTwo.Log(
		event.New().Message("log entry #1 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #2").Build(),
	)
}
```

</details>



_PostgreSQL Snippet_ - [_example_](./examples/datastore/db/postgres/postgres.go)


<details>

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/postgres"

	"os"
)

const (
	_         string = "POSTGRES_USER"     // account for these env variables
	_         string = "POSTGRES_PASSWORD" // account for these env variables
	dbAddrEnv string = "POSTGRES_HOST"
	dbPortEnv string = "POSTGRES_PORT"
	dbNameEnv string = "POSTGRES_DATABASE"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func setupMethodOne(dbAddr, dbPort, dbName string) log.Logger {
	// create a new DB writer
	db, err := postgres.New(dbAddr, dbPort, dbName)

	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	// create a new logger with the general DB config
	logger := log.New(
		log.WithDatabase(db),
	)

	return logger
}

func setupMethodTwo(dbAddr, dbPort, dbName string) log.Logger {
	// create one with the package function
	return log.New(
		postgres.WithPostgres(dbAddr, dbPort, dbName),
	)
}

func main() {
	// load postgres db details from environment variables
	postgresAddr, ok := getEnv(dbAddrEnv)
	if !ok {
		log.Fatalf("Postgres database address not provided, from env variable %s", dbAddrEnv)
	}

	postgresPort, ok := getEnv(dbPortEnv)
	if !ok {
		log.Fatalf("Postgres database port not provided, from env variable %s", dbPortEnv)
	}

	postgresDB, ok := getEnv(dbNameEnv)
	if !ok {
		log.Fatalf("Postgres database name not provided, from env variable %s", dbNameEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne := setupMethodOne(postgresAddr, postgresPort, postgresDB)

	// write a message to the DB
	loggerOne.Log(
		event.New().Message("log entry #1 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #1").Build(),
	)

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(postgresAddr, postgresPort, postgresDB)

	// write a message to the DB
	loggerTwo.Log(
		event.New().Message("log entry #1 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #2").Build(),
	)
}
```

</details>




_MongoDB Snippet_  - [_example_](./examples/datastore/db/mongo/mongo.go)


<details>

```go
package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/mongo"

	"os"
)

const (
	_         string = "MONGO_USER"     // account for these env variables
	_         string = "MONGO_PASSWORD" // account for these env variables
	dbAddrEnv string = "MONGO_HOST"
	dbPortEnv string = "MONGO_PORT"
	dbNameEnv string = "MONGO_DATABASE"
	dbCollEnv string = "MONGO_COLLECTION"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func setupMethodOne(dbAddr, dbPort, dbName, dbColl string) (log.Logger, func()) {
	// create a new DB writer
	db, err := mongo.New(dbAddr+":"+dbPort, dbName, dbColl)

	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	// create a new logger with the general DB config
	logger := log.New(
		log.WithDatabase(db),
	)

	return logger, func() {
		db.Close()
	}
}

func setupMethodTwo(dbAddr, dbPort, dbName, dbColl string) log.Logger {
	// create one with the package function
	return log.New(
		mongo.WithMongo(dbAddr+":"+dbPort, dbName, dbColl),
	)
}

func main() {
	// load mongo db details from environment variables
	mongoAddr, ok := getEnv(dbAddrEnv)
	if !ok {
		log.Fatalf("Mongo database address not provided, from env variable %s", dbAddrEnv)
	}

	mongoPort, ok := getEnv(dbPortEnv)
	if !ok {
		log.Fatalf("Mongo database port not provided, from env variable %s", dbPortEnv)
	}

	mongoDB, ok := getEnv(dbNameEnv)
	if !ok {
		log.Fatalf("Mongo database name not provided, from env variable %s", dbNameEnv)
	}

	mongoColl, ok := getEnv(dbCollEnv)
	if !ok {
		log.Fatalf("Mongo collection name not provided, from env variable %s", dbCollEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne, close := setupMethodOne(mongoAddr, mongoPort, mongoDB, mongoColl)
	defer close()

	// write a message to the DB
	loggerOne.Log(
		event.New().Message("log entry written to sqlite db writer #1").Build(),
	)

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(mongoAddr, mongoPort, mongoDB, mongoColl)

	// write a message to the DB
	n, err := loggerTwo.Output(
		event.New().Message("log entry written to sqlite db writer #2").Build(),
	)
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	if n == 0 {
		log.Fatalf("zero bytes written")
	}

	log.Info("event written to DB writer #2 successfully")
}
```

</details>


</details>

There are packages available in [`store/db`](./store/db) which will allow for a simple setup when preferring to write log events to a database (instead of a file or a buffer).

For this, most DBs leverage GORM to make it seamless to use different databases. Here is the list of supported database types:
- [SQLite](#sqlite)
- [MySQL](#mysql)
- [PostgreSQL](#postgresql)
- [MongoDB](#mongodb)

More information on this topic in [the _Databases_ section](#databases).


_________________


#### Logging events remotely

##### gRPC server / client - [_example_](./examples/grpc/simple_unary_client_server/)


<details>

_Server Snippet_

```go
package main

import (
	"os"

	"github.com/zalgonoise/zlog/log"

	"github.com/zalgonoise/zlog/grpc/server"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func getTLSConf() server.LogServerConfig {
	var tlsConf server.LogServerConfig

	var withCert bool
	var withKey bool
	var withCA bool

	certPath, ok := getEnv("TLS_SERVER_CERT")
	if ok {
		withCert = true
	}

	keyPath, ok := getEnv("TLS_SERVER_KEY")
	if ok {
		withKey = true
	}

	caPath, ok := getEnv("TLS_CA_CERT")
	if ok {
		withCA = true
	}

	if withCert && withKey {
		if withCA {
			tlsConf = server.WithTLS(certPath, keyPath, caPath)
		} else {
			tlsConf = server.WithTLS(certPath, keyPath)
		}
	}

	return tlsConf
}

func main() {

	grpcLogger := server.New(
		server.WithLogger(
			log.New(
				log.WithFormat(log.TextColorLevelFirstSpaced),
			),
		),
		server.WithServiceLogger(
			log.New(
				log.WithFormat(log.TextColorLevelFirstSpaced),
			),
		),
		server.WithAddr("127.0.0.1:9099"),
		server.WithGRPCOpts(),
		getTLSConf(),
	)
	grpcLogger.Serve()
}
```

_Output_

```
[info]          [2022-08-04T11:16:29.624287477Z]                [gRPC]          [listen]                gRPC server is listening to connections         [ addr = "127.0.0.1:9099" ] 
[debug]         [2022-08-04T11:16:29.624489336Z]                [gRPC]          [serve]         gRPC server is running          [ addr = "127.0.0.1:9099" ] 
[debug]         [2022-08-04T11:16:29.624530116Z]                [gRPC]          [handler]               message handler is running
[trace]         [2022-08-04T11:16:29.624588293Z]                [gRPC]          [Done]          listening to done signal
```

_Client snippet_

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func getTLSConf() client.LogClientConfig {
	var tlsConf client.LogClientConfig

	var withCert bool
	var withKey bool
	var withCA bool

	certPath, ok := getEnv("TLS_CLIENT_CERT")
	if ok {
		withCert = true
	}

	keyPath, ok := getEnv("TLS_CLIENT_KEY")
	if ok {
		withKey = true
	}

	caPath, ok := getEnv("TLS_CA_CERT")
	if ok {
		withCA = true
	}

	if withCA {
		if withCert && withKey {
			tlsConf = client.WithTLS(caPath, certPath, keyPath)
		} else {
			tlsConf = client.WithTLS(caPath)
		}
	}

	return tlsConf
}

func main() {
	logger := log.New(
		log.WithFormat(log.TextColorLevelFirst),
	)

	grpcLogger, errCh := client.New(
		client.WithAddr("127.0.0.1:9099"),
		client.UnaryRPC(),
		client.WithLogger(
			logger,
		),
		client.WithGRPCOpts(),
		getTLSConf(),
	)
	_, done := grpcLogger.Channels()

	grpcLogger.Log(event.New().Message("hello from client").Build())

	for i := 0; i < 3; i++ {
		grpcLogger.Log(event.New().Level(event.Level_warn).Message(fmt.Sprintf("warning #%v", i)).Build())
		time.Sleep(time.Millisecond * 50)
	}

	for {
		select {
		case err := <-errCh:
			panic(err)
		case <-done:
			return
		}
	}
}
```

_Client output_

```
[debug] [2022-08-04T11:21:11.36101206Z] [gRPC]  [init]  setting up Unary gRPC client
[debug] [2022-08-04T11:21:11.361184863Z]        [gRPC]  [conn]  connecting to remote    [ addr = "127.0.0.1:9099" ; index = 0 ] 
[debug] [2022-08-04T11:21:11.361299493Z]        [gRPC]  [conn]  connecting to remote    [ index = 0 ; addr = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.361876301Z]        [gRPC]  [conn]  dialed the address successfully [ addr = "127.0.0.1:9099" ; index = 0 ] 
[debug] [2022-08-04T11:21:11.361995075Z]        [gRPC]  [log]   setting up log service with connection  [ remote = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.362029197Z]        [gRPC]  [log]   received a new log message to register  [ timeout = 30 ; remote = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.362151576Z]        [gRPC]  [conn]  dialed the address successfully [ addr = "127.0.0.1:9099" ; index = 0 ] 
[debug] [2022-08-04T11:21:11.36221311Z] [gRPC]  [log]   setting up log service with connection  [ remote = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.362242582Z]        [gRPC]  [log]   received a new log message to register  [ remote = "127.0.0.1:9099" ; timeout = 30 ] 
[debug] [2022-08-04T11:21:11.414202683Z]        [gRPC]  [conn]  connecting to remote    [ addr = "127.0.0.1:9099" ; index = 0 ] 
[debug] [2022-08-04T11:21:11.415095627Z]        [gRPC]  [conn]  dialed the address successfully [ index = 0 ; addr = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.415162552Z]        [gRPC]  [log]   setting up log service with connection  [ remote = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.415192529Z]        [gRPC]  [log]   received a new log message to register  [ timeout = 30 ; remote = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.464857675Z]        [gRPC]  [conn]  connecting to remote    [ index = 0 ; addr = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.465455812Z]        [gRPC]  [conn]  dialed the address successfully [ addr = "127.0.0.1:9099" ; index = 0 ] 
[debug] [2022-08-04T11:21:11.465509873Z]        [gRPC]  [log]   setting up log service with connection  [ remote = "127.0.0.1:9099" ] 
[debug] [2022-08-04T11:21:11.465537439Z]        [gRPC]  [log]   received a new log message to register  [ remote = "127.0.0.1:9099" ; timeout = 30 ]
```

_Server output_

```
(...)
[trace]         [2022-08-04T11:16:29.624588293Z]                [gRPC]          [Done]          listening to done signal
[warn]          [2022-08-04T11:21:11.361248552Z]                [log]           warning #0
[info]          [2022-08-04T11:21:11.361160675Z]                [log]           hello from client
[debug]         [2022-08-04T11:21:11.362648438Z]                [gRPC]          [handler]               input log message parsed and registered
[debug]         [2022-08-04T11:21:11.362701641Z]                [gRPC]          [handler]               input log message parsed and registered
[warn]          [2022-08-04T11:21:11.414114247Z]                [log]           warning #1
[debug]         [2022-08-04T11:21:11.415435299Z]                [gRPC]          [handler]               input log message parsed and registered
[warn]          [2022-08-04T11:21:11.464787592Z]                [log]           warning #2
[debug]         [2022-08-04T11:21:11.465792823Z]                [gRPC]          [handler]               input log message parsed and registered
```


</details>

Setting up a Log Server to register events over a network is made possible with the [gRPC](https://grpc.io/) implementation of both client and server logic. This is to make it both fast and simple to configure (remote) logging when it should not be a center-piece of your application; merely a debugging / development feature.

The [gRPC](https://grpc.io/) framework allows a quick and secure means of exchanging messages as remote procedure calls, leveraging protocol buffers to generate source code on-the-fly. As such, protocol buffers have a huge weight on this library, having the event data structure defined as one.

Not only is it simple to setup and kick off a gRPC Log server, with either unary and stream RPCs and many other neat features -- it's also extensible by using the same service and message `.proto` files and integrate the Logger in your own application, by generating the Log Client code for your environment.

More information on gRPC service / server / client implementations, in [the _gRPC_ section](#grpc).


_________________


#### Building the library

This library uses [Bazel](https://bazel.build/) as a build system. Bazel is an amazing tool that promises a hermetic build -- one that will *always* output the same results. While this is the greatest feature of Bazel (which may not be achievable in other build tools that are dependent on your local environment), it is also very extensible in terms of integration, support for multiple languages, and clear _recipes_ of the dependencies of a particular library or binary.

Bazel also has a dozen features like remote execution, [like with BuildBuddy](https://www.buildbuddy.io/) and [remote caching](https://bazel.build/docs/remote-caching) (with a Docker container in another machine, for instance).

To install [Bazel](https://bazel.build/) on a machine, I prefer to use the [`bazelisk` tool](https://github.com/bazelbuild/bazelisk) which ensures that Bazel is both up-to-date and will not raise any sort of version issue later on. To install it, you can use your system's package manager, or the following Go command:

```shell
go install github.com/bazelbuild/bazelisk@latest
```

Bazel is the build tool of choice for this library as (for me, personally) it becomes very easy to configure a new build target when required; you simply need to setup the `WORKSPACE` file from presets provided by the tools used ([Go rules](https://github.com/bazelbuild/rules_go/releases), [`gazelle`](https://github.com/bazelbuild/bazel-gazelle/releases) and [`buildifier`](https://github.com/bazelbuild/buildtools/releases)); by adding the corresponding `WORKSPACE` code to your repo's file. 

Once the `WORKSPACE` file is configured, you're able to use `gazelle` to auto-generate the build files for your Go code. This is done by running the `//:gazelle` targets listed in the section below:

##### Targets

Command | Description
:--:|:--:
`bazel run //:gazelle --` | Generate the `BUILD.bazel` files for this and all subdirectories. It must be called at the root of the repository, where the `WORKSPACE` file is located
`bazel run //:gazelle -- update-repos -from_file=go.mod` | Update the `WORKSPACE` file with the new dependencies, when they are added to Go code. It must be called at the root of the repository, where the `WORKSPACE` file is located
`bazel build //log/...` | Builds the `log` package and all subpackages
`bazel test //...` | Tests the entire library
`bazel test //log/event:event_test` | Tests the `event` package, in the form of `//path/to:target`
`bazel run //:zlog` | Executes the `:zlog` (`/main.go`) target
`bazel run //examples/logger/simple_logger:simple_logger` | Executes the `simple_logger.go` example, in the form of `//path/to:target`
`bazel run //:buildifier-check` | Executes the `buildifier` target in lint-checking mode for build files
`bazel run //:buildifier-fix` | Executes the `buildifier` target and corrects any build file lint issues

##### Target types

__Go Binary__

If you have a binary (`main.go` file), it will be listed as a binary target in a `BUILD.bazel` file, such as:

```bzl
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_binary(
    name = "zlog",
    embed = [":zlog_lib"],
    visibility = ["//visibility:public"],
)
```

These targets can be called with Bazel with a command such as:

```shell
bazel run //:zlog --
```

...Or if they are in some deeper directory, call the command from the repo's root:

```shell
bazel run //examples/logger/simple_logger:simple_logger
```

__Go Library__

Libraries are non-executable packages of (Go) code which will provide functionality to a binary, as referenced in its imports:

```bzl
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "simple_logger_lib",
    srcs = ["simple_logger.go"],
    importpath = "github.com/zalgonoise/zlog/examples/logger/simple_logger",
    visibility = ["//visibility:private"],
    deps = ["//log"],
)

go_binary(
    name = "simple_logger",
    embed = [":simple_logger_lib"], # uses the lib above
    visibility = ["//visibility:public"],
)
```

__Gazelle__

Used to generate Go build files for Bazel, it's declared in the `WORSKAPCE` file and referenced in the main `BUILD.bazel` (at the root of the repository):

```bzl
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/zalgonoise/zlog
gazelle(name = "gazelle")
```

Usage of `gazelle` just above this section, with both commands used to update Go build files and their dependencies for Bazel.

__Buildifier__

Used to verify (and fix) build files that you may have created or modified. Similar to `go fmt` but for Bazel. it's declared in the `WORSKAPCE` file and referenced in the main `BUILD.bazel` (at the root of the repository), with two targets available:

```bzl
load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

buildifier(
    name = "buildifier-check",
    lint_mode = "warn",
    mode = "check",
    multi_diff = True,
)

buildifier(
    name = "buildifier-fix",
    lint_mode = "fix",
    mode = "fix",
    multi_diff = True,
)
```


_________________



### Features

This library provides a feature-rich structured logger, ready to write to many types of outputs (standard out / error, to buffers, to files and databases) and over-the-wire (via gRPC).

#### Simple API

<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/simple_logger_cli.png" />
</p>


> See the [_Simple Logger_ example](#simple-logger---example)


The [`Logger` interface](./log/logger.go#L95) in this library provides a set complete set of idiomatic methods which allow to either control the logger:


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

...or to use its [Printer interface](./log/print.go#L18) and print messages in the `fmt.Print()` / `fmt.Println()` / `fmt.Printf()` way:


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

The logger configuration methods (listed below) can be used during runtime, to adapt the logger to the application's needs. However special care needs to be taken when calling the log writer-altering methods (`AddOuts()` and `SetOuts()`) considering that they can raise undersirable effects. It's recommended for these methods to be called when the logger is initialized in your app, if called at all.


Method | Description
:--:|:--:
[`SetOuts(...io.Writer) Logger`](./log/logger.go#L180) | sets (replaces) the defined [`io.Writer`](https://pkg.go.dev/io#Writer) in the Logger with the input list of [`io.Writer`](https://pkg.go.dev/io#Writer).
[`AddOuts(...io.Writer) Logger`](./log/logger.go#L207) | adds (appends) the defined [`io.Writer`](https://pkg.go.dev/io#Writer) in the Logger with the input list of [`io.Writer`](https://pkg.go.dev/io#Writer).
[`Prefix(string) Logger`](./log/logger.go#L237) | sets a logger-scoped (as opposed to message-scoped) prefix string to the logger
[`Sub(string) Logger`](./log/logger.go#L259) | sets a logger-scoped (as opposed to message-scoped) sub-prefix string to the logger
[`Fields(map[string]interface{}) Logger`](./log/logger.go#L275) | sets logger-scoped (as opposed to message-scoped) metadata fields to the logger
[`IsSkipExit() bool`](./log/logger.go#L294) | returns a boolean on whether this logger is set to skip os.Exit(1) or panic() calls.

> Note: `SetOuts()` and `AddOuts()` methods will apply the [multi-writer pattern](#multi-everything) to the input list of [`io.Writer`](https://pkg.go.dev/io#Writer). The writers are merged as one.

> Note: Logger-scoped parameters (prefix, sub-prefix and metadata) allow calling either the [`Printer` interface](./log/print.go#18) methods (including event-based methods like `Log()` and `Output()`) without having to define these values. This can be especially useful when registering multiple log events in a certain module of your code -- _however_, the drawback is that these values are persisted in the logger, so you may need to unset them (calling `nil` or empty values on them).

> Note: `IsSkipExit()` is a useful method, used for example to determine wether a [MultiLogger](#multi-everything) should should be presented as a skip-exit-calls logger or not -- if _at least one_ configured logger in a multilogger is __not__ skipping exit calls, its output would be `false`.

#### Highly configurable 


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/logger_new.png" />
</p>


> See the [_Custom Logger_ example](#custom-logger---example)


Creating a new logger with, for example, [`log.New()`](./log/logger.go#L134) takes any number of configurations (including none, for the default configuration). This allows added modularity to the way your logger should behave.

Method / Variable | Description
:--:|:--:
[`NilLogger()`](./log/conf.go#L154) | create a __nil-logger__ (that doesn't write anything to anywhere)
[`WithPrefix(string)`](./log/conf.go#L165) | set a __default prefix__
[`WithSub(string)`](./log/conf.#L172) | set a __default sub-prefix__
[`WithOut(...io.Writer)`](./log/format.go#L179) | set (a) __default writer(s)__
[`WithFormat(LogFormatter)`](./log/format.go#L33) | set the [formatter](#different-formatters) for the log event output content
[`SkipExit` config](./log/conf.go#L78) | set the __skip-exit option__ (to skip `os.Exit(1)` and `panic()` calls)
[`WithFilter(event.Level)`](./log/conf.go#L203) | set a __log-level filter__
[`WithDatabase(...io.WriteCloser)`](./log/conf.go#L211) | set a __database writer__ (if [using a database](#databases))

Beyond the functions and preset configurations above, the package also exposes the following preset for the [default config](./log/conf.go#L56):

```go
var DefaultConfig LoggerConfig = &multiconf{
  confs: []LoggerConfig{
    WithFormat(TextColorLevelFirst),
    WithOut(),
    WithPrefix(event.Default_Event_Prefix),
  },
}
```

...and the following [(initialized) presets](./log/conf.go#L77) for several useful "defaults":

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

It's important to underline that the [`Logger` interface](./log/logger.go#L95) can also be launched in a goroutine without any hassle by using the [`log/logch` package](./log/logch/logch.go), for its [`ChanneledLogger` interface](./log/logch/logch.go#L12). 

> See the [_Channeled Logger_ example](#channeled-logger---example)

This interface provides a narrower set of methods, but instead focuses on setting up controls to interact with the logger and goroutine. Note the list of methods available:


Method | Description
:--:|:--:
[`Log(msg ...*event.Event)`](./log/logch/logch.go#L91) | takes in any number of pointers to event.Event, and iterating through each of them, pushing them to the LogMessage channel.
[`Close()`](./log/logch/logch.go#L116) | sends a signal (an empty `struct{}`) to the done channel, triggering the spawned goroutine to return
[`Channels() (logCh chan *event.Event, done chan struct{})`](./log/logch/logch.go#L132) | returns the LogMessage channel and the done channel, so that they can be used directly with the same channel messaging patterns

The [`ChanneledLogger` interface](./log/logch/logch.go#L12) can be initialized with the [`New(log.Logger)`](./log/logch/logch.go#L48) function, which creates both message and done channels, and then kicks off the goroutine with the input logger listening to messages in it. Note that if you require multiple loggers to be converted to a [`ChanneledLogger`](./log/logch/logch.go#L12), then you should merge them with [`log.Multilogger(...log.Logger)`](#multi-everything), first.


#### Feature-rich events

<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/event_builder.png" />
</p>


> See the [_Modular events_ example](#modular-events---example)


##### Data structure

The events are defined in a protocol buffer format, in [`proto/event.proto`](./proto/event.proto#L20); to give it a seamless integration as a gRPC logger's request message:

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

The event builder will allow chaining methods after [`event.New()`](./log/event/builder.go#L29) until the [`Build()`](./log/event/builder.go#L107) method is called. Below is a list of all available methods to the [`event.EventBuilder`](./log/event/builder.go#L14):

Method signature | Description
:--:|:--:
[`Prefix(p string) *EventBuilder`](./log/event/builder.go#L47) | set the prefix element
[`Sub(s string) *EventBuilder`](./log/event/builder.go#L54) | set the sub-prefix element
[`Message(m string) *EventBuilder`](./log/event/builder.go#L61) | set the message body element
[`Level(l Level) *EventBuilder`](./log/event/builder.go#L68) | set the level element
[`Metadata(m map[string]interface{}) *EventBuilder`](./log/event/builder.go#L75) | set (or add to) the metadata element
[`CallStack(all bool) *EventBuilder`](./log/event/builder.go#L94) | grab the current call stack, and add it as a "callstack" object in the event's metadata
[`Build() *Event`](./log/event/builder.go#L107) | build an event with configured elements, defaults applied where needed, and by adding a timestamp

##### Log levels

Log levels are defined as a protobuf enum, as [`Level` enum](./proto/event.proto#L9):

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

The [generated code](./log/event/event.pb.go) creates a type and two maps which set these levels:

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

The [`Level` type](./log/event/event.pb.go#L25) also has an exposed (custom) method, [`Int() int32`](./log/event/level.go#L6), which acts as a quick converter from the map value to an `int32` value.

##### Structured metadata


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/metadata.png" />
</p>


_output to the call above:_

```json
{
  "timestamp": "2022-07-14T15:48:23.386745176Z",
  "service": "module",
  "module": "service",
  "level": "info",
  "message": "Logger says hi!",
  "metadata": {
    "multi-value": true,
    "nested-metadata": {
      "inner-with-type-field": {
        "ok": true
      }
    },
    "three-numbers": [
      0,
      1,
      2
    ],
    "type": "structured logger"
  }
}
```


Metadata is added to the [`event.Event`](./log/event/event.pb.go#L96) as a `map[string]interface{}` which is compatible with JSON output (for the most part, for most the common data types). This allows a list of key-value pairs where the key is always a string (an identifier) and the value is the data itself, regardless of the type.

The event package also exposes a unique type ([`event.Field`](./log/event/field.go#L11)):

```go
// Field type is a generic type to build Event Metadata
type Field map[string]interface{}
```

The [`event.Field`](./log/event/field.go#L11) type exposes three methods to allow fast / easy conversion to [`structpb.Struct`](https://pkg.go.dev/google.golang.org/protobuf/types/known/structpb#Struct) pointers; needed for the protobuf encoders:

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

#### Callstack in metadata 

```json
{
  "timestamp": "2022-07-17T14:59:41.793879193Z",
  "service": "log",
  "level": "error",
  "message": "operation failed",
  "metadata": {
    "callstack": {
      "goroutine-1": {
        "id": "1",
        "stack": [
          {
            "method": "github.com/zalgonoise/zlog/log/trace.(*stacktrace).getCallStack(...)",
            "reference": "/go/src/github.com/zalgonoise/zlog/log/trace/trace.go:54"
          },
          {
            "method": "github.com/zalgonoise/zlog/log/trace.New(0x30?)",
            "reference": "/go/src/github.com/zalgonoise/zlog/log/trace/trace.go:41 +0x7f"
          },
          {
            "method": "github.com/zalgonoise/zlog/log/event.(*EventBuilder).CallStack(0xc000075fc0, 0x30?)",
            "reference": "/go/src/github.com/zalgonoise/zlog/log/event/builder.go:99 +0x6b"
          },
          {
            "method": "main.subOperation(0x0)",
            "reference": "/go/src/github.com/zalgonoise/zlog/examples/logger/callstack_in_metadata/callstack_md.go:31 +0x34b"
          },
          {
            "method": "main.operation(...)",
            "reference": "/go/src/github.com/zalgonoise/zlog/examples/logger/callstack_in_metadata/callstack_md.go:15"
          },
          {
            "method": "main.main()",
            "reference": "/go/src/github.com/zalgonoise/zlog/examples/logger/callstack_in_metadata/callstack_md.go:42 +0x33"
          }
        ],
        "status": "running"
      }
    },
    "error": "input cannot be zero",
    "input": 0
  }
}
```

> See the [_Callstack in Metadata_ example](#callstack-in-metadata---example)

It's also possible to include the current callstack (at the time of the log event being built / created) as metadata to the log entry, by calling the event's [`Callstack(all bool)`](./log/event/builder.go#L94/) method.

This call will add the `map[string]interface{}` output of a [`trace.New(bool)` call](./log/trace/trace.go#L27), to the event's metadata element, as an object named `callstack`. This `trace` package will fetch the call stack from a [`runtime.Stack([]byte, bool)` call](https://pkg.go.dev/runtime#Stack), where the limit in size is 1 kilobyte (`make([]byte, 1024)`).

This package will parse the contents of this call and build a JSON document (as a `map[string]interface{}`) with key `callstack`, as a list of objects. These objects will have three elements:

- an `id` element, as the numeric identifier for the goroutine in question
- a `status` element, like `running`
- a `stack` element, which is a list of objects, each object contains:
    - a `method` element (package and method / function call)
	- a `reference` element (path in filesystem, with a pointer to the file and line)

#### Multi-everything

> See the [_Multilogger_ example](#multilogger---example)

In this library, there are many implementations of `multiSomething`, following the same logic of [`io.MultiWriter()`](https://pkg.go.dev/io#MultiWriter).

In the reference above, the data structure holds a slice of [`io.Writer` interface](https://pkg.go.dev/io#Writer), and implements the same methods as an [`io.Writer`](https://pkg.go.dev/io#Writer). Its implementation of the `Write()` method will involve iterating through all configured [`io.Writer`](https://pkg.go.dev/io#Writer), and calling its own `Write()` method accordingly.

It is a very useful concept in the sense that you're able to _merge_ a slice of interfaces while working with a single one. It allows greater manouverability with maybe a few downsides or restrictions. It is not a required module but merely a helper, or a wrapper for a simple purpose.

The actual [`io.MultiWriter()`](https://pkg.go.dev/io#MultiWriter) is used when defining a [`io.Writer`](https://pkg.go.dev/io#Writer) for the logger; to allow setting it up with multiple writers:


```go
func (l *logger) SetOuts(outs ...io.Writer) Logger {
	// (...)

	l.out = io.MultiWriter(newouts...)
	return l
}
```
> from [log/logger.go](./log/logger.go#L180)

```go
func (l *logger) AddOuts(outs ...io.Writer) Logger {
	// (...)

	l.out = io.MultiWriter(newouts...)
	return l
}
```
> from [log/logger.go](./log/logger.go#L207)

```go
func WithOut(out ...io.Writer) LoggerConfig {
	// (...)

	if len(out) > 1 {
		return &LCOut{
			out: io.MultiWriter(out...),
		}
	}

	// (...)
}
```
> from [log/conf.go](./log/conf.go#L179)

...but even beyond this useful implementation, it is _mimicked_ in other pars of the code base:

- as a [`LoggerConfig` merger](./log/conf.go#L25):

```go
type LoggerConfig interface {
	Apply(lb *LoggerBuilder)
}

type multiconf struct {
	confs []LoggerConfig
}

func (m multiconf) Apply(lb *LoggerBuilder) {
	for _, c := range m.confs {
		if c != nil {
			c.Apply(lb)
		}
	}
}

func MultiConf(conf ...LoggerConfig) LoggerConfig {
	// (...)
}
```

-  as a [`Logger` interface merger](./log/multilog.go#L11):

```go
type multiLogger struct {
	loggers []Logger
}

// func (l *multiLogger) {every single method in Logger}

func MultiLogger(loggers ...Logger) Logger {
	// (...)
}
```

- as a [`MultiWriteCloser`](./store/db/db.go#L96) for databases:

```go
func MultiWriteCloser(wc ...io.WriteCloser) io.WriteCloser {
	// (...)
}

```

- as a [`LogClientConfig` merger](./grpc/client/conf.go#L61):

```go
type LogClientConfig interface {
	Apply(ls *gRPCLogClientBuilder)
}

type multiconf struct {
	confs []LogClientConfig
}

func (m multiconf) Apply(lb *gRPCLogClientBuilder) {
	for _, c := range m.confs {
		c.Apply(lb)
	}
}

func MultiConf(conf ...LogClientConfig) LogClientConfig {
	// (...)
}
```

- as a [`LogServerConfig` merger](./grpc/server/conf.go#L57):

```go
type LogServerConfig interface {
	Apply(ls *gRPCLogServerBuilder)
}

type multiconf struct {
	confs []LogServerConfig
}

func (m multiconf) Apply(lb *gRPCLogServerBuilder) {
	for _, c := range m.confs {
		c.Apply(lb)
	}
}

func MultiConf(conf ...LogServerConfig) LogServerConfig {
	// (...)
}
```

-  as a [`GRPCLogger` interface (client) merger](./grpc/client/multilog.go#L13):

```go
type multiLogger struct {
	loggers []GRPCLogger
}

// func (l *multiLogger) {every single method in Logger and ChanneledLogger}

func MultiLogger(loggers ...GRPCLogger) GRPCLogger {
	// (...)
}
```

-  as a [`LogServer` interface merger](./grpc/server/multilog.go#L13):

```go
type multiLogger struct {
	loggers []LogServer
}

// func (l *multiLogger) Serve() {}
// func (l *multiLogger) Stop() {}

func MultiLogger(loggers ...LogServer) LogServer {
	// (...)
}
```

#### Different formatters

> See the [_Output formats_ example](#output-formats---example)

The logger can output events in several different formats, listed below:

##### Text

<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/text_formatter.png" />
</p>


The text formatter allows an array of options, with the text formatter sub-package exposing a builder to create a text formatter. Below is the list of methods you can expect when calling [`text.New()....Build()`](./log/format/text/text.go#L87):

Method | Description
:--:|:--:
[`Time(LogTimestamp)`](./log/format/text/text.go#L93) | define the timestamp format, based on the exposed [list of timestamps](#log-timestamps), from the table below
[`LevelFirst()`](./log/format/text/text.go#L100) | place the log level as the first element in the line
[`DoubleSpace()`](./log/format/text/text.go#L107) | place double-tab-spaces between elements (`\t\t`)
[`Color()`](./log/format/text/text.go#L114) | add color to log levels (it is skipped on Windows CLI, as it doesn't support it)
[`Upper()`](./log/format/text/text.go#L121) | make log level, prefix and sub-prefix uppercase
[`NoTimestamp()`](./log/format/text/text.go#L128) | skip adding the timestamp element
[`NoHeaders()`](./log/format/text/text.go#L135) | skip adding the prefix and sub-prefix elements
[`NoLevel()`](./log/format/text/text.go#L142) | skip adding the log level element

##### Log Timestamps

Regarding the timestamp constraints, please note the available timestamps for the text formatter:

Constant | Description
:--:|:--:
[`LTRFC3339Nano`](./log/format/text/text.go#L48) | Follows the standard in [`time.RFC3339Nano`](https://pkg.go.dev/time#pkg-constants)
[`LTRFC3339`](./log/format/text/text.go#L49) | Follows the standard in [`time.RFC3339`](https://pkg.go.dev/time#pkg-constants)
[`LTRFC822Z`](./log/format/text/text.go#L50) | Follows the standard in [`time.RFC822Z`](https://pkg.go.dev/time#pkg-constants)
[`LTRubyDate`](./log/format/text/text.go#L51) | Follows the standard in [`time.RubyDate`](https://pkg.go.dev/time#pkg-constants)
[`LTUnixNano`](./log/format/text/text.go#L52) | Displays a Unix timestamp, in nanos
[`LTUnixMilli`](./log/format/text/text.go#L53) | Displays a Unix timestamp, in millis
[`LTUnixMicro`](./log/format/text/text.go#L54) | Displays a Unix timestamp, in micros

The library also exposes a few initialized preset configurations using text formatters, as in the list below. While these are [`LoggerConfig`](./log/conf.go#L21) presets, they're a wrapper for [the same formatter](./log/format.go#L18), which is also available by not including the `Cfg` prefix:

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


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/json_formatter.png" />
</p>


The JSON formatter allow generating JSON events in different ways. These formatters are already initialized as [`LoggerConfig`](./log/format.go#L69) and [`LogFormatter`](./log/format.go#L39) objects.

[This formatter](./log/format/json/json.go#L13) allows creating JSON events separated by newlines or not, and also to optionally add indentation:

```go
type FmtJSON struct {
	SkipNewline bool
	Indent      bool
}
```

Also note how the [`LoggerConfig`](./log/format.go#L69) presets are exposed. While these are a wrapper for the same formatter, they are also available as [`LogFormatter`](./log/format.go#L39) by not including the `Cfg` prefix:

```go
var (
	CfgFormatJSON                  = WithFormat(&json.FmtJSON{})  // default
	CfgFormatJSONSkipNewline       = WithFormat(&json.FmtJSON{SkipNewline: true}) // with a skip-newline config
	CfgFormatJSONIndentSkipNewline = WithFormat(&json.FmtJSON{SkipNewline: true, Indent: true}) // with a skip-newline and indentation config
	CfgFormatJSONIndent            = WithFormat(&json.FmtJSON{Indent: true}) // with an indentation config
)
```



##### BSON


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/bson_formatter.png" />
</p>



##### CSV


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/csv_formatter.png" />
</p>

##### XML


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/xml_formatter.png" />
</p>


##### Protobuf


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/pb_formatter.png" />
</p>


##### Gob

<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/gob_formatter.png" />
</p>

#### Data stores

##### Writer interface


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/log_writer.png" />
</p>

> Output of the [_Logger as a Writer_ example](#logger-as-a-writer---example)


Not only [`Logger` interface](./log/logger.go#L95) uses the [`io.Writer` interface](https://pkg.go.dev/io#Writer) to write to its outputs with its [`Output()` method](./log/print.go#L88), it also implements it in its own [`Write()` method](./log/logger.go#L307) so it can be used directly as one. This gives the logger more flexibility as it can be vastly integrated with other modules.


The the input slice of bytes is decoded, in case the input is an encoded [`event.Event`](./log/event/event.pb.go#L96). If the conversion is successful, the input event is logged as-is.

If it is not an [`event.Event`](./log/event/event.pb.go#L96) (there will be an error from the [`Decode()` method](./log/event/event.go#L32)), then a new message is created where:
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


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/datastore_file.png" />
</p>

> See the [example in `examples/datastore/file/`](./examples/datastore/file/logfile.go)

This library also provides a simple [`Logfile`](./store/fs/logfile.go#L18) (an actual file in the disk where log entries are written to) configuration with appealing features for simple applications.

```go
type Logfile struct {
	path   string
	file   *os.File
	size   int64
	rotate int
}
```

The [`Logfile`](./store/fs/logfile.go#L18) exposes a few methods that could be helpful to keep the events organized:

Method | Description
:--:|:--:
[`MaxSize(mb int) *Logfile`](./store/fs/logfile.go#L59) | sets the rotation indicator for the Logfile, or, the target size when should the logfile be rotated (in MBs)
[`Size() (int64, error)`](./store/fs/logfile.go#L139) | a wrapper for an [`os.File.Stat()`](https://pkg.go.dev/os#File.Stat) followed by [`fs.FileInfo.Size()`](https://pkg.go.dev/io/fs#FileInfo)
[`IsTooHeavy() bool`](./store/fs/logfile.go#L151) | verify the file's size and rotate it if exceeding the set maximum weight (in the Logfile's rotate element)
[`Write(b []byte) (n int, err error)`](./store/fs/logfile.go#L174) | implement the [`io.Writer` interface](https://pkg.go.dev/io#Writer), for Logfile to be compatible with Logger as an output to be used


##### Databases

It's perfectly possible to write log events to a database instead of the terminal, a buffer, or a file. It makes it more reliable for a larger scale operation or for the long-run.

This library leverages an ORM to handle interactions with most of the databases, for the sake of simplicity and streamlined testing -- these should focus on using a database as a writer, and not re-testing the database connections, configurations, etc. This is why an ORM is being used. This library uses [GORM](https://gorm.io/) for this purpose.

Databases are not configured to loggers as an [`io.Writer` interface](https://pkg.go.dev/io#Writer) using the [`WithOut()` method](./log/conf.go#L179), but with their dedicated [`WithDatabase()` method](./log/conf.go#L211). This takes an [`io.WriterCloser` interface](https://pkg.go.dev/io#WriteCloser).


To create this [`io.WriterCloser`](https://pkg.go.dev/io#WriteCloser), either the database package's appropriate `New()` method can be used; or by using its package function for the same purpose, `WithXxx()`.

Note the available database writers, and their features:

##### SQLite


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/datastore_sqlite.png" />
</p>


> See the [example in `examples/datastore/db/sqlite/`](./examples/datastore/db/sqlite/sqlite.go)

Symbol | Type | Description
:--:|:--:|:--:
[`New(path string) (sqldb io.WriteCloser, err error)`](./store/db/sqlite/sqlite.go#L22) | function | takes in a path to a `.db` file; and create a new instance of a [`SQLite3`](./store/db/sqlite/sqlite.go#L15) object; returning an [`io.WriterCloser` interface](https://pkg.go.dev/io#WriteCloser) and an error.
[`*SQLite.Create(msg ...*event.Event) error`](./store/db/sqlite/sqlite.go#L39) | method | will register any number of [`event.Event`](./log/event/event.pb.go#L96) in the SQLite database, returning an error (exposed method, but it's mostly used internally )
[`*SQLite.Write(p []byte) (n int, err error)`](./store/db/sqlite/sqlite.go#L71) | method | implements the [`io.Writer` interface](https://pkg.go.dev/io#Writer), for SQLite DBs to be used with a [`Logger` interface](./log/logger.go#L95), as its writer.
[`*SQLite.Close() error`](./store/db/sqlite/sqlite.go#L97) | method | method added for compatibility with DBs that require it
[`WithSQLite(path string) log.LoggerConfig`](./store/db/sqlite/sqlite.go#L113) | function | takes in a path to a `.db` file, and a table name; and returns a [`LoggerConfig`](./log/conf.go#L21) so that this type of writer is defined in a [`Logger`](./log/logger.go#L95)

##### MySQL


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/datastore_mysql.png" />
</p>


> See the [example in `examples/datastore/db/mysql/`](./examples/datastore/db/mysql/mysql.go)

> Using this package will require the following environment variables to be set:

Variable | Type | Description
:--:|:--:|:--:
`MYSQL_USER` | string | username for the MySQL database connection
`MYSQL_PASSWORD` | string | password for the MySQL database connection

Symbol | Type | Description
:--:|:--:|:--:
[`New(address, database string) (sqldb io.WriteCloser, err error)`](./store/db/mysql/mysql.go#L32) | function | takes in a MySQL DB address and database name; and create a new instance of a [`MySQL`](./store/db/mysql/mysql.go#L24) object; returning an [`io.WriterCloser` interface](https://pkg.go.dev/io#WriteCloser) and an error.
[`*MySQL.Create(msg ...*event.Event) error`](./store/db/mysql/mysql.go#L50) | method | will register any number of [`event.Event`](./log/event/event.pb.go#L96) in the MySQL database, returning an error (exposed method, but it's mostly used internally )
[`*MySQL.Write(p []byte) (n int, err error)`](./store/db/mysql/mysql.go#L82) | method | implements the [`io.Writer` interface](https://pkg.go.dev/io#Writer), for MySQL DBs to be used with a [`Logger` interface](./log/logger.go#L95), as its writer.
[`*MySQL.Close() error`](./store/db/mysql/mysql.go#L112) | method | method added for compatibility with DBs that require it
[`WithMySQL(addr, database string) log.LoggerConfig`](./store/db/mysql/mysql.go#L166) | function | takes in an address to a MySQL server, and a database name; and returns a [`LoggerConfig`](./log/conf.go#L21) so that this type of writer is defined in a [`Logger`](./log/logger.go#L95)


##### PostgreSQL



<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/datastore_postgres.png" />
</p>

> See the [example in `examples/datastore/db/postgres/`](./examples/datastore/db/postgres/postgres.go)

> Using this package will require the following environment variables to be set:

Variable | Type | Description
:--:|:--:|:--:
`POSTGRES_USER` | string | username for the Postgres database connection
`POSTGRES_PASSWORD` | string | password for the Postgres database connection

Symbol | Type | Description
:--:|:--:|:--:
[`New(address, port, database string) (sqldb io.WriteCloser, err error)`](./store/db/postgres/postgres.go#L33) | function | takes in a Postgres DB address, port and database name; and create a new instance of a [`Postgres`](./store/db/postgres/postgres.go#L24) object; returning an [`io.WriterCloser` interface](https://pkg.go.dev/io#WriteCloser) and an error.
[`*Postgres.Create(msg ...*event.Event) error`](./store/db/postgres/postgres.go#L52) | method | will register any number of [`event.Event`](./log/event/event.pb.go#L96) in the Postgres database, returning an error (exposed method, but it's mostly used internally )
[`*Postgres.Write(p []byte) (n int, err error)`](./store/db/postgres/postgres.go#L84) | method | implements the [`io.Writer` interface](https://pkg.go.dev/io#Writer), for Postgres DBs to be used with a [`Logger` interface](./log/logger.go#L95), as its writer.
[`*Postgres.Close() error`](./store/db/postgres/postgres.go#L118) | method | method added for compatibility with DBs that require it
[`WithPostgres(addr, port, database string) log.LoggerConfig`](./store/db/postgres/postgres.go#L172) | function | takes in an address and port to a Postgres server, and a database name; and returns a [`LoggerConfig`](./log/conf.go#L21) so that this type of writer is defined in a [`Logger`](./log/logger.go#L95)



##### MongoDB


<p align="center">
  <img src="https://github.com/ZalgoNoise/zlog/raw/media/img/datastore_mongo.png" />
</p>

> See the [example in `examples/datastore/db/mongo/`](./examples/datastore/db/mongo/mongo.go)

> Using this package will require the following environment variables to be set:

Variable | Type | Description
:--:|:--:|:--:
`MONGO_USER` | string | username for the Mongo database connection
`MONGO_PASSWORD` | string | password for the Mongo database connection

Symbol | Type | Description
:--:|:--:|:--:
[`New(address, database, collection string) (io.WriteCloser, error)`](./store/db/mongo/mongo.go#L36) | function | takes in a MongoDB address, database and collection names; and create a new instance of a [`Mongo`](./store/db/mongo/mongo.go#L26) object; returning an [`io.WriterCloser` interface](https://pkg.go.dev/io#WriteCloser) and an error.
[`*Mongo.Create(msg ...*event.Event) error`](./store/db/mongo/mongo.go#L85) | method | will register any number of [`event.Event`](./log/event/event.pb.go#L96) in the Mongo database, returning an error (exposed method, but it's mostly used internally )
[`*Mongo.Write(p []byte) (n int, err error)`](./store/db/mongo/mongo.go#L132) | method | implements the [`io.Writer` interface](https://pkg.go.dev/io#Writer), for Mongo DBs to be used with a [`Logger` interface](./log/logger.go#L95), as its writer.
[`*Mongo.Close() error`](./store/db/mongo/mongo.go#L79) | method | used to terminate the live connection to the MongoDB instance
[`WithPostgres(addr, port, database string) log.LoggerConfig`](./store/db/mongo/mongo.go#L163) | function | takes in the address to the mongo server, and a database and collection name; and returns a [`LoggerConfig`](./log/conf.go#L21) so that this type of writer is defined in a [`Logger`](./log/logger.go#L95)

#### gRPC

To provide a solution to loggers that write _over the wire_, this library implements a [gRPC](https://grpc.io/) log server with a number of useful features; as well a [gRPC](https://grpc.io/) log client which will act as a regular logger, but one that writes the log messages to a gRPC log server.

The choice for gRPC was simple. The framework is very solid and provides both fast and secure transmission of messages over a network. This is all that it's needed, right? Nope! There are also protocol buffers which helped in shaping the structure of this library in a more organized way (in my opinion).

Originally, the plan was to create the event data structures in Go (manually), and from that point integrate the logger logic as an HTTP Writer or something -- note this is already possible as the [`Logger` interface](./log/logger.go#L95) implements the [`io.Writer` interface](https://pkg.go.dev/io#Writer) already. But the problem there would be a repetition in defining the event data structure. If gRPC was in fact the choice, it would mean that there would be a data structure for Go and another for gRPC (with generated Go code, for the same thing).

So, easy-peasy: scratch off the Go data structure and keep the protocol buffers, even for (local) events and loggers. This worked great, it was easy enough to switch over, and the logic remained _kinda_ the same way, in the end.

The added benefit is that gRPC and protobuf will create this generated code (from [`proto/event.proto`](./proto/event.proto) and [`proto/service.proto`](./proto/service.proto), to [`log/event/event.pb.go`](./log/event/event.pb.go) and [`proto/service/service.pb.go`](./proto/service/service.pb.go) respectively); which it a huge boost to productivity.

An added bonus is a very lightweight encoded format of the exchanged messages, as you are able to convert the protocol buffer messages into byte slices, too.

Lastly, no network logic implementation headaches as creating a gRPC server and client is super smooth -- it only takes a few hours reading documentation and examples. The benefit is being able to very quickly push out a server-client solution to your app, with zero effort in the _engine_ transmitting those messages, only what you actually do with them.

On this note, it's also possible to easily implement a log client in a different programming language supported by [gRPC](https://grpc.io/). This means that if you really love this library, then you could create a Java-based gRPC log client for your Android app, while you run your gRPC log server in Go, in your big-shot datacenter.

This section will cover features that you will find in both server and client implementations.

##### gRPC Log Service

The service, defined in [`proto/service/service.go`](./proto/service/service.go), is the implementation of the log server core logic, from the [gRPC generated code](./proto/service/service.pb.go).

This file will have the implementation of the [`LogServer` struct](./proto/service/service.go#L34) and its [`Log()` method](./proto/service/service.go#L75) and [`LogStream()` method](./proto/service/service.go#L92) -- on how the server handles the messages exchanged in either configuration, as a unary RPC logger or a stream RPC logger.

It also contains additional methods used within the core logic of the gRPC Log Server; such as its [`Done()` method](./proto/service/service.go#L321) and [`Stop` method](./store/service/service.go#L247).

##### gRPC Log Server


> See the [examples in `examples/grpc/simple_unary_client_server/server/`](./examples/grpc/simple_unary_client_server/server/server.go) and [in `examples/grpc/simple_stream_client_server/server/`](./examples/grpc/simple_stream_client_server/server/server.go)

The [`LogServer` interface](./grpc/server/server.go#L24) is found in [`grpc/server/server.go`](./grpc/server/server.go), which defines how the server is initialized, how can it be configured, and other features. This should be perceived as a simple wrapper for setting up a gRPC server using the logic in [`proto/service/service.go`](./proto/service/service.go), with added features to make it even more useful:

```go
type LogServer interface {
	Serve()
	Stop()
	Channels() (logCh, logSvCh chan *event.Event, errCh chan error)
}
```

A new Log Server is created with the public function [`New(...LogServerConfig)`](./grpc/server/server.go#L102), which parses any number of configurations (covered below). The resulting [`GRPCLogServer` pointer](./grpc/server/server.go#L31) will expose the following methods:

Method | Description
:--:|:--:
[`Serve()`](./grpc/server/server.go#L241) | a long-running, blocking function which will launch the gRPC server 
[`Stop()`](./grpc/server/server.go#L277) | a wrapper for the routine involved to (gracefully) stop this gRPC Log Server.
[`Channels() (logCh, logSvCh chan *event.Event, errCh chan error)`](./grpc/server/server.go#L289) | returns channels for a Log Server's I/O. It returns a channel for log messages (for actual log event writes), a channel for the service logger (the server's own logger), and an error channel to collect Log Server errors from.

##### Log Server Configs

The Log Server can be configured in a number of ways, like specifying exposed address, the output logger for your events, a _service logger_ for the Log Server activity (yes, a logger for your logger), added metadata like timing, and of course TLS. 

Here is the list of exposed functions to allow a granular configuration of your Log Server:

Function | Description 
:--:|:--:
[`WithAddr(string)`](./grpc/server/conf.go#L161) | takes one address for the gRPC Log Server to listen to. Defaults to `localhost:9099`
[`WithLogger(...log.Logger)`](./grpc/server/conf.go#L180) | defines this gRPC Log Server's logger(s)
[`WithServiceLogger(...log.Logger)`](./grpc/server/conf.go#L204) | defines this gRPC Log Server's service logger(s) (for the gRPC Log Server activity)
[`WithServiceLoggerV(...log.Logger)`](./grpc/server/conf.go#230) | defines this gRPC Log Server's service logger(s) (for the gRPC Log Server activity) in verbose mode -- by adding an interceptor that checks each transaction, if OK or not, and for errors (added overhead)
[`WithTiming()`](./grpc/server/conf.go#L254) | sets a gRPC Log Server's service logger to measure the time taken when executing RPCs, as added metadata (added overhead)
[`WithGRPCOpts(...grpc.ServerOption)`](./grpc/server/conf.go#L263) | sets a gRPC Log Server's service logger to measure the time taken when executing RPCs, as added metadata (added overhead)
[`WithTLS(certPath, keyPath string, caPath ...string)`](./grpc/server/conf.go#L288) | allows configuring TLS / mTLS for a gRPC Log Server. If only two parameters are passed (certPath, keyPath), it will run its TLS flow. If three parameters are set (certPath, keyPath, caPath), it will run its mTLS flow.


Lastly, the library also exposes some preset configurations:

```go
var (
	defaultConfig LogServerConfig = &multiconf{
		confs: []LogServerConfig{
			WithAddr(""),
			WithLogger(),
			WithServiceLogger(),
		},
	}

	DefaultCfg        = LogServerConfigs[0] // default LogServerConfig
	ServiceLogDefault = LogServerConfigs[1] // default logger as service logger
	ServiceLogNil     = LogServerConfigs[2] // nil-service-logger LogServerConfig
	ServiceLogColor   = LogServerConfigs[3] // colored, level-first, service logger
	ServiceLogJSON    = LogServerConfigs[4] // JSON service logger
	LoggerDefault     = LogServerConfigs[5] // default logger
	LoggerColor       = LogServerConfigs[6] // colored, level-first logger
	LoggerJSON        = LogServerConfigs[7] // JSON logger
)
```

##### gRPC Log Client


> See the [examples in `examples/grpc/simple_unary_client_server/client/`](./examples/grpc/simple_unary_client_server/client/client.go) and [in `examples/grpc/simple_stream_client_server/client/`](./examples/grpc/simple_stream_client_server/client/client.go)

There is a gRPC Log Client implementation in Go, for the sake of providing an out-of-the-box solution for communicating with the gRPC Log Server; although this can simply serve as a reference for you to implement your own gRPC Log Client -- in any of the gRPC-supported languages.

This client will act just like a regular (channeled) [`Logger` interface](./log/logger.go#L95), with added features (and configurations):

```go
// import (
// 	"github.com/zalgonoise/zlog/log"
// 	"github.com/zalgonoise/zlog/log/logch"
// )

type GRPCLogger interface {
	log.Logger
	logch.ChanneledLogger
}
```

Creating a new gRPC Log Client depends on whether you're setting up a Unary gRPC logger or a Stream gRPC one. The [`New(...LogClientConfig)` function](./grpc/client/client.go#L166) will serve as a factory, where depending on the configuration it will either spawn a [Unary gRPC logger](./grpc/client/client.go#L192) or a [Stream gRPC logger](./grpc/client/client.go#L199). Similar to other modules, the underlying builder pattern as you create a [`GRPCLogger`](./grpc/client/client.go#L62) will apply the default configuration before overwriting it with the user's configs.

This client will expose the public methods as per the interfaces it contains, and nothing else. There are a few things to keep in mind:

Method | Description
:--:|:--:
[`Close()`](./grpc/client/client.go#L582) | iterates through all (alive) connections in the `ConnAddr` map, and close them. After doing so, it sends the done signal to its channel, which causes all open streams to cancel their context and exit gracefully
[`Output(*event.Event) (int, error)`](./grpc/client/client.go#L622) |  pushes the incoming Log Message to the message channel, which is sent to a gRPC Log Server, either via a Unary or Stream RPC. Note that it will always return `1, nil`.
[`SetOuts(...io.Writer) log.Logger`](./grpc/client/client.go#L643) | for compatibility with the Logger interface, this method must take in io.Writers. However, this is not how the gRPC Log Client will work to register messages. Instead, the input io.Writer needs to be of type `ConnAddr`. More info on this type below. This method overwrites the configured addresses.
[`AddOuts(...io.Writer) log.Logger`](./grpc/client/client.go#L703) | for compatibility with the Logger interface, this method must take in io.Writers. However, this is not how the gRPC Log Client will work to register messages. Instead, the input io.Writer needs to be of type `ConnAddr`. More info on this type below. This method adds addresses to the configured ones.
[`Write([]byte) (int, error)`](./grpc/client/client.go#L776) | consider that `Write()` will return a call of `Output()`. This means that you should expect it to return `1, nil`.
[`IsSkipExit() bool`](./grpc/client/client.go#L859) | returns a boolean on whether the gRPC Log Client's __service logger__ is set to skip os.Exit(1) or panic() calls.


##### Log Client Configs

This Log Client can be configured in a number of ways, like specifying exposed address, the output logger for your events, a _service logger_ for the gRPC / Log Server activity (yes, a logger for your logger), added metadata like timing, and of course TLS. 

Here is the list of exposed functions to allow a granular configuration of your Log Server:

Function | Description 
:--:|:--:
[`WithAddr(...string)`](./grpc/client/conf.go#L176) | take in any amount of addresses, and create a [connections map](#connection-addresses) with them, for the gRPC client to connect to the server. Defaults to `localhost:9099`
[`StreamRPC()`](./grpc/client/conf.go#L199) | sets this gRPC Log Client type as Stream RPC
[`UnaryRPC()`](./grpc/client/conf.go#L208) | sets this gRPC Log Client type as Unary RPC
[`WithLogger(...log.Logger)`](./grpc/client/conf.go#L223) | defines this gRPC Log Client's service logger. This logger will register the gRPC Client transactions; and not the log messages it is handling.
[`WithLoggerV(...log.Logger)`](./grpc/client/conf.go#L265) | defines this gRPC Log Client's service logger, in verbose mode. This logger will register the gRPC Client transactions; and not the log messages it is handling. (added overhead)
[`WithBackoff(time.Duration, BackoffFunc)`](./grpc/client/conf.go#L310) | takes in a [`time.Duration`](https://pkg.go.dev/time#Duration) value to set as the exponential backoff module's retry deadline, and a [`BackoffFunc`](./grpc/client/backoff.go#L33) to customize the backoff pattern. [Backoff](#log-client-backoff) is further described in the next section.
[`WithTiming()`](./grpc/client/conf.go#L340) | sets a gRPC Log Client's service logger to measure the time taken when executing RPCs. It is only an option, and is directly tied to the configured service logger. (added overhead)
[`WithGRPCOpts(...grpc.DialOption)`](./grpc/client/conf.go#L349) | allows passing in any number of grpc.DialOption, which are added to the gRPC Log Client.
[`Insecure()`](./grpc/client/conf.go#L372) | allows creating an insecure gRPC connection (maybe for testing purposes) by adding a new option for insecure transport credentials (no TLS / mTLS).
[`WithTLS(string, ...string)`](./grpc/client/conf.go#L372) | allows configuring TLS / mTLS for a gRPC Log Client. If only one parameter is passed (caPath), it will run its TLS flow. If three parameters are set (caPath, certPath, keyPath), it will run its mTLS flow.

Lastly, the library also exposes some preset configurations:

```go
var (
	defaultConfig LogClientConfig = &multiconf{
		confs: []LogClientConfig{
			WithAddr(""),
			WithGRPCOpts(),
			Insecure(),
			WithLogger(),
			WithBackoff(0, BackoffExponential()),
		},
	}

	DefaultCfg     = LogClientConfigs[0] // default LogClientConfig
	BackoffFiveMin = LogClientConfigs[1] // backoff config with 5-minute deadline
	BackoffHalfMin = LogClientConfigs[2] // backoff config with 30-second deadline
)
```


##### Log Client Backoff

There is a backoff module available, in order to retry transactions in case they fail in any (perceivable) way. While this is optional, it was implemented to consider that connections over a network may fail.

This package exposes the following types to serve as a core logic for any backoff implementation:

```go
type BackoffFunc func(uint) time.Duration

type Backoff struct {
	counter     uint
	max         time.Duration
	wait        time.Duration
	call        interface{}
	msg         []*event.Event
	backoffFunc BackoffFunc
	locked      bool
	mu          sync.Mutex
}
```

[`BackoffFunc`](./grpc/client/backoff.go#L33) takes in a(n unsigned) integer representing the attempt counter, and returns a time.Duration value of how much should the module wait before the next attempt / retry.

[`Backoff`](./grpc/client/backoff.go#L44) struct defines the elements of a backoff module, which is configured by setting a [`BackoffFunc`](./grpc/client/backoff.go#L33) to define the interval between each attempt.

Backoff will also try to act as a message buffer in case the server connection cannot be established -- as it will attempt to flush these records to the server as soon as connected. 

Implementing backoff logic is as simple as writing a function which will return a function with the same signature as [`BackoffFunc`](./grpc/client/backoff.go#L33). The parameters that your function takes or how it arrives to the return [`time.Duration`](https://pkg.go.dev/time#Duration) value is completely up to you.

Two examples below, one for [`NoBackoff()`](./grpc/client/backoff.go#L58) and one for [`BackoffExponential()`](./grpc/client/backoff.go#L89):

```go
func NoBackoff() BackoffFunc {
	return func(attempt uint) time.Duration {
		return 0
	}
}

func BackoffExponential() BackoffFunc {
	return func(attempt uint) time.Duration {
		return time.Millisecond * time.Duration(
			int64(math.Pow(2, float64(attempt)))+rand.New(
				rand.NewSource(time.Now().UnixNano())).Int63n(1000),
		)
	}
}
```

With this in mind, regardless if the exposed [`BackoffFunc`](./grpc/client/backoff.go#L33), you may pass a deadline and your own [`BackoffFunc`](./grpc/client/backoff.go#L33) to [`WithBackoff()`](./grpc/client/conf.go#L310), as you create your gRPC Log Client.

Here is a list of the preset [`BackoffFunc`](./grpc/client/backoff.go#L33) factories, available in this library:

Function | Description
:--:|:--:
[`NoBackoff()`](./grpc/client/backoff.go#L58) | returns a [`BackoffFunc`](./grpc/client/backoff.go#L33) that overrides the backoff module by setting a zero wait-between duration. This is detected as a sign that the module should be overriden.
[`BackoffLinear(time.Duration)`](./grpc/client/backoff.go#L67) | returns a [`BackoffFunc`](./grpc/client/backoff.go#L33) that sets a linear backoff according to the input duration. If the input duration is 0, then the default wait-between time is set (3 seconds).
[`BackoffIncremental(time.Duration)`](./grpc/client/backoff.go#L79) | returns a [`BackoffFunc`](./grpc/client/backoff.go#L33) that calculates exponential backoff according to a scalar method
[`BackoffExponential()`](./grpc/client/backoff.go#L89) | returns a [`BackoffFunc`](./grpc/client/backoff.go#L33) that calculates exponential backoff [according to its standard](https://en.wikipedia.org/wiki/Exponential_backoff)

Creating a new [`Backoff`](./grpc/client/backoff.go#L44) instance can be manual, although not necessary considering it is embeded in the gRPC Log Client's logic:

```go
// import "github.com/zalgonoise/zlog/grpc/client"

b := client.NewBackoff()
b.BackoffFunc(
	// your BackoffFunc here
)
```

From this point onwards, the [`Backoff`](./grpc/client/backoff.go#L44) module is called on certain errors. For example:

_From [grpc/client/client.go](./grpc/client/client.go#L230)_

```go
func (c *GRPCLogClient) connect() error {
	// (...)

		// handle dial errors
		if err != nil {
			retryErr := c.backoff.UnaryBackoffHandler(err, c.svcLogger)
			if errors.Is(retryErr, ErrBackoffLocked) {
				return retryErr
			} else if errors.Is(retryErr, ErrFailedConn) {
				return retryErr
			} else {
				// (...)
			}
		// (...)
		}
	// (...)
}
```

While this [`Backoff`](./grpc/client/backoff.go#L44) logic can be taken as a reference for different implementations, it is very specific to the gRPC Log Client's logic considering its [`init()` method](./grpc/client/backoff.go#L109) and [`Register()` method](./grpc/client/backoff.go#L238); which are hooks for being able to work with either Unary or Stream gRPC Log Clients.

##### Connection Addresses

Connection Addresses, or [`ConnAddr` type](./grpc/address/address.go#L8) is a custom type for `map[string]`[`*grpc.ClientConn`](https://pkg.go.dev/google.golang.org/grpc#ClientConn), that also exposes a few handy methods for this particular application. 

Considering that the Logger interface works with [`io.Writer` interfaces](https://pkg.go.dev/io#Writer) to write events, it was becoming pretty obvious that the gRPC Client / Server logic would need to either discard the `Write()`, `AddOuts()` and `SetOuts()` while adding different methods (or configs) in replacement ...or why not keep working with an [`io.Writer` interface](https://pkg.go.dev/io#Writer)?

The [`ConnAddr` type](./grpc/address/address.go#L8) implements this and other useful methods which will allow the gRPC client and server logic to leverage the same methods as in the [`Logger` interface](./log/logger.go#L95) for the purposes that it needs. This, similar to the [`Logger` interface](./log/logger.go#L95), allows one gRPC Log Client to connect to multiple gRPC Log Servers at the same time, writing the same events to different endpoints (as needed).

There is also a careful verification if the input [`io.Writer` interface](https://pkg.go.dev/io#Writer) is actually of [`ConnAddr` type](./grpc/address/address.go#L8), for example in the [client's `SetOuts()` method](./grpc/client/client.go#L643):

```go
func (c *GRPCLogClient) SetOuts(outs ...io.Writer) log.Logger {
	// (...)
	for _, remote := range outs {
		// ensure the input writer is not nil
		if remote == nil {
			continue
		}

		// ensure the input writer is of type *address.ConnAddr
		// if not, skip this writer and register this event
		if r, ok := remote.(*address.ConnAddr); !ok {		
			// (...)
		} else {
			o = append(o, r.Keys()...)
		}
	}
	// (...)
}
```

The core of the [`ConnAddr` type](./grpc/address/address.go#L8) stores addresses prior to _verifying_ them. As the gRPC Log Client starts running, it will iterate through all addresses in the map and connect to them (thus the map of strings and pointers to [`grpc.ClientConn`](https://pkg.go.dev/google.golang.org/grpc#ClientConn)).

A [`ConnAddr`](./grpc/address/address.go#L8) is initialized with any number of parameters (at least one, otherwise it returns nil):

```go
// import "github.com/zalgonoise/zlog/grpc/address"

a := address.New(
	"localhost:9099",
	"mycoolserver.io:9099",
	// more addresses
)
```

This type exposes a few methods that may be useful; although keep in mind that all of this logic is embeded in the gRPC client and server implementations already (in their `WithAddr()` configs and writer-related methods like `AddOuts()` and `SetOuts()`):

Method | Description
:--:|:--:
[`AsMap() map[string]*grpc.ClientConn`](./grpc/address/address.go#L39) | returns a [`ConnAddr` type](./grpc/address/address.go#L8) object in a `map[string]`[`*grpc.ClientConn`](https://pkg.go.dev/google.golang.org/grpc#ClientConn) format
[`Add(...string)`](./grpc/address/address.go#L45) | allocates the input strings as entries in the map, with initialized pointers to [`grpc.ClientConn`](https://pkg.go.dev/google.golang.org/grpc#ClientConn)
[`Keys() []string`](./grpc/address/address.go#L66) | returns a [`ConnAddr` type](./grpc/address/address.go#L8) object's keys (its addresses) in a slice of strings
[`Get(string) *grpc.ClientConn`](./grpc/address/address.go#L76) | returns the pointer to a [`grpc.ClientConn`](https://pkg.go.dev/google.golang.org/grpc#ClientConn), as referenced in the input address `k`
[`Set(string, *grpc.ClientConn)`](./grpc/address/address.go#L86) | allocates the input connection to the input string, within the [`ConnAddr` type](./grpc/address/address.go#L8) object (overwritting it if existing)
[`Len() int`](./grpc/address/address.go#L94) | returns the size of the [`ConnAddr` type](./grpc/address/address.go#L8) object
[`Reset()`](./grpc/address/address.go#L99) | overwrites the existing [`ConnAddr` type](./grpc/address/address.go#L8) object with a new, empty one.
[`Unset(...string)`](./grpc/address/address.go#L110) | removes the input addr strings from the [`ConnAddr` type](./grpc/address/address.go#L8) object, if existing
[`Write(p []byte) (n int, err error)`](./grpc/address/address.go#L134) | an implementation of [`io.Writer` interface](https://pkg.go.dev/io#Writer), so that the [`ConnAddr` type](./grpc/address/address.go#L8) object can be used in a gRPC Log Client's [`SetOuts()`](./grpc/client/client.go#L643) and [`AddOuts()`](./grpc/client/client.go#L703) methods. These need to conform with the [`Logger` interface](./log/logger.go#L95) that implements the same methods. For the same layer of compatibility to be possible in a gRPC Log Client (who will write its log entries in a remote server), it uses these methods to implement its way of altering the existing connections, instead of dismissing this part of the implementation all together. __This is not a regular [`io.Writer` interface](https://pkg.go.dev/io#Writer)__.


_______________

### Integration

This section will cover general procedures when performing changes or general concerns when integrating a certain package or module to work with this logger. Generally there are interfaces defined so you can adapt the feature to your own needs (like the formatters); but may not be as clear as day for other parts of the code base.

#### Protobuf code generation

When making changes to the `.proto` files within the [`proto/` directory](./proto/), such as when adding new log levels or a new element to the Event data structure, you will need to run the appropriate code generation command.

The commands referenced below refer to [Go's GitHub repo for `protoc-gen-go`](https://github.com/golang/protobuf/protoc-gen-go), and not the one from the instructions in the [gRPC Quick Start guide](https://grpc.io/docs/languages/go/quickstart/). It is important to review this document; but the decision to go with Go's `protoc` release is only due to compatibility with (built-in) gRPC code generation. This may change in the future, in adoption of Google's official release versions.

Then, you can regenerate the code by running a command appropriate to your target:

Target | Command
:--:|:--:
[`proto/event.proto`](./proto/event.proto)|`protoc  --go_out=.  proto/event.proto`
[`proto/service.proto`](./proto/event.proto)|`protoc --go_out=plugins=grpc:proto proto/service.proto`

<details>

_Example of editing the `proto/event.proto` file_

1. Make changes to [the `.proto` file](./proto/event.proto), for example changing the default prefix or adding a new field:

```proto
message Event {
    optional google.protobuf.Timestamp time = 1;
    optional string prefix = 2 [ default = "system" ];
    optional string sub = 3;
    optional Level level = 4 [ default = info ];
    required string msg = 5;
    optional google.protobuf.Struct meta = 6;
	required bool breaking = 7;
}
```

2. Execute the appropriate `protoc` command to regenerate the code:

```shell
protoc  --go_out=.  proto/event.proto
```

3. Review changes in the diff. These changes would be reflected in the target `.pb.go` file, in [`log/event/event.pb.go`](./log/event/event.pb.go):


```go
// (...)
const (
	Default_Event_Prefix = string("system")
	Default_Event_Level  = Level_info
)
// (...)
func (x *Event) GetBreaking() bool {
	if x != nil && x.Breaking != nil {
		return *x.Breaking
	}
	return false
}
```

4. Use the new logic within the context of your application

```go
// (...)
if event.GetBreaking() {
	// handle the special type of event
}
```

</details>

> **NOTE**: When performing changes to [`proto/service.proto`](./proto/service.proto) and regenerating the code, you will notice that the [`proto/service/service.pb.go` file](./proto/service/service.pb.go) will have an invalid import for `event`.
>
> To correct this, simply change this import from:
>
>     event "./log/event"
>
> ...to:
>
>     event "github.com/zalgonoise/zlog/log/event"
>
> For your convenience, [the shell script in [`proto/gen_pbgo.sh`](./proto/gen_pbgo.sh) is a simple wrapper for these actions.

___________


#### Adding your own configuration settings

Most of the modules will be configured with a function that takes a variadic parameter for its configuration, which is an interface -- meaning that you can implement it too. This section will cover an example of both adding a new formatter and a new Logger configuration.


_Example of adding a new Formatter_

<details>

1. Inspect the flow of configuring a formatter for a Logger. This is done in [`log/format.go`](./log/format.go) by calling the [`WithFormat(LogFormatter)`](./log/format.go#L33) function, which returns a `LoggerConfig`. This means that to add a new formatter, a new [`LogFormatter` type](./log/format.go#L18) needs to be created, so it can be passed into this function when creating a Logger.

2. Create your own type and implement the `Format(*event.Event) ([]byte, error)` method. This is an example with a compact text formatter:

```go
package compact

import (
	"strings"
	"time"

	"github.com/zalgonoise/zlog/log/event"
)

type Text struct {}

func (Text) Format(log *event.Event) (buf []byte, err error) {
	// uses hypen as separator, no square brackets, simple info only
	var sb strings.Builder

	sb.WriteString(log.GetTime().AsTime().Format(time.RFC3339))
	sb.WriteString(" - ")
	sb.WriteString(log.GetLevel().String())
	sb.WriteString(" - ")
	sb.WriteString(log.GetMsg())
	sb.WriteString("\n")

	buf = []byte(sb.String())
	return
}
```

3. When configuring the Logger in your application, import your package and use it to configure the Logger:

```go
package main 

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"

	"github.com/myuser/myrepo/mypkg/compact" // your package here
)

func main() {
	logger := log.New(
		log.WithFormat(compact.Text{})	
	)

	logger.Info("hi from custom formatter!")
}
```

4. Test it out:

```
2022-08-06T18:02:34Z - info - hi from custom formatter!
```

</details>


_Example of adding a new LoggerConfig_

<details>

1. Inspect the flow of configuring a Logger. This is done in [`log/conf.go`](./log/conf.go) by calling a function implementing the [`LoggerConfig` interface](./log/conf.go#L21), which applies your changes (from the builder object) to the resulting object. This means that to add a new configuration, a new [`LoggerConfig` type](./log/conf.go#21) needs to be created, so it can be passed as a parameter when creating the Logger.

2. Create your own type and implement the `Apply(lb *LoggerBuilder)` method, and optionally a helper function. This is an example with a combo of prefix and subprefix config:

```go
package id

import (
	"github.com/zalgonoise/zlog/log"
)

type id struct {
	prefix string
	sub    string
}

func (i *id) Apply(lb *log.LoggerBuilder) {
	// apply settings to prefix and subprefix
	lb.Prefix = i.prefix
	lb.Sub = i.sub
}

func WithID(prefix, sub string) log.LoggerConfig {
	return &id{
		prefix: prefix,
		sub:    sub,
	}
}

```

3. When configuring the Logger in your application, import your package and use it to configure the Logger:

```go
package main 

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"

	"github.com/myuser/myrepo/mypkg/id" // your package here
)

func main() {
	logger := log.New(
		id.WithID("http", "handlers")
	)

	logger.Info("hi from custom formatter logger!")
}
```

4. Test it out:

```
[info]  [2022-08-06T18:25:07.581200505Z]        [http]  [handlers]      hi from custom formatter logger!
```

</details>

______________

#### Adding methods to a Builder pattern

Currently, the [`EventBuilder`](./log/event/builder.go#L14) is the only data structure with  a _true_ builder pattern. This is to allow a message to be modified on each particular field independently, while wrapping up all of these modifications with the final [`Build()`](./log/event/builder.go#L107) method, that spits out the actual [`Event` object](./log/event/event.pb.go#96).

The methods in the [`EventBuilder`](./log/event/builder.go#L14) are chained together and closed with the [`Build()`](./log/event/builder.go#L107) call. This means that new methods can be added to the [`EventBuilder`](./log/event/builder.go#L14); to be chained-in on creation.

> **NOTE**: this is a very specific case where you're creating a wrapper for the EventBuilder, so that you don't have to define all the existing methods again. This particular module (the log events) are based on the autogenerated data structure from the corresponding `.proto` file. 


_Example of adding a new method to the Event building process_

<details>

1. Inspect the flow of creating an Event. This is done in [`log/event/builder.go`](./log/event/builder.go) by calling a function to initialize the builder object, and then chaining any number of methods (at least `Message()`) until it is closed with the `Build()` call, that outputs an event (with timestamp). This is a very strict and linear pattern but you're able to create your own custom chains by wrapping the builder object. The only concern is to at a certain point returning the (original) [`EventBuilder` type](./log/event/builder.go#L14) to gain access to its helper methods.

2. Create your own builder type and implement the chaining methods you need, and optionally a helper function to initialize this object. The new builder type will only be a wrapper for the original one. This is an example with a combo of prefix and subprefix in a single method:

```go
package newbuilder

import "github.com/zalgonoise/zlog/log/event"

type NewEventBuilder struct {
	*event.EventBuilder
}

func NewB() *NewEventBuilder { // initializing defaults just like the original New()
	var (
		prefix   string = event.Default_Event_Prefix
		sub      string
		level    event.Level = event.Default_Event_Level
		metadata map[string]interface{}
	)

	return &NewEventBuilder{
		EventBuilder: &event.EventBuilder{
			BPrefix:   &prefix,
			BSub:      &sub,
			BLevel:    &level,
			BMetadata: &metadata,
		},
	}
}

func (e *NewEventBuilder) PrefixSub(prefix, sub string) *event.EventBuilder {
	*e.BPrefix = prefix
	*e.BSub = sub

	return e.EventBuilder
}
```

3. When configuring the Logger in your application, import your package and use it to configure the Logger:

```go
package main

import (
	"github.com/zalgonoise/zlog/log"

	"github.com/myuser/myrepo/mypkg/newbuilder" // your package here
)

func main() {
	event := newbuilder.NewB(). // initialize new builder
				PrefixSub("http", "handlers"). // use custom methods, returning original builder
				Message("hi from a custom event message!"). // use its methods
				Build() // output a valid message to log

	log.Log(event)
}
```

4. Test it out:

```
[info]  [2022-08-07T18:53:16.683574327Z]        [http]  [handlers]      hi from a custom event message!
```

</details>


______________

#### Adding interceptors to gRPC server / client

gRPC allows very granular middleware to be added to either client or server implementations, to suit your most peculiar needs. This is simply a means to execute a function before (or after) a message is sent or received by either end.

To add multiple interceptors, this library uses the [`go-grpc-middleware`](https://github.com/grpc-ecosystem/go-grpc-middleware) library, which allows easy chaining of multiple interceptors (not being limited to just one). This library is very interesting and provides a number of useful examples that could suit your logger's needs -- and basically it is an interceptor capable of registering (and calling) other interceptors you need to configure.


_Example of adding a simple server interceptor to add a UUID_

<details>

1. Inspect the flow of an existing interceptor, such as [the server logging interceptor](./grpc/server/logging.go). This consists of implementing two wrapper functions that return other functions, with the following signature:

Purpose | Signature
:--:|:--:
Unary RPCs | `func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)`
Stream RPCs | `func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error`

These functions are executed on the corresponding action, take an example of a stripped version (no actions taken) of these functions:

__Unary RPC -- stripped__

```go
func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := handler(ctx, req)

	return res, err
}
```

__Stream RPC -- stripped__

```go
func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, wStream)

	return err
}
```

> __NOTE__: while the approach for Unary RPCs is usually very simple (one message, one interceptor call), the stream RPCs are persistent connections. This means that if you want to handle the underlying messages being exchanged, you need to implement a wrapper for the stream, as shown below in the context of this example.

2. Create your own interceptor(s) (for your use-case's type of RPCs) by wrapping the functions in the table above. This example will add UUIDs to log events as the server replies back to the client. This is already implemented by default but serves as an example. Both Unary and Stream RPC interceptors are implemented, note how the Stream wraps up the stream handler:

```go

// UnaryServerIDModifier sets a preset UUID to log responses
func UnaryServerIDModifier(id string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// regular flow
		res, err := handler(ctx, req)
		
		// modify response's ID
		res.(*pb.LogResponse).ReqID = id

		// continue as normal
		return res, err
	}
}

// StreamServerIDModifier sets a preset UUID to log responses
func StreamServerIDModifier(id string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// wrap the stream 
		wStream := modIDStream{
			id: id,
		}

		// handle the messages using the wrapped stream
		err := handler(srv, wStream)
		return err
	}
}

// modIDStream is a wrapper for grpc.ServerStream, that will replace
// the response's ID with the preset `id` string value
type modIDStream struct {
	id string
}


// Header method is a wrapper for the grpc.ServerStream.Header() method
func (w loggingStream) SetHeader(m metadata.MD) error { return w.stream.SetHeader(m) }

// Trailer method is a wrapper for the grpc.ServerStream.Trailer() method
func (w loggingStream) SendHeader(m metadata.MD) error { return w.stream.SendHeader(m) }

// CloseSend method is a wrapper for the grpc.ServerStream.CloseSend() method
func (w loggingStream) SetTrailer(m metadata.MD) { w.stream.SetTrailer(m) }

// Context method is a wrapper for the grpc.ServerStream.Context() method
func (w loggingStream) Context() context.Context { return w.stream.Context() }

// SendMsg method is a wrapper for the grpc.ServerStream.SendMsg(m) method, for which the
// configured logger will set a predefined ID
func (w loggingStream) SendMsg(m interface{}) error {
	// set the custom ID 
	m.(*pb.LogResponse).ReqID = w.id

	// continue as normal
	return w.stream.SendMsg(m)
}

// RecvMsg method is a wrapper for the grpc.ServerStream.RecvMsg(m) method
func (w loggingStream) RecvMsg(m interface{}) error { return w.stream.RecvMsg(m) }
```

3. When configuring the Logger in your application, import your package and use it to configure the Logger:

```go
package main

import (
	"github.com/zalgonoise/zlog/log/grpc/server"

	"github.com/myuser/myrepo/mypkg/fixedid" // your package here
)

func main() {
	grpcLogger := server.New(
		server.WithAddr("example.com:9099"),
		server.WithGRPCOpts(),
		server.WithStreamInterceptor(
			"fixed-id",
			fixedid.StreamServerIDModifier("staging")
		)
	)
	grpcLogger.Serve()
}
```

4. Launch a client, and send some messages back to the server. The response in the client should contain the fixed ID in its metadata.

</details>

_______________


### Benchmarks

Tests for speed and performance are done with benchmark tests, where different approaches to the many loggers is measured so it's clear where the library excels and lacks. This is done with multiple configs of individual features and as well a comparison with other Go loggers.

Benchmark tests are added and posted under the [`/benchmark` directory](./benchmark/), to ensure that its imports is separate from the rest of the library. If by any means this approach still makes it _bloated_, it can be moved to an entirely separate branch, at any point.

As changes are made, the benchmarks' raw results are posted in [the `/benchmark/README.md` file](./benchmark/README.md).

______________

### Contributing

To contribute to this repository, open an issue in GitHub's Issues section to kick off a discussion with your proposal. It's best that proposals are complete with a (brief) description, current and expected behavior, proposed solution (if any), as well as your view on the impact it causes. In time there will be a preset for both bugs, feature requests, and anything in between.
