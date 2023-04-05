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
	"testing"

	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestLookup_ApplyOverrides(t *testing.T) {
	testLogging("WARN")
	test := testLookup(true, true)
	// GIVEN various json strings to parse as parts of a Lookup
	tests := map[string]struct {
		accessToken        *string
		allowInvalidCerts  *string
		require            *string
		semanticVersioning *string
		typeStr            *string
		url                *string
		urlCommands        *string
		usePreRelease      *string
		previous           *Lookup
		errRegex           string
		want               *Lookup
	}{
		"all nil": {
			previous: testLookup(true, true),
			want:     testLookup(true, true),
		},
		"access token": {
			accessToken: stringPtr("foo"),
			previous:    testLookup(true, true),
			want: &Lookup{
				AccessToken: stringPtr("foo"),

				Type:              test.Type,
				URL:               test.URL,
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           test.Options,
				Status:            test.Status,
			},
		},
		"allow invalid certs": {
			allowInvalidCerts: stringPtr("false"),
			previous:          testLookup(true, true),
			want: &Lookup{
				AllowInvalidCerts: boolPtr(false),

				Type:        test.Type,
				URL:         test.URL,
				URLCommands: test.URLCommands,
				Require:     test.Require,
				Options:     test.Options,
				Status:      test.Status,
			},
		},
		"require": {
			require:  stringPtr(`{"docker":{"type": "ghcr", "image": "release-argus/Argus", "tag": "latest"}}`),
			previous: testLookup(true, true),
			want: &Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Type:  "ghcr",
						Image: "release-argus/Argus",
						Tag:   "latest"}},

				Type:              test.Type,
				URL:               test.URL,
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Options:           test.Options,
				Status:            test.Status,
			},
		},
		"require - default docker.type to 'hub'": {
			require:  stringPtr(`{"docker":{"type": "", "image": "release-argus/Argus", "tag": "latest"}}`),
			previous: testLookup(true, true),
			errRegex: `^$`,
		},
		"require - invalid": {
			require:  stringPtr(`{"docker":{"type": "foo", "image": "release-argus/Argus", "tag": "latest"}}`),
			previous: testLookup(true, true),
			errRegex: `type: ".*" <invalid>`,
		},
		"semantic versioning": {
			semanticVersioning: stringPtr("false"),
			previous:           testLookup(true, true),
			want: &Lookup{
				Options: &opt.Options{
					SemanticVersioning: boolPtr(false),
				},

				Type:              test.Type,
				URL:               test.URL,
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Status:            test.Status,
			},
		},
		"url": {
			url:      stringPtr("https://valid.release-argus.io/json"),
			previous: testLookup(true, true),
			want: &Lookup{
				URL: "https://valid.release-argus.io/json",

				Type:              test.Type,
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           test.Options,
				Status:            test.Status,
			},
		},
		"url commands": {
			urlCommands: stringPtr(`[{"type": "regex", "regex": "v?([0-9.]})"}]`),
			previous:    testLookup(true, true),
			want: &Lookup{
				URLCommands: filter.URLCommandSlice{
					{Type: "regex", Regex: stringPtr("v?([0-9.]})")}},

				Type:              test.Type,
				URL:               test.URL,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           test.Options,
				Status:            test.Status,
			},
		},
		"url commands - invalid": {
			urlCommands: stringPtr(`[{"type": "foo", "regex": "v?([0-9.]})"}]`),
			previous:    testLookup(true, true),
			want:        nil,
			errRegex:    `type: .* <invalid>`,
		},
		"use prerelease": {
			usePreRelease: stringPtr("true"),
			previous:      testLookup(true, true),
			want: &Lookup{
				UsePreRelease: boolPtr(true),

				Type:              test.Type,
				URL:               test.URL,
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           &opt.Options{},
				Status:            test.Status,
			},
		},
		"type github defaulted if not set": {
			url: stringPtr("release-argus/Argus"),
			previous: &Lookup{
				Options: &opt.Options{
					Defaults:     &opt.Options{},
					HardDefaults: &opt.Options{}},
				Status: &svcstatus.Status{
					ServiceID: stringPtr("test")}},
			want: &Lookup{
				Type:       "github",
				URL:        "release-argus/Argus",
				GitHubData: &GitHubData{}},
		},
		"type github carries over Releases and ETag": {
			url: stringPtr("release-argus/other"),
			previous: &Lookup{
				Type:              "github",
				URL:               "release-argus/Argus",
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           test.Options,
				Status:            test.Status,
				GitHubData: &GitHubData{
					ETag: "123",
					Releases: []github_types.Release{
						{TagName: "v1.0.0"}}}},
			want: &Lookup{
				Type:              "github",
				URL:               "release-argus/other",
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           test.Options,
				Status:            test.Status,
				GitHubData: &GitHubData{
					ETag: "123",
					Releases: []github_types.Release{
						{TagName: "v1.0.0"}}}},
		},
		"GitHubData removed if type changed from github": {
			url:     stringPtr("https://valid.release-argus.io/json"),
			typeStr: stringPtr("url"),
			previous: &Lookup{
				Type:              "github",
				URL:               "release-argus/Argus",
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           test.Options,
				Status:            test.Status,
				GitHubData: &GitHubData{
					ETag: "123",
					Releases: []github_types.Release{
						{TagName: "v1.0.0"}}}},
			want: &Lookup{
				Type:              "url",
				URL:               "https://valid.release-argus.io/json",
				URLCommands:       test.URLCommands,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Require:           test.Require,
				Options:           test.Options,
				Status:            test.Status},
		},
		"override with invalid (empty) url": {
			url:      stringPtr(""),
			previous: testLookup(true, true),
			want:     nil,
			errRegex: "url: <required>",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we call applyOverrides
			got, err := tc.previous.applyOverrides(
				tc.accessToken,
				tc.allowInvalidCerts,
				tc.require,
				tc.semanticVersioning,
				tc.typeStr,
				tc.url,
				tc.urlCommands,
				tc.usePreRelease,
				&name,
				&util.LogFrom{Primary: name})

			// THEN we get an error if expected
			if tc.errRegex != "" || err != nil {
				// No error expected
				if tc.errRegex == "" {
					tc.errRegex = "^$"
				}
				e := util.ErrorToString(err)
				re := regexp.MustCompile(tc.errRegex)
				match := re.MatchString(e)
				if !match {
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
				return
			}
			// AND we get the expected result otherwise
			if tc.want.String() != got.String() {
				t.Errorf("expected:\n%v\nbut got:\n%v", tc.want, got)
			}
			// AND the GitHubData is only carried over to github types
			if tc.want.GitHubData.String() != got.GitHubData.String() {
				t.Errorf("expected:\n%v\nbut got:\n%v",
					tc.want.GitHubData, got.GitHubData)
			}
		})
	}
}

func TestLookup_Refresh(t *testing.T) {
	testLogging("DEBUG")
	testURL := testLookup(true, true)
	testURL.Query(true, &util.LogFrom{})
	testVersionURL := testURL.Status.GetLatestVersion()
	testGitHub := testLookup(false, false)
	testGitHub.Query(true, &util.LogFrom{})
	testVersionGitHub := testGitHub.Status.GetLatestVersion()

	// GIVEN a Lookup and various json strings to override parts of it
	tests := map[string]struct {
		accessToken        *string
		allowInvalidCerts  *string
		require            *string
		semanticVersioning *string
		typeStr            *string
		url                *string
		urlCommands        *string
		usePreRelease      *string
		latestVersion      string
		previous           *Lookup
		errRegex           string
		want               string
		announce           bool
	}{
		"Change of URL": {
			url:      stringPtr("https://valid.release-argus.io/plain"),
			previous: testLookup(true, true),
			want:     testVersionURL,
		},
		"Removal of URL": {
			url:      stringPtr(""),
			previous: testLookup(true, true),
			errRegex: "url: <required>",
			want:     "",
		},
		"Change of a few vars": {
			urlCommands:        stringPtr(`[{"type": "regex", "regex": "beta: \"v?([^\"]+)"}]`),
			semanticVersioning: stringPtr("false"),
			previous:           testLookup(true, true),
			want:               testVersionURL + "-beta",
		},
		"Change of vars that fail Query": {
			allowInvalidCerts: stringPtr("false"),
			previous:          testLookup(true, true),
			errRegex:          `x509 \(certificate invalid\)`,
		},
		"URL - Refresh new version": {
			previous: &Lookup{
				URL:               testURL.URL,
				Type:              testURL.Type,
				AllowInvalidCerts: testURL.AllowInvalidCerts,
				URLCommands:       testURL.URLCommands,
				Require:           testURL.Require,
				Options:           testURL.Options,
				Status: &svcstatus.Status{
					ServiceID: stringPtr("Refresh new version")},
				Defaults:     testURL.Defaults,
				HardDefaults: testURL.HardDefaults,
			},
			want:     testVersionURL,
			announce: true,
		},
		"GitHub - Refresh new version": {
			previous: &Lookup{
				URL:               testGitHub.URL,
				Type:              testGitHub.Type,
				AllowInvalidCerts: testGitHub.AllowInvalidCerts,
				UsePreRelease:     testGitHub.UsePreRelease,
				URLCommands:       testGitHub.URLCommands,
				Require:           testGitHub.Require,
				Options:           testGitHub.Options,
				GitHubData:        testGitHub.GitHubData,
				Status: &svcstatus.Status{
					ServiceID: stringPtr("Refresh new version")},
				Defaults:     testGitHub.Defaults,
				HardDefaults: testGitHub.HardDefaults,
			},
			latestVersion: "0.0.0",
			want:          testVersionGitHub,
			announce:      true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Copy the starting status
			tc.previous.Status.Init(
				0, 0, 0,
				&name,
				nil)
			// Set the latest version
			if tc.latestVersion != "" {
				tc.previous.Status.SetLatestVersion(tc.latestVersion, false)
			}
			previousStatus := tc.previous.Status

			// WHEN we call Refresh
			got, gotAnnounce, err := tc.previous.Refresh(
				tc.accessToken,
				tc.allowInvalidCerts,
				tc.require,
				tc.semanticVersioning,
				tc.typeStr,
				tc.url,
				tc.urlCommands,
				tc.usePreRelease)

			// THEN we get an error if expected
			if tc.errRegex != "" || err != nil {
				// No error expected
				if tc.errRegex == "" {
					tc.errRegex = "^$"
				}
				e := util.ErrorToString(err)
				re := regexp.MustCompile(tc.errRegex)
				match := re.MatchString(e)
				if !match {
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
			}
			// AND announce is only true when expected
			if tc.announce != gotAnnounce {
				t.Errorf("expected announce of %t, not %t",
					tc.announce, gotAnnounce)
			}
			// AND we get the expected result otherwise
			if tc.want != got {
				t.Errorf("expected version %q, not %q", tc.want, got)
			}
			// AND the timestamp only changes if the version changed
			if previousStatus.GetLatestVersionTimestamp() != "" {
				// If the possible query-changing overrides are nil
				if tc.require == nil && tc.semanticVersioning == nil && tc.url == nil && tc.urlCommands == nil {
					// The timestamp should change only if the version changed
					if previousStatus.GetLatestVersion() != tc.previous.Status.GetLatestVersion() &&
						previousStatus.GetLatestVersionTimestamp() == tc.previous.Status.GetLatestVersionTimestamp() {
						t.Errorf("expected timestamp to change from %q, but got %q",
							previousStatus.GetLatestVersionTimestamp(), tc.previous.Status.GetLatestVersionTimestamp())
						// The timestamp shouldn't change as the version didn't change
					} else if previousStatus.GetLatestVersionTimestamp() != tc.previous.Status.GetLatestVersionTimestamp() {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.GetLatestVersionTimestamp(), tc.previous.Status.GetLatestVersionTimestamp())
					}
					// If the overrides are not nil
				} else {
					// The timestamp shouldn't change
					if previousStatus.GetLatestVersionTimestamp() != tc.previous.Status.GetLatestVersionTimestamp() {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.GetLatestVersionTimestamp(), tc.previous.Status.GetLatestVersionTimestamp())
					}
				}
			}
		})
	}
}
