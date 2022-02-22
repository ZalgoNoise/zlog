package log

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
)

var mockLogfileOps = map[string]interface{}{
	"init": func() string {
		path, err := os.MkdirTemp("/tmp/", "zlog-test-*")
		if err != nil {
			panic(err)
		}
		return path
	},
	"createFiles": func(path string) []string {
		var files []string

		for i := 0; i < 5; i++ {
			new := fmt.Sprintf("%s/%s-%v", path, "buffer", i)
			_, err := os.Create(new)
			if err != nil {
				panic(err)
			}
			files = append(files, new)
		}
		return files
	},
	"createOversize": func(path string) error {
		b := []byte("this base string will be 64 characters in length as multiplier!\n")

		f, err := os.Create(path)
		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}

		// go for >50mb
		for i := 0; i < 850000; i++ {
			n, err := buf.Write(b)
			if err != nil {
				return err
			}
			if n != 64 {
				return errors.New("write did not satisfy input buffer length")
			}

		}

		_, err = f.Write(buf.Bytes())

		if err != nil {
			return err
		}

		return nil
	},
}

func TestNewLogFile(t *testing.T) {

	var mockPaths = []string{
		"/test-buffer",
		"/test-file",
		"/test-logger",
	}

	var mockInvalidPaths = []string{
		"//test",
		"/var/log/private/logfile",
		"/var/log/private/",
	}

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	f, err := os.Create(targetDir + "/no-permissions")
	if err != nil {
		panic(err)
	}
	f.Close()
	if err := os.Chmod(targetDir+"/no-permissions", 0100); err != nil {
		panic(err)
	}

	fileMap := mockLogfileOps["createFiles"].(func(path string) []string)(targetDir)
	err = mockLogfileOps["createOversize"].(func(string) error)(targetDir + "/oversize")
	if err != nil {
		panic(err)
	}

	var allPaths []string
	allPaths = append(allPaths, fileMap...)
	for _, v := range mockPaths {
		allPaths = append(allPaths, targetDir+v)
	}
	allPaths = append(allPaths, targetDir+"/oversize")

	type test struct {
		path string
		ok   bool
	}

	var tests []test

	for a := 0; a < len(allPaths); a++ {
		tests = append(tests, test{
			path: allPaths[a],
			ok:   true,
		})
	}
	for a := 0; a < len(mockInvalidPaths); a++ {
		tests = append(tests, test{
			path: mockInvalidPaths[a],
			ok:   false,
		})
	}
	tests = append(tests, test{
		path: targetDir + "/no-permissions",
		ok:   false,
	})

	var verify = func(id int, test test) {
		_, err := NewLogfile(test.path)
		if err != nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] NewLogfile(%s) -- operation failed with an error: %s",
				id,
				test.path,
				err,
			)
			return
		}
		t.Logf(
			"#%v -- PASSED -- [Logfile] NewLogfile(%s)",
			id,
			test.path,
		)
	}

	for id, test := range tests {
		verify(id, test)
	}

}

func TestLogfileMaxSize(t *testing.T) {}

func TestLogfileNew(t *testing.T) {}

func TestLogfileLoad(t *testing.T) {}

func TestLogfileMove(t *testing.T) {}

func TestLogfileHasExt(t *testing.T) {}

func TestLogfileAddExt(t *testing.T) {}

func TestLogfileCutExt(t *testing.T) {}

func TestLogfileSize(t *testing.T) {}

func TestLogfileIsTooHeavy(t *testing.T) {}

func TestLogfileRotate(t *testing.T) {}

func TestLogfileWrite(t *testing.T) {}
