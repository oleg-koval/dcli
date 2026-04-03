# Cmd Package Mocking Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development to execute this plan task-by-task. Each task produces self-contained changes with tests. Fresh subagent per task with spec compliance + code quality review.

**Goal:** Mock Docker and Git helpers in cmd tests to achieve 70%+ code coverage by testing actual command execution paths without requiring Docker/Git to be running.

**Architecture:** Create mock interfaces for docker and git helpers, inject them into commands via a test helper, then write behavioral tests that exercise success and error paths.

**Tech Stack:** Go testing (standard library), interface-based mocking, table-driven tests

---

## Current State
- cmd package: 22% coverage (26 tests, but mostly structural)
- Total project coverage: 45.8% (93 of 203 statements)
- Issue: Tests verify metadata but don't execute command logic

## File Structure

**New files:**
- `cmd/docker.go` - Add helper interface and injection point (2-3 lines)
- `cmd/git.go` - Add helper interface and injection point (2-3 lines)
- `cmd/mocks_test.go` - Mock implementations for Docker and Git helpers
- `cmd/docker_clean_test.go` - Rewrite tests with mocks (expand from 8 to 16+ tests)
- `cmd/docker_restart_test.go` - Rewrite tests with mocks (expand from 8 to 16+ tests)
- `cmd/git_reset_test.go` - Rewrite tests with mocks (expand from 10 to 20+ tests)

---

## Task 1: Create Mock Interfaces and Injection Points

**Files:**
- Modify: `cmd/docker_clean.go`
- Modify: `cmd/docker_restart.go`
- Create: `cmd/mocks_test.go`

- [ ] **Step 1: Review current docker_clean.go structure**

Run: `head -30 cmd/docker_clean.go` to see how it calls helpers

Expected: Command calls `docker.GetServices()`, `docker.RunCommand()`

- [ ] **Step 2: Add interface and injection to docker_clean.go**

At the top of docker_clean.go (after imports, before init), add:

```go
// DockerHelper defines the interface for Docker operations
type DockerHelper interface {
	GetServices(projectDir string) ([]string, error)
	RunCommand(projectDir string, args ...string) error
	GetContainers() ([]string, error)
}

// Global helper - will be overridden in tests
var dockerHelper DockerHelper = &defaultDockerHelper{}

type defaultDockerHelper struct{}

func (d *defaultDockerHelper) GetServices(projectDir string) ([]string, error) {
	return docker.GetServices(projectDir)
}

func (d *defaultDockerHelper) RunCommand(projectDir string, args ...string) error {
	return docker.RunCommand(projectDir, args...)
}

func (d *defaultDockerHelper) GetContainers() ([]string, error) {
	return docker.GetContainers()
}
```

Then replace all `docker.GetServices()` calls with `dockerHelper.GetServices()` and all `docker.RunCommand()` calls with `dockerHelper.RunCommand()` in the RunE function.

- [ ] **Step 3: Add interface and injection to docker_restart.go**

Repeat Step 2 for docker_restart.go (same interface, same pattern)

- [ ] **Step 4: Create mocks_test.go**

```go
package cmd

import "fmt"

// MockDockerHelper implements DockerHelper for testing
type MockDockerHelper struct {
	GetServicesFn   func(projectDir string) ([]string, error)
	RunCommandFn    func(projectDir string, args ...string) error
	GetContainersFn func() ([]string, error)
	Calls           struct {
		GetServices []struct {
			ProjectDir string
		}
		RunCommand []struct {
			ProjectDir string
			Args       []string
		}
		GetContainers []struct{}
	}
}

func (m *MockDockerHelper) GetServices(projectDir string) ([]string, error) {
	m.Calls.GetServices = append(m.Calls.GetServices, struct {
		ProjectDir string
	}{projectDir})
	if m.GetServicesFn != nil {
		return m.GetServicesFn(projectDir)
	}
	return []string{}, nil
}

func (m *MockDockerHelper) RunCommand(projectDir string, args ...string) error {
	m.Calls.RunCommand = append(m.Calls.RunCommand, struct {
		ProjectDir string
		Args       []string
	}{projectDir, args})
	if m.RunCommandFn != nil {
		return m.RunCommandFn(projectDir, args...)
	}
	return nil
}

func (m *MockDockerHelper) GetContainers() ([]string, error) {
	m.Calls.GetContainers = append(m.Calls.GetContainers, struct{}{})
	if m.GetContainersFn != nil {
		return m.GetContainersFn()
	}
	return []string{}, nil
}

// MockGitHelper implements GitHelper for testing
type MockGitHelper struct {
	IsGitRepoFn      func(path string) bool
	CheckoutBranchFn func(path, branch string) error
	ResetHardFn      func(path, branch string) error
	FetchOriginFn    func(path string) error
	Calls            struct {
		IsGitRepo      []struct{ Path string }
		CheckoutBranch []struct {
			Path   string
			Branch string
		}
		ResetHard []struct {
			Path   string
			Branch string
		}
		FetchOrigin []struct{ Path string }
	}
}

func (m *MockGitHelper) IsGitRepo(path string) bool {
	m.Calls.IsGitRepo = append(m.Calls.IsGitRepo, struct{ Path string }{path})
	if m.IsGitRepoFn != nil {
		return m.IsGitRepoFn(path)
	}
	return true
}

func (m *MockGitHelper) CheckoutBranch(path, branch string) error {
	m.Calls.CheckoutBranch = append(m.Calls.CheckoutBranch, struct {
		Path   string
		Branch string
	}{path, branch})
	if m.CheckoutBranchFn != nil {
		return m.CheckoutBranchFn(path, branch)
	}
	return nil
}

func (m *MockGitHelper) ResetHard(path, branch string) error {
	m.Calls.ResetHard = append(m.Calls.ResetHard, struct {
		Path   string
		Branch string
	}{path, branch})
	if m.ResetHardFn != nil {
		return m.ResetHardFn(path, branch)
	}
	return nil
}

func (m *MockGitHelper) FetchOrigin(path string) error {
	m.Calls.FetchOrigin = append(m.Calls.FetchOrigin, struct{ Path string }{path})
	if m.FetchOriginFn != nil {
		return m.FetchOriginFn(path)
	}
	return nil
}

// Helper functions for tests
func setDockerHelper(helper DockerHelper) {
	dockerHelper = helper
}

func resetDockerHelper() {
	dockerHelper = &defaultDockerHelper{}
}

func setGitHelper(helper GitHelper) {
	gitHelper = helper
}

func resetGitHelper() {
	gitHelper = &defaultGitHelper{}
}
```

- [ ] **Step 5: Commit**

```bash
git add cmd/docker_clean.go cmd/docker_restart.go cmd/git_reset.go cmd/mocks_test.go
git commit -m "refactor: add helper interfaces and injection points for testing"
```

---

## Task 2: Update docker_clean_test.go with Mocking

**Files:**
- Modify: `cmd/docker_clean_test.go`

- [ ] **Step 1: Replace test file with mocked behavioral tests**

Replace entire docker_clean_test.go content with:

```go
package cmd

import (
	"os"
	"testing"
)

func TestDockerCleanWithValidServices(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web", "db"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	dockerCleanCmd.SetArgs([]string{"web", "db"})
	err := dockerCleanCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify docker helper was called
	if len(mockHelper.Calls.RunCommand) == 0 {
		t.Error("expected RunCommand to be called")
	}
}

func TestDockerCleanWithNoServices(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web", "db", "cache"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	dockerCleanCmd.SetArgs([]string{})
	err := dockerCleanCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify GetServices was called to get all services
	if len(mockHelper.Calls.GetServices) == 0 {
		t.Error("expected GetServices to be called when no services specified")
	}
}

func TestDockerCleanRunCommandCalled(t *testing.T) {
	runCommandCalled := false
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			runCommandCalled = true
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	dockerCleanCmd.SetArgs([]string{"web"})
	err := dockerCleanCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !runCommandCalled {
		t.Error("expected RunCommand to be called")
	}
}

func TestDockerCleanCommandMetadata(t *testing.T) {
	if dockerCleanCmd.Use != "clean [services...]" {
		t.Errorf("expected Use 'clean [services...]', got %s", dockerCleanCmd.Use)
	}

	if dockerCleanCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if dockerCleanCmd.Long == "" {
		t.Error("expected non-empty Long description")
	}

	if dockerCleanCmd.RunE == nil {
		t.Error("expected RunE function to be defined")
	}
}

func TestDockerCleanProjectDirFromEnv(t *testing.T) {
	oldProjectDir := os.Getenv("DCLI_PROJECT_DIR")
	defer func() {
		if oldProjectDir != "" {
			os.Setenv("DCLI_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("DCLI_PROJECT_DIR")
		}
	}()

	os.Setenv("DCLI_PROJECT_DIR", "/test/path")

	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			if projectDir != "/test/path" {
				t.Errorf("expected projectDir '/test/path', got %s", projectDir)
			}
			return []string{"web"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	dockerCleanCmd.SetArgs([]string{"web"})
	err := dockerCleanCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDockerCleanHelp(t *testing.T) {
	dockerCleanCmd.SetArgs([]string{"--help"})
	err := dockerCleanCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `go test ./cmd -run TestDockerClean -v`

Expected: All TestDockerClean* tests pass

- [ ] **Step 3: Check coverage improved**

Run: `go test ./cmd -coverprofile=cmd-coverage.out && go tool cover -func=cmd-coverage.out | grep docker_clean`

Expected: docker_clean.go coverage > 50%

- [ ] **Step 4: Commit**

```bash
git add cmd/docker_clean_test.go
git commit -m "test: add mocked behavioral tests for docker clean command"
```

---

## Task 3: Update docker_restart_test.go with Mocking

**Files:**
- Modify: `cmd/docker_restart_test.go`

Same pattern as Task 2:
- Add 6-8 mocked behavioral tests
- Test: services processed, RunCommand called with correct args, env vars handled
- Verify > 50% coverage on docker_restart.go

---

## Task 4: Add GitHelper interface to git.go and Update git_reset_test.go

**Files:**
- Modify: `cmd/git.go` - Add GitHelper interface and injection (same pattern as Task 1)
- Modify: `cmd/git_reset_test.go` - Replace with mocked behavioral tests (8-12 tests)

Tests should cover:
- Valid branch reset (develop, acceptance)
- Invalid branch rejection
- Config loading with repos
- Git operations called with correct paths
- Error propagation

---

## Success Criteria

After all tasks:
- [ ] cmd package coverage > 70% (was 22%)
- [ ] All cmd tests pass (60+ tests total)
- [ ] Overall project coverage > 75% (was 45.8%)
- [ ] Mocking pattern is consistent and reusable
- [ ] No flaky tests (no actual Docker/Git required)
- [ ] All tests pass locally and in CI
