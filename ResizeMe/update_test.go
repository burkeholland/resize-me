package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLatestCompatibleWindowsReleaseSkipsIneligibleReleases(t *testing.T) {
	releases := []githubRelease{
		{
			TagName: "v9.0.0-windows",
			Draft:   true,
			Assets:  []githubReleaseAsset{{Name: "ResizeMe-windows-amd64.exe"}},
		},
		{
			TagName:    "v8.0.0-windows",
			Prerelease: true,
			Assets:     []githubReleaseAsset{{Name: "ResizeMe-windows-amd64.exe"}},
		},
		{
			TagName: "v7.0.0-windows",
			Assets:  []githubReleaseAsset{{Name: "ResizeMe-windows-arm64.exe"}},
		},
		{
			TagName: "v0.4.0-mac",
			Assets:  []githubReleaseAsset{{Name: "ResizeMe-windows-amd64.exe"}},
		},
		{
			TagName: "v0.2.3-windows",
			Assets:  []githubReleaseAsset{{Name: "ResizeMe-windows-amd64.exe"}},
		},
		{
			TagName: "v0.3.0-windows",
			Assets:  []githubReleaseAsset{{Name: "ResizeMe-windows-amd64.exe"}},
		},
	}

	release, found := latestCompatibleWindowsRelease(releases, "ResizeMe-windows-amd64.exe")

	if !found {
		t.Fatal("expected a compatible Windows release")
	}
	if got := release.version.String(); got != "0.3.0" {
		t.Fatalf("version = %q, want 0.3.0", got)
	}
}

func TestGitHubReleaseCheckerReportsCompatibleUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("Accept = %q", got)
		}
		if got := request.Header.Get("User-Agent"); got != "ResizeMe-update-checker" {
			t.Errorf("User-Agent = %q", got)
		}
		_, _ = response.Write([]byte(`[
			{"tag_name":"v0.3.0-windows","html_url":"https://github.com/burkeholland/resize-me/releases/tag/v0.3.0-windows","assets":[{"name":"ResizeMe-windows-amd64.exe"}]},
			{"tag_name":"v0.4.0-windows","assets":[{"name":"ResizeMe-windows-arm64.exe"}]}
		]`))
	}))
	defer server.Close()

	checker := newGitHubReleaseChecker(server.Client(), server.URL, "amd64")
	update, err := checker.Check("0.2.3")

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !update.Available {
		t.Fatal("Available = false, want true")
	}
	if update.CurrentVersion != "0.2.3" || update.LatestVersion != "0.3.0" {
		t.Fatalf("versions = %q -> %q, want 0.2.3 -> 0.3.0", update.CurrentVersion, update.LatestVersion)
	}
	if update.AssetName != "ResizeMe-windows-amd64.exe" {
		t.Fatalf("AssetName = %q", update.AssetName)
	}
	if update.UpdateCommand != wingetUpgradeCommand {
		t.Fatalf("UpdateCommand = %q", update.UpdateCommand)
	}
}

func TestGitHubReleaseCheckerReportsCurrentRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte(`[
			{"tag_name":"v0.2.3-windows","assets":[{"name":"ResizeMe-windows-amd64.exe"}]}
		]`))
	}))
	defer server.Close()

	checker := newGitHubReleaseChecker(server.Client(), server.URL, "amd64")
	update, err := checker.Check("0.2.3")

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if update.Available {
		t.Fatal("Available = true, want false")
	}
	if update.LatestVersion != "0.2.3" {
		t.Fatalf("LatestVersion = %q, want 0.2.3", update.LatestVersion)
	}
}

func TestGitHubReleaseCheckerRejectsFailuresAndUnsupportedResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Error(response, `{"message":"rate limit exceeded"}`, http.StatusForbidden)
	}))
	defer server.Close()

	checker := newGitHubReleaseChecker(server.Client(), server.URL, "amd64")
	if _, err := checker.Check("0.2.3"); err == nil || !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Fatalf("Check() error = %v, want GitHub failure", err)
	}

	unsupported := newGitHubReleaseChecker(server.Client(), server.URL, "386")
	if _, err := unsupported.Check("0.2.3"); err == nil || !strings.Contains(err.Error(), "do not support") {
		t.Fatalf("Check() error = %v, want unsupported architecture", err)
	}
}

func TestGitHubReleaseCheckerRejectsMissingCompatibleAsset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte(`[
			{"tag_name":"v0.3.0-windows","assets":[{"name":"ResizeMe-windows-arm64.exe"}]}
		]`))
	}))
	defer server.Close()

	checker := newGitHubReleaseChecker(server.Client(), server.URL, "amd64")
	if _, err := checker.Check("0.2.3"); err == nil || !strings.Contains(err.Error(), "ResizeMe-windows-amd64.exe") {
		t.Fatalf("Check() error = %v, want missing x64 asset", err)
	}
}

func TestWindowsAssetNameSupportsReleaseArchitectures(t *testing.T) {
	tests := map[string]string{
		"amd64": "ResizeMe-windows-amd64.exe",
		"arm64": "ResizeMe-windows-arm64.exe",
	}

	for architecture, want := range tests {
		t.Run(architecture, func(t *testing.T) {
			got, err := windowsAssetName(architecture)
			if err != nil {
				t.Fatalf("windowsAssetName() error = %v", err)
			}
			if got != want {
				t.Fatalf("windowsAssetName() = %q, want %q", got, want)
			}
		})
	}
}

func TestReleaseVersionComparisonHandlesExecutableRevision(t *testing.T) {
	current, err := parseReleaseVersion("v0.2.3.4")
	if err != nil {
		t.Fatalf("parseReleaseVersion() error = %v", err)
	}
	latest, err := parseReleaseVersion("0.2.4")
	if err != nil {
		t.Fatalf("parseReleaseVersion() error = %v", err)
	}
	if compareReleaseVersions(latest, current) <= 0 {
		t.Fatalf("compareReleaseVersions(%s, %s) <= 0", latest.String(), current.String())
	}
	if _, err := parseReleaseVersion("Development"); err == nil {
		t.Fatal("parseReleaseVersion(Development) succeeded")
	}
}
