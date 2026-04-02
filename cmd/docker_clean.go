package cmd

import (
	"fmt"
	"os"

	"github.com/oleg-koval/dcli/internal/docker"
	"github.com/spf13/cobra"
)

var dockerCleanCmd = &cobra.Command{
	Use:   "clean [services...]",
	Short: "Clean up and rebuild Docker containers and volumes",
	Long: `Clean removes containers, volumes, and images for specified services,
then rebuilds and restarts them.

If no services are specified, all services are cleaned.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project directory
		projectDir := os.Getenv("DCLI_PROJECT_DIR")
		if projectDir == "" {
			projectDir = "."
		}

		// Get services to clean
		var services []string
		if len(args) > 0 {
			services = args
		} else {
			// Get all available services
			availableServices, err := docker.GetServices(projectDir)
			if err != nil {
				return fmt.Errorf("failed to get services: %w", err)
			}
			services = availableServices
		}

		if len(services) == 0 {
			return fmt.Errorf("no services found to clean")
		}

		// Print target services
		fmt.Println("🎯  Target services:")
		for _, service := range services {
			fmt.Printf("  -  %s\n", service)
		}
		fmt.Println()

		// Remove containers and volumes
		fmt.Println("🧹  Removing containers and volumes...")
		rmArgs := append([]string{"compose", "rm", "-sfv"}, services...)
		if err := docker.RunCommand(projectDir, rmArgs...); err != nil {
			return fmt.Errorf("failed to remove containers: %w", err)
		}
		fmt.Println("✓ Containers and volumes removed")
		fmt.Println()

		// Rebuild images
		fmt.Println("🔨  Building images...")
		buildArgs := append([]string{"compose", "build"}, services...)
		if err := docker.RunCommand(projectDir, buildArgs...); err != nil {
			return fmt.Errorf("failed to build images: %w", err)
		}
		fmt.Println("✓ Images built")
		fmt.Println()

		// Start services
		fmt.Println("🚀  Starting services...")
		upArgs := append([]string{"compose", "up", "-d"}, services...)
		if err := docker.RunCommand(projectDir, upArgs...); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}
		fmt.Println("✓ Services started")
		fmt.Println()

		fmt.Println("✨  Clean and rebuild complete!")
		return nil
	},
}
