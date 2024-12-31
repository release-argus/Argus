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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestAccessToken(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		"root overrides all": {
			rootValue:        ("a"),
			defaultValue:     ("b"),
			hardDefaultValue: ("c"),
			want:             ("a"),
		},
		"default overrides hardDefault": {
			defaultValue:     ("b"),
			hardDefaultValue: ("c"),
			want:             ("b"),
		},
		"hardDefault is last resort": {
			hardDefaultValue: ("c"),
			want:             ("c"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.AccessToken = tc.rootValue
			lookup.Defaults.AccessToken = tc.defaultValue
			lookup.HardDefaults.AccessToken = tc.hardDefaultValue

			// WHEN accessToken is called
			got := lookup.accessToken()

			// THEN the expected value is returned
			if got != tc.want {
				t.Errorf("accessToken() mismatch:\nwant: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		url         string
		tagFallback bool
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
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.URL = tc.url
			if tc.tagFallback {
				lookup.GetGitHubData().SetTagFallback()
			}

			// WHEN url is called
			got := lookup.url()

			// THEN the expected value is returned
			if got != tc.want {
				t.Errorf("url() mismatch\nwant: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestUsePreRelease(t *testing.T) {
	// GIVEN a Lookup
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

			// WHEN usePreRelease is called
			got := lookup.usePreRelease()

			// THEN the expected value is returned
			gotStr := fmt.Sprint(got)
			wantStr := test.StringifyPtr(tc.want)
			if gotStr != wantStr {
				t.Errorf("usePreRelease mismatch\nwant: %q\ngot:  %q",
					wantStr, gotStr)
			}
		})
	}
}

func TestServiceURL(t *testing.T) {
	type args struct {
		ignoreWebURL bool
	}
	// GIVEN a Lookup
	tests := map[string]struct {
		url, webURL   string
		latestVersion string
		args          args
		want          string
	}{
		"not ignoreWebURL, have webURL, have latestVersion, template webURL": {
			url:           "release-argus/Argus",
			webURL:        "https://example.com/{{ version }}",
			latestVersion: "1.0.0",
			args:          args{ignoreWebURL: false},
			want:          "https://example.com/1.0.0",
		},
		"not ignoreWebURL, have webURL, have latestVersion, no template webURL": {
			url:           "release-argus/Argus",
			webURL:        "https://example.com/version",
			latestVersion: "1.0.0",
			args:          args{ignoreWebURL: false},
			want:          "https://github.com/release-argus/Argus",
		},
		"not ignoreWebURL, have webURL, no latestVersion, template webURL": {
			url:           "release-argus/Argus",
			webURL:        "https://example.com/{{ version }}",
			latestVersion: "",
			args:          args{ignoreWebURL: false},
			want:          "https://github.com/release-argus/Argus",
		},
		"not ignoreWebURL, no webURL, have latestVersion, template webURL": {
			url:           "release-argus/Argus",
			webURL:        "",
			latestVersion: "1.0.0",
			args:          args{ignoreWebURL: false},
			want:          "https://github.com/release-argus/Argus",
		},
		"ignoreWebURL, repo": {
			url:           "release-argus/Argus",
			webURL:        "https://example.com/{{ version }}",
			latestVersion: "1.0.0",
			args:          args{ignoreWebURL: true},
			want:          "https://github.com/release-argus/Argus",
		},
		"ignoreWebURL, url": {
			url:           "https://github.com/release-argus/Argus/releases",
			webURL:        "https://example.com/{{ version }}",
			latestVersion: "1.0.0",
			args:          args{ignoreWebURL: true},
			want:          "https://github.com/release-argus/Argus/releases",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.URL = tc.url
			lookup.Status.WebURL = test.StringPtr(tc.webURL)
			lookup.Status.SetLatestVersion(tc.latestVersion, "", false)

			// WHEN ServiceURL is called
			got := lookup.ServiceURL(tc.args.ignoreWebURL)

			// THEN the expected value is returned
			if got != tc.want {
				t.Errorf("ServiceURL(%t) mismatch\nwant: %q\ngot:  %q",
					tc.args.ignoreWebURL, tc.want, got)
			}
		})
	}
}
