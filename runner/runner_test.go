package runner_test

import (
	"testing"

	"github.com/wirebase/wire/runner"
)

func TestRunningOfQuickProcess(t *testing.T) {
	r := runner.New()
	err := r.Run("go", runner.Config{})
	if err != nil {
		t.Fatalf("expected run to succeed, got: %v", err)
	}

	err = r.Run("go", runner.Config{})
	if err != nil {
		t.Fatalf("expected second run to succeed, got: %v", err)
	}

	t.Run("blocking process", func(t *testing.T) {
		err := r.Run("sleep", runner.Config{Args: []string{"300"}})
		if err != nil {
			t.Fatalf("expected run to succeed, got: %v", err)
		}

		err = r.Run("sleep", runner.Config{Args: []string{"300"}})
		if err != nil {
			t.Fatalf("expected second run to succeed, got: %v", err)
		}

		err = r.Kill()
		if err != nil {
			t.Fatalf("failed to kill: %v", err)
		}
	})
}
