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
	"strconv"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
)

func TestLookup_GetType(t *testing.T) {
	l := &Lookup{}
	l.Type = "other"

	got := l.GetType()

	if got != Type {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, Type)
	}
}

func TestLookup_AccessToken(t *testing.T) {
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

			lookup := testLookup(t)
			lookup.AccessToken = tc.rootValue
			lookup.typeDefaults.AccessToken = tc.defaultValue
			lookup.typeHardDefaults.AccessToken = tc.hardDefaultValue

			got := lookup.accessToken()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_URL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		tagFallback bool
		page        int
		foundOnPage int
		want        string
	}{
		{
			name: "Repo escaping/a repo cannot climb out of the /repos prefix",
			url:  "../..",
			want: "https://api.github.com/repos/%2E%2E/%2E%2E/releases",
		},
		{
			name: "Repo/releases",
			url:  test.ArgusGitHubRepo,
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/releases",
		},
		{
			name:        "Repo/tags",
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
			name: "Repo/page 1",
			url:  test.ArgusGitHubRepo,
			page: 1,
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/releases",
		},
		{
			name: "Repo/page >1",
			url:  test.ArgusGitHubRepo,
			page: 2,
			want: "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/releases?page=2",
		},
		{
			name:        "Repo/per_page",
			url:         test.ArgusGitHubRepo,
			foundOnPage: 2,
			want: fmt.Sprintf(
				"https://api.github.com/repos/%s/releases?per_page=%d",
				test.ArgusGitHubRepo, 2*defaultPerPage,
			),
		},
		{
			name:        "Repo/page >1 and per_page",
			url:         test.ArgusGitHubRepo,
			page:        2,
			foundOnPage: 4,
			want: fmt.Sprintf(
				"https://api.github.com/repos/%s/releases?page=2&per_page=%d",
				test.ArgusGitHubRepo, 4*defaultPerPage,
			),
		},
		{
			name:        "Repo/tag fallback and page >1",
			url:         test.ArgusGitHubRepo,
			tagFallback: true,
			page:        3,
			want:        "https://api.github.com/repos/" + test.ArgusGitHubRepo + "/tags?page=3",
		},
		{
			name:        "Repo/tag fallback, page >1, and per_page",
			url:         test.ArgusGitHubRepo,
			tagFallback: true,
			page:        3,
			foundOnPage: 7,
			want: fmt.Sprintf(
				"https://api.github.com/repos/%s/tags?page=3&per_page=%d",
				test.ArgusGitHubRepo, 7*defaultPerPage,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t)
			lookup.URL = tc.url
			if tc.tagFallback {
				lookup.GetGitHubData().SetTagFallback()
			}
			lookup.data.SetPerPage(tc.foundOnPage)

			got := lookup.url(tc.page)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_UsePreRelease(t *testing.T) {
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      *bool
	}{
		{
			name:             "root overrides all",
			rootValue:        new(false),
			defaultValue:     new(true),
			hardDefaultValue: new(true),
			want:             new(false),
		},
		{
			name:             "default overrides hardDefault",
			defaultValue:     new(true),
			hardDefaultValue: new(false),
			want:             new(true),
		},
		{
			name:             "hardDefault is last resort",
			hardDefaultValue: new(false),
			want:             new(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{UsePreRelease: tc.rootValue}
			lookup.SetTypeDefaults(
				&Defaults{UsePreRelease: tc.defaultValue},
				&Defaults{UsePreRelease: tc.hardDefaultValue},
			)

			result := lookup.usePreRelease()

			wantStr := test.StringifyPtr(tc.want)
			if got := strconv.FormatBool(result); got != wantStr {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, wantStr)
			}
		})
	}
}

func TestLookup_UseTagsAPI(t *testing.T) {
	tests := []struct {
		name          string
		usePreRelease *bool
		require       *filter.Require
		want          bool
	}{
		{
			name:    "require/nil - allowed",
			require: nil,
			want:    true,
		},
		{
			name:    "require/empty - allowed",
			require: &filter.Require{},
			want:    true,
		},
		{
			name:    "require/regex_content=empty - allowed",
			require: &filter.Require{RegexContent: ""},
			want:    true,
		},
		{
			name:    "require/regex_content=set - blocked",
			require: &filter.Require{RegexContent: "some-pattern"},
			want:    false, // regex_content needs release assets.
		},
		{
			name:          "use_prerelease/false - allowed",
			usePreRelease: new(false),
			want:          true,
		},
		{
			name:          "use_prerelease/true - allowed, as tag names carry the label",
			usePreRelease: new(true),
			want:          true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t)
			lookup.UsePreRelease = tc.usePreRelease
			lookup.Require = tc.require

			got := lookup.useTagsAPI()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestServiceURL(t *testing.T) {
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

			lookup := testLookup(t)
			lookup.URL = tc.url

			got := lookup.ServiceURL()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_GetGitHubData(t *testing.T) {
	lookup := testLookup(t)

	got := lookup.GetGitHubData()

	want := &lookup.data
	if got != want {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, want)
	}
}
