package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cli "github.com/urfave/cli/v2"
)

// TestHandlerSetters ensures fluent setters set internal fields correctly.
func TestHandlerSetters(t *testing.T) {
	h := New().
		SetConfigDir("cfg").
		SetMode("m").
		SetFocus("f").
		SetMute("u").
		SetTUI(true).
		SetSPMMode(true)
	if h.configDir != "cfg" {
		t.Errorf("configDir: expected %q, got %q", "cfg", h.configDir)
	}
	if h.mode != "m" {
		t.Errorf("mode: expected %q, got %q", "m", h.mode)
	}
	if h.focus != "f" {
		t.Errorf("focus: expected %q, got %q", "f", h.focus)
	}
	if h.mute != "u" {
		t.Errorf("mute: expected %q, got %q", "u", h.mute)
	}
	// SetTUI sets the noTUI flag to disable the TUI when true
	if !h.noTUI {
		t.Error("noTUI: expected true, got false")
	}
	// SetSPMMode sets the spmMode flag
	if !h.spmMode {
		t.Error("spmMode: expected true, got false")
	}
}

// helper to suppress stdout and stderr during test
func suppressOutput(f func()) {
	origOut, origErr := os.Stdout, os.Stderr
	null, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	os.Stdout, os.Stderr = null, null
	defer func() {
		null.Close()
		os.Stdout, os.Stderr = origOut, origErr
	}()
	f()
}

// TestRun_Success verifies Run sets environment and returns nil on valid config.
func TestRun_Success(t *testing.T) {
	dir := t.TempDir()
	// Create config file
	key := "__TEST_APP_RUN__"
	val := "VALUE"
	config := `core_services:
  svc:
    command: "echo ok"
    env:
      ` + key + `: "` + val + `"
`
	configFile := filepath.Join(dir, "treehouse.yaml")
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}
	defer os.Unsetenv(key)

	// disable the TUI to exercise the services-runner path
	h := New().SetConfigDir(dir).SetMode("test").SetTUI(true)
	suppressOutput(func() {
		if err := h.Run(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
	if got := os.Getenv(key); got != val {
		t.Errorf("env var %s: expected %q, got %q", key, val, got)
	}
}

// TestRun_SPMMode verifies Run works correctly in SPM mode.
func TestRun_SPMMode(t *testing.T) {
	dir := t.TempDir()
	// Create config file with multiple services
	config := `core_services:
  svc1:
    command: "echo ok"
    health_check:
      url: "http://localhost:8080"
      codes: [200]
  svc2:
    command: "echo ok"
    health_check:
      url: "http://localhost:8081"
      codes: [200]
`
	configFile := filepath.Join(dir, "treehouse.yaml")
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	// Run in SPM mode focusing on svc1
	h := New().
		SetConfigDir(dir).
		SetMode("test").
		SetTUI(true).
		SetSPMMode(true).
		SetFocus("svc1")
	suppressOutput(func() {
		if err := h.Run(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

// TestRun_ConfigMissing verifies Run returns ExitError when config is missing.
func TestRun_ConfigMissing(t *testing.T) {
	dir := t.TempDir()
	// disable the TUI to exercise the services-runner path
	h := New().SetConfigDir(dir).SetMode("nonexistent").SetTUI(true)
	var exitCoder cli.ExitCoder
	suppressOutput(func() {
		err := h.Run()
		if !errors.As(err, &exitCoder) {
			t.Fatalf("expected cli.ExitCoder, got %v", err)
		}
		if exitCoder.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitCoder.ExitCode())
		}
	})
}
