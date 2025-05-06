package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/simiancreative/treehouse"
	"os"
)

func main() {
	var coreService treehouse.Service = &treehouse.CoreService{}

	var rootCmd = &cobra.Command{
		Use:   "treehouse",
		Short: "Treehouse is a CLI tool for orchestrating local development services",
	}

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start all core services",
		Run: func(cmd *cobra.Command, args []string) {
			if err := coreService.Start(); err != nil {
				fmt.Println("Error starting core services:", err)
			} else {
				fmt.Println("Core services started successfully.")
			}
		},
	}

	var spmCmd = &cobra.Command{
		Use:   "spm",
		Short: "Start a single process mode",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting single process mode...")
			// Placeholder for starting a single service logic
		},
	}

	var configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Configure services",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Configuring services...")
			// Placeholder for configuring services logic
		},
	}

	rootCmd.AddCommand(startCmd, spmCmd, configureCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
