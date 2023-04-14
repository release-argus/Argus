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
	"strings"
	"testing"

	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
)

func TestLookup_String(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		lookup Lookup
		want   string
	}{
		"empty": {
			lookup: Lookup{},
			want:   "{}\n",
		},
		"filled": {
			lookup: Lookup{
				Type:              "github",
				URL:               "https://test.com",
				AccessToken:       stringPtr("token"),
				AllowInvalidCerts: boolPtr(true),
				UsePreRelease:     boolPtr(true),
				URLCommands: []filter.URLCommand{
					{Type: "regex", Regex: stringPtr("v([0-9.]+)")}},
				Require: &filter.Require{
					RegexContent: "foo.tar.gz",
				},
				Options: &opt.Options{
					Interval: "1h2m3s",
				},
				GitHubData: &GitHubData{
					ETag: "etag",
				},
				Defaults: &Lookup{
					Type: "gitlab",
				},
				HardDefaults: &Lookup{
					AllowInvalidCerts: boolPtr(true),
				},
			},
			want: `
type: github
url: https://test.com
access_token: token
allow_invalid_certs: true
use_prerelease: true
url_commands:
    - type: regex
      regex: v([0-9.]+)
require:
    regex_content: foo.tar.gz
`,
		},
		"quotes otherwise invalid yaml strings": {
			lookup: Lookup{
				AccessToken: stringPtr(">123"),
				URLCommands: filter.URLCommandSlice{
					{Type: "regex", Regex: stringPtr("{2}([0-9.]+)")}}},
			want: `
access_token: '>123'
url_commands:
    - type: regex
      regex: '{2}([0-9.]+)'
`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Lookup is stringified with String
			got := tc.lookup.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestGitHubData_String(t *testing.T) {
	// GIVEN a GitHubData
	tests := map[string]struct {
		githubData GitHubData
		want       string
	}{
		"empty": {
			githubData: GitHubData{},
			want:       `{"etag":""}`},
		"filled": {
			githubData: GitHubData{
				ETag: "argus",
				Releases: []github_types.Release{
					{URL: "https://test.com/1.2.3"},
					{URL: "https://test.com/3.2.1"},
				}},
			want: `{"etag":"argus","releases":[{"url":"https://test.com/1.2.3"},{"url":"https://test.com/3.2.1"}]}`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the GitHubData is stringified with String
			got := tc.githubData.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestLookup_IsEqual(t *testing.T) {
	// GIVEN two Lookups
	tests := map[string]struct {
		a, b *Lookup
		want bool
	}{
		"empty": {
			a:    &Lookup{},
			b:    &Lookup{},
			want: true,
		},
		"defaults ignored": {
			a: &Lookup{
				Defaults: &Lookup{
					AllowInvalidCerts: boolPtr(false)}},
			b:    &Lookup{},
			want: true,
		},
		"hard_defaults ignored": {
			a: &Lookup{
				HardDefaults: &Lookup{
					AllowInvalidCerts: boolPtr(false)}},
			b:    &Lookup{},
			want: true,
		},
		"equal": {
			a: &Lookup{
				Type:              "github",
				URL:               "https://example.com",
				AccessToken:       stringPtr("token"),
				AllowInvalidCerts: boolPtr(false),
				UsePreRelease:     boolPtr(true),
				Require: &filter.Require{
					RegexContent: "foo.tar.gz"},
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				GitHubData: &GitHubData{
					ETag: "etag"},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
			b: &Lookup{
				Type:              "github",
				URL:               "https://example.com",
				AccessToken:       stringPtr("token"),
				AllowInvalidCerts: boolPtr(false),
				UsePreRelease:     boolPtr(true),
				Require: &filter.Require{
					RegexContent: "foo.tar.gz"},
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				GitHubData: &GitHubData{
					ETag: "etag"},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
			want: true,
		},
		"not equal": {
			a: &Lookup{
				Type:              "github",
				URL:               "https://example.com",
				AccessToken:       stringPtr("token"),
				AllowInvalidCerts: boolPtr(false),
				UsePreRelease:     boolPtr(true),
				Require: &filter.Require{
					RegexContent: "foo.tar.gz"},
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				GitHubData: &GitHubData{
					ETag: "etag"},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
			b: &Lookup{
				Type:              "github",
				URL:               "https://example.com/other",
				AccessToken:       stringPtr("token"),
				AllowInvalidCerts: boolPtr(false),
				UsePreRelease:     boolPtr(true),
				Require: &filter.Require{
					RegexContent: "foo.tar.gz"},
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				GitHubData: &GitHubData{
					ETag: "etag"},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
			want: false,
		},
		"not equal with nil": {
			a: nil,
			b: &Lookup{
				URL: "https://example.com"},
			want: false,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Set Status vars just to ensure they're not printed
			if tc.a != nil {
				tc.a.Status = &svcstatus.Status{}
				tc.a.Status.Init(
					0, 0, 0,
					&name,
					stringPtr("http://example.com"))
				tc.a.Status.SetLatestVersion("foo", false)
			}

			// WHEN the two Lookups are compared
			got := tc.a.IsEqual(tc.b)

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}
