package runner

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/simiancreative/treehouse/app/colors"
	"github.com/simiancreative/treehouse/app/config"
	"github.com/simiancreative/treehouse/app/health"
	"github.com/simiancreative/treehouse/app/service"

	"github.com/charmbracelet/lipgloss"
)

// default health check settings (seconds)
const (
	defaultHealthInterval = 2
	defaultHealthTimeout  = 30
)

// Options configures a Runner.
type Options struct {
	ConfigDir             string
	Mode                  string
	Focus, Mute           string
	Colors                []string
	DefaultHealthInterval int
	DefaultHealthTimeout  int
	HTTPClient            health.HTTPClient
	SPMMode               bool // When true, only run health checks for the focused service
}

// Runner orchestrates services and health checks.
type Runner struct {
	opts Options
}

// NewRunner creates a Runner with provided options, filling defaults.
func New(opts Options) *Runner {
	if len(opts.Colors) == 0 {
		opts.Colors = colors.ServiceColors
	}
	if opts.DefaultHealthInterval <= 0 {
		opts.DefaultHealthInterval = defaultHealthInterval
	}
	if opts.DefaultHealthTimeout <= 0 {
		opts.DefaultHealthTimeout = defaultHealthTimeout
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}
	return &Runner{opts: opts}
}

// Run executes the environment setup, starts services, performs health checks, and waits.
func (r *Runner) Run(ctx context.Context) error {
	// Load the consolidated configuration
	cfg, err := config.LoadConfig(r.opts.ConfigDir + "/treehouse.yaml")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Set environment variables
	for k, v := range cfg.GlobalEnv {
		os.Setenv(k, v)
	}

	// Convert core services to service configs
	var svcs []config.ServiceConfig
	for name, svc := range cfg.CoreServices {
		serviceConfig, err := cfg.GetServiceConfig(name, r.opts.Mode)
		if err != nil {
			return fmt.Errorf("getting service config: %w", err)
		}
		svcs = append(svcs, *serviceConfig)

		// Set service-specific environment variables
		for k, v := range svc.Env {
			os.Setenv(k, v)
		}
	}

	// Start all services concurrently
	var svcWg sync.WaitGroup
	svcWg.Add(len(svcs))
	for i, svc := range svcs {
		color := r.opts.Colors[i%len(r.opts.Colors)]
		go func(s config.ServiceConfig, color string) {
			defer svcWg.Done()
			r.startService(ctx, s, color)
		}(svc, color)
	}

	// Perform health checks for services
	var hcWg sync.WaitGroup
	for i, svc := range svcs {
		// In SPM mode, only run health checks for the focused service
		if r.opts.SPMMode && svc.Name != r.opts.Focus {
			continue
		}

		hc, err := cfg.GetHealthCheck(svc.Name)
		if err != nil {
			continue // Skip health check if not configured
		}
		color := r.opts.Colors[i%len(r.opts.Colors)]
		hcWg.Add(1)
		go func(name string, entry config.HealthEntry, color string) {
			defer hcWg.Done()
			r.startHealth(ctx, name, entry, color)
		}(svc.Name, *hc, color)
	}
	hcWg.Wait()

	// Wait for all service processes to exit before returning
	svcWg.Wait()
	return nil
}

func serviceTextHandler(svc config.ServiceConfig, color string) func(string) {
	// Prepare a lipgloss style for this service
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	// Prefix each line with styled [service]
	return func(text string) {
		fmt.Println(style.Render("["+svc.Name+"]") + " " + text)
	}
}

// startService starts a service process and streams its output.
//   - Launches the command as a subprocess tied to the provided context (cancellable).
//   - Attaches to both stdout and stderr pipes.
//   - Processes each output line with service.ProcessStream (applying focus/mute filters).
//   - Prints each line prefixed with the service name and colored output.
//   - Waits for all output to be drained and the process to exit.
func (r *Runner) startService(ctx context.Context, svc config.ServiceConfig, color string) {
	textHandler := serviceTextHandler(svc, color)

	err := service.
		New().
		SetConfig(svc).
		SetStdOutCallback(textHandler).
		SetStdErrCallback(textHandler).
		SetFocus(r.opts.Focus).
		SetMute(r.opts.Mute).
		Start(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error for %s: %v\n", svc.Name, err)
	}
}

// startHealth performs health checks for a service until success, timeout, or context done.
//   - Reads or defaults the polling interval and timeout duration.
//   - Repeatedly invokes health.CheckStatus against the configured URL and expected codes.
//   - On first successful status, prints a success message and returns.
//   - If the timeout duration elapses, prints a failure message and returns.
//   - If the context is canceled, prints an aborted message and returns immediately.
func (r *Runner) startHealth(ctx context.Context, svcName string, entry config.HealthEntry, color string) {
	interval := entry.IntervalSeconds
	if interval <= 0 {
		interval = r.opts.DefaultHealthInterval
	}
	timeout := entry.TimeoutSeconds
	if timeout <= 0 {
		timeout = r.opts.DefaultHealthTimeout
	}
	start := time.Now()
	// Prepare a lipgloss style for health messages
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	for {
		select {
		case <-ctx.Done():
			// Context canceled: abort health check
			fmt.Printf("%s aborted\n", style.Render(fmt.Sprintf("[health][%s]", svcName)))
			return
		default:
		}
		ok, code, err := health.CheckStatus(r.opts.HTTPClient, entry.URL, entry.Codes)
		if err == nil && ok {
			// Success
			fmt.Printf("%s success (%d)\n", style.Render(fmt.Sprintf("[health][%s]", svcName)), code)
			return
		}
		if time.Since(start) > time.Duration(timeout)*time.Second {
			// Timeout
			fmt.Printf("%s failure (timeout)\n", style.Render(fmt.Sprintf("[health][%s]", svcName)))
			return
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
