package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

// Execute runs a command definition step by step.
func Execute(ctx context.Context, command Command, args []string, stdout, stderr io.Writer) error {
	if err := command.Validate(); err != nil {
		return err
	}

	runnable := command.Clone()
	if len(args) > 0 {
		last := lastExecStepIndex(runnable.Steps)
		if last < 0 {
			return fmt.Errorf("command %s does not accept arguments", runnable.Key())
		}
		runnable.Steps[last].Command = append(runnable.Steps[last].Command, args...)
	}

	for _, step := range runnable.Steps {
		if err := executeStep(ctx, step, stdout, stderr); err != nil {
			return err
		}
	}
	return nil
}

func executeStep(ctx context.Context, step Step, stdout, stderr io.Writer) error {
	switch step.Type {
	case StepTypeExec:
		return runExec(ctx, step, stdout, stderr)
	case StepTypeShell:
		return runShell(ctx, step, stdout, stderr)
	default:
		return fmt.Errorf("unsupported step type %q", step.Type)
	}
}

func runExec(ctx context.Context, step Step, stdout, stderr io.Writer) error {
	//nolint:gosec // G204: argv comes from the user's own command pack; running it is the feature.
	cmd := exec.CommandContext(ctx, step.Command[0], step.Command[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), envPairs(step.Env)...)
	if step.Dir != "" {
		cmd.Dir = step.Dir
	}
	return cmd.Run()
}

func runShell(ctx context.Context, step Step, stdout, stderr io.Writer) error {
	shell, shellArgs := defaultShell(step.Script)
	//nolint:gosec // G204: shell and script come from the user's pack or OS defaults.
	cmd := exec.CommandContext(ctx, shell, shellArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), envPairs(step.Env)...)
	if step.Dir != "" {
		cmd.Dir = step.Dir
	}
	return cmd.Run()
}

func defaultShell(script string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", script}
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	return shell, []string{"-lc", script}
}

func envPairs(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		pairs = append(pairs, key+"="+value)
	}
	return pairs
}

func lastExecStepIndex(steps []Step) int {
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].Type == StepTypeExec {
			return i
		}
	}
	return -1
}

