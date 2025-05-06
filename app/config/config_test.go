package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	content := `core_services:
  web:
    command: "run-web"
    health_check:
      url: "http://localhost:3000"
      codes: [200]
      interval_seconds: 5
      timeout_seconds: 2
    env:
      PORT: "3000"
  worker:
    command: "run-worker"
    health_check:
      url: "http://localhost:3001"
      codes: [200]
      interval_seconds: 5
      timeout_seconds: 2

optional_services:
  cache:
    command: "run-cache"
    health_check:
      url: "http://localhost:6379"
      codes: [200]
      interval_seconds: 5
      timeout_seconds: 2

global_env:
  ENVIRONMENT: "test"
  LOG_LEVEL: "debug"
`
	fname := filepath.Join(dir, "treehouse.yaml")
	if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	config, err := LoadConfig(fname)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test core services
	if len(config.CoreServices) != 2 {
		t.Fatalf("expected 2 core services, got %d", len(config.CoreServices))
	}

	// Test web service
	web, ok := config.CoreServices["web"]
	if !ok {
		t.Fatal("missing web service")
	}
	if web.Command != "run-web" {
		t.Errorf("web command: got %s, want run-web", web.Command)
	}
	if web.HealthCheck.URL != "http://localhost:3000" {
		t.Errorf("web health check URL: got %s, want http://localhost:3000", web.HealthCheck.URL)
	}
	if len(web.HealthCheck.Codes) != 1 || web.HealthCheck.Codes[0] != 200 {
		t.Errorf("web health check codes: got %v, want [200]", web.HealthCheck.Codes)
	}
	if web.Env["PORT"] != "3000" {
		t.Errorf("web env PORT: got %s, want 3000", web.Env["PORT"])
	}

	// Test worker service
	worker, ok := config.CoreServices["worker"]
	if !ok {
		t.Fatal("missing worker service")
	}
	if worker.Command != "run-worker" {
		t.Errorf("worker command: got %s, want run-worker", worker.Command)
	}

	// Test optional services
	if len(config.OptionalServices) != 1 {
		t.Fatalf("expected 1 optional service, got %d", len(config.OptionalServices))
	}
	cache, ok := config.OptionalServices["cache"]
	if !ok {
		t.Fatal("missing cache service")
	}
	if cache.Command != "run-cache" {
		t.Errorf("cache command: got %s, want run-cache", cache.Command)
	}

	// Test global env
	if len(config.GlobalEnv) != 2 {
		t.Fatalf("expected 2 global env vars, got %d", len(config.GlobalEnv))
	}
	if config.GlobalEnv["ENVIRONMENT"] != "test" {
		t.Errorf("global env ENVIRONMENT: got %s, want test", config.GlobalEnv["ENVIRONMENT"])
	}
	if config.GlobalEnv["LOG_LEVEL"] != "debug" {
		t.Errorf("global env LOG_LEVEL: got %s, want debug", config.GlobalEnv["LOG_LEVEL"])
	}
}

func TestGetServiceConfig(t *testing.T) {
	config := &Config{
		CoreServices: map[string]Service{
			"web": {
				Command: "run-web",
				Modes: map[string]string{
					"prod": "run-web --prod",
				},
			},
		},
		OptionalServices: map[string]Service{
			"worker": {
				Command: "run-worker",
			},
		},
	}

	// Test core service
	svc, err := config.GetServiceConfig("web", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.Name != "web" || svc.Cmd != "run-web" {
		t.Errorf("got %+v, want {Name: web, Cmd: run-web}", svc)
	}

	// Test core service with mode
	svc, err = config.GetServiceConfig("web", "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.Name != "web" || svc.Cmd != "run-web --prod" {
		t.Errorf("got %+v, want {Name: web, Cmd: run-web --prod}", svc)
	}

	// Test optional service
	svc, err = config.GetServiceConfig("worker", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.Name != "worker" || svc.Cmd != "run-worker" {
		t.Errorf("got %+v, want {Name: worker, Cmd: run-worker}", svc)
	}

	// Test non-existent service
	_, err = config.GetServiceConfig("nonexistent", "")
	if err == nil {
		t.Fatal("expected error for non-existent service")
	}
}

func TestGetHealthCheck(t *testing.T) {
	config := &Config{
		CoreServices: map[string]Service{
			"web": {
				HealthCheck: HealthEntry{
					URL:             "http://localhost:3000",
					Codes:           []int{200},
					IntervalSeconds: 5,
					TimeoutSeconds:  2,
				},
			},
		},
		OptionalServices: map[string]Service{
			"worker": {
				HealthCheck: HealthEntry{
					URL:             "http://localhost:3001",
					Codes:           []int{200},
					IntervalSeconds: 5,
					TimeoutSeconds:  2,
				},
			},
		},
	}

	// Test core service health check
	hc, err := config.GetHealthCheck("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hc.URL != "http://localhost:3000" {
		t.Errorf("got %s, want http://localhost:3000", hc.URL)
	}

	// Test optional service health check
	hc, err = config.GetHealthCheck("worker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hc.URL != "http://localhost:3001" {
		t.Errorf("got %s, want http://localhost:3001", hc.URL)
	}

	// Test non-existent service
	_, err = config.GetHealthCheck("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent service")
	}
}

func TestGetEnv(t *testing.T) {
	config := &Config{
		CoreServices: map[string]Service{
			"web": {
				Env: map[string]string{
					"PORT": "3000",
				},
			},
		},
		OptionalServices: map[string]Service{
			"worker": {
				Env: map[string]string{
					"PORT": "3001",
				},
			},
		},
		GlobalEnv: map[string]string{
			"ENVIRONMENT": "test",
		},
	}

	// Test core service env
	env := config.GetEnv("web", "")
	if env["PORT"] != "3000" {
		t.Errorf("got %s, want 3000", env["PORT"])
	}
	if env["ENVIRONMENT"] != "test" {
		t.Errorf("got %s, want test", env["ENVIRONMENT"])
	}

	// Test optional service env
	env = config.GetEnv("worker", "")
	if env["PORT"] != "3001" {
		t.Errorf("got %s, want 3001", env["PORT"])
	}
	if env["ENVIRONMENT"] != "test" {
		t.Errorf("got %s, want test", env["ENVIRONMENT"])
	}

	// Test non-existent service
	env = config.GetEnv("nonexistent", "")
	if env["ENVIRONMENT"] != "test" {
		t.Errorf("got %s, want test", env["ENVIRONMENT"])
	}
}
