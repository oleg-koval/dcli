package cmd

import (
	"fmt"
	"os"

	"github.com/oleg-koval/dcli/internal/config"
	"github.com/oleg-koval/dcli/internal/git"
	"github.com/spf13/cobra"
)

var gitResetCmd = &cobra.Command{
	Use:   "reset [develop|acceptance]",
	Short: "Reset all configured repositories to a branch",
	Long:  "Fetch, checkout, and hard reset all repositories to specified branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]
		if err := git.ValidateBranchTarget(branch); err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repositories configured in ~/.dcli/config.yaml")
		}

		fmt.Printf("🔄 Resetting repositories to %s branch...\n", branch)
		fmt.Println("")

		successCount := 0
		var hasFailures bool
		for _, repo := range cfg.Repositories {
			fmt.Printf("🔄 Processing %s...\n", repo.Name)

			if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
				fmt.Printf("  ⚠️  Directory not found: %s\n", repo.Path)
				fmt.Println("")
				hasFailures = true
				continue
			}

			if !gitHelper.IsGitRepo(repo.Path) {
				fmt.Printf("  ⚠️  Not a git repository: %s\n", repo.Path)
				fmt.Println("")
				hasFailures = true
				continue
			}

			fmt.Println("  📥 Fetching from origin...")
			if err := gitHelper.FetchOrigin(repo.Path); err != nil {
				fmt.Printf("  ❌ Failed to fetch from origin: %v\n", err)
				fmt.Println("")
				hasFailures = true
				continue
			}

			fmt.Printf("  🔀 Checking out %s...\n", branch)
			if err := gitHelper.CheckoutBranch(repo.Path, branch); err != nil {
				fmt.Printf("  ❌ Failed to checkout %s: %v\n", branch, err)
				fmt.Println("")
				hasFailures = true
				continue
			}

			fmt.Printf("  🔄 Resetting to origin/%s...\n", branch)
			if err := gitHelper.ResetHard(repo.Path, branch); err != nil {
				fmt.Printf("  ❌ Failed to reset to origin/%s: %v\n", branch, err)
				fmt.Println("")
				hasFailures = true
				continue
			}

			fmt.Printf("  ✅ %s reset to origin/%s\n", repo.Name, branch)
			fmt.Println("")
			successCount++
		}

		fmt.Printf("🎉 Done! Successfully reset %d/%d repositories\n", successCount, len(cfg.Repositories))
		if hasFailures {
			return fmt.Errorf("some repositories failed to reset")
		}
		return nil
	},
}
