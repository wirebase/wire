package bundle_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/advanderveer/9d587b277/line/bundle"
)

func TestBundling(t *testing.T) {
	b, err := bundle.New()
	if err != nil {
		t.Fatalf("failed to create bundle, got: %v", err)
	}

	dir, _ := ioutil.TempDir("", "bundle_test")
	p := filepath.Join(dir, "assets.go")
	err = b.Write(p)
	if err != nil {
		t.Fatalf("failed to write bundle, got: %v", err)
	}

	err = b.Clear()
	if err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	_, err = os.Stat(b.Dir())
	if !os.IsNotExist(err) {
		t.Fatalf("expected file to no longer exist due to clear")
	}
}
