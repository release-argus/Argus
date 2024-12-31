// Copyright [2024] [Argus]
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

// Package github provides a github-based lookup type.
package github

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/release-argus/Argus/service/latest_version/filter"
	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/test"
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

func TestLookup_FilterGitHubReleases(t *testing.T) {
	// GIVEN a bunch of releases
	tests := map[string]struct {
		releases                           []github_types.Release
		semanticVersioning, usePreReleases bool
		urlCommands                        *filter.URLCommandSlice
		want                               []string
	}{
		"use Name if no TagName (/tags vs /releases API)": {
			releases: []github_types.Release{
				{Name: "0.99.0"},
				{Name: "0.3.0"},
				{Name: "0.0.1"},
			},
			want: []string{"0.99.0", "0.3.0", "0.0.1"},
		},
		"handle leading v's": {
			usePreReleases: true,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "v0.3.0"},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "v0.3.0", "0.0.1"},
		},
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
			},
			want: []string{"0.99.0", "0.0.1"},
		},
		"exclude non-semantic": {
			usePreReleases:     true,
			semanticVersioning: true,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "version 0.2.0", PreRelease: true},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "0.3.0", "0.0.1"},
		},
		"keep pre-release non-semantic": {
			usePreReleases: true,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "v0.2.0", PreRelease: true},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "0.3.0", "v0.2.0", "0.0.1"},
		},
		"exclude pre-release non-semantic": {
			usePreReleases: false,
			releases: []github_types.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "v0.2.0", PreRelease: true},
				{TagName: "v0.0.2"},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "v0.0.2", "0.0.1"},
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
			},
			want: []string{"0.3.0", "0.2.0", "0.0.2", "0.0.1", "0.0.0"},
		},
		"filter releases with failed urlCommand": {
			usePreReleases:     false,
			semanticVersioning: true,
			releases: []github_types.Release{
				{TagName: "0.0.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.2.0", PreRelease: true},
				{TagName: "0.0.2-0.0.2"},
				{TagName: "0.0.1-0.0.1"},
			},
			urlCommands: &filter.URLCommandSlice{
				{Type: "regex", Regex: `-(.*)`}},
			want: []string{"0.0.2", "0.0.1"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lv := testLookup(false)
			lv.URLCommands = nil
			if tc.urlCommands != nil {
				lv.URLCommands = *tc.urlCommands
			}
			lv.UsePreRelease = &tc.usePreReleases
			lv.Options.SemanticVersioning = &tc.semanticVersioning
			lv.GetGitHubData().SetReleases(tc.releases)

			// WHEN filterGitHubReleases is called on this body
			filteredReleases := lv.filterGitHubReleases(util.LogFrom{})

			// THEN only the expected releases are kept
			if len(tc.want) != len(filteredReleases) {
				t.Fatalf("Length not the same\nwant: %v (%d)\ngot:  %v (%d)",
					tc.want, len(tc.want), filteredReleases, len(filteredReleases))
			}
			for i := range tc.want {
				if tc.want[i] != filteredReleases[i].TagName {
					t.Fatalf("got unexpected release %v\nwant: %q (%v)\ngot:  %v",
						filteredReleases[i], tc.want[i], tc.want, filteredReleases)
				}
			}
		})
	}
}

func TestLookup_CheckGitHubReleasesBody(t *testing.T) {
	// GIVEN a body
	tests := map[string]struct {
		body     string
		errRegex string
	}{
		"invalid json": {
			body:     strings.Repeat("something something something", 100),
			errRegex: `unmarshal .* failed`},
		"1 release": {
			body: test.TrimJSON(`
				[
					{"tag_name":"0.18.0","name":"0.18.0","prerelease":true,"published_at":"2024-05-07T13:10:29Z",
						"assets":[
							{"id": 9,"name":"Argus-0.18.0.linux-amd64","created_at":"2024-05-07T13:11:30Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.18.0/Argus-0.18.0.darwin-amd64"}
						]}
				]`),
			errRegex: `^$`,
		},
		"test releases": {
			body:     string(testBody),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			body := []byte(tc.body)
			lv := Lookup{}

			// WHEN filterGitHubReleases is called on this body
			releases, err := lv.checkGitHubReleasesBody(body, util.LogFrom{})

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			// ELSE the releases marshal correctly
			if tc.errRegex == "^$" {
				releasesYAML, _ := json.Marshal(releases)
				if string(releasesYAML) != tc.body {
					t.Errorf("releases mismatch\nwant: %q\ngot:  %q",
						testBody, string(releasesYAML))
				}
			}
		})
	}
}