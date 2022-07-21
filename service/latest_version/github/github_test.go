// Copyright [2022] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"strings"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/release-argus/Argus/utils"
)

func TestInsertionSortWithNewestVersion(t *testing.T) {
	// GIVEN a list of releases
	releases := []GitHubRelease{
		{TagName: "0.99.0"},
		{TagName: "0.3.0"},
		{TagName: "0.1.0"},
		{TagName: "0.0.1"},
		{TagName: "0.0.0"},
	}
	for i := range releases {
		semVer, _ := semver.NewVersion(releases[i].TagName)
		releases[i].SemanticVersion = semVer
	}

	// WHEN insertionSort is called with a release newer than the head
	release := GitHubRelease{
		TagName: "1.0.0",
	}
	semVer, _ := semver.NewVersion(release.TagName)
	release.SemanticVersion = semVer
	insertionSort(release, &releases)

	// THEN it can be found at the first index
	if releases[0].TagName != release.TagName {
		t.Errorf("Expected %v to be inserted at the head of releases. Got %v",
			release, releases[0])
	}
}

func TestInsertionSortWithNotNewestVersion(t *testing.T) {
	// GIVEN a list of releases
	releases := []GitHubRelease{
		{TagName: "0.99.0"},
		{TagName: "0.3.0"},
		{TagName: "0.1.0"},
		{TagName: "0.0.1"},
		{TagName: "0.0.0"},
	}
	for i := range releases {
		semVer, _ := semver.NewVersion(releases[i].TagName)
		releases[i].SemanticVersion = semVer
	}

	// WHEN insertionSort is called with a release somewhere in the middle of the versions so far
	release := GitHubRelease{
		TagName: "0.2.0",
	}
	semVer, _ := semver.NewVersion(release.TagName)
	release.SemanticVersion = semVer
	insertionSort(release, &releases)

	// THEN it can be found at the third index
	if releases[2].TagName != release.TagName {
		t.Errorf("Expected %v to be inserted at the third index of releases. Got %v",
			release, releases[2])
	}
}

func TestInsertionSortWithNotOldestVersion(t *testing.T) {
	// GIVEN a list of releases
	releases := []GitHubRelease{
		{TagName: "0.99.0"},
		{TagName: "0.3.0"},
		{TagName: "0.2.0"},
		{TagName: "0.1.0"},
		{TagName: "0.0.1"},
	}
	for i := range releases {
		semVer, _ := semver.NewVersion(releases[i].TagName)
		releases[i].SemanticVersion = semVer
	}

	// WHEN insertionSort is called with a release older than the versions so far
	release := GitHubRelease{
		TagName: "0.0.0",
	}
	semVer, _ := semver.NewVersion(release.TagName)
	release.SemanticVersion = semVer
	insertionSort(release, &releases)

	// THEN it can be found at the last index
	if releases[len(releases)-1].TagName != release.TagName {
		t.Errorf("Expected %v to be inserted at the third index of releases. Got %v",
			release, releases[len(releases)-1].TagName)
	}
}

func TestCheckGitHubReleasesBodyWithRateLimit(t *testing.T) {
	// GIVEN a body detailing a rate limit
	jLog = utils.NewJLog("WARN", false)
	body := []byte("something rate limit something")
	svc := Service{}

	// WHEN filterGitHubReleases is called on this body
	_, err := svc.checkGitHubReleasesBody(&body, utils.LogFrom{})

	// THEN we receive a nerr informing of this rate limit
	if utils.ErrorToString(err) != "rate limit reached for GitHub" {
		t.Errorf("Expected an error about rate limit being reached but got %q",
			utils.ErrorToString(err))
	}
}

func TestCheckGitHubReleasesBodyWithBadCredentials(t *testing.T) {
	// GIVEN a body with no tag_name's
	jLog = utils.NewJLog("WARN", false)
	body := []byte("something Bad credentials something")
	url := "https://example.com"
	svc := Service{URL: &url}
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(utils.ErrorToString(r.(error)), "github access token is invalid") {
			t.Errorf("Expected an error about rate limit being reached but got %q",
				r.(string))
		}
	}()

	// WHEN filterGitHubReleases is called on this body
	_, err := svc.checkGitHubReleasesBody(&body, utils.LogFrom{})

	// THEN we receive a nerr informing of this rate limit
	t.Errorf("Shouldn't reach this as we should Fatal on %q\nerr = %s",
		body, utils.ErrorToString(err))
}

func TestCheckGitHubReleasesBodyWithNoTagNames(t *testing.T) {
	// GIVEN a body with no tag_name's
	jLog = utils.NewJLog("WARN", false)
	body := []byte("something something something")
	url := "https://example.com"
	svc := Service{URL: &url}

	// WHEN filterGitHubReleases is called on this body
	_, err := svc.checkGitHubReleasesBody(&body, utils.LogFrom{})

	// THEN we receive a nerr informing of this rate limit
	if !strings.HasPrefix(utils.ErrorToString(err), "tag_name not found at") {
		t.Errorf("Expected an error about rate limit being reached but got %q",
			utils.ErrorToString(err))
	}
}

func TestCheckGitHubReleasesBodyWithInvalidJSON(t *testing.T) {
	// GIVEN a body with invalid JSON
	jLog = utils.NewJLog("WARN", false)
	body := []byte(strings.Repeat("something something something", 100))
	url := "https://example.com"
	svc := Service{URL: &url}

	// WHEN filterGitHubReleases is called on this body
	_, err := svc.checkGitHubReleasesBody(&body, utils.LogFrom{})

	// THEN we receive a nerr informing of this rate limit
	if !strings.HasPrefix(utils.ErrorToString(err), "unmarshal of GitHub API data failed") {
		t.Errorf("Expected an error about unmarshal failing but got %q",
			utils.ErrorToString(err))
	}
}

func TestFilterGitHubReleasesDoesFilterPreReleases(t *testing.T) {
	// GIVEN a list of releases
	jLog = utils.NewJLog("WARN", false)
	var (
		url                = "https://example.com"
		semanticVersioning = true
		usePreRelease      = true
	)
	svc := Service{
		URL:                &url,
		SemanticVersioning: &semanticVersioning,
		UsePreRelease:      &usePreRelease,
		Defaults:           &Service{},
		HardDefaults:       &Service{},
	}
	releases := []GitHubRelease{
		{TagName: "0.99.0"},
		{TagName: "0.3.0", PreRelease: true},
		{TagName: "0.2.0"},
		{TagName: "0.0.1"},
	}

	// WHEN filterGitHubReleases is called on these releases
	// with a Service that wants pre_release's
	wantKept := releases[1]
	filteredReleases := svc.filterGitHubReleases(releases, utils.LogFrom{})

	// THEN the pre_release is filtered out
	if len(filteredReleases) != 4 || filteredReleases[1].TagName != wantKept.TagName {
		t.Errorf("Didn't expect %v to be removed from the releases after filter. Got %v",
			wantKept, filteredReleases)
	}
}

func TestFilterGitHubReleasesDoesntFilterPreReleases(t *testing.T) {
	// GIVEN a list of releases
	jLog = utils.NewJLog("WARN", false)
	var (
		url                = "https://example.com"
		semanticVersioning = true
		usePreRelease      = false
	)
	svc := Service{
		URL:                &url,
		SemanticVersioning: &semanticVersioning,
		UsePreRelease:      &usePreRelease,
		Defaults:           &Service{},
		HardDefaults:       &Service{},
	}
	releases := []GitHubRelease{
		{TagName: "0.99.0"},
		{TagName: "0.3.0", PreRelease: true},
		{TagName: "0.2.0"},
		{TagName: "0.0.1"},
	}

	// WHEN filterGitHubReleases is called on these releases
	wantGone := releases[1]
	filteredReleases := svc.filterGitHubReleases(releases, utils.LogFrom{})

	// THEN the pre_release is filtered out
	if len(filteredReleases) != 3 || filteredReleases[1].TagName == wantGone.TagName {
		t.Errorf("Expected %v to be removed from the releases after filter. Got %v",
			wantGone, filteredReleases)
	}
}

func TestFilterGitHubReleasesWithNotCareSemantic(t *testing.T) {
	// GIVEN a list of non-semantic versioned releases
	jLog = utils.NewJLog("WARN", false)
	var (
		url                = "https://example.com"
		semanticVersioning = false
		usePreRelease      = false
	)
	svc := Service{
		URL:                &url,
		SemanticVersioning: &semanticVersioning,
		UsePreRelease:      &usePreRelease,
		Defaults:           &Service{},
		HardDefaults:       &Service{},
	}
	releases := []GitHubRelease{
		{TagName: "990"},
		{TagName: "30"},
		{TagName: "20"},
		{TagName: "01"},
	}

	// WHEN filterGitHubReleases is called on these releases
	// with no semantic versioning not wanted
	filteredReleases := svc.filterGitHubReleases(releases, utils.LogFrom{})

	// THEN all releases are returned
	if len(filteredReleases) != 4 {
		t.Errorf("Expected all releases to be kept after filter. Got %v",
			filteredReleases)
	}
}

func TestFilterGitHubReleasesWithSomeNonSemantic(t *testing.T) {
	// GIVEN a list of sementic and non-semantic releases
	jLog = utils.NewJLog("WARN", false)
	var (
		url                = "https://example.com"
		semanticVersioning = true
		usePreRelease      = false
	)
	svc := Service{
		URL:                &url,
		SemanticVersioning: &semanticVersioning,
		UsePreRelease:      &usePreRelease,
		Defaults:           &Service{},
		HardDefaults:       &Service{},
	}
	releases := []GitHubRelease{
		{TagName: "990"},
		{TagName: "30"},
		{TagName: "0.0.0"},
		{TagName: "20"},
		{TagName: "01"},
	}

	// WHEN filterGitHubReleases is called on these releases
	// with semantic versioning wanted
	want := releases[2]
	filteredReleases := svc.filterGitHubReleases(releases, utils.LogFrom{})

	// THEN the non-semantic releases are filtered out
	if len(filteredReleases) != 1 || filteredReleases[0].TagName != want.TagName {
		t.Errorf("Expected all the non-semantic releases to be removed with the filter. Got %v",
			filteredReleases)
	}
}

func TestFilterGitHubReleasesWithSomeNonSemanticDidSort(t *testing.T) {
	// GIVEN a list of sementic and non-semantic releases
	jLog = utils.NewJLog("WARN", false)
	var (
		url                = "https://example.com"
		semanticVersioning = true
		usePreRelease      = false
	)
	svc := Service{
		URL:                &url,
		SemanticVersioning: &semanticVersioning,
		UsePreRelease:      &usePreRelease,
		Defaults:           &Service{},
		HardDefaults:       &Service{},
	}
	releases := []GitHubRelease{
		{TagName: "990"},
		{TagName: "0.2.0"},
		{TagName: "30"},
		{TagName: "0.0.0"},
		{TagName: "20"},
		{TagName: "0.1.0"},
		{TagName: "01"},
	}

	// WHEN filterGitHubReleases is called on these releases
	// with semantic versioning wanted
	want := []string{
		"0.2.0",
		"0.1.0",
		"0.0.0",
	}
	filteredReleases := svc.filterGitHubReleases(releases, utils.LogFrom{})

	// THEN the non-semantic releases are filtered out

	if len(filteredReleases) != 3 ||
		filteredReleases[0].TagName != "0.2.0" ||
		filteredReleases[1].TagName != "0.1.0" ||
		filteredReleases[2].TagName != "0.0.0" {
		t.Errorf("Expected the releases to be sorted %v. Got %v",
			want, filteredReleases)
	}
}
