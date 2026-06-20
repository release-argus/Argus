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
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestLookup_GetType(t *testing.T) {
	// GIVEN: a Lookup with a Type.
	tests := []struct {
		name  string
		lType string
	}{
		{name: "empty", lType: ""},
		{name: "test", lType: "test"},
		{name: "x", lType: "x"},
		{name: "y", lType: "y"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &Lookup{}
			l.Type = tc.lType

			// WHEN: GetType is called.
			got := l.GetType()

			wantType := Type
			// THEN: the Type is returned.
			if got != wantType {
				t.Errorf(
					"%s\nLookup.GetType() mismatch\ngot:  %q\nwant: %q",
					packageName, got, wantType,
				)
			}
		})
	}
}

func TestLookup_AccessToken(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		{
			name:             "root overrides all",
			rootValue:        "a",
			defaultValue:     "b",
			hardDefaultValue: "c",
			want:             "a",
		},
		{
			name:             "default overrides hardDefault",
			defaultValue:     "b",
			hardDefaultValue: "c",
			want:             "b",
		},
		{
			name:             "hardDefault is last resort",
			hardDefaultValue: "c",
			want:             "c",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.AccessToken = tc.rootValue
			lookup.Defaults.AccessToken = tc.defaultValue
			lookup.HardDefaults.AccessToken = tc.hardDefaultValue

			// WHEN: accessToken is called.
			got := lookup.accessToken()

			// THEN: the expected value is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.accessToken() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLookup_URL(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name        string
		url         string
		tagFallback bool
		page        int
		perPage     int
		want        string
	}{
		{
			name: "Repo",
			url:  test.ArgusGitHubRepo,
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/releases",
		},
		{
			name:        "Repo with tag fallback",
			url:         test.ArgusGitHubRepo,
			tagFallback: true,
			want:        "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/tags",
		},
		{
			name: "API URL",
			url:  "https://api.github.com/repos/" + test.ArgusGitHubRepo,
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo,
		},
		{
			name: "Repo with page 1",
			url:  test.ArgusGitHubRepo,
			page: 1,
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/releases",
		},
		{
			name: "Repo with page >1",
			url:  test.ArgusGitHubRepo,
			page: 2,
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/releases?page=2",
		},
		{
			name:    "Repo with per_page",
			url:     test.ArgusGitHubRepo,
			perPage: 2,
			want: fmt.Sprintf(
				"https://api.github.com/repos/"+test.ArgusGitHubRepo+"/releases?per_page=%d",
				2*defaultPerPage,
			),
		},
		{
			name:    "Repo with page >1 and per_page",
			url:     test.ArgusGitHubRepo,
			page:    2,
			perPage: 4,
			want: fmt.Sprintf(
				"https://api.github.com/repos/"+test.ArgusGitHubRepo+"/releases?page=2&per_page=%d",
				4*defaultPerPage,
			),
		},
		{
			name:        "Repo with tag fallback and page >1",
			url:         test.ArgusGitHubRepo,
			tagFallback: true,
			page:        3,
			want:        "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/tags?page=3",
		},
		{
			name:        "Repo with tag fallback, page >1, and per_page",
			url:         test.ArgusGitHubRepo,
			tagFallback: true,
			page:        3,
			perPage:     7,
			want: fmt.Sprintf(
				"https://api.github.com/repos/"+test.ArgusGitHubRepo+"/tags?page=3&per_page=%d",
				7*defaultPerPage,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.URL = tc.url
			if tc.tagFallback {
				lookup.GetGitHubData().SetTagFallback()
			}
			lookup.data.SetPerPage(tc.perPage)

			// WHEN: url is called with the page argument.
			got := lookup.url(tc.page)

			// THEN: the expected value is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.url(%d) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.page,
					got, tc.want,
				)
			}
		})
	}
}

func TestLookup_UsePreRelease(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      *bool
	}{
		{
			name:             "root overrides all",
			rootValue:        test.Ptr(false),
			defaultValue:     test.Ptr(true),
			hardDefaultValue: test.Ptr(true),
			want:             test.Ptr(false),
		},
		{
			name:             "default overrides hardDefault",
			defaultValue:     test.Ptr(true),
			hardDefaultValue: test.Ptr(false),
			want:             test.Ptr(true),
		},
		{
			name:             "hardDefault is last resort",
			hardDefaultValue: test.Ptr(false),
			want:             test.Ptr(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.UsePreRelease = tc.rootValue
			lookup.Defaults.UsePreRelease = tc.defaultValue
			lookup.HardDefaults.UsePreRelease = tc.hardDefaultValue

			// WHEN: usePreRelease is called.
			result := lookup.usePreRelease()

			// THEN: the expected value is returned.
			wantStr := test.StringifyPtr(tc.want)
			if got := fmt.Sprint(result); got != wantStr {
				t.Errorf(
					"%s\nLookup.usePreRelease() mismatch\ngot:  %q\nwant: %q",
					packageName, wantStr, got,
				)
			}
		})
	}
}

func TestServiceURL(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "owner-repo",
			url:  test.ArgusGitHubRepo,
			want: "https://github.com/" + test.ArgusGitHubRepo,
		},
		{
			name: "GitHub url",
			url:  "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/tags",
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/tags",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.URL = tc.url

			// WHEN: ServiceURL is called.
			got := lookup.ServiceURL()

			// THEN: the expected value is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.ServiceURL() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLookup_GetGitHubData(t *testing.T) {
	// GIVEN: a Lookup.
	lookup := testLookup(t, false)

	// WHEN: GetGitHubData is called.
	got := lookup.GetGitHubData()

	// THEN: the address of the data is returned.
	want := &lookup.data
	if got != want {
		t.Errorf(
			"%s\nLookup.GetGitHubData() pointer mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}
