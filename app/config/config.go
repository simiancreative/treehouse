package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServiceConfig holds the name and command for a service.
type ServiceConfig struct {
	Name string
	Cmd  string
}

// HealthEntry defines a health check configuration for a service.
type HealthEntry struct {
	URL             string `yaml:"url"`
	Codes           []int  `yaml:"codes"`
	IntervalSeconds int    `yaml:"interval_seconds"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
}

// ServiceMode represents a specific mode configuration for a service
type ServiceMode struct {
	Command string            `yaml:"command"`
	Env     map[string]string `yaml:"env,omitempty"`
}

// Service represents a complete service configuration
type Service struct {
	Command     string            `yaml:"command"`
	Modes       map[string]string `yaml:"modes,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	HealthCheck HealthEntry       `yaml:"health_check,omitempty"`
}

// Config represents the complete configuration structure
type Config struct {
	CoreServices     map[string]Service `yaml:"core_services"`
	OptionalServices map[string]Service `yaml:"optional_services"`
	GlobalEnv        map[string]string  `yaml:"global_env,omitempty"`
}

// LoadConfig loads the consolidated configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetServiceConfig returns the service configuration for a given mode
func (c *Config) GetServiceConfig(serviceName, mode string) (*ServiceConfig, error) {
	// Check core services first
	if svc, ok := c.CoreServices[serviceName]; ok {
		cmd := svc.Command
		if mode != "" {
			if modeCmd, ok := svc.Modes[mode]; ok {
				cmd = modeCmd
			}
		}
		return &ServiceConfig{
			Name: serviceName,
			Cmd:  cmd,
		}, nil
	}

	// Check optional services
	if svc, ok := c.OptionalServices[serviceName]; ok {
		cmd := svc.Command
		if mode != "" {
			if modeCmd, ok := svc.Modes[mode]; ok {
				cmd = modeCmd
			}
		}
		return &ServiceConfig{
			Name: serviceName,
			Cmd:  cmd,
		}, nil
	}

	return nil, fmt.Errorf("service %s not found", serviceName)
}

// GetHealthCheck returns the health check configuration for a service
func (c *Config) GetHealthCheck(serviceName string) (*HealthEntry, error) {
	if svc, ok := c.CoreServices[serviceName]; ok {
		return &svc.HealthCheck, nil
	}
	if svc, ok := c.OptionalServices[serviceName]; ok {
		return &svc.HealthCheck, nil
	}
	return nil, fmt.Errorf("health check for service %s not found", serviceName)
}

// GetEnv returns the combined environment variables for a service and mode
func (c *Config) GetEnv(serviceName, mode string) map[string]string {
	env := make(map[string]string)

	// Add global environment variables
	for k, v := range c.GlobalEnv {
		env[k] = v
	}

	// Add service-specific environment variables
	if svc, ok := c.CoreServices[serviceName]; ok {
		for k, v := range svc.Env {
			env[k] = v
		}
	} else if svc, ok := c.OptionalServices[serviceName]; ok {
		for k, v := range svc.Env {
			env[k] = v
		}
	}

	return env
}
