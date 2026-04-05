package cmd

import (
	"fmt"
	"strings"

	"github.com/oleg-koval/dcli/internal/commands"
	"github.com/spf13/cobra"
)

func splitPathAndStep(cmd *cobra.Command, args []string) ([]string, []string, error) {
	dash := cmd.ArgsLenAtDash()
	if dash < 0 {
		return nil, nil, fmt.Errorf("use -- to separate the command path from the step command")
	}
	if dash == 0 {
		return nil, nil, fmt.Errorf("command path is required before --")
	}
	if dash >= len(args) {
		return nil, nil, fmt.Errorf("step command is required after --")
	}

	path := append([]string(nil), args[:dash]...)
	step := append([]string(nil), args[dash:]...)
	return path, step, nil
}

func buildSingleStepCommand(path []string, description string, step []string, shell bool) commands.Command {
	command := commands.Command{
		Path:        append([]string(nil), path...),
		Description: description,
		Enabled:     true,
		Revision:    1,
		Steps: []commands.Step{
			{Type: commands.StepTypeExec, Command: append([]string(nil), step...)},
		},
	}
	if shell {
		command.Steps[0] = commands.Step{Type: commands.StepTypeShell, Script: strings.Join(step, " ")}
	}
	return command
}

func scopeFromFlags(shared bool) string {
	if shared {
		return commands.ScopeShared
	}
	return commands.ScopeLocal
}
