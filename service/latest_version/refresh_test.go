// Copyright [2023] [Argus]
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
	"os"
	"regexp"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestLookup_ApplyOverrides(t *testing.T) {
	test := testLookup(true, true)
	// GIVEN various json strings to parse as parts of a Lookup
	tests := map[string]struct {
		accessToken         *string
		allowInvalidCerts   *string
		require             *string
		semanticVersioning  *string
		typeStr             *string
		url                 *string
		urlCommands         *string
		usePreRelease       *string
		previous            *Lookup
		gitHubData          *GitHubData
		carryOverGitHubData bool
		errRegex            string
		want                *Lookup
	}{
		"all nil": {
			previous: testLookup(true, true),
			want:     testLookup(true, true),
		},
		"access token": {
			accessToken: stringPtr("foo"),
			previous:    testLookup(true, true),
			want: New(
				stringPtr("foo"),
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				test.Type,
				test.URL,
				&test.URLCommands,
				test.UsePreRelease,
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"allow invalid certs": {
			allowInvalidCerts: stringPtr("false"),
			previous:          testLookup(true, true),
			want: New(
				test.AccessToken,
				boolPtr(false),
				nil,
				test.Options,
				test.Require,
				nil,
				test.Type,
				test.URL,
				&test.URLCommands,
				test.UsePreRelease,
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"require": {
			require: stringPtr(`{
				"docker": {
					"type": "ghcr",
					"image": "release-argus/Argus",
					"tag": "latest"}}`),
			previous: testLookup(true, true),
			want: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"ghcr",
						"release-argus/Argus",
						"latest",
						"", "", "", time.Now(), nil)},
				nil,
				test.Type,
				test.URL,
				&test.URLCommands,
				test.UsePreRelease,
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"require - no docker.type fail": {
			require: stringPtr(`{
				"docker": {
					"type": "",
					"image": "release-argus/Argus",
					"tag": "latest"}}`),
			previous: testLookup(true, true),
			errRegex: `^require:[^ ]+  docker:[^ ]+    type: <required>`,
		},
		"require - invalid": {
			require: stringPtr(`{
				"docker": {
					"type": "foo",
					"image": "release-argus/Argus",
					"tag": "latest"}}`),
			previous: testLookup(true, true),
			errRegex: `type: ".*" <invalid>`,
		},
		"semantic versioning": {
			semanticVersioning: stringPtr("false"),
			previous:           testLookup(true, true),
			want: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				opt.New(
					nil, "", boolPtr(false),
					nil, nil),
				test.Require,
				nil,
				test.Type,
				test.URL,
				&test.URLCommands,
				test.UsePreRelease,
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"url": {
			url:      stringPtr("https://valid.release-argus.io/json"),
			previous: testLookup(true, true),
			want: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				test.Type,
				"https://valid.release-argus.io/json",
				&test.URLCommands,
				test.UsePreRelease,
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"url commands": {
			urlCommands: stringPtr(`[
					{"type": "regex", "regex": "v?([0-9.]})"}
				]`),
			previous: testLookup(true, true),
			want: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				test.Type,
				test.URL,
				&filter.URLCommandSlice{
					{Type: "regex", Regex: stringPtr("v?([0-9.]})")}},
				test.UsePreRelease,
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"url commands - invalid": {
			urlCommands: stringPtr(`[
					{"type": "foo", "regex": "v?([0-9.]})"}]`),
			previous: testLookup(true, true),
			want:     nil,
			errRegex: `type: .* <invalid>`,
		},
		"use prerelease": {
			usePreRelease: stringPtr("true"),
			previous:      testLookup(true, true),
			want: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				test.Type,
				test.URL,
				&test.URLCommands,
				boolPtr(true),
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"type github defaulted if not set": {
			url: stringPtr("release-argus/Argus"),
			previous: &Lookup{
				Options: &opt.Options{},
				Status: &svcstatus.Status{
					ServiceID: stringPtr("test")}},
			want: &Lookup{
				Type:       "github",
				URL:        "release-argus/Argus",
				GitHubData: NewGitHubData("", nil)},
		},
		"type github carries over Releases and ETag": {
			url: stringPtr("release-argus/other"),
			previous: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				"github",
				"release-argus/Argus",
				nil,
				boolPtr(false),
				&LookupDefaults{},
				&LookupDefaults{}),
			carryOverGitHubData: true,
			gitHubData: &GitHubData{
				eTag: "123",
				releases: []github_types.Release{
					{TagName: "v1.0.0"}}},
			want: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				"github",
				"release-argus/other",
				nil,
				boolPtr(false),
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"GitHubData removed if type changed from github": {
			url:     stringPtr("https://valid.release-argus.io/json"),
			typeStr: stringPtr("url"),
			previous: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				"github",
				"release-argus/Argus",
				nil,
				boolPtr(false),
				&LookupDefaults{},
				&LookupDefaults{}),
			carryOverGitHubData: false,
			gitHubData: &GitHubData{
				eTag: "123",
				releases: []github_types.Release{
					{TagName: "v1.0.0"}}},
			want: New(
				test.AccessToken,
				test.AllowInvalidCerts,
				nil,
				test.Options,
				test.Require,
				nil,
				"url",
				"https://valid.release-argus.io/json",
				nil,
				boolPtr(false),
				&LookupDefaults{},
				&LookupDefaults{}),
		},
		"override with invalid (empty) url": {
			url:      stringPtr(""),
			previous: testLookup(true, true),
			want:     nil,
			errRegex: "url: <required>",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.previous.Status.ServiceID = &name
			tc.previous.GitHubData = tc.gitHubData
			if tc.carryOverGitHubData {
				tc.want.GitHubData = tc.gitHubData
			}

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
			if tc.want.String("") != got.String("") {
				t.Errorf("expected:\n%v\nbut got:\n%v",
					tc.want, got)
			}
			// AND the GitHubData is only carried over to github types
			if tc.want.GitHubData.String() != got.GitHubData.String() {
				t.Errorf("expected:\n%s\nbut got:\n%s",
					tc.want.GitHubData, got.GitHubData)
			}
		})
	}
}

func TestLookup_Refresh(t *testing.T) {
	testURL := testLookup(true, true)
	testURL.Query(true, &util.LogFrom{})
	testVersionURL := testURL.Status.LatestVersion()
	testGitHub := testLookup(false, false)
	testGitHub.AccessToken = stringPtr(os.Getenv("GITHUB_TOKEN"))
	testGitHub.Query(true, &util.LogFrom{})
	testVersionGitHub := testGitHub.Status.LatestVersion()

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
			urlCommands: stringPtr(`[
					{"type": "regex", "regex": "beta: \"v?([^\"]+)"}]`),
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
			previous: New(
				nil,
				testURL.AllowInvalidCerts,
				nil,
				testURL.Options,
				testURL.Require,
				nil,
				testURL.Type,
				testURL.URL,
				&testURL.URLCommands,
				testURL.UsePreRelease,
				testURL.Defaults,
				testURL.HardDefaults),
			want:     testVersionURL,
			announce: true,
		},
		"GitHub - Refresh new version": {
			previous: New(
				testGitHub.AccessToken,
				testGitHub.AllowInvalidCerts,
				nil,
				testGitHub.Options,
				testGitHub.Require,
				nil,
				testGitHub.Type,
				testGitHub.URL,
				&testGitHub.URLCommands,
				testGitHub.UsePreRelease,
				testGitHub.Defaults,
				testGitHub.HardDefaults),
			latestVersion: "0.0.0",
			want:          testVersionGitHub,
			announce:      true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.previous.AccessToken = stringPtr(os.Getenv("GITHUB_TOKEN"))
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
			if previousStatus.LatestVersionTimestamp() != "" {
				// If the possible query-changing overrides are nil
				if tc.require == nil && tc.semanticVersioning == nil && tc.url == nil && tc.urlCommands == nil {
					// The timestamp should change only if the version changed
					if previousStatus.LatestVersion() != tc.previous.Status.LatestVersion() &&
						previousStatus.LatestVersionTimestamp() == tc.previous.Status.LatestVersionTimestamp() {
						t.Errorf("expected timestamp to change from %q, but got %q",
							previousStatus.LatestVersionTimestamp(), tc.previous.Status.LatestVersionTimestamp())
						// The timestamp shouldn't change as the version didn't change
					} else if previousStatus.LatestVersionTimestamp() != tc.previous.Status.LatestVersionTimestamp() {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.LatestVersionTimestamp(), tc.previous.Status.LatestVersionTimestamp())
					}
					// If the overrides are not nil
				} else {
					// The timestamp shouldn't change
					if previousStatus.LatestVersionTimestamp() != tc.previous.Status.LatestVersionTimestamp() {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.LatestVersionTimestamp(), tc.previous.Status.LatestVersionTimestamp())
					}
				}
			}
		})
	}
}

func TestLookup_updateFromRefresh(t *testing.T) {
	// GIVEN a Lookup and a refreshed version of that Lookup
	tests := map[string]struct {
		previous          *Lookup
		now               *Lookup
		changingOverrides bool
		want              *Lookup
	}{
		"No changes": {
			previous: New(
				nil, nil, nil, nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil, nil, nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: false,
			want: New(
				nil, nil, nil, nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
		"new ETag - changing overrides": {
			previous: New(
				nil, nil,
				&GitHubData{
					eTag: "old-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: true,
			want: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
		"new ETag - no changing overrides - update last_queried": {
			previous: New(
				nil, nil,
				&GitHubData{
					eTag: "old-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: false,
			want: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
		"new version not set/announced if changingOverrides": {
			previous: New(
				nil, nil,
				&GitHubData{
					eTag: "old-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: true,
			want: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
		"new version set and announced if no changingOverrides": {
			previous: New(
				nil, nil,
				&GitHubData{
					eTag: "old-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: false,
			want: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
		"docker queryToken not copied over if Docker is new": {
			previous: New(
				nil, nil,
				&GitHubData{
					eTag: "old-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"hub",
						"", "",
						"argus", "SECRET",
						"", time.Time{}, nil)},
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: false,
			want: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil, nil,
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
		"docker queryToken copied over": {
			previous: New(
				nil, nil,
				&GitHubData{
					eTag: "old-etag"},
				nil,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"hub",
						"", "",
						"argus", "SECRET",
						"oldQueryToken", time.Time{}, nil)},
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"hub",
						"", "",
						"argus", "SECRET",
						"newToken", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), nil)},
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: false,
			want: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"hub",
						"", "",
						"argus", "SECRET",
						"newToken", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), nil)},
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
		"docker queryToken not copied over if Docker target changed": {
			previous: New(
				nil, nil,
				&GitHubData{
					eTag: "old-etag"},
				nil,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"hub",
						"argus", "",
						"argus", "SECRET",
						"oldQueryToken", time.Time{}, nil)},
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.4.0", "2020-01-01T00:00:00Z",
					"2021-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			now: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"hub",
						"somethingElse", "",
						"argus", "SECRET",
						"newToken", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						nil)},
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
			changingOverrides: false,
			want: New(
				nil, nil,
				&GitHubData{
					eTag: "new-etag"},
				nil,
				&filter.Require{
					Docker: filter.NewDockerCheck(
						"hub",
						"argus", "",
						"argus", "SECRET",
						"oldQueryToken", time.Time{},
						nil)},
				svcstatus.New(
					nil, nil, nil,
					"0.0.0",
					"0.1.0", "2020-01-01T00:00:00Z",
					"0.6.0", "2022-01-01T00:00:00Z",
					"2022-01-01T00:00:00Z"),
				"github",
				"release-argus/Argus",
				nil, nil, nil, nil),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbChannel := make(chan dbtype.Message, 2)
			tc.previous.Status.DatabaseChannel = &dbChannel
			tc.now.Status.ServiceID = &name
			tc.previous.Status.ServiceID = &name
			tc.want.Status.ServiceID = &name
			expectNewLatestVersion := tc.previous.Status.LatestVersion() != tc.want.Status.LatestVersion()

			// WHEN we call updateFromRefresh
			tc.previous.updateFromRefresh(tc.now, tc.changingOverrides)

			// THEN we get the expected result
			if tc.previous.String("") != tc.want.String("") {
				t.Errorf("expected\n%q\ngot\n%q",
					tc.want.String(""), tc.previous.String(""))
			}
			// ETag copied when expected
			if tc.previous.GitHubData != nil && tc.want.GitHubData == nil ||
				tc.previous.GitHubData == nil && tc.want.GitHubData != nil {
				t.Errorf("expected GitHubData %v, not %v",
					tc.want.GitHubData, tc.previous.GitHubData)
			} else if tc.previous.GitHubData != nil &&
				tc.previous.GitHubData.ETag() != tc.want.GitHubData.ETag() {
				t.Errorf("expected ETag %q, not %q",
					tc.want.GitHubData.ETag(), tc.previous.GitHubData.ETag())
			}
			// LastQueried copied over when expected
			if tc.previous.Status.LastQueried() != tc.want.Status.LastQueried() {
				t.Errorf("expected LastQueried %q, not %q",
					tc.want.Status.LastQueried(), tc.previous.Status.LastQueried())
			}
			// Docker queryToken copied over when expected
			if tc.want.Require == nil && tc.previous.Require != nil ||
				tc.want.Require != nil && tc.previous.Require == nil {
				t.Errorf("expected Require %v, not %v",
					tc.want.Require, tc.previous.Require)
			} else if tc.want.Require == nil { // No Require in either, skip
			} else if tc.want.Require.Docker == nil && tc.previous.Require.Docker != nil ||
				tc.want.Require.Docker != nil && tc.previous.Require.Docker == nil {
				t.Errorf("expected Docker %v, not %v",
					tc.want.Require.Docker, tc.previous.Require.Docker)
			} else if tc.want.Require != nil && tc.want.Require.Docker != nil {
				wantToken, wantValidUntil := tc.want.Require.Docker.CopyQueryToken()
				gotToken, gotValidUntil := tc.previous.Require.Docker.CopyQueryToken()
				if wantToken != gotToken {
					t.Errorf("expected Docker queryToken %q, not %q",
						wantToken, gotToken)
				}
				if wantValidUntil != gotValidUntil {
					t.Errorf("expected Docker queryToken validUntil %q, not %q",
						wantValidUntil, gotValidUntil)
				}
			}
			// latest_version copied over when expected
			if tc.want.Status.LatestVersion() != tc.previous.Status.LatestVersion() {
				t.Errorf("expected LatestVersion %q, not %q",
					tc.want.Status.LatestVersion(), tc.previous.Status.LatestVersion())
			} else if expectNewLatestVersion {
				if len(*tc.previous.Status.DatabaseChannel) != 1 {
					t.Errorf("expected 1 message on database channel, not %d",
						len(*tc.previous.Status.DatabaseChannel))
				}
			}
		})
	}
}
