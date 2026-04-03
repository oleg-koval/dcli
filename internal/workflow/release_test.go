package workflow_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// workflowPath returns the absolute path to .github/workflows/release.yml
// by resolving relative to this test file's location.
func workflowPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	// internal/workflow/release_test.go → ../../.github/workflows/release.yml
	return filepath.Join(filepath.Dir(filename), "..", "..", ".github", "workflows", "release.yml")
}

// Minimal YAML structs for parsing the GitHub Actions workflow.

type workflow struct {
	Name string              `yaml:"name"`
	On   workflowTrigger     `yaml:"on"`
	Jobs map[string]job      `yaml:"jobs"`
}

type workflowTrigger struct {
	Push pushTrigger `yaml:"push"`
}

type pushTrigger struct {
	Branches []string `yaml:"branches"`
}

type job struct {
	Name  string   `yaml:"name"`
	Needs string   `yaml:"needs"`
	Steps []step   `yaml:"steps"`
}

type step struct {
	Name string            `yaml:"name"`
	ID   string            `yaml:"id"`
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
	Env  map[string]string `yaml:"env"`
	Run  string            `yaml:"run"`
}

// readWorkflow reads and parses the release.yml file into a workflow struct.
func readWorkflow(t *testing.T) *workflow {
	t.Helper()
	data, err := os.ReadFile(workflowPath(t))
	if err != nil {
		t.Fatalf("failed to read release.yml: %v", err)
	}
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("failed to parse release.yml as YAML: %v", err)
	}
	return &wf
}

// findGoreleaserStep returns the goreleaser-action step from the release job, or nil.
func findGoreleaserStep(steps []step) *step {
	for i := range steps {
		if strings.HasPrefix(steps[i].Uses, "goreleaser/goreleaser-action") {
			return &steps[i]
		}
	}
	return nil
}

// TestReleaseWorkflowFileExists verifies the workflow file can be read.
func TestReleaseWorkflowFileExists(t *testing.T) {
	path := workflowPath(t)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("release.yml does not exist at %s: %v", path, err)
	}
}

// TestReleaseWorkflowHasNoConflictMarkers checks that the file contains no
// unresolved git merge-conflict markers. Such markers (<<<<<<, =======, >>>>>>>)
// make the file invalid YAML and would cause the release job to fail at runtime.
// This is a regression test for the conflict introduced in this PR.
func TestReleaseWorkflowHasNoConflictMarkers(t *testing.T) {
	data, err := os.ReadFile(workflowPath(t))
	if err != nil {
		t.Fatalf("failed to read release.yml: %v", err)
	}
	content := string(data)

	conflictMarkers := []string{"<<<<<<<", "=======", ">>>>>>>"}
	for _, marker := range conflictMarkers {
		if strings.Contains(content, marker) {
			t.Errorf("release.yml contains unresolved merge conflict marker %q; resolve the conflict before merging", marker)
		}
	}
}

// TestReleaseWorkflowIsValidYAML verifies the file parses as valid YAML.
// This fails while merge-conflict markers are present.
func TestReleaseWorkflowIsValidYAML(t *testing.T) {
	data, err := os.ReadFile(workflowPath(t))
	if err != nil {
		t.Fatalf("failed to read release.yml: %v", err)
	}
	var out interface{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatalf("release.yml is not valid YAML: %v", err)
	}
}

// TestReleaseWorkflowName checks the workflow has the correct display name.
func TestReleaseWorkflowName(t *testing.T) {
	wf := readWorkflow(t)
	if wf.Name != "Release" {
		t.Errorf("expected workflow name %q, got %q", "Release", wf.Name)
	}
}

// TestReleaseWorkflowTriggersOnMainPush verifies the workflow only fires on
// pushes to the main branch, preventing accidental releases from other branches.
func TestReleaseWorkflowTriggersOnMainPush(t *testing.T) {
	wf := readWorkflow(t)
	branches := wf.On.Push.Branches
	if len(branches) == 0 {
		t.Fatal("expected at least one branch trigger, got none")
	}
	found := false
	for _, b := range branches {
		if b == "main" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected push trigger to include %q, got %v", "main", branches)
	}
}

// TestReleaseWorkflowJobsExist validates that the three required jobs are present.
func TestReleaseWorkflowJobsExist(t *testing.T) {
	wf := readWorkflow(t)
	required := []string{"test-gate", "release", "update-homebrew"}
	for _, name := range required {
		if _, ok := wf.Jobs[name]; !ok {
			t.Errorf("expected job %q to exist in release.yml", name)
		}
	}
}

// TestReleaseJobNeedsTestGate ensures the release job depends on test-gate,
// so releases are blocked when tests or linting fail.
func TestReleaseJobNeedsTestGate(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	if releaseJob.Needs != "test-gate" {
		t.Errorf("release job should depend on test-gate, got %q", releaseJob.Needs)
	}
}

// TestUpdateHomebrewNeedsRelease ensures the update-homebrew job runs only
// after the release job completes successfully.
func TestUpdateHomebrewNeedsRelease(t *testing.T) {
	wf := readWorkflow(t)
	homebrew, ok := wf.Jobs["update-homebrew"]
	if !ok {
		t.Fatal("update-homebrew job not found")
	}
	if homebrew.Needs != "release" {
		t.Errorf("update-homebrew job should depend on release, got %q", homebrew.Needs)
	}
}

// TestReleaseWorkflowGoreleaserStepPresent checks the goreleaser-action step
// exists in the release job steps.
func TestReleaseWorkflowGoreleaserStepPresent(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser/goreleaser-action step not found in release job")
	}
}

// TestReleaseWorkflowGoreleaserArgsContainRelease verifies the args value starts
// with "release", which is the required goreleaser sub-command for publishing.
func TestReleaseWorkflowGoreleaserArgsContainRelease(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser step not found")
	}
	args := s.With["args"]
	if !strings.HasPrefix(args, "release") {
		t.Errorf("goreleaser args should start with %q, got %q", "release", args)
	}
}

// TestReleaseWorkflowGoreleaserArgsContainClean verifies the --clean flag is
// present, ensuring the dist directory is wiped before building.
func TestReleaseWorkflowGoreleaserArgsContainClean(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser step not found")
	}
	args := s.With["args"]
	if !strings.Contains(args, "--clean") {
		t.Errorf("goreleaser args should include --clean flag, got %q", args)
	}
}

// TestReleaseWorkflowGoreleaserArgsVerboseFlag verifies the -v (verbose) flag
// is present in the goreleaser args. This flag was added in the feature branch
// to expose detailed build output in CI logs, which aids debugging release failures.
func TestReleaseWorkflowGoreleaserArgsVerboseFlag(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser step not found")
	}
	args := s.With["args"]
	fields := strings.Fields(args)
	found := false
	for _, f := range fields {
		if f == "-v" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("goreleaser args should include -v verbose flag, got %q", args)
	}
}

// TestReleaseWorkflowGoreleaserArgsExactValue validates the complete goreleaser
// args string matches the intended value from the feature branch: "release --clean -v".
func TestReleaseWorkflowGoreleaserArgsExactValue(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser step not found")
	}
	const wantArgs = "release --clean -v"
	args := s.With["args"]
	if args != wantArgs {
		t.Errorf("goreleaser args: want %q, got %q", wantArgs, args)
	}
}

// TestReleaseWorkflowGoreleaserEnvGithubToken checks that GITHUB_TOKEN is
// configured on the goreleaser step, which is required for publishing releases.
func TestReleaseWorkflowGoreleaserEnvGithubToken(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser step not found")
	}
	if _, ok := s.Env["GITHUB_TOKEN"]; !ok {
		t.Error("goreleaser step is missing GITHUB_TOKEN env var")
	}
}

// TestReleaseWorkflowGoreleaserEnvCurrentTag verifies GORELEASER_CURRENT_TAG is
// set on the goreleaser step, propagating the version generated from git tags.
func TestReleaseWorkflowGoreleaserEnvCurrentTag(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser step not found")
	}
	val, ok := s.Env["GORELEASER_CURRENT_TAG"]
	if !ok {
		t.Error("goreleaser step is missing GORELEASER_CURRENT_TAG env var")
		return
	}
	// The value must reference the version step output.
	if !strings.Contains(val, "steps.version.outputs.version") {
		t.Errorf("GORELEASER_CURRENT_TAG should reference steps.version.outputs.version, got %q", val)
	}
}

// TestReleaseWorkflowVersionStepExists confirms the version-generation step
// (id: version) is present before the goreleaser step in the release job.
func TestReleaseWorkflowVersionStepExists(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	for _, s := range releaseJob.Steps {
		if s.ID == "version" {
			return
		}
	}
	t.Error("release job is missing a step with id: version")
}

// TestReleaseWorkflowVersionStepBeforeGoreleaser verifies that the version step
// appears before the goreleaser step so the tag is available when goreleaser runs.
func TestReleaseWorkflowVersionStepBeforeGoreleaser(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	versionIdx := -1
	goreleaserIdx := -1
	for i, s := range releaseJob.Steps {
		if s.ID == "version" {
			versionIdx = i
		}
		if strings.HasPrefix(s.Uses, "goreleaser/goreleaser-action") {
			goreleaserIdx = i
		}
	}
	if versionIdx == -1 {
		t.Fatal("version step not found")
	}
	if goreleaserIdx == -1 {
		t.Fatal("goreleaser step not found")
	}
	if versionIdx >= goreleaserIdx {
		t.Errorf("version step (index %d) must appear before goreleaser step (index %d)", versionIdx, goreleaserIdx)
	}
}

// TestReleaseWorkflowGoreleaserArgsNoExtraWhitespace is a boundary check
// ensuring the args string has no leading/trailing whitespace that could cause
// unexpected goreleaser CLI parsing errors.
func TestReleaseWorkflowGoreleaserArgsNoExtraWhitespace(t *testing.T) {
	wf := readWorkflow(t)
	releaseJob, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("release job not found")
	}
	s := findGoreleaserStep(releaseJob.Steps)
	if s == nil {
		t.Fatal("goreleaser step not found")
	}
	args := s.With["args"]
	if args != strings.TrimSpace(args) {
		t.Errorf("goreleaser args has unexpected leading/trailing whitespace: %q", args)
	}
}