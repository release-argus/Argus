// Copyright [2025] [Argus]
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
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestAccessToken(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		"root overrides all": {
			rootValue:        "a",
			defaultValue:     "b",
			hardDefaultValue: "c",
			want:             "a",
		},
		"default overrides hardDefault": {
			defaultValue:     "b",
			hardDefaultValue: "c",
			want:             "b",
		},
		"hardDefault is last resort": {
			hardDefaultValue: "c",
			want:             "c",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.AccessToken = tc.rootValue
			lookup.Defaults.AccessToken = tc.defaultValue
			lookup.HardDefaults.AccessToken = tc.hardDefaultValue

			// WHEN accessToken is called.
			got := lookup.accessToken()

			// THEN the expected value is returned.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestURL(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		url         string
		tagFallback bool
		page        int
		perPage     int
		want        string
	}{
		"Repo": {
			url:  "release-argus/Argus",
			want: "https://api.github.com/repos/release-argus/Argus/releases",
		},
		"Repo with tag fallback": {
			url:         "release-argus/Argus",
			tagFallback: true,
			want:        "https://api.github.com/repos/release-argus/Argus/tags",
		},
		"API URL": {
			url:  "https://api.github.com/repos/release-argus/Argus",
			want: "https://api.github.com/repos/release-argus/Argus",
		},
		"Repo with page 1": {
			url:  "release-argus/Argus",
			page: 1,
			want: "https://api.github.com/repos/release-argus/Argus/releases",
		},
		"Repo with page >1": {
			url:  "release-argus/Argus",
			page: 2,
			want: "https://api.github.com/repos/release-argus/Argus/releases?page=2",
		},
		"Repo with per_page": {
			url:     "release-argus/Argus",
			perPage: 2,
			want:    fmt.Sprintf("https://api.github.com/repos/release-argus/Argus/releases?per_page=%d", 2*defaultPerPage),
		},
		"Repo with page >1 and per_page": {
			url:     "release-argus/Argus",
			page:    2,
			perPage: 4,
			want:    fmt.Sprintf("https://api.github.com/repos/release-argus/Argus/releases?page=2&per_page=%d", 4*defaultPerPage),
		},
		"Repo with tag fallback and page >1": {
			url:         "release-argus/Argus",
			tagFallback: true,
			page:        3,
			want:        "https://api.github.com/repos/release-argus/Argus/tags?page=3",
		},
		"Repo with tag fallback, page >1, and per_page": {
			url:         "release-argus/Argus",
			tagFallback: true,
			page:        3,
			perPage:     7,
			want:        fmt.Sprintf("https://api.github.com/repos/release-argus/Argus/tags?page=3&per_page=%d", 7*defaultPerPage),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.URL = tc.url
			if tc.tagFallback {
				lookup.GetGitHubData().SetTagFallback()
			}
			lookup.data.SetPerPage(tc.perPage)

			// WHEN url is called with the page argument.
			got := lookup.url(tc.page)

			// THEN the expected value is returned.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestUsePreRelease(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      *bool
	}{
		"root overrides all": {
			rootValue:        test.BoolPtr(false),
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(true),
			want:             test.BoolPtr(false),
		},
		"default overrides hardDefault": {
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(false),
			want:             test.BoolPtr(true),
		},
		"hardDefault is last resort": {
			hardDefaultValue: test.BoolPtr(false),
			want:             test.BoolPtr(false),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.UsePreRelease = tc.rootValue
			lookup.Defaults.UsePreRelease = tc.defaultValue
			lookup.HardDefaults.UsePreRelease = tc.hardDefaultValue

			// WHEN usePreRelease is called.
			got := lookup.usePreRelease()

			// THEN the expected value is returned.
			gotStr := fmt.Sprint(got)
			wantStr := test.StringifyPtr(tc.want)
			if gotStr != wantStr {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, wantStr, gotStr)
			}
		})
	}
}

func TestServiceURL(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		url  string
		want string
	}{
		"owner/repo": {
			url:  "release-argus/Argus",
			want: "https://github.com/release-argus/Argus",
		},
		"GitHub url": {
			url:  "https://api.github.com/repos/release-argus/Argus/tags",
			want: "https://api.github.com/repos/release-argus/Argus/tags",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.URL = tc.url

			// WHEN ServiceURL is called.
			got := lookup.ServiceURL()

			// THEN the expected value is returned.
			if got != tc.want {
				t.Errorf("%s\nServiceURL mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}
