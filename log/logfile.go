package log

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	bytes int64 = 1 << (10 * iota)
	kilobytes
	megabytes
)

type Logfile struct {
	path   string
	file   *os.File
	rotate int
}

func NewLogfile(path string) (*Logfile, error) {
	f := &Logfile{rotate: 50}

	_, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return f.new(path)
		} else if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied -- %s", err)
		} else {
			return nil, fmt.Errorf("unknown error -- %s", err)
		}
	} else {
		err := f.load(path)
		if err != nil {
			return nil, err
		}
		err = f.Rotate()
		if err != nil {
			return nil, err
		}
		return f, nil
	}

}

func (f *Logfile) MaxSize(mb int) *Logfile {
	f.rotate = mb
	return f
}

func (f *Logfile) new(path string) (*Logfile, error) {
	if !f.hasExt(path) {
		path = f.addExt(path)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	f.path = path
	f.file = file
	return f, nil
}

func (f *Logfile) load(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	f.file = file
	f.path = path

	return nil
}

func (f *Logfile) move(path string) error {
	now := time.Now().Format("2006-01-02-15-04-05")
	var p string
	if f.hasExt(path) {
		p = f.cutExt(path)
		p += "_" + now
		p = f.addExt(p)
	} else {
		p = path + "_" + now
		p = f.addExt(p)
	}
	return os.Rename(path, p)

}

func (f *Logfile) hasExt(path string) (ok bool) {
	ok = false
	// fast check - check last 4 chars for ".log"
	sub := path[:4]
	if sub == ".log" {
		ok = true
		return
	}

	// slow check - split ".", check last slice
	subslice := strings.Split(path, ".")
	if subslice[len(subslice)-1] == "log" {
		ok = true
		return
	}
	return
}

func (f *Logfile) addExt(path string) string {
	return path + ".log"
}

func (f *Logfile) cutExt(path string) string {
	return path[:len(path)-4]
}

func (f *Logfile) Size() (int64, error) {
	s, err := f.file.Stat()
	if err != nil {
		return -1, err
	}
	return s.Size(), nil
}

func (f *Logfile) IsTooHeavy() bool {
	s, err := f.Size()

	if err != nil {
		// errored; will rotate
		return true
	}

	if s > int64(f.rotate)*megabytes {
		return true
	}
	return false
}

func (f *Logfile) Rotate() error {
	if f.IsTooHeavy() {
		err := f.move(f.path)
		if err != nil {
			return err
		}
		newF, err := f.new(f.path)
		if err != nil {
			return err
		}
		f.file = newF.file
	}
	return nil
}

func (f *Logfile) Write(b []byte) (n int, err error) {
	return f.file.Write(b)
}
