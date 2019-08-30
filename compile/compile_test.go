package compile_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/wirebase/wire/compile"
)

func TestCompilerCreation(t *testing.T) {
	dir, err := ioutil.TempDir("", "tl_comp_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Run("no module", func(t *testing.T) {
		_, err = compile.New(dir, "js", "wasm")
		if !errors.Is(err, compile.ErrNoModule) {
			t.Fatalf("expected error about no module in '%s', got: %v", dir, err)
		}
	})

	ioutil.WriteFile(filepath.Join(dir, "go.mod"), []byte("module app\n"), 0777)

	t.Run("no package", func(t *testing.T) {
		_, err = compile.New(dir, "js", "wasm")
		if !errors.Is(err, compile.ErrNoGoPackage) {
			t.Fatalf("expected error about no go package in '%s', got: %v", dir, err)
		}
	})

	ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(
		`// +build !js

	   package main

	   func main(){ println("hello") }`), 0777)

	// t.Run("all excluded", func(t *testing.T) {
	// 	_, err = compile.New(dir, "js", "wasm")
	// 	if !errors.Is(err, compile.ErrAllExcluded) {
	// 		t.Fatalf("expected error about all go go files excluded in '%s', got: %v", dir, err)
	// 	}
	// })

	_, err = compile.New(dir, "", "")
	if err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}

	ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(
		`// +build !js

    package app

    func main(){ println("hello") }`), 0777)

	t.Run("not a main", func(t *testing.T) {
		_, err = compile.New(dir, "", "")
		if !errors.Is(err, compile.ErrNotAProgram) {
			t.Fatalf("expected error about not a program, got: %v", err)
		}
	})

	t.Run("go not found", func(t *testing.T) {
		path := os.Getenv("PATH")
		defer os.Setenv("PATH", path)

		os.Setenv("PATH", "")
		_, err = compile.New(dir, "", "")
		if err != compile.ErrGoNotFound {
			t.Fatalf("expected go not found error, got: %v", err)
		}
	})
}

func TestCompileBuilding(t *testing.T) {
	dir, err := ioutil.TempDir("", "tl_comp_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	ioutil.WriteFile(filepath.Join(dir, "go.mod"), []byte("module app\n"), 0777)
	ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(
		`// +build !js

    package main

    func main(){ println("hello) }`), 0777)

	c, err := compile.New(dir, "", "")
	if err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}

	t.Run("failed build", func(t *testing.T) {
		err = c.Build(filepath.Join(dir, "app"), time.Second)
		if be, ok := err.(compile.BuildErr); !ok || be.Dir != dir {
			t.Fatalf("expected build error for dir '%s' , got: %v", dir, err)
		}
	})

	t.Run("correct build", func(t *testing.T) {
		ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(
			`// +build !js

      package main

      func main(){ println("hello") }`), 0777)

		p := filepath.Join(dir, "app")
		err = c.Build(p, time.Second)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		out := bytes.NewBuffer(nil)
		cmd := exec.Command(p)
		cmd.Stderr = out
		err = cmd.Run()
		if err != nil {
			t.Fatalf("expected bin to run without error, got: %v", err)
		}

		if v := out.String(); v != "hello\n" {
			t.Fatalf("expected program to output correctyl, got: %v", v)
		}
	})
}
