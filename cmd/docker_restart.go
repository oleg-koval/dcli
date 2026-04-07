package cmd

import (
	"fmt"
	"os"

	"github.com/oleg-koval/dcli/internal/docker"
	"github.com/spf13/cobra"
)

var dockerRestartCmd = &cobra.Command{
	Use:   "restart [services...]",
	Short: "Restart Docker Compose services (preserves data)",
	Long: `Restart stops and starts services without removing volumes or data.

If no services are specified, all services are restarted.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectDir, err := resolveProjectDir()
		if err != nil {
			return err
		}

		// Get services to restart
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
			return fmt.Errorf("no services found to restart")
		}

		// Print target services
		fmt.Println("🔄 Restarting Docker Compose services...")
		fmt.Println("──────────────────────────────")
		fmt.Println("🎯  Target services:")
		for _, service := range services {
			fmt.Printf("  -  %s\n", service)
		}
		fmt.Println("──────────────────────────────")

		stopArgs, upArgs := docker.BuildRestartCommandArgs(services, dockerProfiles...)

		// Stop containers
		fmt.Println("⏸️  Stopping containers (preserving all volumes & data)...")
		if err := dockerHelper.RunCommand(projectDir, stopArgs...); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to stop containers: %v\n", err)
		}

		// Start services
		fmt.Println("──────────────────────────────")
		fmt.Println("🚀 Starting services...")
		if err := dockerHelper.RunCommand(projectDir, upArgs...); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}

		fmt.Println("──────────────────────────────")
		fmt.Println("✅ All services restarted (data preserved)")
		fmt.Println("[✓] Done.")
		return nil
	},
}
