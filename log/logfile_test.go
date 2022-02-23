package log

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
)

var underweightIters = []int{
	20000,
	50000,
	100000,
	300000,
	500000,
	800000,
}

var overweightIters = []int{
	850000,
	1000000,
	900000,
}

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
	"createOversize": func(path string, size int) error {
		b := []byte("this base string will be 64 characters in length as multiplier!\n")

		f, err := os.Create(path)
		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}

		// go for >50mb
		for i := 0; i < size; i++ {
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
	err = mockLogfileOps["createOversize"].(func(string, int) error)(targetDir+"/oversize", overweightIters[0])
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

func TestLogfileMaxSize(t *testing.T) {
	type test struct {
		path string
		iter int
		size int
	}

	targetDir := mockLogfileOps["init"].(func() string)()
	// defer func() { os.RemoveAll(targetDir) }()

	var tests []test

	for a := 0; a < len(underweightIters); a++ {
		newfile := fmt.Sprintf("%s%s-%v", targetDir, "/undersize", a)
		err := mockLogfileOps["createOversize"].(func(string, int) error)(newfile, underweightIters[a])
		if err != nil {
			panic(err)
		}

		obj := test{
			path: newfile,
			iter: underweightIters[a],
			size: 1,
		}

		tests = append(tests, obj)
	}

	// for a := 0; a < len(overweightIters); a++ {
	// 	newfile := fmt.Sprintf("%s%s-%v", targetDir, "/oversize", a)
	// 	err := mockLogfileOps["createOversize"].(func(string, int) error)(newfile, overweightIters[a])
	// 	if err != nil {

	// 		return
	// 	}
	// 	obj := test{
	// 		path:    newfile,
	// 		iter:    overweightIters[a],
	// 		size:    25,
	// 		rotates: true,
	// 	}

	// 	tests = append(tests, obj)
	// }

	var verify = func(id int, test test, logf *Logfile) {
		isHeavyBefore := logf.IsTooHeavy()

		if isHeavyBefore {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.MaxSize(%v) -- item cannot be declared as too heavy prior to setting a new size",
				id,
				test.size,
			)
			return
		}

		logf.MaxSize(test.size)

		isHeavyAfter := logf.IsTooHeavy()

		if !isHeavyAfter {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.MaxSize(%v) -- item should be heavier than limit after the change",
				id,
				test.size,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.MaxSize(%v)",
			id,
			test.size,
		)

	}

	for id, test := range tests {
		t.Logf(
			"#%v -- [Logfile] Logfile.MaxSize(%v) on %s",
			id,
			test.size,
			test.path,
		)
		logfile, err := NewLogfile(test.path)
		if err != nil {
			panic(err)
		}
		verify(id, test, logfile)

	}

}

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
