package update

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckForUpdate_DevVersion(t *testing.T) {
	// Dev versions skip the check entirely
	result := CheckForUpdate(context.Background(), "dev")
	if result != nil {
		t.Error("expected nil for dev version")
	}

	result = CheckForUpdate(context.Background(), "")
	if result != nil {
		t.Error("expected nil for empty version")
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
		{"0.1.0", "v0.1.0"},
	}

	for _, tt := range tests {
		got := normalizeVersion(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNewChecker(t *testing.T) {
	c := NewChecker()
	if c.HTTPClient == nil {
		t.Error("expected HTTPClient to be set")
	}
	if c.ReleasesURL != DefaultReleasesURL {
		t.Errorf("expected ReleasesURL = %q, got %q", DefaultReleasesURL, c.ReleasesURL)
	}
	if c.Timeout != CheckTimeout {
		t.Errorf("expected Timeout = %v, got %v", CheckTimeout, c.Timeout)
	}
}

func TestChecker_NewerVersionAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Errorf("expected Accept header, got %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("User-Agent") != "airwallex-cli/1.0.0" {
			t.Errorf("expected User-Agent header, got %q", r.Header.Get("User-Agent"))
		}

		release := Release{
			TagName: "v2.0.0",
			HTMLURL: "https://github.com/example/repo/releases/tag/v2.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		Timeout:     5 * time.Second,
	}

	result := checker.Check(context.Background(), "1.0.0")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.UpdateAvailable {
		t.Error("expected UpdateAvailable to be true")
	}
	if result.CurrentVersion != "1.0.0" {
		t.Errorf("expected CurrentVersion = 1.0.0, got %q", result.CurrentVersion)
	}
	if result.LatestVersion != "2.0.0" {
		t.Errorf("expected LatestVersion = 2.0.0, got %q", result.LatestVersion)
	}
	if result.UpdateURL != "https://github.com/example/repo/releases/tag/v2.0.0" {
		t.Errorf("unexpected UpdateURL: %q", result.UpdateURL)
	}
}

func TestChecker_CurrentVersionIsLatest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v1.0.0",
			HTMLURL: "https://github.com/example/repo/releases/tag/v1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		Timeout:     5 * time.Second,
	}

	result := checker.Check(context.Background(), "1.0.0")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.UpdateAvailable {
		t.Error("expected UpdateAvailable to be false")
	}
}

func TestChecker_CurrentVersionNewer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v1.0.0",
			HTMLURL: "https://github.com/example/repo/releases/tag/v1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		Timeout:     5 * time.Second,
	}

	result := checker.Check(context.Background(), "2.0.0")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.UpdateAvailable {
		t.Error("expected UpdateAvailable to be false when current > latest")
	}
}

// mockHTTPClient is a mock HTTP client for testing error scenarios
type mockHTTPClient struct {
	err error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return nil, m.err
}

func TestChecker_NetworkError(t *testing.T) {
	checker := &Checker{
		HTTPClient:  &mockHTTPClient{err: errors.New("network error")},
		ReleasesURL: "http://localhost:9999/nonexistent",
		Timeout:     5 * time.Second,
	}

	result := checker.Check(context.Background(), "1.0.0")
	if result != nil {
		t.Error("expected nil result on network error")
	}
}

func TestChecker_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		Timeout:     5 * time.Second,
	}

	result := checker.Check(context.Background(), "1.0.0")
	if result != nil {
		t.Error("expected nil result on invalid JSON")
	}
}

func TestChecker_Non200StatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
		{"Forbidden", http.StatusForbidden},
		{"RateLimited", http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			checker := &Checker{
				HTTPClient:  server.Client(),
				ReleasesURL: server.URL,
				Timeout:     5 * time.Second,
			}

			result := checker.Check(context.Background(), "1.0.0")
			if result != nil {
				t.Errorf("expected nil result for status %d", tt.statusCode)
			}
		})
	}
}

func TestChecker_VersionComparisonEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestTagName  string
		wantUpdate     bool
		wantLatest     string
	}{
		{
			name:           "v prefix in current only",
			currentVersion: "v1.0.0",
			latestTagName:  "2.0.0",
			wantUpdate:     true,
			wantLatest:     "2.0.0",
		},
		{
			name:           "v prefix in latest only",
			currentVersion: "1.0.0",
			latestTagName:  "v2.0.0",
			wantUpdate:     true,
			wantLatest:     "2.0.0",
		},
		{
			name:           "both with v prefix",
			currentVersion: "v1.0.0",
			latestTagName:  "v2.0.0",
			wantUpdate:     true,
			wantLatest:     "2.0.0",
		},
		{
			name:           "neither with v prefix",
			currentVersion: "1.0.0",
			latestTagName:  "2.0.0",
			wantUpdate:     true,
			wantLatest:     "2.0.0",
		},
		{
			name:           "patch version update",
			currentVersion: "1.0.0",
			latestTagName:  "v1.0.1",
			wantUpdate:     true,
			wantLatest:     "1.0.1",
		},
		{
			name:           "minor version update",
			currentVersion: "1.0.0",
			latestTagName:  "v1.1.0",
			wantUpdate:     true,
			wantLatest:     "1.1.0",
		},
		{
			name:           "prerelease vs stable",
			currentVersion: "1.0.0-beta.1",
			latestTagName:  "v1.0.0",
			wantUpdate:     true,
			wantLatest:     "1.0.0",
		},
		{
			name:           "stable vs higher prerelease",
			currentVersion: "1.0.0",
			latestTagName:  "v1.0.1-alpha.1",
			wantUpdate:     true, // 1.0.1-alpha.1 > 1.0.0 in semver
			wantLatest:     "1.0.1-alpha.1",
		},
		{
			name:           "stable vs same prerelease",
			currentVersion: "1.0.0",
			latestTagName:  "v1.0.0-alpha.1",
			wantUpdate:     false, // 1.0.0-alpha.1 < 1.0.0 in semver
			wantLatest:     "1.0.0-alpha.1",
		},
		{
			name:           "same prerelease version",
			currentVersion: "1.0.0-beta.1",
			latestTagName:  "v1.0.0-beta.1",
			wantUpdate:     false,
			wantLatest:     "1.0.0-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				release := Release{
					TagName: tt.latestTagName,
					HTMLURL: "https://github.com/example/repo/releases/tag/" + tt.latestTagName,
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(release); err != nil {
					t.Fatal(err)
				}
			}))
			defer server.Close()

			checker := &Checker{
				HTTPClient:  server.Client(),
				ReleasesURL: server.URL,
				Timeout:     5 * time.Second,
			}

			result := checker.Check(context.Background(), tt.currentVersion)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.UpdateAvailable != tt.wantUpdate {
				t.Errorf("UpdateAvailable = %v, want %v", result.UpdateAvailable, tt.wantUpdate)
			}
			if result.LatestVersion != tt.wantLatest {
				t.Errorf("LatestVersion = %q, want %q", result.LatestVersion, tt.wantLatest)
			}
		})
	}
}

func TestChecker_InvalidSemver(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestTagName  string
	}{
		{
			name:           "invalid current version",
			currentVersion: "not-a-version",
			latestTagName:  "v1.0.0",
		},
		{
			name:           "invalid latest version",
			currentVersion: "1.0.0",
			latestTagName:  "invalid",
		},
		{
			name:           "both invalid",
			currentVersion: "foo",
			latestTagName:  "bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				release := Release{
					TagName: tt.latestTagName,
					HTMLURL: "https://github.com/example/repo/releases/tag/" + tt.latestTagName,
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(release); err != nil {
					t.Fatal(err)
				}
			}))
			defer server.Close()

			checker := &Checker{
				HTTPClient:  server.Client(),
				ReleasesURL: server.URL,
				Timeout:     5 * time.Second,
			}

			result := checker.Check(context.Background(), tt.currentVersion)
			if result == nil {
				t.Fatal("expected non-nil result even with invalid semver")
			}
			// With invalid semver, UpdateAvailable should be false
			if result.UpdateAvailable {
				t.Error("expected UpdateAvailable to be false with invalid semver")
			}
		})
	}
}

func TestChecker_DevVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for dev versions")
	}))
	defer server.Close()

	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		Timeout:     5 * time.Second,
	}

	tests := []struct {
		version string
	}{
		{"dev"},
		{""},
	}

	for _, tt := range tests {
		result := checker.Check(context.Background(), tt.version)
		if result != nil {
			t.Errorf("expected nil for version %q", tt.version)
		}
	}
}

func TestChecker_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		release := Release{TagName: "v2.0.0", HTMLURL: "https://example.com"}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		Timeout:     10 * time.Millisecond, // Very short timeout
	}

	result := checker.Check(context.Background(), "1.0.0")
	// Should return nil due to timeout
	if result != nil {
		t.Error("expected nil result due to timeout")
	}
}

func TestChecker_DefaultsWhenFieldsNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{TagName: "v2.0.0", HTMLURL: "https://example.com"}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	// Checker with some fields unset
	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		// Timeout is zero, should use default
	}

	result := checker.Check(context.Background(), "1.0.0")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.UpdateAvailable {
		t.Error("expected UpdateAvailable to be true")
	}
}

func TestChecker_EmptyTagName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "",
			HTMLURL: "https://github.com/example/repo/releases",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	checker := &Checker{
		HTTPClient:  server.Client(),
		ReleasesURL: server.URL,
		Timeout:     5 * time.Second,
	}

	result := checker.Check(context.Background(), "1.0.0")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Empty tag is not valid semver, so UpdateAvailable should be false
	if result.UpdateAvailable {
		t.Error("expected UpdateAvailable to be false with empty tag")
	}
}

func TestGitHubReleasesURLConstant(t *testing.T) {
	// Ensure backwards compatibility constant exists and matches default
	if GitHubReleasesURL != DefaultReleasesURL {
		t.Errorf("GitHubReleasesURL (%q) should equal DefaultReleasesURL (%q)", GitHubReleasesURL, DefaultReleasesURL)
	}
}
