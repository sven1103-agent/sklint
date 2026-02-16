package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExitWithError(t *testing.T) {
	if os.Getenv("SKLINT_EXIT_HELPER") == "1" {
		exitWithError("boom")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExitWithError")
	cmd.Env = append(os.Environ(), "SKLINT_EXIT_HELPER=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", exitErr.ExitCode())
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Fatalf("expected stderr to contain message, got %q", stderr.String())
	}
}
