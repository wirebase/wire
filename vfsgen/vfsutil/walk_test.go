package vfsutil_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/advanderveer/9d587b277/line/vfsgen/vfsutil"
)

func testFilesystem(tb testing.TB, files map[*[]string]string) (fs http.FileSystem, clean func()) {
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

	return http.Dir(dir), func() {
		err = os.RemoveAll(dir)
		if err != nil {
			tb.Fatalf("failed to remove test fs: %v", err)
		}
	}
}

func TestWalk(t *testing.T) {
	fs, clean := testFilesystem(t, map[*[]string]string{
		&[]string{"zzz-last-file.txt"}:      "It should be visited last.",
		&[]string{"a-file.txt"}:             "It has stuff.",
		&[]string{"another-file.txt"}:       "Also stuff.",
		&[]string{"folderA", "entry-A.txt"}: "Alpha.",
		&[]string{"folderA", "entry-B.txt"}: "Beta.",
		&[]string{"skip-me", "entry-C.txt"}: "Gamma.",
	})
	defer clean()

	t.Run("walk", func(t *testing.T) {
		var visits []string
		wfn := func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if path == "/skip-me" {
				return filepath.SkipDir
			}

			visits = append(visits, path)
			return nil
		}

		err := vfsutil.Walk(fs, "/", wfn)
		if err != nil {
			t.Fatalf("walk shouldnt fail, got: %v", err)
		}

		exp := []string{
			"/",
			"/a-file.txt",
			"/another-file.txt",
			"/folderA",
			"/folderA/entry-A.txt",
			"/folderA/entry-B.txt",
			"/zzz-last-file.txt",
		}

		if !reflect.DeepEqual(visits, exp) {
			t.Fatalf("expected visits: %v, got: %v", exp, visits)
		}
	})

	t.Run("walk files", func(t *testing.T) {
		var visits []string
		wfn := func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
			if err != nil {
				return err
			}

			if path == "/skip-me" {
				return filepath.SkipDir
			}

			if !fi.IsDir() {
				b, err := ioutil.ReadAll(r)
				if err != nil {
					t.Fatalf("can't read file %s: %v\n", path, err)
					return nil
				}

				if len(b) < 1 {
					t.Fatal("expected some bytes to be read")
				}
			}

			visits = append(visits, path)
			return nil
		}

		err := vfsutil.WalkFiles(fs, "/", wfn)
		if err != nil {
			t.Fatalf("walk shouldnt fail, got: %v", err)
		}

		exp := []string{
			"/",
			"/a-file.txt",
			"/another-file.txt",
			"/folderA",
			"/folderA/entry-A.txt",
			"/folderA/entry-B.txt",
			"/zzz-last-file.txt",
		}

		if !reflect.DeepEqual(visits, exp) {
			t.Fatalf("expected visits: %v, got: %v", exp, visits)
		}
	})
}
