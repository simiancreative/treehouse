package cmd

import (
	"fmt"
	"os"

	"github.com/simiancreative/treehouse/app"

	"github.com/urfave/cli/v2"
)

// Execute initializes and runs the CLI application.
func Execute() {
	app := &cli.App{
		Name:  "treehouse",
		Usage: "Development control tool",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-dir", Aliases: []string{"c"}, Value: "configs", Usage: "Directory containing config files"},
			&cli.StringFlag{Name: "mode", Aliases: []string{"m"}, Value: "dev", Usage: "Mode to run (e.g., dev, prod)"},
			&cli.StringFlag{Name: "focus", Aliases: []string{"f"}, Value: "", Usage: "Service to focus on"},
			&cli.StringFlag{Name: "mute", Value: "", Usage: "Service to mute"},
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start all services with full TUI",
				Action: func(c *cli.Context) error {
					return runWithOptions(c, false)
				},
			},
			{
				Name:      "spm",
				Usage:     "Run a single service without TUI",
				UsageText: "treehouse spm SERVICE_NAME",
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return cli.Exit("spm command requires exactly one service name argument", 1)
					}
					serviceName := c.Args().Get(0)
					return runSingleService(c, serviceName)
				},
			},
			{
				Name:  "compose",
				Usage: "Open TUI menu to select services and modes",
				Action: func(c *cli.Context) error {
					return runComposeMode(c)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

// runWithOptions runs the application with the given options
func runWithOptions(c *cli.Context, noTUI bool) error {
	err := app.New().
		SetConfigDir(c.String("config-dir")).
		SetMode(c.String("mode")).
		SetFocus(c.String("focus")).
		SetMute(c.String("mute")).
		SetTUI(noTUI).
		Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return cli.Exit("", 1)
	}

	return nil
}

// runSingleService runs a single service in non-TUI mode
func runSingleService(c *cli.Context, serviceName string) error {
	// Create a new app instance with SPM mode enabled
	err := app.New().
		SetConfigDir(c.String("config-dir")).
		SetMode(c.String("mode")).
		SetFocus(serviceName). // Use focus to select the single service
		SetMute("").           // Mute all other services
		SetTUI(true).          // Disable TUI
		SetSPMMode(true).      // Enable SPM mode to only run health checks for the focused service
		Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return cli.Exit("", 1)
	}

	return nil
}

// runComposeMode runs the application in compose mode with service selection
func runComposeMode(c *cli.Context) error {
	// TODO: Implement compose mode with service selection
	// This will require:
	// 1. Loading the config
	// 2. Creating a TUI menu for service selection
	// 3. Creating a TUI menu for mode selection per service
	// 4. Running the selected services with their chosen modes
	fmt.Println("Compose mode not yet implemented")
	return cli.Exit("Not implemented", 1)
}
