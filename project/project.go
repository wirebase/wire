package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/advanderveer/9d587b277/line/bundle"
	"github.com/advanderveer/9d587b277/line/compile"
	"github.com/advanderveer/9d587b277/line/poller"
	"github.com/advanderveer/9d587b277/line/runner"
)

// Config configures the development server
type Config struct {

	// EmbedFilename is the filename relative to the project directory
	// to which the bundle filesystem will be written. Defaults to 'bundle.go'
	EmbedFilename string

	// WasmFilename is the name under which the webassembly binary will
	// be stored in the bundle directory.
	WasmFilename string

	// MaxWasmBuildTime configures how long the wasm build is allowed
	// to take on each change. Defaults to 5s
	MaxWasmBuildTime time.Duration

	// MaxServeBuildTime configures how long building of the serving
	// binary is allowed to take, Defaults to 30s.
	MaxServeBuildTime time.Duration

	// Poller holds configuration for the poller
	Poller poller.Config

	// Runner holds configuration for the runner
	Runner runner.Config
}

// DefaultConfig returns a sensible default config
func DefaultConfig() (cfg Config) {
	return Config{
		EmbedFilename:     "bundle.go",
		WasmFilename:      "main.wasm",
		MaxWasmBuildTime:  time.Second * 5,
		MaxServeBuildTime: time.Second * 30,
	}
}

// Project describes a source code directory that is being developed
type Project struct {
	dir   string
	pollf time.Duration
}

// New will setup the project
func New(dir string, pollf time.Duration) (b *Project) {
	b = &Project{dir: dir, pollf: pollf}
	return
}

// Run will block and start polling for changes and bundle, build and run
// the application whenever this happens. Whenever the context is cancelled
// the polling will stop.
func (p *Project) Run(ctx context.Context) error {
	runner := runner.New()
	poller := poller.New(ctx, p.dir, p.pollf)
	ui := NewTerseTerminal(os.Stderr)

	// start initial bundle, build and run
	err := BundleBuildAndRun(ui, p.dir, runner, poller)
	if err != nil {
		return err
	}

	// then perform the same on every change
	for poller.Next() {
		err := BundleBuildAndRun(ui, p.dir, runner, poller)
		if err != nil {
			return err
		}
	}

	return nil
}

// BundleBuildAndRun will attempt to build the project in 'dir' and run it using the
// provided runner. It will re-load the configuration from disk and update the
// poller and runner with it.
func BundleBuildAndRun(ui UI, dir string, runner *runner.Runner, poller *poller.Poller) (err error) {
	ui.ShowRebuildStarted()

	// setup and laod configuration
	cfg := DefaultConfig()
	cfg.Poller.Ignore = append(cfg.Poller.Ignore, cfg.EmbedFilename)
	poller.Update(cfg.Poller)
	ui.ShowConfigLoaded()

	// bundle frontend code
	err = bundleFrontend(ui, dir, cfg)
	if err != nil {
		return fmt.Errorf("failed to bundle: %w", err)
	}

	ui.ShowBundlingDone()

	// build the backend
	binp, err := buildBackend(ui, dir, cfg)
	if err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	ui.ShowBuildingDone()

	// run the (new) binary, if build was successfull
	if binp != "" {
		err = runner.Run(binp, cfg.Runner)
		if err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}

		ui.ShowRunningDone()
	}

	ui.ShowRebuildDone()
	fmt.Println("cmx,m,cx")
	return
}

// Bundle will gather all the frontend code and assets and produce an filesystem
// that can be embedded to serve them
func bundleFrontend(ui UI, dir string, cfg Config) (err error) {

	// init a new bundle
	b, err := bundle.New()
	if err != nil {
		return fmt.Errorf("failed to start bundle: %w", err)
	}

	defer b.Clear()

	ui.ShowBundleCreated()

	// try to compile wasm to bundle
	wasmc, err := compile.New(dir, "js", "wasm")
	if err == nil {

		//there is some wasm to build, do so
		err = wasmc.Build(filepath.Join(b.Dir(), cfg.WasmFilename), cfg.MaxWasmBuildTime)
		if err != nil {
			return fmt.Errorf("failed to build wasm: %w", err)
		}

		ui.ShowWasmBundled()
	}

	// turn bundle into an embeddable go file, write to project dir
	embedp := filepath.Join(dir, cfg.EmbedFilename)
	err = b.Write(embedp)
	if err != nil {
		return fmt.Errorf("failed to write embed file: %w", err)
	}

	ui.ShowEmbedFileWritten()

	return
}

func buildBackend(ui UI, dir string, cfg Config) (binp string, err error) {

	// start serve compile
	servec, err := compile.New(dir, "", "")
	if err != nil {
		return "", nil //nothing to do
	}

	// compile to binary
	binp = filepath.Join(os.TempDir(), "serve_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	err = servec.Build(binp, cfg.MaxServeBuildTime)
	if err != nil {
		return "", fmt.Errorf("failed to build program: %w", err)
	}

	return
}
