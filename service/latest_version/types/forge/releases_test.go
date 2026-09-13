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

// Package forge provides the release handling shared by git forge lookup types.
package forge

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestUnmarshalReleases(t *testing.T) {
	// GIVEN: a response body.
	tests := []struct {
		name     string
		body     string
		want     []string
		errRegex string
	}{
		{
			name:     "valid/list of releases",
			body:     string(testBody),
			want:     []string{"0.18.0", "0.17.4"},
			errRegex: `^$`,
		},
		{
			name:     "valid/empty list",
			body:     `[]`,
			want:     []string{},
			errRegex: `^$`,
		},
		{
			name:     "valid/empty list with trailing newline",
			body:     "[]\n",
			want:     []string{},
			errRegex: `^$`,
		},
		{
			name:     "valid/null decodes to no releases",
			body:     `null`,
			want:     nil,
			errRegex: `^$`,
		},
		{
			name: "valid/unknown fields ignored",
			body: test.TrimJSON(`[
				{
					"author": {
						"login":"someone"
					},
					"tag_name":"1.2.3",
					"draft":true
				}
			]`),
			want:     []string{"1.2.3"},
			errRegex: `^$`,
		},
		{
			name:     "invalid/not JSON",
			body:     `not json`,
			errRegex: `invalid character`,
		},
		{
			name:     "invalid/object rather than a list",
			body:     `{"tag_name":"1.2.3"}`,
			errRegex: `cannot unmarshal`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: UnmarshalReleases is called on it.
			got, err := UnmarshalReleases([]byte(tc.body))

			prefix := fmt.Sprintf(
				"%s\nUnmarshalReleases(%q)",
				packageName, tc.body,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex)
			}

			// AND: the releases are as expected.
			if err := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a forgetypes.Release, b string) bool { return a.TagName == b },
				prefix,
				"",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFilterReleases(t *testing.T) {
	// GIVEN: a set of releases and the settings to filter them with.
	tests := []struct {
		name                               string
		releases                           []forgetypes.Release
		semanticVersioning, usePreReleases bool
		urlCommands                        filter.URLCommands
		want                               []string
	}{
		{
			name:     "no releases",
			releases: []forgetypes.Release{},
			want:     []string{},
		},
		{
			name: "Name used when TagName is empty (the /tags endpoint)",
			releases: []forgetypes.Release{
				{Name: "0.99.0"},
				{Name: "0.3.0"},
			},
			want: []string{
				"0.99.0",
				"0.3.0",
			},
		},
		{
			name:           "pre-releases excluded by default",
			usePreReleases: false,
			releases: []forgetypes.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.0.1"},
			},
			want: []string{
				"0.99.0",
				"0.0.1",
			},
		},
		{
			name:           "pre-releases kept when allowed",
			usePreReleases: true,
			releases: []forgetypes.Release{
				{TagName: "0.99.0"},
				{TagName: "0.3.0", PreRelease: true},
				{TagName: "0.0.1"},
			},
			want: []string{
				"0.99.0",
				"0.3.0",
				"0.0.1",
			},
		},
		{
			name:               "non-semver dropped when semantic versioning wanted",
			semanticVersioning: true,
			usePreReleases:     true,
			releases: []forgetypes.Release{
				{TagName: "0.99.0"},
				{TagName: "version 0.2.0"},
				{TagName: "0.0.1"},
			},
			want: []string{
				"0.99.0",
				"0.0.1",
			},
		},
		{
			name:               "non-semver kept when semantic versioning not required",
			semanticVersioning: false,
			usePreReleases:     true,
			releases: []forgetypes.Release{
				{TagName: "0.99.0"},
				{TagName: "version 0.2.0"},
				{TagName: "0.0.1"},
			},
			want: []string{
				"0.99.0",
				"version 0.2.0",
				"0.0.1",
			},
		},
		{
			name:               "semver sort newest to oldest",
			semanticVersioning: true,
			usePreReleases:     true,
			releases: []forgetypes.Release{
				{TagName: "0.0.0"},
				{TagName: "0.3.0"},
				{TagName: "0.10.0"},
				{TagName: "0.0.2"},
				{TagName: "1.0.0"},
				{TagName: "0.0.1"},
			},
			want: []string{
				"1.0.0",
				"0.10.0",
				"0.3.0",
				"0.0.2",
				"0.0.1",
				"0.0.0",
			},
		},
		{
			name:               "source order kept when non-semver",
			semanticVersioning: false,
			usePreReleases:     true,
			releases: []forgetypes.Release{
				{TagName: "0.0.0"},
				{TagName: "0.3.0"},
				{TagName: "0.0.2"},
			},
			want: []string{
				"0.0.0",
				"0.3.0",
				"0.0.2",
			},
		},
		{
			name:               "leading v parses as semver, and the tag keeps the prefix",
			semanticVersioning: true,
			usePreReleases:     true,
			releases: []forgetypes.Release{
				{TagName: "v0.9.0"},
				{TagName: "1.0.0"},
			},
			want: []string{
				"1.0.0",
				"v0.9.0",
			},
		},
		{
			name:               "duplicate versions are not deduplicated",
			semanticVersioning: false,
			usePreReleases:     true,
			releases: []forgetypes.Release{
				{TagName: "1.0.0"},
				{TagName: "1.0.0"},
			},
			want: []string{
				"1.0.0", "1.0.0",
			},
		},
		{
			name:               "url_commands applied, and drop what they fail on",
			semanticVersioning: true,
			usePreReleases:     true,
			releases: []forgetypes.Release{
				{TagName: "0.0.2-0.0.2"},
				{TagName: "0.0.1-0.0.1"},
				{TagName: "0.0.0"},
			},
			urlCommands: filter.URLCommands{
				{Type: "regex", Regex: `-([0-9.]+)$`},
			},
			want: []string{
				"0.0.2", "0.0.1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: FilterReleases is called on them.
			got := FilterReleases(
				tc.releases,
				tc.urlCommands,
				tc.semanticVersioning,
				tc.usePreReleases,
				logx.LogFrom{Primary: t.Name()},
			)

			prefix := fmt.Sprintf(
				"%s\nFilterReleases(%+v)",
				packageName, tc.releases,
			)

			// THEN: only the expected releases are kept, in the expected order.
			if err := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a forgetypes.Release, b string) bool { return a.TagName == b },
				prefix,
				"",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReleaseSortsBefore(t *testing.T) {
	// GIVEN: two releases to order.
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{
			name: "higher version sorts first",
			a:    "2.0.0",
			b:    "1.9.9",
			want: true,
		},
		{
			name: "lower version does not sort first",
			a:    "1.0.0",
			b:    "1.1.0",
			want: false,
		},
		{
			name: "equal versions do not sort before each other",
			a:    "1.2.3",
			b:    "1.2.3",
			want: false,
		},
		{
			name: "release sorts ahead of its pre-release",
			a:    "1.2.3",
			b:    "1.2.3-alpha",
			want: true,
		},
		{
			name: "build metadata is not an ordering",
			a:    "1.2.3+build2",
			b:    "1.2.3+build1",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := forgetypes.Release{SemanticVersion: semver.MustParse(tc.a)}
			b := forgetypes.Release{SemanticVersion: semver.MustParse(tc.b)}

			// WHEN: they are compared.
			got := releaseSortsBefore(a, b)

			// THEN: the newer of the two sorts first.
			if got != tc.want {
				t.Errorf(
					"%s\nreleaseSortsBefore(a=%q, b=%q) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.a, tc.b,
					got, tc.want,
				)
			}
		})
	}
}
