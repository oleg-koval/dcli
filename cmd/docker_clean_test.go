package cmd

import (
	"testing"
)

func TestDockerCleanHelp(t *testing.T) {
	dockerCleanCmd.SetArgs([]string{"--help"})
	err := dockerCleanCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDockerCleanNoArgs(t *testing.T) {
	dockerCleanCmd.SetArgs([]string{})
	if dockerCleanCmd.Name() != "clean" {
		t.Fatalf("expected command name 'clean', got %s", dockerCleanCmd.Name())
	}
}

func TestDockerCleanWithServiceArgs(t *testing.T) {
	dockerCleanCmd.SetArgs([]string{"service1", "service2"})
	if dockerCleanCmd.Name() != "clean" {
		t.Fatalf("expected command name 'clean', got %s", dockerCleanCmd.Name())
	}
}
