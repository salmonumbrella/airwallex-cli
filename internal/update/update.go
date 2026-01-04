package update

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	// DefaultReleasesURL is the default API endpoint for releases
	DefaultReleasesURL = "https://api.github.com/repos/salmonumbrella/airwallex-cli/releases/latest"
	// GitHubReleasesURL is kept for backwards compatibility
	GitHubReleasesURL = DefaultReleasesURL
	// CheckTimeout is the timeout for version check
	CheckTimeout = 5 * time.Second
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckResult contains the result of a version check
type CheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateURL       string
	UpdateAvailable bool
}

// Checker provides configurable update checking functionality
type Checker struct {
	HTTPClient  HTTPClient
	ReleasesURL string
	Timeout     time.Duration
}

// NewChecker creates a new Checker with default settings
func NewChecker() *Checker {
	return &Checker{
		HTTPClient:  http.DefaultClient,
		ReleasesURL: DefaultReleasesURL,
		Timeout:     CheckTimeout,
	}
}

// defaultChecker is the package-level checker instance
var defaultChecker = NewChecker()

// CheckForUpdate checks if a newer version is available on GitHub.
// Returns nil if the check fails (network error, etc.) - never blocks the CLI.
// This is a convenience function that uses the default checker.
func CheckForUpdate(ctx context.Context, currentVersion string) *CheckResult {
	return defaultChecker.Check(ctx, currentVersion)
}

// Check checks if a newer version is available on GitHub.
// Returns nil if the check fails (network error, etc.) - never blocks the CLI.
func (c *Checker) Check(ctx context.Context, currentVersion string) *CheckResult {
	// Don't check dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return nil
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = CheckTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	releasesURL := c.ReleasesURL
	if releasesURL == "" {
		releasesURL = DefaultReleasesURL
	}

	req, err := http.NewRequestWithContext(ctx, "GET", releasesURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "airwallex-cli/"+currentVersion)

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}

	// Normalize versions for comparison (add 'v' prefix if missing)
	current := normalizeVersion(currentVersion)
	latest := normalizeVersion(release.TagName)

	result := &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  strings.TrimPrefix(release.TagName, "v"),
		UpdateURL:      release.HTMLURL,
	}

	// Compare versions using semver
	if semver.IsValid(current) && semver.IsValid(latest) {
		result.UpdateAvailable = semver.Compare(latest, current) > 0
	}

	return result
}

func normalizeVersion(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
