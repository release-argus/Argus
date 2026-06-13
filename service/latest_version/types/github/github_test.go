// Copyright [2026] [Argus]
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
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_FilterGitHubReleases(t *testing.T) {
	// GIVEN: a bunch of releases.
	tests := []struct {
		name                               string
		releases                           []ghtypes.Release
		semanticVersioning, usePreReleases bool
		urlCommands                        *filter.URLCommands
		want                               []string
	}{
		{
			name: "use Name if no TagName (/tags vs /releases API)",
			releases: []ghtypes.Release{
				{Name: "0.99.0"},
				{Name: "0.3.0"},
				{Name: "0.0.1"},
			},
			want: []string{"0.99.0", "0.3.0", "0.0.1"},
		},
		{
			name:           "handle leading v's",
			usePreReleases: true,
			releases: []ghtypes.Release{
				{TagName: "0.99.0"},
				{TagName: "v0.3.0"},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "v0.3.0", "0.0.1"},
		},
		{
			name:           "keep pre-releases",
			usePreReleases: true,
			releases: []ghtypes.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.0.1"},
			}, want: []string{"0.99.0", "0.3.0", "0.0.1"},
		},
		{
			name:           "exclude pre-releases",
			usePreReleases: false,
			releases: []ghtypes.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "0.0.1"},
		},
		{
			name:               "exclude non-semantic",
			usePreReleases:     true,
			semanticVersioning: true,
			releases: []ghtypes.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "version 0.2.0", PreRelease: true},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "0.3.0", "0.0.1"},
		},
		{
			name:           "keep pre-release non-semantic",
			usePreReleases: true,
			releases: []ghtypes.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "v0.2.0", PreRelease: true},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "0.3.0", "v0.2.0", "0.0.1"},
		},
		{
			name:           "exclude pre-release non-semantic",
			usePreReleases: false,
			releases: []ghtypes.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "v0.2.0", PreRelease: true},
				{TagName: "v0.0.2"},
				{TagName: "0.0.1"},
			},
			want: []string{"0.99.0", "v0.0.2", "0.0.1"},
		},
		{
			name:               "does sort releases",
			usePreReleases:     true,
			semanticVersioning: true,
			releases: []ghtypes.Release{
				{TagName: "0.0.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.2.0", PreRelease: true},
				{TagName: "0.0.2"},
				{TagName: "0.0.1"},
			},
			want: []string{"0.3.0", "0.2.0", "0.0.2", "0.0.1", "0.0.0"},
		},
		{
			name:               "filter releases with failed urlCommand",
			usePreReleases:     false,
			semanticVersioning: true,
			releases: []ghtypes.Release{
				{TagName: "0.0.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.2.0", PreRelease: true},
				{TagName: "0.0.2-0.0.2"},
				{TagName: "0.0.1-0.0.1"},
			},
			urlCommands: &filter.URLCommands{
				{Type: "regex", Regex: `-(.*)`},
			},
			want: []string{"0.0.2", "0.0.1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lv := testLookup(t, false)
			lv.URLCommands = nil
			if tc.urlCommands != nil {
				lv.URLCommands = *tc.urlCommands
			}
			lv.UsePreRelease = &tc.usePreReleases
			lv.Options.SemanticVersioning = &tc.semanticVersioning
			lv.GetGitHubData().SetReleases(tc.releases)

			// WHEN: filterGitHubReleases is called on this body.
			filteredReleases := lv.filterGitHubReleases(logx.LogFrom{})

			// THEN: only the expected releases are kept.
			if err := test.AssertSlicesEqualFunc(
				t,
				filteredReleases,
				tc.want,
				func(a ghtypes.Release, b string) bool { return a.TagName == b },
				fmt.Sprintf("%s\nLookup.filterGitHubReleases()", packageName),
				"",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLookup_UnmarshalGitHubReleasesBody(t *testing.T) {
	// GIVEN: a URL body.
	tests := []struct {
		name     string
		body     string
		errRegex string
	}{
		{
			name: "invalid JSON",
			body: strings.Repeat("something something something", 100),
			errRegex: test.TrimYAML(`
				^unmarshal .* failed:
					jsontext:
						invalid character.*$`,
			),
		},
		{
			name: "1 release",
			body: test.TrimJSON(`[
				{
					"tag_name": "0.19.0",
					"name": "0.19.0",
					"prerelease": true,
					"published_at": "2024-05-07T13:10:29Z",
					"assets": [
						{
							"id": 9,
							"name": "Argus-0.19.0.linux-amd64",
							"created_at": "2024-05-07T13:11:30Z",
							"browser_download_url": "https://github.com/release-argus/Argus/releases/download/0.19.0/Argus-0.19.0.darwin-amd64"
						}
					]
				}
			]`),
			errRegex: `^$`,
		},
		{
			name:     "test releases",
			body:     string(testBody),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body := []byte(tc.body)
			lv := Lookup{}

			// WHEN: filterGitHubReleases is called on this body.
			releases, err := lv.unmarshalGitHubReleasesBody(body)

			prefix := fmt.Sprintf(
				"%s\nLookup.checkGitHubReleasesBody(%q)",
				packageName, tc.body,
			)

			// THEN: it errors when expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}
			// ELSE: the releases marshal correctly.
			releasesYAML, _ := decode.Marshal("json", releases)
			if got := string(releasesYAML); got != tc.body {
				t.Errorf(
					"%s releases mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.body,
				)
			}
		})
	}
}
