package sqlite

import (
	"os"
	"regexp"
	"testing"

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

func TestNew(t *testing.T) {
	module := "SQLite"
	funcname := "New()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		envPath string
		ok      bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       SQLITE_PATH={path to an existing sqlite DB} \
	//       SQLITE_NEW_PATH={path to new sqlite DB} \
	//       SQLITE_INVALID_PATH=/log_invalid.db
	//
	var tests = []test{
		{
			name:    "working database",
			envPath: "SQLITE_PATH",
			ok:      true,
		},
		{
			name:    "new database",
			envPath: "SQLITE_NEW_PATH",
			ok:      true,
		},
		{
			name:    "new database",
			envPath: "SQLITE_INVALID_PATH",
			ok:      false,
		},
	}

	var verify = func(idx int, test test) {

		p, ok := getEnv(test.envPath)

		if !ok {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] no %s env defined; skipping test as no DB will be used -- action: %s",
				idx,
				module,
				funcname,
				test.envPath,
				test.name,
			)
			return
		}

		_, err := New(p)

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestCreate(t *testing.T) {
	module := "SQLite"
	funcname := "Create()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		envPath string
		e       []*event.Event
		ok      bool
	}

	// paths are fetched with os.Getenv(); you may need to export this variable
	// before executing the tests:
	//
	//     export SQLITE_PATH={path to an existing sqlite DB}
	//
	var tests = []test{
		{
			name:    "create a single event",
			envPath: "SQLITE_PATH",
			e: []*event.Event{
				event.New().Message("null").Build(),
			},
			ok: true,
		},
		{
			name:    "create multiple events",
			envPath: "SQLITE_PATH",
			e: []*event.Event{
				event.New().Message("null_0").Build(),
				event.New().Message("null_1").Build(),
				event.New().Message("null_2").Build(),
			},
			ok: true,
		},
		{
			name:    "create with empty list of events",
			envPath: "SQLITE_PATH",
			e:       []*event.Event{},
			ok:      true,
		},
		{
			name:    "create with nil events",
			envPath: "SQLITE_PATH",
			e:       nil,
			ok:      true,
		},
		{
			name:    "create multiple events with some nil messages",
			envPath: "SQLITE_PATH",
			e: []*event.Event{
				event.New().Message("null_0").Build(),
				nil,
				nil,
			},
			ok: true,
		},
		{
			name:    "create multiple events with all nil messages",
			envPath: "SQLITE_PATH",
			e:       []*event.Event{nil, nil, nil},
			ok:      true,
		},
		{
			name:    "create invalid / empty events",
			envPath: "SQLITE_PATH",
			e:       []*event.Event{{}, {}, {}},
			ok:      false,
		},
	}

	var verify = func(idx int, test test) {

		p, ok := getEnv(test.envPath)

		if !ok {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] no %s env defined; skipping test as no DB will be used -- action: %s",
				idx,
				module,
				funcname,
				test.envPath,
				test.name,
			)
			return
		}

		db, err := New(p)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		sql, ok := db.(*SQLite)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] DB creation did not return a SQLite pointer -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		if test.e == nil {
			err = sql.Create(nil)
		} else {
			err = sql.Create(test.e...)
		}

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWrite(t *testing.T) {
	module := "SQLite"
	funcname := "Write()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		envPath string
		e       []byte
		ok      bool
	}

	// paths are fetched with os.Getenv(); you may need to export this variable
	// before executing the tests:
	//
	//     export SQLITE_PATH={path to an existing sqlite DB} \
	//       SQLITE_INVALID_PATH=/log_invalid.db
	//
	var tests = []test{
		{
			name:    "write a single event",
			envPath: "SQLITE_PATH",
			e:       event.New().Message("null").Build().Encode(),
			ok:      true,
		},
		{
			name:    "write invalid event",
			envPath: "SQLITE_PATH",
			e:       []byte("null"),
			ok:      false,
		},
		{
			name:    "write nil event",
			envPath: "SQLITE_PATH",
			e:       nil,
			ok:      false,
		},
		{
			name:    "new database",
			envPath: "SQLITE_INVALID_PATH",
			e:       event.New().Message("null").Build().Encode(),
			ok:      false,
		},
	}

	var verify = func(idx int, test test) {

		p, ok := getEnv(test.envPath)

		if !ok {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] no %s env defined; skipping test as no DB will be used -- action: %s",
				idx,
				module,
				funcname,
				test.envPath,
				test.name,
			)
			return
		}

		db, err := New(p)

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		} else if !test.ok {
			return
		}

		if test.e == nil {
			_, err = db.Write(nil)
		} else {
			_, err = db.Write(test.e)
		}

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

	}

	var verifyNewFromPath = func(idx int, test test) {

		p, ok := getEnv(test.envPath)

		if !ok {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] no %s env defined; skipping test as no DB will be used -- action: %s",
				idx,
				module,
				funcname,
				test.envPath,
				test.name,
			)
			return
		}

		db := &SQLite{
			path: p,
		}

		var err error

		if test.e == nil {
			_, err = db.Write(nil)
		} else {
			_, err = db.Write(test.e)
		}

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

	}

	var verifyInvalidMessage = func(idx int, test test) {

		p, ok := getEnv(test.envPath)

		if !ok {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] no %s env defined; skipping test as no DB will be used -- action: %s",
				idx,
				module,
				funcname,
				test.envPath,
				test.name,
			)
			return
		}

		db := &SQLite{
			path: p,
		}

		var err error

		empty := ""
		buf := test.e
		e, _ := event.Decode(buf)

		e.Msg = &empty

		_, err = db.Write(e.Encode())

		if err == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] expected an error; didn't get one -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

	// verify new from path
	for idx, test := range tests {
		verifyNewFromPath(idx, test)
	}

	// verify invalid message parsing
	verifyInvalidMessage(0, tests[0])

}

func TestClose(t *testing.T) {
	module := "SQLite"
	funcname := "Close()"

	_ = module
	_ = funcname

	type test struct {
	}

	// paths are fetched with os.Getenv(); you may need to export this variable
	// before executing the tests:
	//
	//     export SQLITE_PATH={path to an existing sqlite DB}
	//
	var tests = []test{
		{},
	}

	var verify = func(idx int, test test) {

		db := &SQLite{}

		err := db.Close()

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v",
				idx,
				module,
				funcname,
				err,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestWithSQLite(t *testing.T) {
	module := "SQLite"
	funcname := "WithSQLite()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		envPath string
		ok      bool
	}

	// paths are fetched with os.Getenv(); you may need to export this variable
	// before executing the tests:
	//
	//     export SQLITE_PATH={path to an existing sqlite DB} \
	//       SQLITE_INVALID_PATH=/log_invalid.db
	//
	var tests = []test{
		{
			name:    "write a single event",
			envPath: "SQLITE_PATH",
			ok:      true,
		},
		{
			name:    "write a single event",
			envPath: "SQLITE_INVALID_PATH",
			ok:      false,
		},
	}

	var catchPanic = func(idx int, test test) {
		r := recover()

		if r == nil {
			return
		}

		regexStr := `^failed to create logger config -- database creation failed:.+`
		regex := regexp.MustCompile(regexStr)

		err, ok := r.(error)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] panic value is not an error: %v -- %T -- action: %s",
				idx,
				module,
				funcname,
				r,
				r,
				test.name,
			)
			return
		}

		if !regex.MatchString(err.Error()) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error mismatch: wanted to match %s; got %v -- action: %s",
				idx,
				module,
				funcname,
				regexStr,
				err.Error(),
				test.name,
			)
			return
		}
	}

	var verify = func(idx int, test test) {
		defer catchPanic(idx, test)

		p, ok := getEnv(test.envPath)

		if !ok {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] no %s env defined; skipping test as no DB will be used -- action: %s",
				idx,
				module,
				funcname,
				test.envPath,
				test.name,
			)
			return
		}

		c := WithSQLite(p)

		cfg, ok := c.(*log.LCDatabase)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] operation didn't return expected type: %T -- action: %s",
				idx,
				module,
				funcname,
				cfg,
				test.name,
			)
			return

		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}
