package project_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wirebase/wire/poller"
	"github.com/wirebase/wire/project"
	"github.com/wirebase/wire/runner"
)

var _ project.UI = &project.TerseTerminal{}

func setupTestProject(tb testing.TB) (dir string, f func()) {
	dir, err := ioutil.TempDir("", "project_test_")
	if err != nil {
		tb.Fatalf("failed to create test project dir: %v", err)
	}

	return dir, func() {
		err = os.RemoveAll(dir)
		if err != nil {
			tb.Fatalf("failed to remove test project directory: %v", err)
		}
	}
}

func writeWorkingProjectFiles(tb testing.TB, dir string) {
	ioutil.WriteFile(filepath.Join(dir, "go.mod"), []byte("module app\n"), 0777)
	ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(
		`// +build wasm

    package main

    func main(){ }`), 0777)

	ioutil.WriteFile(filepath.Join(dir, "serve.go"), []byte(
		`// +build !wasm

    package main

    func main(){ }`), 0777)
}

func TestBuildAndRun(t *testing.T) {
	dir, clean := setupTestProject(t)
	defer clean()
	writeWorkingProjectFiles(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	buf := bytes.NewBuffer(nil)
	runner := runner.New()
	poller := poller.New(ctx, dir, time.Millisecond*10)
	ui := project.NewTerseTerminal(buf)
	err := project.BundleBuildAndRun(ui, dir, runner, poller)
	if err != nil {
		t.Fatalf("should build successfully, got: %v", err)
	}

	if buf.String() != "rebuilding.......done\n" {
		t.Fatalf("expected this output, got: %v", buf.String())
	}
}
