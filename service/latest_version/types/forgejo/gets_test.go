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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

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

func TestLookup_ServiceURL(t *testing.T) {
	// GIVEN: a host and a repository.
	tests := []struct {
		name string
		host string
		url  string
		want string
	}{
		{
			name: "scheme+host and repository",
			host: "https://codeberg.org",
			url:  "owner/repo",
			want: "https://codeberg.org/owner/repo",
		},
		{
			name: "host and repository",
			host: "codeberg.org",
			url:  "owner/repo",
			want: "https://codeberg.org/owner/repo",
		},
		{
			name: "scheme+host+sub-path and repository",
			host: "https://example.com/git",
			url:  "owner/repo",
			want: "https://example.com/git/owner/repo",
		},
		{
			name: "scheme+host+port and repository",
			host: "http://forge.example.com:3000",
			url:  "owner/repo",
			want: "http://forge.example.com:3000/owner/repo",
		},
		{
			name: "scheme+host+port+sub-path and repository",
			host: "http://forge.example.com:3000/sub/path",
			url:  "owner/repo",
			want: "http://forge.example.com:3000/sub/path/owner/repo",
		},
		{
			name: "unusable host falls back to the repository",
			host: "",
			url:  "owner/repo",
			want: "owner/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{Host: tc.host}
			lookup.URL = tc.url

			// WHEN: ServiceURL is called.
			got := lookup.ServiceURL()

			// THEN: it addresses the repository on the configured instance.
			if got != tc.want {
				t.Fatalf(
					"%s\nLookup{host: %q, url: %q}.ServiceURL() mismatch\ngot:  %q\nwant: %q",
					packageName, tc.host, tc.url,
					got, tc.want,
				)
			}
		})
	}
}

func TestLookup_UsePreRelease(t *testing.T) {
	// GIVEN: a service value, and the Defaults behind it.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		{
			name:             "root overrides all",
			rootValue:        new(false),
			defaultValue:     new(true),
			hardDefaultValue: new(true),
			want:             false,
		},
		{
			name:             "default overrides hardDefault",
			defaultValue:     new(true),
			hardDefaultValue: new(false),
			want:             true,
		},
		{
			name:             "hardDefault is last resort",
			hardDefaultValue: new(false),
			want:             false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{UsePreRelease: tc.rootValue}
			lookup.SetTypeDefaults(
				&Defaults{Common: CommonDefaults{UsePreRelease: tc.defaultValue}},
				&Defaults{Common: CommonDefaults{UsePreRelease: tc.hardDefaultValue}},
			)

			// WHEN: usePreRelease is called.
			got := lookup.usePreRelease()

			// THEN: the expected value is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nLookup.usePreRelease() mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLookup_UseTagsAPI(t *testing.T) {
	// GIVEN: a Lookup whose filters may or may not need what only a release may carry.
	tests := []struct {
		name          string
		usePreRelease *bool
		requireYAML   string
		want          bool
	}{
		{
			name: "usable/no prereleases wanted, and no content filter",
			want: true,
		},
		{
			name:          "unusable/prereleases wanted",
			usePreRelease: new(true),
			want:          false,
		},
		{
			name:        "unusable/regex_content set",
			requireYAML: "regex_content: 'Argus-{{ version }}.linux-amd64'",
			want:        false,
		},
		{
			name:          "unusable/both prereleases and regex_content",
			usePreRelease: new(true),
			requireYAML:   "regex_content: 'Argus'",
			want:          false,
		},
		{
			name:        "usable/a require without regex_content",
			requireYAML: `regex_version: '^[0-9.]+$'`,
			want:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookupYAML := test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo`,
			)
			if tc.usePreRelease != nil {
				lookupYAML += fmt.Sprintf("\nuse_prerelease: %t", *tc.usePreRelease)
			}
			if tc.requireYAML != "" {
				lookupYAML += "\nrequire:\n  " + tc.requireYAML
			}
			lookup := testLookup(t, lookupYAML)

			// WHEN: useTagsAPI is called.
			got := lookup.useTagsAPI()

			// THEN: the /tags fallback is allowed only when tags could satisfy the filters.
			if got != tc.want {
				t.Fatalf(
					"%s\nLookup.useTagsAPI() mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}
