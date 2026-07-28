package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	githubReleasesEndpoint = "https://api.github.com/repos/burkeholland/resize-me/releases?per_page=100"
	wingetUpgradeCommand  = "winget upgrade --id BurkeHolland.ResizeMe --exact"
	updateRequestTimeout  = 10 * time.Second
	maxUpdateResponseSize = 2 << 20
)

var (
	windowsReleaseTagPattern = regexp.MustCompile(`^v([0-9]+\.[0-9]+\.[0-9]+)-windows$`)
	versionPattern           = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:\.([0-9]+))?$`)
)

type UpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Available      bool   `json:"available"`
	ReleaseURL     string `json:"releaseURL"`
	AssetName      string `json:"assetName"`
	UpdateCommand  string `json:"updateCommand"`
}

type updateChecker interface {
	Check(currentVersion string) (UpdateInfo, error)
}

type githubReleaseChecker struct {
	client       *http.Client
	endpoint     string
	architecture string
}

type githubRelease struct {
	Draft      bool                 `json:"draft"`
	Prerelease bool                 `json:"prerelease"`
	TagName    string               `json:"tag_name"`
	HTMLURL    string               `json:"html_url"`
	Assets     []githubReleaseAsset `json:"assets"`
}

type githubReleaseAsset struct {
	Name string `json:"name"`
}

type releaseVersion struct {
	parts [4]uint64
	count int
}

type compatibleRelease struct {
	version releaseVersion
	release githubRelease
	asset   string
}

func defaultUpdateHTTPClient() *http.Client {
	return &http.Client{Timeout: updateRequestTimeout}
}

func runtimeArchitecture() string {
	return runtime.GOARCH
}

func newGitHubReleaseChecker(client *http.Client, endpoint, architecture string) *githubReleaseChecker {
	return &githubReleaseChecker{
		client:       client,
		endpoint:     endpoint,
		architecture: architecture,
	}
}

func (c *githubReleaseChecker) Check(currentVersion string) (UpdateInfo, error) {
	current, err := parseReleaseVersion(currentVersion)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("check for updates: current version %q is not a numbered release: %w", currentVersion, err)
	}

	assetName, err := windowsAssetName(c.architecture)
	if err != nil {
		return UpdateInfo{}, err
	}

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.endpoint, nil)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("create GitHub update request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "ResizeMe-update-checker")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := c.client.Do(request)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("check GitHub releases: %w", err)
	}
	defer response.Body.Close()

	body, err := readUpdateResponse(response.Body)
	if err != nil {
		return UpdateInfo{}, err
	}
	if response.StatusCode != http.StatusOK {
		return UpdateInfo{}, githubResponseError(response.Status, body)
	}

	var releases []githubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return UpdateInfo{}, fmt.Errorf("decode GitHub releases: %w", err)
	}

	latest, found := latestCompatibleWindowsRelease(releases, assetName)
	if !found {
		return UpdateInfo{}, fmt.Errorf("no published Windows release includes the %s asset", assetName)
	}

	return UpdateInfo{
		CurrentVersion: current.String(),
		LatestVersion:  latest.version.String(),
		Available:      compareReleaseVersions(latest.version, current) > 0,
		ReleaseURL:     latest.release.HTMLURL,
		AssetName:      latest.asset,
		UpdateCommand:  wingetUpgradeCommand,
	}, nil
}

func readUpdateResponse(body io.Reader) ([]byte, error) {
	response, err := io.ReadAll(io.LimitReader(body, maxUpdateResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read GitHub release response: %w", err)
	}
	if len(response) > maxUpdateResponseSize {
		return nil, fmt.Errorf("GitHub release response exceeds %d bytes", maxUpdateResponseSize)
	}
	return response, nil
}

func githubResponseError(status string, body []byte) error {
	var apiError struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &apiError); err == nil && apiError.Message != "" {
		return fmt.Errorf("GitHub update check returned %s: %s", status, apiError.Message)
	}
	return fmt.Errorf("GitHub update check returned %s", status)
}

func latestCompatibleWindowsRelease(releases []githubRelease, assetName string) (compatibleRelease, bool) {
	var latest compatibleRelease
	found := false

	for _, release := range releases {
		if release.Draft || release.Prerelease || !releaseHasAsset(release, assetName) {
			continue
		}

		version, ok := windowsReleaseVersion(release.TagName)
		if !ok {
			continue
		}
		if !found || compareReleaseVersions(version, latest.version) > 0 {
			latest = compatibleRelease{
				version: version,
				release: release,
				asset:   assetName,
			}
			found = true
		}
	}

	return latest, found
}

func releaseHasAsset(release githubRelease, assetName string) bool {
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return true
		}
	}
	return false
}

func windowsReleaseVersion(tag string) (releaseVersion, bool) {
	match := windowsReleaseTagPattern.FindStringSubmatch(strings.TrimSpace(tag))
	if match == nil {
		return releaseVersion{}, false
	}
	version, err := parseReleaseVersion(match[1])
	if err != nil {
		return releaseVersion{}, false
	}
	return version, true
}

func windowsAssetName(architecture string) (string, error) {
	switch architecture {
	case "amd64":
		return "ResizeMe-windows-amd64.exe", nil
	case "arm64":
		return "ResizeMe-windows-arm64.exe", nil
	default:
		return "", fmt.Errorf("Windows update checks do not support the %q architecture", architecture)
	}
}

func parseReleaseVersion(raw string) (releaseVersion, error) {
	match := versionPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil {
		return releaseVersion{}, fmt.Errorf("expected a version such as 0.2.3")
	}

	version := releaseVersion{count: 3}
	for index := 1; index <= 3; index++ {
		value, err := strconv.ParseUint(match[index], 10, 64)
		if err != nil {
			return releaseVersion{}, fmt.Errorf("parse version component %q: %w", match[index], err)
		}
		version.parts[index-1] = value
	}
	if match[4] != "" {
		value, err := strconv.ParseUint(match[4], 10, 64)
		if err != nil {
			return releaseVersion{}, fmt.Errorf("parse version component %q: %w", match[4], err)
		}
		version.parts[3] = value
		version.count = 4
	}
	return version, nil
}

func compareReleaseVersions(left, right releaseVersion) int {
	for index := range left.parts {
		switch {
		case left.parts[index] > right.parts[index]:
			return 1
		case left.parts[index] < right.parts[index]:
			return -1
		}
	}
	return 0
}

func (v releaseVersion) String() string {
	components := []string{
		strconv.FormatUint(v.parts[0], 10),
		strconv.FormatUint(v.parts[1], 10),
		strconv.FormatUint(v.parts[2], 10),
	}
	if v.count == 4 && v.parts[3] != 0 {
		components = append(components, strconv.FormatUint(v.parts[3], 10))
	}
	return strings.Join(components, ".")
}
