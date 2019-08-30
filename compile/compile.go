package compile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	// ErrGoNotFound is returned when we couldn't find the go executable on the system
	ErrGoNotFound = errors.New("couldn't find 'go' executable in PATH")

	// ErrNoModule is returned when the the go tool expects a module to be defined
	ErrNoModule = errors.New("no module defined, make sure you've added a go.mod file")

	// ErrNoGoPackage is returned when the tool expected some go files at least
	ErrNoGoPackage = errors.New("no go package defined in any Go files, or no Go files at all")

	// ErrAllExcluded is returned when the tool expected some go files to included for building
	ErrAllExcluded = errors.New("no go files to build, all excluded by build tags")

	// ErrNotAProgram is returned when the tool expected a main package
	ErrNotAProgram = errors.New("no program to build, not a 'main' package")
)

// listErrs maps `go list` stderr message to error values we can easier assert
// and allow for a more friendly message
var listErrs = map[*regexp.Regexp]error{

	// go: cannot find main module; see 'go help modules'
	regexp.MustCompile(`.*cannot find main module.*`): ErrNoModule,

	// can't load package: package app: unknown import path "app": package app is not in the main module (app)
	regexp.MustCompile(`.*cannot find module for path.*`): ErrNoGoPackage,
}

// BuildErr is returned when the build command fails
type BuildErr struct {
	Dir string
	Msg string
}

func (e BuildErr) Error() string {
	return fmt.Sprintf("failed to build '%s':\n%s", e.Dir, e.Msg)
}

// A Compile will compile Go programs in a directory
type Compile struct {
	exe  string
	dir  string
	os   string
	arch string
}

// New initate a compiler by inspecting a directory for buildable Go files.
func New(dir string, goos, goarch string) (c *Compile, err error) {
	c = &Compile{dir: dir, os: goos, arch: goarch}

	c.exe, err = exec.LookPath("go")
	if err != nil {
		return nil, ErrGoNotFound
	}

	main, err := c.inspectDir(time.Second)
	if err != nil {
		return nil, err
	}

	if !main {
		return nil, ErrNotAProgram
	}

	return
}

// Build the source code into binary file 'o'
func (c *Compile) Build(o string, to time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	_, stde, err := c.runGo(ctx, "build", "-o", o)
	if err != nil {
		return BuildErr{Dir: c.dir, Msg: stde.String()}
	}

	return
}

func (c *Compile) runGo(ctx context.Context, args ...string) (stdo, stde *bytes.Buffer, err error) {
	stde = bytes.NewBuffer(nil)
	stdo = bytes.NewBuffer(nil)

	cmd := exec.CommandContext(ctx, c.exe, args...)
	cmd.Dir = c.dir
	cmd.Stdout = stdo
	cmd.Stderr = stde
	cmd.Env = os.Environ()
	if c.os != "" {
		cmd.Env = append(cmd.Env, "GOOS="+c.os)
	}

	if c.arch != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+c.arch)
	}

	err = cmd.Run()
	if err != nil {
		return stdo, stde, fmt.Errorf("failed to run '%s': \n\t%s\n %w", strings.Join(cmd.Args, " "), stde.String(), err)
	}

	return stdo, stde, nil
}

// inspectDir will use `go list` to inspect the directory for buildable files.
// The command will be cancelled if it takes longer then timeout 'to'
func (c *Compile) inspectDir(to time.Duration) (main bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	stdo, stde, err := c.runGo(ctx, "list", "-json")
	if err != nil {
		for exp, err := range listErrs {
			if exp.Match(stde.Bytes()) {
				return false, fmt.Errorf("inspecting '%s': %w", c.dir, err)
			}
		}

		return false, err
	}

	var info struct {
		Name string
	}

	dec := json.NewDecoder(stdo)
	err = dec.Decode(&info)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal `go list -json` output\n: %w", err)
	}

	return info.Name == "main", nil
}
