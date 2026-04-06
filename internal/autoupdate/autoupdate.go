package autoupdate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"golang.org/x/mod/semver"
)

const DisableEnvVar = "DCLI_DISABLE_AUTO_UPDATE"

const TimeoutEnvVar = "DCLI_AUTO_UPDATE_TIMEOUT"

// ChannelEnvVar selects which GitHub Releases channel the startup updater considers.
// stable (default): GitHub "latest" non-prerelease only.
// prerelease: newest semver among GitHub prereleases with a matching asset.
// beta / alpha: like prerelease but tag must contain "-beta" or "-alpha" (case-insensitive).
const ChannelEnvVar = "DCLI_AUTO_UPDATE_CHANNEL"

const defaultTimeout = 1 * time.Second

const releasesListPerPage = 30

const releasesListMaxPages = 5

// ReleaseChannel controls which GitHub releases auto-update may install.
type ReleaseChannel string

const (
	ChannelStable     ReleaseChannel = "stable"
	ChannelPrerelease ReleaseChannel = "prerelease"
	ChannelBeta       ReleaseChannel = "beta"
	ChannelAlpha      ReleaseChannel = "alpha"
)

const githubAPIBaseURL = "https://api.github.com"

var ErrReleaseNotFound = errors.New("latest release not found")

type Repository struct {
	Owner string
	Name  string
}

type Release struct {
	Version   string
	AssetURL  string
	AssetName string
}

type Client interface {
	LatestRelease(ctx context.Context, repository Repository, channel ReleaseChannel) (*Release, error)
	UpdateTo(ctx context.Context, release *Release, executable string) error
}

type Runner struct {
	Client         Client
	Repository     Repository
	ReleaseChannel ReleaseChannel
	DisableEnvVar  string
	Timeout        time.Duration
	Executable     func() (string, error)
	Environment    func() []string
	Restart        func(exe string, args []string, env []string) error
}

func NewRunner(repository Repository) *Runner {
	return &Runner{
		Client:         NewGitHubClient(),
		Repository:     repository,
		ReleaseChannel: releaseChannelFromEnv(),
		DisableEnvVar:  DisableEnvVar,
		Timeout:        defaultRunnerTimeout(),
		Executable:     os.Executable,
		Environment:    os.Environ,
		Restart:        restartBinary,
	}
}

func releaseChannelFromEnv() ReleaseChannel {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(ChannelEnvVar)))
	switch raw {
	case "", string(ChannelStable):
		return ChannelStable
	case string(ChannelPrerelease), "pre":
		return ChannelPrerelease
	case string(ChannelBeta):
		return ChannelBeta
	case string(ChannelAlpha):
		return ChannelAlpha
	default:
		return ChannelStable
	}
}

func defaultRunnerTimeout() time.Duration {
	raw, ok := os.LookupEnv(TimeoutEnvVar)
	if !ok {
		return defaultTimeout
	}

	timeout, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || timeout <= 0 {
		return defaultTimeout
	}

	return timeout
}

func (r *Runner) Run(ctx context.Context, currentVersion string, args []string) {
	if r == nil || r.Client == nil {
		return
	}

	if r.DisableEnvVar != "" {
		if _, ok := os.LookupEnv(r.DisableEnvVar); ok {
			return
		}
	}

	currentVersion = normalizeVersion(currentVersion)
	if currentVersion == "" {
		return
	}

	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	ch := r.ReleaseChannel
	if ch == "" {
		ch = ChannelStable
	}

	release, err := r.Client.LatestRelease(ctx, r.Repository, ch)
	if err != nil || release == nil {
		return
	}

	latestVersion := normalizeVersion(release.Version)
	if latestVersion == "" || semver.Compare(latestVersion, currentVersion) <= 0 {
		return
	}

	if r.Executable == nil || r.Restart == nil {
		return
	}

	executable, err := r.Executable()
	if err != nil || executable == "" {
		return
	}

	if err := r.Client.UpdateTo(ctx, release, executable); err != nil {
		return
	}

	env := []string{}
	if r.Environment != nil {
		env = append(env, r.Environment()...)
	}
	if r.DisableEnvVar != "" {
		env = append(env, r.DisableEnvVar+"=1")
	}

	if err := r.Restart(executable, args, env); err != nil {
		log.Printf(
			"auto-update restart failed for %s/%s version=%s executable=%q args=%q: %v",
			r.Repository.Owner,
			r.Repository.Name,
			release.Version,
			executable,
			args,
			err,
		)
	}
}

type GitHubClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		BaseURL: githubAPIBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *GitHubClient) LatestRelease(ctx context.Context, repository Repository, channel ReleaseChannel) (*Release, error) {
	if repository.Owner == "" || repository.Name == "" {
		return nil, fmt.Errorf("repository owner and name are required")
	}

	if channel == "" {
		channel = ChannelStable
	}

	var release *githubRelease
	var err error
	switch channel {
	case ChannelStable:
		release, err = c.fetchLatestStableRelease(ctx, repository)
	case ChannelPrerelease, ChannelBeta, ChannelAlpha:
		release, err = c.pickBestListedRelease(ctx, repository, channel)
	default:
		release, err = c.pickBestListedRelease(ctx, repository, channel)
	}
	if err != nil {
		return nil, err
	}

	asset, err := selectReleaseAsset(release, repository.Name, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}

	return &Release{
		Version:   release.TagName,
		AssetURL:  asset.BrowserDownloadURL,
		AssetName: asset.Name,
	}, nil
}

func (c *GitHubClient) UpdateTo(ctx context.Context, release *Release, executable string) error {
	if release == nil {
		return errors.New("release is required")
	}
	return selfupdate.UpdateTo(ctx, release.AssetURL, release.AssetName, executable)
}

func (c *GitHubClient) fetchLatestStableRelease(ctx context.Context, repository Repository) (*githubRelease, error) {
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = githubAPIBaseURL
	}

	endpoint := fmt.Sprintf("%s/repos/%s/%s/releases/latest", baseURL, url.PathEscape(repository.Owner), url.PathEscape(repository.Name))
	return c.decodeSingleRelease(ctx, endpoint)
}

func (c *GitHubClient) decodeSingleRelease(ctx context.Context, endpoint string) (*githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "dcli-auto-update")

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusNotFound:
		return nil, ErrReleaseNotFound
	default:
		return nil, fmt.Errorf("github release lookup failed: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	if release.TagName == "" {
		return nil, ErrReleaseNotFound
	}

	return &release, nil
}

func (c *GitHubClient) fetchReleasesPage(ctx context.Context, repository Repository, page, perPage int) ([]githubRelease, error) {
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = githubAPIBaseURL
	}

	endpoint := fmt.Sprintf(
		"%s/repos/%s/%s/releases?per_page=%d&page=%d",
		baseURL,
		url.PathEscape(repository.Owner),
		url.PathEscape(repository.Name),
		perPage,
		page,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "dcli-auto-update")

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github releases list failed: %s", resp.Status)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

func releaseMatchesChannel(rel *githubRelease, channel ReleaseChannel) bool {
	if rel == nil || rel.Draft {
		return false
	}

	switch channel {
	case ChannelStable:
		return !rel.Prerelease
	case ChannelPrerelease:
		return rel.Prerelease
	case ChannelBeta:
		if !rel.Prerelease {
			return false
		}
		return strings.Contains(strings.ToLower(rel.TagName), "-beta")
	case ChannelAlpha:
		if !rel.Prerelease {
			return false
		}
		return strings.Contains(strings.ToLower(rel.TagName), "-alpha")
	default:
		return !rel.Prerelease
	}
}

func (c *GitHubClient) pickBestListedRelease(ctx context.Context, repository Repository, channel ReleaseChannel) (*githubRelease, error) {
	var best *githubRelease
	var bestCanon string

	for page := 1; page <= releasesListMaxPages; page++ {
		releases, err := c.fetchReleasesPage(ctx, repository, page, releasesListPerPage)
		if err != nil {
			return nil, err
		}
		if len(releases) == 0 {
			break
		}

		for i := range releases {
			rel := &releases[i]
			if !releaseMatchesChannel(rel, channel) {
				continue
			}
			ver := normalizeVersion(rel.TagName)
			if ver == "" {
				continue
			}
			if _, err := selectReleaseAsset(rel, repository.Name, runtime.GOOS, runtime.GOARCH); err != nil {
				continue
			}
			if best == nil || semver.Compare(ver, bestCanon) > 0 {
				best = rel
				bestCanon = ver
			}
		}
	}

	if best == nil {
		return nil, ErrReleaseNotFound
	}

	return best, nil
}

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func selectReleaseAsset(release *githubRelease, project, goos, goarch string) (*githubAsset, error) {
	if release == nil {
		return nil, errors.New("release is required")
	}
	if release.Draft {
		return nil, ErrReleaseNotFound
	}

	for _, candidate := range assetCandidates(project, release.TagName, goos, goarch) {
		for _, asset := range release.Assets {
			if asset.Name == candidate {
				asset := asset
				return &asset, nil
			}
		}
	}

	return nil, fmt.Errorf("release asset not found for %s/%s (%s/%s)", project, release.TagName, goos, goarch)
}

func assetCandidates(project, version, goos, goarch string) []string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	trimmedVersion := strings.TrimPrefix(version, "v")
	if trimmedVersion == version {
		trimmedVersion = ""
	}

	candidates := []string{
		fmt.Sprintf("%s-%s-%s-%s%s", project, version, goos, goarch, ext),
		fmt.Sprintf("%s_%s_%s_%s%s", project, version, goos, goarch, ext),
	}

	if trimmedVersion != "" {
		candidates = append(candidates,
			fmt.Sprintf("%s-%s-%s-%s%s", project, trimmedVersion, goos, goarch, ext),
			fmt.Sprintf("%s_%s_%s_%s%s", project, trimmedVersion, goos, goarch, ext),
		)
	}

	return candidates
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return ""
	}

	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if !semver.IsValid(version) {
		return ""
	}
	return semver.Canonical(version)
}
