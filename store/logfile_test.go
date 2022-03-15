package store

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

var mockByteChunk64 = []byte(
	"this base string will be 64 characters in length as multiplier!\n",
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

var newLogfiles = []string{
	"/new-logfile.log",
	"/logfile.log",
	"/test-logfile.log",
	"/service.log",
	"/new.log",
}

var newValidExts = []string{
	"/new-logfile.textlog",
	"/logfile.jsonlog",
	"/test-logfile.tmplog",
	"/service.srvlog",
	"/new.svclog",
}

var newExtlessFiles = []string{
	"/extensionless-new-logfile",
	"/extensionless-logfile",
	"/extensionless-test-logfile",
	"/extensionless-service",
	"/extensionless-new",
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
	"createLogfiles": func(path string) []string {
		var files []string

		for i := 0; i < 5; i++ {
			new := fmt.Sprintf("%s/%s-%v.log", path, "buffer", i)
			_, err := os.Create(new)
			if err != nil {
				panic(err)
			}
			files = append(files, new)
		}
		return files
	},
	"createNamedFiles": func(path string, files ...string) []string {
		var fileList []string

		for _, k := range files {
			new := fmt.Sprintf("%s%s", path, k)
			_, err := os.Create(new)
			if err != nil {
				panic(err)
			}
			fileList = append(fileList, new)
		}
		return fileList
	},
	"createOversize": func(path string, size int) error {
		b := mockByteChunk64

		f, err := os.Create(path)
		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}

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
	defer func() { os.RemoveAll(targetDir) }()

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

func TestLogfileNew(t *testing.T) {
	type test struct {
		path string
		ext  bool
	}

	var tests []test

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	for a := 0; a < len(newLogfiles); a++ {
		tests = append(tests, test{
			path: targetDir + newLogfiles[a],
			ext:  true,
		})
	}

	for a := 0; a < len(newExtlessFiles); a++ {
		tests = append(tests, test{
			path: targetDir + newExtlessFiles[a],
			ext:  false,
		})
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}

		lf, err := f.new(test.path)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.new() -- failed to create logfile with an error: %s",
				id,
				err,
			)
			return
		}

		lf.file.Close()

		var statErr error

		if !test.ext {
			_, statErr = os.Stat(test.path + ".log")
		} else {
			_, statErr = os.Stat(test.path)
		}

		if statErr != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.new() -- failed to stat logfile with an error: %s",
				id,
				err,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.new()",
			id,
		)
	}

	for id, test := range tests {
		verify(id, test)

	}

}

func TestLogfileLoad(t *testing.T) {

	type test struct {
		path string
		ext  bool
	}

	var tests []test

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	for a := 0; a < len(newLogfiles); a++ {
		tests = append(tests, test{
			path: targetDir + newLogfiles[a],
			ext:  true,
		})
	}

	for a := 0; a < len(newExtlessFiles); a++ {
		tests = append(tests, test{
			path: targetDir + newExtlessFiles[a],
			ext:  false,
		})
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}

		lf, err := f.new(test.path)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.new() -- failed to create logfile with an error: %s",
				id,
				err,
			)
			return
		}

		lf.file.Close()

		var statErr error

		if !test.ext {
			_, statErr = os.Stat(test.path + ".log")
		} else {
			_, statErr = os.Stat(test.path)
		}

		if statErr != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.new() -- failed to stat logfile with an error: %s",
				id,
				err,
			)
			return
		}

		nf := &Logfile{rotate: 50}

		var loadErr error

		if !test.ext {
			loadErr = nf.load(test.path + ".log")
		} else {
			loadErr = nf.load(test.path)
		}

		if loadErr != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.load() -- failed to load logfile with an error: %s",
				id,
				loadErr,
			)
			return
		}

		if !test.ext {
			if nf.path != test.path+".log" {
				t.Errorf(
					"#%v -- FAILED -- [Logfile] Logfile.load() -- file path mismatch: wanted %s ; got %s",
					id,
					nf.path,
					test.path+".log",
				)
				return
			}
		} else {
			if nf.path != test.path {
				t.Errorf(
					"#%v -- FAILED -- [Logfile] Logfile.load() -- file path mismatch: wanted %s ; got %s",
					id,
					nf.path,
					test.path,
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.load()",
			id,
		)
	}

	for id, test := range tests {
		verify(id, test)

	}
}

func TestLogfileMove(t *testing.T) {
	type test struct {
		path string
		rgx  *regexp.Regexp
	}

	var tests []test

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	for a := 0; a < len(newLogfiles); a++ {
		tests = append(tests, test{
			path: targetDir + newLogfiles[a],
			rgx:  regexp.MustCompile(`^(.*)_\d{4}(-\d{2}){5}.log$`),
		})
	}

	for a := 0; a < len(newExtlessFiles); a++ {
		tests = append(tests, test{
			path: targetDir + newExtlessFiles[a] + ".log",
			rgx:  regexp.MustCompile(`^(.*)_\d{4}(-\d{2}){5}.log$`),
		})
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}

		lf, err := f.new(test.path)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.new() -- failed to create logfile with an error: %s",
				id,
				err,
			)
			return
		}

		lf.file.Close()

		_, err = os.Stat(test.path)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.new() -- failed to stat logfile with an error: %s",
				id,
				err,
			)
			return
		}

		err = lf.move(test.path)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.move() -- failed to move logfile with an error: %s",
				id,
				err,
			)
			return
		}

		files, err := ioutil.ReadDir(targetDir)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.move() -- failed to list files in test dir with an error: %s",
				id,
				err,
			)
			return
		}

		var passed bool = false

		for _, f := range files {
			if test.rgx.MatchString(f.Name()) {
				match := test.rgx.FindStringSubmatch(f.Name())[1]

				if targetDir+"/"+match+".log" == test.path {
					passed = true
					break
				}

			}
		}

		if !passed {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.move() -- failed to find renamed logfile in test dir",
				id,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.move()",
			id,
		)
	}

	for id, test := range tests {
		verify(id, test)

	}
}

func TestLogfileHasExt(t *testing.T) {
	type test struct {
		path   string
		hasExt bool
	}

	var tests []test

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	mockLogfiles := mockLogfileOps["createNamedFiles"].(func(string, ...string) []string)(targetDir, newLogfiles...)
	mockValidExtFiles := mockLogfileOps["createNamedFiles"].(func(string, ...string) []string)(targetDir, newValidExts...)
	mockExtlessLogfiles := mockLogfileOps["createNamedFiles"].(func(string, ...string) []string)(targetDir, newExtlessFiles...)

	for a := 0; a < len(mockLogfiles); a++ {
		tests = append(tests, test{
			path:   mockLogfiles[a],
			hasExt: true,
		})
	}

	for a := 0; a < len(mockValidExtFiles); a++ {
		tests = append(tests, test{
			path:   mockValidExtFiles[a],
			hasExt: true,
		})
	}

	for a := 0; a < len(mockExtlessLogfiles); a++ {
		tests = append(tests, test{
			path:   mockExtlessLogfiles[a],
			hasExt: false,
		})
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}

		ok := f.hasExt(test.path)
		if ok != test.hasExt {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.hasExt() -- expected results mismatch: wanted %v ; got %v",
				id,
				test.hasExt,
				ok,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.hasExt()",
			id,
		)
	}

	for id, test := range tests {
		verify(id, test)

	}
}

func TestLogfileAddExt(t *testing.T) {
	type test struct {
		path  string
		wants string
	}

	var tests []test

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	mockExtlessLogfiles := mockLogfileOps["createNamedFiles"].(func(string, ...string) []string)(targetDir, newExtlessFiles...)

	for a := 0; a < len(mockExtlessLogfiles); a++ {
		tests = append(tests, test{
			path:  mockExtlessLogfiles[a],
			wants: mockExtlessLogfiles[a] + ".log",
		})
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}

		newPath := f.addExt(test.path)
		if newPath != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.addExt() -- expected results mismatch: wanted %v ; got %v",
				id,
				test.wants,
				newPath,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.addExt()",
			id,
		)
	}

	for id, test := range tests {
		verify(id, test)

	}
}

func TestLogfileCutExt(t *testing.T) {
	type test struct {
		path  string
		wants string
	}

	var tests []test

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	mockLogfiles := mockLogfileOps["createNamedFiles"].(func(string, ...string) []string)(targetDir, newLogfiles...)

	var expectedLogfiles = []string{
		targetDir + "/new-logfile",
		targetDir + "/logfile",
		targetDir + "/test-logfile",
		targetDir + "/service",
		targetDir + "/new",
	}

	for a := 0; a < len(mockLogfiles); a++ {
		tests = append(tests, test{
			path:  mockLogfiles[a],
			wants: expectedLogfiles[a],
		})
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}

		newPath := f.cutExt(test.path)
		if newPath != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.addExt() -- expected results mismatch: wanted %v ; got %v",
				id,
				test.wants,
				newPath,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.addExt()",
			id,
		)
	}

	for id, test := range tests {
		verify(id, test)

	}
}

func TestLogfileSize(t *testing.T) {
	type test struct {
		path string
		ok   bool
	}

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	var tests []test

	for a := 0; a < len(underweightIters); a++ {
		newfile := fmt.Sprintf("%s%s-%v", targetDir, "/with-data", a)
		err := mockLogfileOps["createOversize"].(func(string, int) error)(newfile, underweightIters[a])
		if err != nil {
			panic(err)
		}

		obj := test{
			path: newfile,
			ok:   true,
		}

		tests = append(tests, obj)
	}

	blankFiles := mockLogfileOps["createNamedFiles"].(func(string, ...string) []string)(targetDir, newLogfiles...)

	for a := 0; a < len(blankFiles); a++ {
		obj := test{
			path: blankFiles[a],
			ok:   false,
		}

		tests = append(tests, obj)
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}
		err := f.load(test.path)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.load(%s) -- failed to load test logfile with an error: %s",
				id,
				test.path,
				err,
			)
			return
		}

		n, err := f.Size()
		if err != nil {
			if err == os.ErrNotExist {
				t.Errorf(
					"#%v -- FAILED -- [Logfile] Logfile.Size() -- failed to get the size from the logfile, since it doesn't exist: %s",
					id,
					err,
				)
				return
			}
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.Size() -- failed to get the size from the logfile with an error: %s",
				id,
				err,
			)

			return
		}

		if n < 1 && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.Size() -- unexpectedly empty file, has %v bytes in size",
				id,
				n,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.Size()",
			id,
		)

	}

	for id, test := range tests {

		verify(id, test)

	}
}

func TestLogfileIsTooHeavy(t *testing.T) {
	type test struct {
		path    string
		isHeavy bool
	}

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	var tests []test

	for a := 0; a < len(underweightIters); a++ {
		newfile := fmt.Sprintf("%s%s-%v", targetDir, "/underweight", a)
		err := mockLogfileOps["createOversize"].(func(string, int) error)(newfile, underweightIters[a])
		if err != nil {
			panic(err)
		}

		obj := test{
			path:    newfile,
			isHeavy: false,
		}

		tests = append(tests, obj)
	}
	for a := 0; a < len(overweightIters); a++ {
		newfile := fmt.Sprintf("%s%s-%v", targetDir, "/overweight", a)
		err := mockLogfileOps["createOversize"].(func(string, int) error)(newfile, overweightIters[a])
		if err != nil {
			panic(err)
		}

		obj := test{
			path:    newfile,
			isHeavy: true,
		}

		tests = append(tests, obj)
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}
		err := f.load(test.path)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.load(%s) -- failed to load test logfile with an error: %s",
				id,
				test.path,
				err,
			)
			return
		}

		isHeavy := f.IsTooHeavy()

		if isHeavy && !test.isHeavy {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.IsTooHeavy() -- generated file is unexpectedly too heavy",
				id,
			)
			return
		}

		if !isHeavy && test.isHeavy {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.IsTooHeavy() -- generated file is unexpectedly too light",
				id,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.IsTooHeavy()",
			id,
		)

	}

	for id, test := range tests {

		verify(id, test)

	}
}

func TestLogfileRotate(t *testing.T) {
	type test struct {
		path    string
		isHeavy bool
	}

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	var tests []test

	for a := 0; a < len(underweightIters); a++ {
		newfile := fmt.Sprintf("%s%s-%v", targetDir, "/underweight", a)
		err := mockLogfileOps["createOversize"].(func(string, int) error)(newfile, underweightIters[a])
		if err != nil {
			panic(err)
		}

		obj := test{
			path:    newfile,
			isHeavy: false,
		}

		tests = append(tests, obj)
	}
	for a := 0; a < len(overweightIters); a++ {
		newfile := fmt.Sprintf("%s%s-%v", targetDir, "/overweight", a)
		err := mockLogfileOps["createOversize"].(func(string, int) error)(newfile, overweightIters[a])
		if err != nil {
			panic(err)
		}

		obj := test{
			path:    newfile,
			isHeavy: true,
		}

		tests = append(tests, obj)
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}
		err := f.load(test.path)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.load(%s) -- failed to load test logfile with an error: %s",
				id,
				test.path,
				err,
			)
			return
		}

		// Logfile.move() and Logfile.new() are already tested individually
		rotateErr := f.Rotate()

		if rotateErr != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.Rotate() -- operation failed with an error: %s",
				id,
				err,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.Rotate()",
			id,
		)

	}

	for id, test := range tests {

		verify(id, test)

	}
}

func TestLogfileWrite(t *testing.T) {
	type test struct {
		path string
		iter int
	}

	targetDir := mockLogfileOps["init"].(func() string)()
	defer func() { os.RemoveAll(targetDir) }()

	var tests []test

	blankFiles := mockLogfileOps["createNamedFiles"].(func(string, ...string) []string)(targetDir, newLogfiles...)

	for a := 0; a < len(blankFiles); a++ {
		obj := test{
			path: blankFiles[a],
			iter: a + 1,
		}

		tests = append(tests, obj)
	}

	var verify = func(id int, test test) {
		f := &Logfile{rotate: 50}
		err := f.load(test.path)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [Logfile] Logfile.load(%s) -- failed to load test logfile with an error: %s",
				id,
				test.path,
				err,
			)
			return
		}

		buf := &bytes.Buffer{}

		for i := 0; i < test.iter; i++ {
			n, err := buf.Write(mockByteChunk64)
			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- [Logfile] bytes.Buffer.Write() -- writting to test buffer failed with an error: %s",
					id,
					err,
				)
				return
			}

			if n != len(mockByteChunk64) {
				t.Errorf(
					"#%v -- FAILED -- [Logfile] bytes.Buffer.Write() -- expected to write %v bytes, wrote %v bytes instead",
					id,
					len(mockByteChunk64),
					n,
				)
				return
			}
		}

		n, err := f.Write(buf.Bytes())

		if err != nil {
			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- [Logfile] Logfile.Write() -- writting to Logfile failed with an error: %s",
					id,
					err,
				)
				return
			}

			if n != buf.Len() {
				t.Errorf(
					"#%v -- FAILED -- [Logfile] Logfile.Write() -- expected to write %v bytes, wrote %v bytes instead",
					id,
					buf.Len(),
					n,
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED -- [Logfile] Logfile.Write()",
			id,
		)

	}

	for id, test := range tests {

		verify(id, test)

	}
}
