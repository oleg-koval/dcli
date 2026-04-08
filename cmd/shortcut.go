package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

var shortcutCmd = &cobra.Command{
	Use:   "shortcut",
	Short: "Manage shell shortcuts for dcli",
	Long:  `Install or remove short aliases so you can type a single letter instead of "dcli".`,
}

var shortcutInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a shell alias pointing to dcli",
	Long: `Install a shell alias so you can type a short name instead of "dcli".

Examples:
  dcli shortcut install d       # lets you type: d docker restart
  dcli shortcut install ddd     # lets you type: ddd monorepo start

The alias is written to your shell config file (~/.zshrc, ~/.bashrc, etc.).
Run "source ~/.zshrc" (or restart your terminal) after installing.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := validateShortcutName(name); err != nil {
			return err
		}

		configFile, shellName, err := detectShellConfig()
		if err != nil {
			return err
		}

		marker := shortcutMarker(name)
		if hasAlias(configFile, marker) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Shortcut %q already installed in %s.\n", name, configFile)
			return nil
		}

		if err := appendAlias(configFile, name, marker); err != nil {
			return fmt.Errorf("write to %s: %w", configFile, err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(),
			"Installed shortcut %q in %s (%s).\nRun: source %s\nThen use: %s docker restart\n",
			name, configFile, shellName, configFile, name,
		)
		return nil
	},
}

var shortcutUninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Remove a shell alias installed by dcli",
	Long: `Remove a shortcut alias previously installed with "dcli shortcut install".

Example:
  dcli shortcut uninstall d`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := validateShortcutName(name); err != nil {
			return err
		}

		configFile, _, err := detectShellConfig()
		if err != nil {
			return err
		}

		marker := shortcutMarker(name)
		removed, err := removeAlias(configFile, marker)
		if err != nil {
			return fmt.Errorf("update %s: %w", configFile, err)
		}
		if !removed {
			return fmt.Errorf("shortcut %q not found in %s", name, configFile)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(),
			"Removed shortcut %q from %s.\nRun: source %s\n",
			name, configFile, configFile,
		)
		return nil
	},
}

var shortcutListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dcli shortcuts installed in the shell config",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _, err := detectShellConfig()
		if err != nil {
			return err
		}

		shortcuts, err := listAliases(configFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", configFile, err)
		}

		if len(shortcuts) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No dcli shortcuts installed.")
			return nil
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Installed dcli shortcuts in %s:\n", configFile)
		for _, s := range shortcuts {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", s)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shortcutCmd)
	shortcutCmd.AddCommand(shortcutInstallCmd)
	shortcutCmd.AddCommand(shortcutUninstallCmd)
	shortcutCmd.AddCommand(shortcutListCmd)
}

// detectShellConfig returns the path to the user's shell config file and a
// human-readable shell name. It reads $SHELL and falls back to zsh.
func detectShellConfig() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("resolve home directory: %w", err)
	}

	shell := os.Getenv("SHELL")
	base := filepath.Base(shell)
	switch base {
	case "bash":
		// Prefer .bash_profile on macOS, .bashrc elsewhere
		candidate := filepath.Join(home, ".bash_profile")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, "bash", nil
		}
		return filepath.Join(home, ".bashrc"), "bash", nil
	case "fish":
		fishDir := filepath.Join(home, ".config", "fish")
		if err := os.MkdirAll(fishDir, 0755); err != nil {
			return "", "", fmt.Errorf("create fish config directory: %w", err)
		}
		return filepath.Join(fishDir, "config.fish"), "fish", nil
	default:
		return filepath.Join(home, ".zshrc"), "zsh", nil
	}
}

func validateShortcutName(name string) error {
	if name == "" {
		return errors.New("shortcut name cannot be empty")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return fmt.Errorf("shortcut name %q contains invalid character %q", name, string(r))
		}
	}
	if name == "dcli" {
		return errors.New("shortcut name cannot be \"dcli\"")
	}
	return nil
}

const dcliShortcutTag = "# dcli-shortcut:"

func shortcutMarker(name string) string {
	return dcliShortcutTag + name
}

func hasAlias(configFile, marker string) bool {
	f, err := os.Open(configFile)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), marker) {
			return true
		}
	}
	return false
}

func appendAlias(configFile, name, marker string) (err error) {
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	line := fmt.Sprintf("\nalias %s='dcli' %s\n", name, marker)
	_, err = fmt.Fprint(f, line)
	return err
}

func removeAlias(configFile, marker string) (bool, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	lines := strings.Split(string(data), "\n")
	filtered := make([]string, 0, len(lines))
	removed := false
	for _, line := range lines {
		if strings.Contains(line, marker) {
			removed = true
			continue
		}
		filtered = append(filtered, line)
	}

	if !removed {
		return false, nil
	}

	// Collapse multiple consecutive blank lines left by removal
	clean := collapseBlankLines(filtered)
	return true, os.WriteFile(configFile, []byte(strings.Join(clean, "\n")), 0600)
}

func collapseBlankLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	prev := false
	for _, l := range lines {
		blank := strings.TrimSpace(l) == ""
		if blank && prev {
			continue
		}
		out = append(out, l)
		prev = blank
	}
	return out
}

func listAliases(configFile string) ([]string, error) {
	f, err := os.Open(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var shortcuts []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, dcliShortcutTag); idx >= 0 {
			shortcuts = append(shortcuts, strings.TrimSpace(line[:idx]))
		}
	}
	return shortcuts, scanner.Err()
}
