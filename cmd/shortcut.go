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

// init adds the `shortcut` command and its `install`, `uninstall`, and `list` subcommands to the root command.
func init() {
	rootCmd.AddCommand(shortcutCmd)
	shortcutCmd.AddCommand(shortcutInstallCmd)
	shortcutCmd.AddCommand(shortcutUninstallCmd)
	shortcutCmd.AddCommand(shortcutListCmd)
}

// detectShellConfig returns the path to the user's shell config file and a
// detectShellConfig returns the path to the user's shell configuration file and a human-readable shell name.
// It inspects the SHELL environment variable to choose the config: for bash it prefers ~/.bash_profile if present
// otherwise ~/.bashrc; for fish it ensures ~/.config/fish exists and returns ~/.config/fish/config.fish; for any
// other shell it returns ~/.zshrc. It returns an error if the home directory cannot be resolved or if creating the
// fish config directory fails.
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

// validateShortcutName reports an error if name is not a valid shortcut alias.
// It enforces a non-empty name, allows only Unicode letters, digits, '_' and '-', and disallows the reserved name "dcli".
// The returned error describes the first validation violation encountered.
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

// shortcutMarker returns the marker string used to identify a dcli-managed alias for the given shortcut name.
func shortcutMarker(name string) string {
	return dcliShortcutTag + name
}

// hasAlias reports whether the specified configFile contains a line with the given marker.
// It returns true if any line contains marker, and false if the file cannot be opened or no matching line is found.
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

// appendAlias appends an alias entry for the given shortcut name to the specified
// shell configuration file, creating the file if it does not exist.
//
// The appended line includes the provided marker so the alias can be identified
// and managed later. It returns any error encountered while opening or writing
// the file.
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

// removeAlias removes all lines containing marker from the file at configFile.
// If one or more lines are removed the function rewrites the file with
// consecutive blank lines collapsed.
// It returns true if the file was modified, false if the file did not exist
// or no matching lines were found. Any read or write error is returned.
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

// collapseBlankLines collapses consecutive blank lines into a single blank line.
// A line is considered blank when trimming whitespace yields an empty string.
// Non-blank lines and single blank lines are preserved, and the relative order of lines is unchanged.
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

// listAliases returns the dcli shortcut names found in the given shell config file.
// It scans the file for lines containing the package marker prefix (dcliShortcutTag) and
// collects the text preceding the marker (trimmed of surrounding space) for each match.
// If the file does not exist, it returns (nil, nil). If a scan or read error occurs,
// that error is returned.
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
