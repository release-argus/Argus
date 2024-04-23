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
	"strings"
	"sync"
	"testing"

	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
)

var emptyListETagTestMutex = sync.Mutex{}

func TestGetEmptyListETag(t *testing.T) {
	// GIVEN emptyListETag exists
	emptyListETagTestMutex.Lock()
	defer emptyListETagTestMutex.Unlock()
	emptyListETagMutex.RLock()
	defer emptyListETagMutex.RUnlock()

	// WHEN getEmptyListETag is called
	got := getEmptyListETag()

	// THEN the emptyListETag is returned
	if got != emptyListETag {
		t.Errorf("getEmptyListETag() = %q, want %q", got, emptyListETag)
	}
}

func TestSetEmptyListETag(t *testing.T) {
	// GIVEN emptyListETag exists
	emptyListETagTestMutex.Lock()
	defer emptyListETagTestMutex.Unlock()

	// WHEN setEmptyListETag is called
	newValue := "foo"
	setEmptyListETag(newValue)

	// THEN the emptyListETag is set
	if emptyListETag != newValue {
		t.Errorf("setEmptyListETag() = %q, want %q",
			emptyListETag, newValue)
	}
}

func TestFindEmptyListETag(t *testing.T) {
	// GIVEN emptyListETag is set to the incorrect value
	emptyListETagTestMutex.Lock()
	defer emptyListETagTestMutex.Unlock()
	incorrectValue := "foo"
	setEmptyListETag(incorrectValue)

	// WHEN FindEmptyListETag is called
	FindEmptyListETag(os.Getenv("GITHUB_TOKEN"))

	// THEN the emptyListETag is set
	setTo := getEmptyListETag()
	if setTo == incorrectValue {
		t.Errorf("emptyListETag wasn't updated. Got %q, want %q",
			setTo, emptyListETag)
	}
	if setTo != initialEmptyListETag {
		t.Errorf("Empty list ETag has changed from %q to %q",
			initialEmptyListETag, setTo)
	}
}

func TestNewGitHubData(t *testing.T) {
	emptyListETagTestMutex.Lock()
	defer emptyListETagTestMutex.Unlock()
	startingEmptyListETag := getEmptyListETag()
	// GIVEN a GitHubData is wanted with/without an eTag/releases
	tests := map[string]struct {
		eTag     string
		releases *[]github_types.Release
		want     *GitHubData
	}{
		"no eTag or releases": {
			eTag:     "",
			releases: nil,
			want: &GitHubData{
				eTag:     startingEmptyListETag,
				releases: []github_types.Release{},
			},
		},
		"eTag but no releases": {
			eTag:     "foo",
			releases: nil,
			want: &GitHubData{
				eTag:     "foo",
				releases: []github_types.Release{},
			},
		},
		"no eTag but releases": {
			eTag: "",
			releases: &[]github_types.Release{
				{TagName: "bar"}},
			want: &GitHubData{
				eTag: startingEmptyListETag,
				releases: []github_types.Release{
					{TagName: "bar"}},
			},
		},
		"eTag and releases": {
			eTag: "zing",
			releases: &[]github_types.Release{
				{TagName: "zap"}},
			want: &GitHubData{
				eTag: "zing",
				releases: []github_types.Release{
					{TagName: "zap"}},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN NewGitHubData is called
			got := NewGitHubData(tc.eTag, tc.releases)

			// THEN the correct GitHubData is returned
			if got.eTag != tc.want.eTag {
				t.Errorf("eTag: got %q, want %q",
					got.eTag, tc.want.eTag)
			}
			if len(got.releases) != len(tc.want.releases) {
				t.Errorf("releases: got %v, want %v",
					got.releases, tc.want.releases)
			} else {
				for i, release := range got.releases {
					if release.TagName != tc.want.releases[i].TagName {
						t.Errorf("%d: TagName, got %q (%v), want %q (%v)",
							i, got.releases[i].TagName, got.releases, tc.want.releases[i].TagName, tc.want.releases)
					}
				}
			}
		})
	}
}

func TestGitHubData_hasReleases(t *testing.T) {
	// GIVEN a GitHubData
	tests := map[string]struct {
		gd   *GitHubData
		want bool
	}{
		"no releases": {
			gd:   &GitHubData{},
			want: false,
		},
		"1 release": {
			gd: &GitHubData{
				releases: []github_types.Release{
					{TagName: "foo"}}},
			want: true,
		},
		"multiple releases": {
			gd: &GitHubData{
				releases: []github_types.Release{
					{TagName: "foo"},
					{TagName: "bar"}}},
			want: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN hasReleases is called
			got := tc.gd.hasReleases()

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGitHubData_TagFallback(t *testing.T) {
	// GIVEN a GitHubData
	gd := &GitHubData{}
	tests := []bool{
		true, false, true, false, true}

	if gd.tagFallback != false {
		t.Fatalf("tagFallback wasn't set to false initially")
	}

	for _, tc := range tests {
		gd.SetTagFallback()

		// WHEN TagFallback is called
		got := gd.TagFallback()

		// THEN the correct value is returned
		if got != tc {
			t.Errorf("got %t, want %t", got, tc)
		}
	}
}

func TestGitHubData_Copy(t *testing.T) {
	// GIVEN a fresh GitHubData and a GitHubData to copy from
	tests := map[string]struct {
		fresh *GitHubData
		gd    *GitHubData
	}{
		"empty": {
			gd: &GitHubData{},
		},
		"filled": {
			gd: &GitHubData{
				eTag: "foo",
				releases: []github_types.Release{
					{TagName: "bar"}}},
		},
		"filled with data to overwrite": {
			fresh: &GitHubData{
				eTag: "fizz",
				releases: []github_types.Release{
					{TagName: "bang"}}},
			gd: &GitHubData{
				eTag: "foo",
				releases: []github_types.Release{
					{TagName: "bar"}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.fresh == nil {
				tc.fresh = &GitHubData{}
			}

			// WHEN Copy is called
			tc.fresh.Copy(tc.gd)

			// THEN the correct GitHubData is returned
			if tc.fresh.eTag != tc.gd.eTag {
				t.Errorf("eTag: got %q, want %q",
					tc.fresh.eTag, tc.gd.eTag)
			}
			if len(tc.fresh.releases) != len(tc.gd.releases) {
				t.Errorf("releases: got %v, want %v",
					tc.fresh.releases, tc.gd.releases)
			} else {
				for i, release := range tc.fresh.releases {
					if release.TagName != tc.gd.releases[i].TagName {
						t.Errorf("%d: TagName, got %q (%v), want %q (%v)",
							i, tc.fresh.releases[i].TagName, tc.fresh.releases, tc.gd.releases[i].TagName, tc.gd.releases)
					}
				}
			}
		})
	}
}

func TestLookup_String(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		lookup *Lookup
		want   string
	}{
		"empty": {
			lookup: &Lookup{},
			want:   "{}\n",
		},
		"filled": {
			lookup: New(
				test.StringPtr("token"),
				test.BoolPtr(true),
				nil,
				opt.New(
					nil, "1h2m3s", nil,
					nil, nil),
				&filter.Require{
					RegexContent: "foo.tar.gz",
				},
				nil,
				"github",
				"https://test.com",
				&filter.URLCommandSlice{
					{Type: "regex", Regex: test.StringPtr("v([0-9.]+)")}},
				test.BoolPtr(true),
				NewDefaults(
					test.StringPtr("foo"), nil, nil, nil),
				NewDefaults(
					nil, test.BoolPtr(true), nil, nil)),
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
			lookup: New(
				test.StringPtr(">123"),
				nil, nil, nil, nil, nil, "", "",
				&filter.URLCommandSlice{
					{Type: "regex", Regex: test.StringPtr("{2}([0-9.]+)")}},
				nil, nil, nil),
			want: `
access_token: '>123'
url_commands:
  - type: regex
    regex: '{2}([0-9.]+)'
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Lookup is stringified with String
			got := tc.lookup.String("")

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
		githubData *GitHubData
		want       string
	}{
		"nil": {
			githubData: nil,
			want:       ""},
		"empty": {
			githubData: &GitHubData{},
			want:       "{}"},
		"filled": {
			githubData: &GitHubData{
				eTag: "argus",
				releases: []github_types.Release{
					{URL: "https://test.com/1.2.3"},
					{URL: "https://test.com/3.2.1"},
				}},
			want: `
				{
					"etag": "argus",
					"releases": [
						{"url": "https://test.com/1.2.3"},
						{"url": "https://test.com/3.2.1"}
					]
				}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

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
				Defaults: NewDefaults(
					nil, test.BoolPtr(false), nil, nil)},
			b:    &Lookup{},
			want: true,
		},
		"hard_defaults ignored": {
			a: &Lookup{
				HardDefaults: NewDefaults(
					nil, test.BoolPtr(false), nil, nil)},
			b:    &Lookup{},
			want: true,
		},
		"equal": {
			a: New(
				test.StringPtr("token"),
				test.BoolPtr(false),
				nil,
				opt.New(
					nil, "", test.BoolPtr(true),
					nil, nil),
				&filter.Require{
					RegexContent: "foo.tar.gz"},
				nil,
				"github",
				"https://example.com",
				nil, nil,
				NewDefaults(
					test.StringPtr("foo"), nil, nil, nil),
				NewDefaults(
					nil, test.BoolPtr(true), nil, nil)),
			b: New(
				test.StringPtr("token"),
				test.BoolPtr(false),
				nil,
				opt.New(
					nil, "", test.BoolPtr(true),
					nil, nil),
				&filter.Require{
					RegexContent: "foo.tar.gz"},
				nil,
				"github",
				"https://example.com",
				nil, nil,
				NewDefaults(
					test.StringPtr("foo"), nil, nil, nil),
				NewDefaults(
					nil, test.BoolPtr(true), nil, nil)),
			want: true,
		},
		"not equal": {
			a: New(
				test.StringPtr("token"),
				test.BoolPtr(false),
				nil,
				opt.New(
					nil, "", test.BoolPtr(true),
					nil, nil),
				&filter.Require{
					RegexContent: "foo.tar.gz"},
				nil,
				"github",
				"https://example.com",
				nil,
				test.BoolPtr(true),
				NewDefaults(
					test.StringPtr("foo"), nil, nil, nil),
				NewDefaults(
					nil, test.BoolPtr(true), nil, nil)),
			b: New(
				test.StringPtr("token"),
				test.BoolPtr(false),
				nil,
				opt.New(
					nil, "", test.BoolPtr(true),
					nil, nil),
				&filter.Require{
					RegexContent: "foo.tar.gz"},
				nil,
				"github",
				"https://example.com/other",
				nil,
				test.BoolPtr(true),
				NewDefaults(
					test.StringPtr("foo"), nil, nil, nil),
				NewDefaults(
					nil, test.BoolPtr(true), nil, nil)),
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Set Status vars just to ensure they're not printed
			if tc.a != nil {
				tc.a.Status = &svcstatus.Status{}
				tc.a.Status.Init(
					0, 0, 0,
					&name,
					test.StringPtr("http://example.com"))
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
