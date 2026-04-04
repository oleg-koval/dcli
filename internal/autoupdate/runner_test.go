package autoupdate

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"
)

type fakeClient struct {
	latest           *Release
	latestErr        error
	updateErr        error
	latestCalled     bool
	updateCalled     bool
	updateRelease    *Release
	updateExecutable string
}

func (f *fakeClient) LatestRelease(_ context.Context, _ Repository) (*Release, error) {
	f.latestCalled = true
	return f.latest, f.latestErr
}

func (f *fakeClient) UpdateTo(_ context.Context, release *Release, executable string) error {
	f.updateCalled = true
	f.updateRelease = release
	f.updateExecutable = executable
	return f.updateErr
}

type fakeRestarter struct {
	called bool
	exe    string
	args   []string
	env    []string
	err    error
}

func (f *fakeRestarter) restart(exe string, args []string, env []string) error {
	f.called = true
	f.exe = exe
	f.args = append([]string{}, args...)
	f.env = append([]string{}, env...)
	return f.err
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   string
		want string
	}{
		"empty":      {in: "", want: ""},
		"plain":      {in: "1.2.3", want: "v1.2.3"},
		"prefixed":   {in: "v1.2.3", want: "v1.2.3"},
		"whitespace": {in: " 1.2.3 ", want: "v1.2.3"},
		"invalid":    {in: "latest", want: ""},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeVersion(tc.in); got != tc.want {
				t.Fatalf("normalizeVersion(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNewRunnerDefaultTimeout(t *testing.T) {
	t.Setenv(TimeoutEnvVar, "")

	runner := NewRunner(Repository{Owner: "oleg-koval", Name: "dcli"})
	if runner.Timeout != defaultTimeout {
		t.Fatalf("expected default timeout %s, got %s", defaultTimeout, runner.Timeout)
	}
}

func TestNewRunnerTimeoutFromEnv(t *testing.T) {
	t.Setenv(TimeoutEnvVar, "250ms")

	runner := NewRunner(Repository{Owner: "oleg-koval", Name: "dcli"})
	if runner.Timeout != 250*time.Millisecond {
		t.Fatalf("expected timeout 250ms, got %s", runner.Timeout)
	}
}

func TestNewRunnerTimeoutFromEnvInvalidFallsBack(t *testing.T) {
	t.Setenv(TimeoutEnvVar, "not-a-duration")

	runner := NewRunner(Repository{Owner: "oleg-koval", Name: "dcli"})
	if runner.Timeout != defaultTimeout {
		t.Fatalf("expected default timeout %s, got %s", defaultTimeout, runner.Timeout)
	}
}

func TestSelectReleaseAsset(t *testing.T) {
	t.Parallel()

	project := "dcli"
	version := "v1.2.3"
	assetName := assetCandidates(project, version, runtime.GOOS, runtime.GOARCH)[0]
	release := &githubRelease{
		TagName: version,
		Assets: []githubAsset{
			{
				Name:               assetName,
				BrowserDownloadURL: "https://example.com/download",
			},
		},
	}

	asset, err := selectReleaseAsset(release, project, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("selectReleaseAsset returned error: %v", err)
	}
	if asset.Name != assetName {
		t.Fatalf("expected asset %q, got %q", assetName, asset.Name)
	}
	if asset.BrowserDownloadURL != "https://example.com/download" {
		t.Fatalf("unexpected asset url %q", asset.BrowserDownloadURL)
	}
}

func TestSelectReleaseAssetMissing(t *testing.T) {
	t.Parallel()

	release := &githubRelease{
		TagName: "v1.2.3",
		Assets: []githubAsset{
			{Name: "other-asset.tar.gz"},
		},
	}

	if _, err := selectReleaseAsset(release, "dcli", runtime.GOOS, runtime.GOARCH); err == nil {
		t.Fatal("expected error when asset is missing")
	}
}

func TestGitHubClientLatestRelease(t *testing.T) {
	t.Parallel()

	assetName := assetCandidates("dcli", "v1.2.3", runtime.GOOS, runtime.GOARCH)[0]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/oleg-koval/dcli/releases/latest" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v1.2.3",
			"assets": [
				{
					"name": "` + assetName + `",
					"browser_download_url": "https://example.com/download"
				}
			]
		}`))
	}))
	defer server.Close()

	client := &GitHubClient{BaseURL: server.URL, HTTPClient: server.Client()}
	release, err := client.LatestRelease(context.Background(), Repository{Owner: "oleg-koval", Name: "dcli"})
	if err != nil {
		t.Fatalf("LatestRelease returned error: %v", err)
	}
	if release.Version != "v1.2.3" {
		t.Fatalf("expected version v1.2.3, got %q", release.Version)
	}
	if release.AssetName != assetName {
		t.Fatalf("expected asset name %q, got %q", assetName, release.AssetName)
	}
}

func TestGitHubClientLatestReleaseNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &GitHubClient{BaseURL: server.URL, HTTPClient: server.Client()}
	if _, err := client.LatestRelease(context.Background(), Repository{Owner: "oleg-koval", Name: "dcli"}); !errors.Is(err, ErrReleaseNotFound) {
		t.Fatalf("expected ErrReleaseNotFound, got %v", err)
	}
}

func TestRunnerSkipsWhenDisabled(t *testing.T) {
	t.Setenv(DisableEnvVar, "1")

	client := &fakeClient{latest: &Release{Version: "v2.0.0"}}
	restarter := &fakeRestarter{}
	runner := &Runner{
		Client:        client,
		Repository:    Repository{Owner: "oleg-koval", Name: "dcli"},
		DisableEnvVar: DisableEnvVar,
		Timeout:       0,
		Executable: func() (string, error) {
			t.Fatal("Executable should not be called when disabled")
			return "", nil
		},
		Environment: os.Environ,
		Restart:     restarter.restart,
	}

	runner.Run(context.Background(), "1.0.0", []string{"dcli", "git", "reset", "develop"})
	if client.latestCalled {
		t.Fatal("expected updater not to be called when disabled")
	}
	if restarter.called {
		t.Fatal("expected restart not to be called when disabled")
	}
}

func TestRunnerNoUpdate(t *testing.T) {
	t.Parallel()

	client := &fakeClient{latest: &Release{Version: "v1.0.0"}}
	restarter := &fakeRestarter{}
	runner := &Runner{
		Client:        client,
		Repository:    Repository{Owner: "oleg-koval", Name: "dcli"},
		DisableEnvVar: DisableEnvVar,
		Timeout:       0,
		Executable: func() (string, error) {
			return "/tmp/dcli", nil
		},
		Environment: os.Environ,
		Restart:     restarter.restart,
	}

	runner.Run(context.Background(), "1.0.0", []string{"dcli"})
	if !client.latestCalled {
		t.Fatal("expected updater to be called")
	}
	if client.updateCalled {
		t.Fatal("expected update not to be called when current version is latest")
	}
	if restarter.called {
		t.Fatal("expected restart not to be called")
	}
}

func TestRunnerUpdateTriggersRestart(t *testing.T) {
	t.Parallel()

	client := &fakeClient{
		latest: &Release{
			Version:   "v2.0.0",
			AssetURL:  "https://example.com/download",
			AssetName: "dcli-v2.0.0-linux-amd64.tar.gz",
		},
	}
	restarter := &fakeRestarter{}
	runner := &Runner{
		Client:        client,
		Repository:    Repository{Owner: "oleg-koval", Name: "dcli"},
		DisableEnvVar: DisableEnvVar,
		Timeout:       0,
		Executable: func() (string, error) {
			if runtime.GOOS == "windows" {
				return `C:\tmp\dcli.exe`, nil
			}
			return "/tmp/dcli", nil
		},
		Environment: func() []string {
			return []string{"PATH=/usr/bin"}
		},
		Restart: restarter.restart,
	}

	runner.Run(context.Background(), "1.0.0", []string{"dcli", "docker", "clean"})
	if !client.updateCalled {
		t.Fatal("expected update to be called")
	}
	if !restarter.called {
		t.Fatal("expected restart to be called after update")
	}
	if len(restarter.args) == 0 || restarter.args[0] != "dcli" {
		t.Fatalf("expected restart args to preserve argv, got %v", restarter.args)
	}
	if len(restarter.env) == 0 {
		t.Fatal("expected restart env to be populated")
	}
	found := false
	for _, env := range restarter.env {
		if env == DisableEnvVar+"=1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected restart env to include %s=1", DisableEnvVar)
	}
}

func TestRunnerUpdateFailureIgnored(t *testing.T) {
	t.Parallel()

	client := &fakeClient{
		latest: &Release{
			Version:   "v2.0.0",
			AssetURL:  "https://example.com/download",
			AssetName: "dcli-v2.0.0-linux-amd64.tar.gz",
		},
		updateErr: errors.New("network down"),
	}
	restarter := &fakeRestarter{}
	runner := &Runner{
		Client:        client,
		Repository:    Repository{Owner: "oleg-koval", Name: "dcli"},
		DisableEnvVar: DisableEnvVar,
		Timeout:       0,
		Executable: func() (string, error) {
			return "/tmp/dcli", nil
		},
		Environment: os.Environ,
		Restart:     restarter.restart,
	}

	runner.Run(context.Background(), "1.0.0", []string{"dcli"})
	if !client.updateCalled {
		t.Fatal("expected update to be called")
	}
	if restarter.called {
		t.Fatal("expected restart not to be called after update failure")
	}
}

func TestRunnerLatestReleaseErrorIgnored(t *testing.T) {
	t.Parallel()

	client := &fakeClient{latestErr: errors.New("network down")}
	restarter := &fakeRestarter{}
	runner := &Runner{
		Client:        client,
		Repository:    Repository{Owner: "oleg-koval", Name: "dcli"},
		DisableEnvVar: DisableEnvVar,
		Timeout:       0,
		Executable: func() (string, error) {
			t.Fatal("Executable should not be called when lookup fails")
			return "", nil
		},
		Environment: os.Environ,
		Restart:     restarter.restart,
	}

	runner.Run(context.Background(), "1.0.0", []string{"dcli"})
	if !client.latestCalled {
		t.Fatal("expected latest release lookup to be attempted")
	}
	if client.updateCalled {
		t.Fatal("expected update not to be called when lookup fails")
	}
	if restarter.called {
		t.Fatal("expected restart not to be called")
	}
}
