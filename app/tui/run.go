package tui

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/simiancreative/treehouse/app/config"
	"github.com/simiancreative/treehouse/app/health"
	"github.com/simiancreative/treehouse/app/service"

	tea "github.com/charmbracelet/bubbletea"
)

// LogMsg carries a single log line from a service.
type LogMsg struct {
	Service string
	Line    string
}

// StatusMsg updates the status of a service.
type StatusMsg struct {
	Service string
	Status  string
}

func serviceTextHandler(svcName string, p *tea.Program) func(string) {
	return func(line string) {
		p.Send(LogMsg{Service: svcName, Line: line})
	}
}

func statusCallbackHandler(svcName string, p *tea.Program) func(string) {
	return func(status string) {
		p.Send(StatusMsg{Service: svcName, Status: status})
	}
}

// Run initializes and runs the interactive TUI, orchestrating service processes and health checks.
//
// 1. Load services and health entries
// 2. Initialize Bubble Tea model and program
// 3. Setup cancellation context for subprocesses
// 4. Launch each service in its own goroutine:
//   - Send status updates (starting, running, crashed, exited)
//   - Stream stdout/stderr as log messages
//
// 5. Launch health check goroutines:
//   - Poll URLs until healthy or timeout, sending status updates
//
// 6. Start the TUI event loop (blocking)
func Run(configDir, mode, focus, mute string) error {
	// Load the consolidated configuration
	cfg, err := config.LoadConfig(configDir + "/treehouse.yaml")
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Set environment variables
	for k, v := range cfg.GlobalEnv {
		os.Setenv(k, v)
	}

	// Convert core services to service configs
	var services []config.ServiceConfig
	var healthChecks = make(map[string]config.HealthEntry)

	for name, svc := range cfg.CoreServices {
		serviceConfig, err := cfg.GetServiceConfig(name, mode)
		if err != nil {
			return fmt.Errorf("getting service config: %w", err)
		}
		services = append(services, *serviceConfig)

		// Set service-specific environment variables
		for k, v := range svc.Env {
			os.Setenv(k, v)
		}

		// Get health check if configured
		if hc, err := cfg.GetHealthCheck(name); err == nil {
			healthChecks[name] = *hc
		}
	}

	// Initialize the TUI model and program
	model := NewModel(services, healthChecks, focus, mute)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Setup cancellation context for subprocesses
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch each service process and stream its output to the TUI
	var svcWg sync.WaitGroup
	svcWg.Add(len(services))

	for _, svc := range services {
		textHandler := serviceTextHandler(svc.Name, p)
		statusHandler := statusCallbackHandler(svc.Name, p)

		svc := svc // Capture the loop variable

		go func() {
			defer svcWg.Done()

			err := service.
				New().
				SetConfig(svc).
				SetStdOutCallback(textHandler).
				SetStdErrCallback(textHandler).
				SetStatusCallback(statusHandler).
				Start(ctx)

			if err != nil {
				p.Send(StatusMsg{Service: svc.Name, Status: "Error"})
			}
		}()
	}

	// Launch health check goroutines that send status updates to the TUI
	for name, entry := range healthChecks {
		name := name
		entry := entry
		go func() {
			interval := entry.IntervalSeconds
			if interval <= 0 {
				interval = health.DefaultHealthInterval
			}
			timeout := entry.TimeoutSeconds
			if timeout <= 0 {
				timeout = health.DefaultHealthTimeout
			}
			start := time.Now()
			for {
				ok, _, err := health.CheckStatus(http.DefaultClient, entry.URL, entry.Codes)
				if err == nil && ok {
					p.Send(StatusMsg{Service: name, Status: "Healthy"})
					return
				}
				if time.Since(start) > time.Duration(timeout)*time.Second {
					p.Send(StatusMsg{Service: name, Status: "Unhealthy"})
					return
				}
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}()
	}

	// Run the Bubble Tea event loop (blocks until the user exits)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error starting TUI: %w", err)
	}

	cancel() // Cancel the context to stop all subprocesses
	svcWg.Wait()

	return nil
}
