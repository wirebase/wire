package vfsgen_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/advanderveer/9d587b277/line/vfsgen"
)

func testFilesystem(tb testing.TB, files map[*[]string]string) (dir string, fs http.FileSystem, clean func()) {
	dir, err := ioutil.TempDir("", "vfsgen_test_")
	if err != nil {
		tb.Fatalf("failed to setup test fs: %v", err)
	}

	for parts, data := range files {
		if parts == nil {
			continue
		}

		p := filepath.Join(*parts...)
		p = filepath.Join(dir, p)
		err = os.MkdirAll(filepath.Dir(p), 0777)
		if err != nil {
			tb.Fatalf("failed to make dir for '%s': %v", p, err)
		}

		err = ioutil.WriteFile(p, []byte(data), 0777)
		if err != nil {
			tb.Fatalf("failed to write file '%s': %v", p, err)
		}
	}

	return dir, http.Dir(dir), func() {
		err = os.RemoveAll(dir)
		if err != nil {
			tb.Fatalf("failed to remove test fs: %v", err)
		}
	}
}

// func testDir(tb testing.TB) (dir string, clean func()) {
// 	dir, err := ioutil.TempDir("", "vfsgen_test_")
// 	if err != nil {
// 		tb.Fatalf("failed to create tempdir: %v", err)
// 	}
//
// 	return dir, func() {
// 		if err := os.RemoveAll(dir); err != nil {
// 			tb.Fatalf("failed to remove: %v", err)
// 		}
// 	}
// }

func TestGenerate_buildAndGofmt(t *testing.T) {
	tempDir, _, clean := testFilesystem(t, nil)
	defer clean()

	_, emptyfs, clean1 := testFilesystem(t, nil)
	defer clean1()

	_, notcompfs, clean2 := testFilesystem(t, map[*[]string]string{
		&[]string{"not-compressable-file.txt"}: "Not compressable.",
	})
	defer clean2()

	_, compfs, clean3 := testFilesystem(t, map[*[]string]string{
		&[]string{"compressable-file.txt"}: "This text compresses easily. " + strings.Repeat(" Go!", 128),
	})
	defer clean3()

	cases := []struct {
		filename  string
		fs        http.FileSystem
		wantError func(error) bool // Nil function means want nil error.
	}{
		{
			// Empty.
			filename: "empty.go",
			fs:       emptyfs,
		},
		{
			// Test that vfsgen.Generate returns an error when there is
			// an error reading from the input filesystem.
			filename:  "notexist.go",
			fs:        http.Dir("notexist"),
			wantError: os.IsNotExist,
		},
		{
			// No compressed files.
			filename: "nocompressed.go",
			fs:       notcompfs,
		},
		{
			// Only compressed files.
			filename: "onlycompressed.go",
			fs:       compfs,
		},
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			filename := filepath.Join(tempDir, c.filename)
			err := vfsgen.Generate(c.fs, vfsgen.Options{
				Filename:    filename,
				PackageName: "test",
			})

			switch {
			case c.wantError == nil && err != nil:
				t.Fatalf("%s: vfsgen.Generate returned non-nil error: %v", c.filename, err)
			case c.wantError != nil && !c.wantError(err):
				t.Fatalf("%s: vfsgen.Generate returned wrong error: %v", c.filename, err)
			}
			if c.wantError != nil {
				return
			}

			if out, err := exec.Command("go", "build", filename).CombinedOutput(); err != nil {
				t.Errorf("err: %v\nout: %s", err, out)
			}
			if out, err := exec.Command("gofmt", "-d", "-s", filename).Output(); err != nil || len(out) != 0 {
				t.Errorf("gofmt issue\nerr: %v\nout: %s", err, out)
			}
		})
	}

	_ = tempDir

}
