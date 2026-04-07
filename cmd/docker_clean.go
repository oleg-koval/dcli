package cmd

import (
	"fmt"

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
		projectDir, err := resolveProjectDir()
		if err != nil {
			return err
		}

		// Get services to clean
		var services []string
		if len(args) > 0 {
			services = args
		} else {
			availableServices, err := dockerHelper.GetServices(projectDir, dockerProfiles...)
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

		rmArgs, buildArgs, upArgs := docker.BuildCleanCommandArgs(services, dockerProfiles...)

		// Remove containers and volumes
		fmt.Println("🧹  Removing containers and volumes...")
		if err := dockerHelper.RunCommand(projectDir, rmArgs...); err != nil {
			return fmt.Errorf("failed to remove containers: %w", err)
		}
		fmt.Println("✓ Containers and volumes removed")
		fmt.Println()

		// Rebuild images
		fmt.Println("🔨  Building images...")
		if err := dockerHelper.RunCommand(projectDir, buildArgs...); err != nil {
			return fmt.Errorf("failed to build images: %w", err)
		}
		fmt.Println("✓ Images built")
		fmt.Println()

		// Start services
		fmt.Println("🚀  Starting services...")
		if err := dockerHelper.RunCommand(projectDir, upArgs...); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}
		fmt.Println("✓ Services started")
		fmt.Println()

		fmt.Println("✨  Clean and rebuild complete!")
		return nil
	},
}
