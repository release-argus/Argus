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

//go:build unit

package latestver

import (
	"regexp"
	"strings"
	"testing"

	"github.com/coreos/go-semver/semver"
	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/util"
)

func TestInsertionSort(t *testing.T) {
	// GIVEN a list of releases and a release to add
	tests := map[string]struct {
		release  string
		expectAt int
	}{
		"newer release": {
			release: "1.0.0", expectAt: 0},
		"middle release": {
			release: "0.2.0", expectAt: 2},
		"oldest release": {
			release: "0.0.0", expectAt: 5},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			releases := []github_types.Release{
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

			// WHEN insertionSort is called with a release
			release := github_types.Release{TagName: tc.release}
			semVer, _ := semver.NewVersion(release.TagName)
			release.SemanticVersion = semVer
			insertionSort(release, &releases)

			// THEN it can be found at the expected index
			if releases[tc.expectAt].TagName != release.TagName {
				t.Errorf("Expected %v to be inserted at index %d. Got %v",
					release, tc.expectAt, release)
			}
		})
	}
}

func TestLookup_CheckGitHubReleasesBody(t *testing.T) {
	// GIVEN a body
	testLogging("WARN")
	tests := map[string]struct {
		body     string
		errRegex string
	}{
		"rate limit": {
			body:     "something rate limit something",
			errRegex: "rate limit reached"},
		"bad credentials": {
			body:     "something Bad credentials something",
			errRegex: "tag_name not found at"},
		"no tag_name found": {
			body:     "bish bash bosh",
			errRegex: "tag_name not found at"},
		"invalid json": {
			body:     strings.Repeat("something something something", 100),
			errRegex: "unmarshal .* failed"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			body := []byte(tc.body)
			lv := Lookup{}

			// WHEN filterGitHubReleases is called on this body
			_, err := lv.checkGitHubReleasesBody(&body, &util.LogFrom{})

			// THEN it err's when expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestLookup_FilterGitHubReleases(t *testing.T) {
	// GIVEN a bunch of releases
	testLogging("WARN")
	tests := map[string]struct {
		releases           []github_types.Release
		semanticVersioning bool
		usePreReleases     bool
		want               []string
	}{
		"keep pre-releases": {
			usePreReleases: true,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.0.1"},
			}, want: []string{"0.99.0", "0.3.0", "0.0.1"},
		},
		"exclude pre-releases": {
			usePreReleases: false,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.0.1"},
			}, want: []string{"0.99.0", "0.0.1"},
		},
		"exclude non-semantic": {
			usePreReleases:     true,
			semanticVersioning: true,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "v0.2.0", PreRelease: true},
				{TagName: "0.0.1"},
			}, want: []string{"0.99.0", "0.3.0", "0.0.1"},
		},
		"keep pre-release non-semantic": {
			usePreReleases: true,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "v0.2.0", PreRelease: true},
				{TagName: "0.0.1"},
			}, want: []string{"0.99.0", "0.3.0", "v0.2.0", "0.0.1"},
		},
		"exclude pre-release non-semantic": {
			usePreReleases: false,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "v0.2.0", PreRelease: true},
				{TagName: "v0.0.2"},
				{TagName: "0.0.1"},
			}, want: []string{"0.99.0", "v0.0.2", "0.0.1"},
		},
		"does sort releases": {
			usePreReleases:     true,
			semanticVersioning: true,
			releases: []github_types.Release{
				{TagName: "0.0.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.2.0", PreRelease: true},
				{TagName: "0.0.2"},
				{TagName: "0.0.1"},
			}, want: []string{"0.3.0", "0.2.0", "0.0.2", "0.0.1", "0.0.0"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lv := Lookup{
				Options: &opt.Options{
					SemanticVersioning: &tc.semanticVersioning,
					Defaults:           &opt.Options{},
					HardDefaults:       &opt.Options{},
				},
				GitHubData:    &GitHubData{},
				UsePreRelease: &tc.usePreReleases,
				Defaults:      &Lookup{},
				HardDefaults:  &Lookup{}}

			// WHEN filterGitHubReleases is called on this body
			filteredReleases := lv.filterGitHubReleases(tc.releases, &util.LogFrom{})

			// THEN only the expected releases are kept
			if len(tc.want) != len(filteredReleases) {
				t.Fatalf("Length not the same\nwant: %v\ngot:  %v",
					tc.want, filteredReleases)
			}
			for i := range tc.want {
				if tc.want[i] != filteredReleases[i].TagName {
					t.Fatalf("got unexpected release %v\nwant: %v\ngot:  %v",
						filteredReleases[i], tc.want, filteredReleases)
				}
			}
		})
	}
}
