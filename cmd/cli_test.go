package cmd

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	cli "github.com/urfave/cli/v2"
)

// makeContext builds a cli.Context with provided flag values.
func makeContext(configDir, mode, focus, mute string, command string, args ...string) *cli.Context {
	app := &cli.App{}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("config-dir", "", "")
	set.String("mode", "", "")
	set.String("focus", "", "")
	set.String("mute", "", "")
	// Build args
	cmdArgs := []string{"--config-dir", configDir, "--mode", mode}
	if focus != "" {
		cmdArgs = append(cmdArgs, "--focus", focus)
	}
	if mute != "" {
		cmdArgs = append(cmdArgs, "--mute", mute)
	}
	if command != "" {
		cmdArgs = append([]string{command}, cmdArgs...)
	}
	cmdArgs = append(cmdArgs, args...)
	set.Parse(cmdArgs)
	return cli.NewContext(app, set, nil)
}

// suppressOutput silences stdout and stderr during f.
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

// TestStart_Success ensures start command returns nil when config exists and is valid.
func TestStart_Success(t *testing.T) {
	dir := t.TempDir()
	// Write treehouse.yaml with test configuration
	config := `core_services:
  svc:
    command: "echo ok"
    env:
      X: "1"
    modes:
      test:
        command: "echo ok"
`
	if err := os.WriteFile(filepath.Join(dir, "treehouse.yaml"), []byte(config), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}
	c := makeContext(dir, "test", "", "", "start")
	suppressOutput(func() {
		if err := runWithOptions(c, false); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// TestSPM_Success ensures spm command returns nil when config exists and is valid.
func TestSPM_Success(t *testing.T) {
	dir := t.TempDir()
	// Write treehouse.yaml with test configuration
	config := `core_services:
  svc:
    command: "echo ok"
    env:
      X: "1"
    modes:
      test:
        command: "echo ok"
`
	if err := os.WriteFile(filepath.Join(dir, "treehouse.yaml"), []byte(config), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}
	c := makeContext(dir, "test", "", "", "spm", "svc")
	suppressOutput(func() {
		if err := runSingleService(c, "svc"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// TestSPM_MissingService ensures spm command returns error when service name is missing.
func TestSPM_MissingService(t *testing.T) {
	dir := t.TempDir()
	c := makeContext(dir, "test", "", "", "spm")
	var exitCoder cli.ExitCoder
	suppressOutput(func() {
		err := runSingleService(c, "")
		if !errors.As(err, &exitCoder) {
			t.Fatalf("expected cli.ExitCoder, got %v", err)
		}
		if exitCoder.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitCoder.ExitCode())
		}
	})
}

// TestConfigMissing ensures commands return ExitCoder when config is absent.
func TestConfigMissing(t *testing.T) {
	dir := t.TempDir()
	var exitCoder cli.ExitCoder

	// Test start command
	c := makeContext(dir, "test", "", "", "start")
	suppressOutput(func() {
		err := runWithOptions(c, false)
		if !errors.As(err, &exitCoder) {
			t.Fatalf("expected cli.ExitCoder, got %v", err)
		}
		if exitCoder.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitCoder.ExitCode())
		}
	})

	// Test spm command
	c = makeContext(dir, "test", "", "", "spm", "svc")
	suppressOutput(func() {
		err := runSingleService(c, "svc")
		if !errors.As(err, &exitCoder) {
			t.Fatalf("expected cli.ExitCoder, got %v", err)
		}
		if exitCoder.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitCoder.ExitCode())
		}
	})
}

// TestCompose_NotImplemented ensures compose command returns not implemented error.
func TestCompose_NotImplemented(t *testing.T) {
	dir := t.TempDir()
	c := makeContext(dir, "test", "", "", "compose")
	var exitCoder cli.ExitCoder
	suppressOutput(func() {
		err := runComposeMode(c)
		if !errors.As(err, &exitCoder) {
			t.Fatalf("expected cli.ExitCoder, got %v", err)
		}
		if exitCoder.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitCoder.ExitCode())
		}
	})
}
