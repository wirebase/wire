package poller_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/wirebase/wire/poller"
)

func TestPollFileCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	dir, err := ioutil.TempDir("", "tl_poller_")
	if err != nil {
		t.Fatalf("failed to create tempdir: %v", err)
	}

	ioutil.WriteFile(filepath.Join(dir, "bar.txt"), nil, 0777)
	os.MkdirAll(filepath.Join(dir, "x", "y"), 0777)
	ioutil.WriteFile(filepath.Join(dir, "x", "y", "z.txt"), nil, 0777)

	p := poller.New(ctx, dir, time.Millisecond*15)
	go func() {
		time.Sleep(time.Millisecond * 20)
		ioutil.WriteFile(filepath.Join(dir, "foo.txt"), nil, 0777)

		p.Update(poller.Config{Ignore: []string{"bar.txt", filepath.Join("x", "y")}})
		time.Sleep(time.Millisecond * 20)
		ioutil.WriteFile(filepath.Join(dir, "bar.txt"), nil, 0777)
		ioutil.WriteFile(filepath.Join(dir, "x", "y", "z.txt"), nil, 0777)
	}()

	var i int
	for p.Next() {
		i++
	}

	if p.Err() != nil {
		t.Fatalf("expected no error to have occured, got: %v", p.Err())
	}

	if i != 1 {
		t.Fatalf("expected this many changed, got: %v", i)
	}
}

func TestPollOfNonExistingDir(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	path := filepath.Join(os.TempDir(), "tl_text_"+strconv.FormatInt(time.Now().UnixNano(), 10))

	p := poller.New(ctx, path, time.Millisecond*15)
	for p.Next() {
	}

	if !os.IsNotExist(p.Err()) {
		t.Fatalf("expected error to be a non-exist error, got: %v", p.Err())
	}
}
