package runner

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// captureStderr redirects os.Stderr for the duration of f and returns the captured output.
func captureStderr(f func()) string {
	// Redirect stderr to a pipe
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	// Run the function
	f()
	// Restore stderr and close writer to unblock reader
	w.Close()
	os.Stderr = orig
	// Read any output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestRun_Success verifies that a valid configuration runs successfully.
func TestRun_Success(t *testing.T) {
	dir := t.TempDir()
	// Create a minimal treehouse.yaml
	config := `core_services:
  svc:
    command: "true"
    env:
      FOO: "bar"
      BAZ: "qux"
    health_check:
      url: "http://localhost:8080"
      codes: [200]
      interval_seconds: 1
      timeout_seconds: 1
`
	if err := os.WriteFile(filepath.Join(dir, "treehouse.yaml"), []byte(config), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	// Ensure vars are unset before Run
	os.Unsetenv("FOO")
	os.Unsetenv("BAZ")

	r := New(Options{ConfigDir: dir, Mode: "test"})
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify environment variables were set
	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO: expected 'bar', got %q", got)
	}
	if got := os.Getenv("BAZ"); got != "qux" {
		t.Errorf("BAZ: expected 'qux', got %q", got)
	}
}

// TestRun_MissingConfig verifies that missing config file aborts Run with an error.
func TestRun_MissingConfig(t *testing.T) {
	dir := t.TempDir()
	r := New(Options{ConfigDir: dir, Mode: "test"})
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected file not found error, got %v", err)
	}
}

// TestRun_InvalidConfig verifies that invalid YAML config aborts Run with an error.
func TestRun_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	// Write invalid YAML
	if err := os.WriteFile(filepath.Join(dir, "treehouse.yaml"), []byte("invalid: [yaml"), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	r := New(Options{ConfigDir: dir, Mode: "test"})
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

// TestRun_ModeOverride verifies that mode-specific commands are used when available.
func TestRun_ModeOverride(t *testing.T) {
	dir := t.TempDir()
	config := `core_services:
  svc:
    command: "false"
    modes:
      test: "true"
`
	if err := os.WriteFile(filepath.Join(dir, "treehouse.yaml"), []byte(config), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	r := New(Options{ConfigDir: dir, Mode: "test"})
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
