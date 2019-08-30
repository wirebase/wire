package runner

import (
	"fmt"
	"os"
	"os/exec"
)

// Config configures the running of processes
type Config struct {

	// Args holds arguments passed to the binary that is run
	Args []string

	// Env configures environment variables that will be appended
	// the to existing environment variables before being passed to
	// the process
	Env []string
}

// Runner manages (re)running the serving binary whenever something changes
type Runner struct{ cmd *exec.Cmd }

// New initiales a new runner
func New() *Runner {
	return &Runner{}
}

// Kill the currently running process, if there is no process running this
// method is a no-op
func (r *Runner) Kill() (err error) {
	if r.cmd != nil {
		err = r.cmd.Process.Kill()
		if err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}

		// wait for process to end, we do not care what happened to the process
		r.cmd.Wait()
	}

	return
}

// Run a binary at the provided location, if there is already a binary runnig
// shut it down before starting the new process.
func (r *Runner) Run(binp string, cfg Config) (err error) {
	err = r.Kill()
	if err != nil {
		return err
	}

	r.cmd = exec.Command(binp, cfg.Args...)
	r.cmd.Env = append(os.Environ(), cfg.Env...)
	r.cmd.Stderr = os.Stderr
	err = r.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	return nil
}
