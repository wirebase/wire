package poller

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Config configures the poller
type Config struct {

	// Ignore holds a list of filepath.Match patterns, if a file matches this
	// pattern it will be ignored during a scan. If it matches a directory
	// all files in the directory will be ignored
	Ignore []string
}

// A Poller will scan a directory for changes by repeatedly checking the
// modification time of any directory or file has changed.
type Poller struct {
	ctx  context.Context
	mods chan time.Time
	errs chan error
	last error
	freq time.Duration
	dir  string
	cfgs chan Config
}

// New will create and start a new poller that scans the directory tree 'dir'
// every 'f' amount of time. Changes can be observed by using the poller as an
// iterator, iteration stops when the provided ctx is cancelled
func New(ctx context.Context, dir string, f time.Duration) (p *Poller) {
	p = &Poller{
		ctx:  ctx,
		mods: make(chan time.Time, 0),
		errs: make(chan error),
		cfgs: make(chan Config, 1),
		freq: f,
		dir:  dir,
	}

	p.cfgs <- Config{}

	go p.start()
	return
}

// Next blocks until a new change has been detected
func (p *Poller) Next() bool {
	select {
	case mt := <-p.mods:
		if mt.IsZero() {
			return false
		}

		return true
	case err := <-p.errs:
		p.last = err
		return true
	}
}

// Update the poller configuration, overwriting it on the next polling cycle
func (p *Poller) Update(cfg Config) {
	p.cfgs <- cfg
}

// Err returns the last error that occured
func (p *Poller) Err() error {
	return p.last
}

// repeatedly scan directory 'dir' for changed files or directories, if something
// has changed a signal is send over 'c'. It will then wait for 'w' amount of time
// before performing a new scan. It will close 'c' if the context is cancelled.
func (p *Poller) start() {
	for t := time.Now(); ; {

		// read the lastest config, this is always available
		cfg := <-p.cfgs

		nt, err := p.scan(t, cfg)
		if err != nil {
			p.errs <- err
		} else if !nt.IsZero() {
			p.mods <- nt
			t = nt
		}

		select {
		case <-time.After(p.freq):

			// try to push the current config back to read on the next iteration. If
			// thats blocks it means new config was pushed and has precedence
			select {
			case p.cfgs <- cfg:
			default:
			}

			continue
		case <-p.ctx.Done():
			close(p.mods)
			return
		}
	}
}

// scan for a file or directory that has a newer mod date then 't' and
// return this newer time. If no newer time was found, it returns a zero time.
func (p *Poller) scan(t time.Time, cfg Config) (nt time.Time, err error) {
	stop := errors.New("stop")
	err = filepath.Walk(p.dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err // stop the walk on any error
		}

		// if the file or directory matches an ignore pattern, skip it
		rel, _ := filepath.Rel(p.dir, path)
		for _, pattern := range cfg.Ignore {
			m, _ := filepath.Match(pattern, rel)
			if !m {
				continue
			}

			if fi.IsDir() {
				return filepath.SkipDir
			}

			return nil //skip just this file
		}

		// if we found any file that is newer, then our last scan we're done
		if fi.ModTime().After(t) {
			nt = fi.ModTime()
			return stop
		}

		return nil
	})

	if err == nil || err == stop {
		return nt, nil
	}

	return
}
