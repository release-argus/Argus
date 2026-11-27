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
	"strings"
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
			errRegex: `^json: .* unmarshal JSON object`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := UnmarshalReleases([]byte(tc.body))
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a forgetypes.Release, b string) bool { return a.TagName == b },
				"releases",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestIsOwnerRepo(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "addressable/owner and repo",
			path: "owner/repo",
			want: true,
		},
		{
			name: "addressable/dots, dashes and underscores in both segments",
			path: "my-org.io/some_repo.v2",
			want: true,
		},
		{
			name: "unaddressable/no separator",
			path: "repo",
			want: false,
		},
		{
			name: "unaddressable/empty",
			path: "",
			want: false,
		},
		{
			name: "unaddressable/no owner",
			path: "/repo",
			want: false,
		},
		{
			name: "unaddressable/no repo",
			path: "owner/",
			want: false,
		},
		{
			name: "unaddressable/a third segment",
			path: "owner/repo/extra",
			want: false,
		},
		{
			name: "unaddressable/a full address",
			path: "https://codeberg.org/owner/repo",
			want: false,
		},
		{
			name: "unaddressable/both segments escape the prefix",
			path: "../..",
			want: false,
		},
		{
			name: "unaddressable/the owner escapes the prefix",
			path: "../admin",
			want: false,
		},
		{
			name: "unaddressable/the repo escapes the prefix",
			path: "owner/..",
			want: false,
		},
		{
			name: "unaddressable/a single dot as the repo",
			path: "owner/.",
			want: false,
		},
		{
			name: "unaddressable/a single dot as the owner",
			path: "./repo",
			want: false,
		},
		{
			name: "unaddressable/surrounded by whitespace",
			path: " owner/repo ",
			want: false,
		},
		{
			name: "unaddressable/a space inside a segment",
			path: "own er/repo",
			want: false,
		},
		{
			name: "unaddressable/a query that redirects the request",
			path: "owner/repo?ref=main",
			want: false,
		},
		{
			name: "unaddressable/a fragment",
			path: "owner/repo#main",
			want: false,
		},
		{
			name: "unaddressable/an encoded separator",
			path: "owner/re%2Fpo",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsOwnerRepo(tc.path)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestIsRepoSegment(t *testing.T) {
	tests := []struct {
		name    string
		segment string
		want    bool
	}{
		{
			name:    "names/letters",
			segment: "owner",
			want:    true,
		},
		{
			name:    "names/digits",
			segment: "2024",
			want:    true,
		},
		{
			name:    "names/dots, dashes and underscores",
			segment: "some_repo.v2-rc",
			want:    true,
		},
		{
			name:    "names/a leading dot followed by characters",
			segment: ".config",
			want:    true,
		},
		{
			name:    "escapes/parent directory",
			segment: "..",
			want:    false,
		},
		{
			name:    "escapes/current directory",
			segment: ".",
			want:    false,
		},
		{
			name:    "not a name/empty",
			segment: "",
			want:    false,
		},
		{
			name:    "not a name/a separator",
			segment: "owner/repo",
			want:    false,
		},
		{
			name:    "not a name/a space",
			segment: "own er",
			want:    false,
		},
		{
			name:    "not a name/a query separator",
			segment: "repo?ref",
			want:    false,
		},
		{
			name:    "not a name/a percent escape",
			segment: "re%2Fpo",
			want:    false,
		},
		{
			name:    "not a name/a colon",
			segment: "owner:repo",
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isRepoSegment(tc.segment)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestEscapedOwnerRepo(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "plain/owner and repo",
			path: "owner/repo",
			want: "owner/repo",
		},
		{
			name: "plain/dots, dashes and underscores pass through",
			path: "my-org.io/some_repo.v2",
			want: "my-org.io/some_repo.v2",
		},
		{
			name: "escaped/dot segments cannot climb out of the prefix",
			path: "../..",
			want: "%2E%2E/%2E%2E",
		},
		{
			name: "escaped/a single dot segment",
			path: "./repo",
			want: "%2E/repo",
		},
		{
			name: "plain/a leading dot is not a dot segment, so it is left alone",
			path: ".config/repo",
			want: ".config/repo",
		},
		{
			name: "escaped/a query cannot redirect the request",
			path: "owner/repo?ref=main",
			want: "owner/repo%3Fref=main",
		},
		{
			name: "escaped/a fragment cannot truncate the path",
			path: "owner/repo#main",
			want: "owner/repo%23main",
		},
		{
			name: "escaped/a space",
			path: "own er/repo",
			want: "own%20er/repo",
		},
		{
			name: "escaped/a third segment stays inside the repo segment",
			path: "owner/repo/extra",
			want: "owner/repo%2Fextra",
		},
		{
			name: "empty/no separator leaves the repo empty",
			path: "repo",
			want: "repo/",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := EscapedOwnerRepo(tc.path)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestBodyExcerpt(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "nothing/empty",
			body: "",
			want: "",
		},
		{
			name: "nothing/whitespace",
			body: "  \n\t ",
			want: "",
		},
		{
			name: "quoted/JSON message",
			body: `{"message":"token is required"}`,
			want: "\n" + `{"message":"token is required"}`,
		},
		{
			name: "quoted/newlines collapse to one line",
			body: "502 Bad Gateway\nnginx/1.24.0\n",
			want: "\n502 Bad Gateway nginx/1.24.0",
		},
		{
			name: "quoted/runs of whitespace collapse to one space",
			body: "proxy      error",
			want: "\nproxy error",
		},
		{
			name: "quoted/non-printables are dropped, so a body cannot forge a log line",
			body: "oops\x00\x07 \x1b[31mred\x1b[0m",
			want: "\noops [31mred[0m",
		},
		{
			name: "bounded/long body is truncated",
			body: strings.Repeat("a", maxErrorBody+50),
			want: "\n" + strings.Repeat("a", maxErrorBody) + "...",
		},
		{
			name: "bounded/truncation does not split a rune",
			body: strings.Repeat("a", maxErrorBody-1) + "\u00e9zz",
			want: "\n" + strings.Repeat("a", maxErrorBody-1) + "...",
		},
		{
			name: "bounded/a body far over the scan bound is still excerpted",
			body: strings.Repeat("a", maxErrorBody*4+1),
			want: "\n" + strings.Repeat("a", maxErrorBody) + "...",
		},
		{
			name: "bounded/content past the scan bound is not reached",
			body: strings.Repeat(" ", maxErrorBody*4) + "boom",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := BodyExcerpt([]byte(tc.body))

			if got != tc.want {
				t.Errorf("excerpt mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestFilterReleases(t *testing.T) {
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
			name: "release without a tag or name dropped",
			releases: []forgetypes.Release{
				{TagName: "0.99.0"},
				{TagName: " "},
				{},
			},
			want: []string{
				"0.99.0",
			},
		},
		{
			name: "url_commands giving an empty version dropped",
			releases: []forgetypes.Release{
				{TagName: "v"},
				{TagName: "v0.99.0"},
			},
			urlCommands: filter.URLCommands{
				{Type: "regex", Regex: `v([0-9.]*)`},
			},
			want: []string{
				"0.99.0",
			},
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
			name:               "semver pre-release tag is excluded even where its flag is unset",
			semanticVersioning: true,
			usePreReleases:     false,
			releases: []forgetypes.Release{
				{TagName: "v2.0.0-rc1"},
				{TagName: "v1.9.0"},
			},
			want: []string{
				"v1.9.0",
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

			got := FilterReleases(
				tc.releases,
				FilterOptions{
					URLCommands:        tc.urlCommands,
					SemanticVersioning: tc.semanticVersioning,
					UsePreReleases:     tc.usePreReleases,
				},
				logx.LogFrom{Primary: t.Name()},
			)

			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a forgetypes.Release, b string) bool { return a.TagName == b },
				"releases",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestFilterReleases__PreReleaseFromTag(t *testing.T) {
	tests := []struct {
		name           string
		releases       []forgetypes.Release
		urlCommands    filter.URLCommands
		usePreReleases bool
		noSemVer       bool
		want           []string
	}{
		{
			name: "a platform suffix survives url_commands",
			releases: []forgetypes.Release{
				{Name: "1.2.3-linux"},
				{Name: "1.2.2"},
			},
			urlCommands: filter.URLCommands{{Type: "regex", Regex: `^([0-9.]+)`}},
			want:        []string{"1.2.3", "1.2.2"},
		},
		{
			name: "a prefixed pre-release is caught once the prefix is stripped",
			releases: []forgetypes.Release{
				{Name: "release-1.2.3-rc1"},
				{Name: "release-1.2.2"},
			},
			urlCommands: filter.URLCommands{{Type: "regex", Regex: `release-(.+)`}},
			want:        []string{"1.2.2"},
		},
		{
			name: "a genuine pre-release with no url_commands is dropped",
			releases: []forgetypes.Release{
				{Name: "0.13.0-rc1"},
				{Name: "0.12.1"},
			},
			want: []string{"0.12.1"},
		},
		{
			name: "kept when pre-releases are wanted",
			releases: []forgetypes.Release{
				{Name: "0.13.0-rc1"},
				{Name: "0.12.1"},
			},
			usePreReleases: true,
			want:           []string{"0.13.0-rc1", "0.12.1"},
		},
		{
			name: "build metadata is not a pre-release",
			releases: []forgetypes.Release{
				{Name: "1.2.3+build5"},
			},
			want: []string{"1.2.3+build5"},
		},
		{
			name: "a tag that is not semver-shaped stays stable",
			releases: []forgetypes.Release{
				{Name: "1.2.3rc1"},
				{Name: "nightly"},
			},
			noSemVer: true,
			want:     []string{"1.2.3rc1", "nightly"},
		},
		{
			name: "a tag that parses as a semver pre-release stays stable without semantic versioning",
			releases: []forgetypes.Release{
				{Name: "2026-11"},
				{Name: "2026-09a"},
			},
			noSemVer: true,
			want:     []string{"2026-11", "2026-09a"},
		},
		{
			name: "TagName wins over Name",
			releases: []forgetypes.Release{
				{TagName: "3.0.0-rc1", Name: "3.0.0"},
			},
			want: []string{},
		},
		{
			name: "the tag names a pre-release even where the forge's flag is unset",
			releases: []forgetypes.Release{
				{TagName: "1.0.0-rc1", PreRelease: false},
				{TagName: "0.9.0"},
			},
			want: []string{"0.9.0"},
		},
		{
			name: "a flag the forge set to true is honoured",
			releases: []forgetypes.Release{
				{TagName: "1.0.0-rc1", PreRelease: true},
				{TagName: "0.9.0"},
			},
			want: []string{"0.9.0"},
		},
		{
			name:     "no releases",
			releases: []forgetypes.Release{},
			want:     []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterReleases(
				tc.releases,
				FilterOptions{
					URLCommands:        tc.urlCommands,
					SemanticVersioning: !tc.noSemVer,
					UsePreReleases:     tc.usePreReleases,
				},
				logx.LogFrom{Primary: t.Name()},
			)

			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a forgetypes.Release, b string) bool { return a.TagName == b },
				"releases",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestReleaseSortsBefore(t *testing.T) {
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

			got := releaseSortsBefore(a, b)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestNextPage(t *testing.T) {
	tests := []struct {
		name string
		link string
		want int
	}{
		{
			name: "next and last",
			link: `<https://api.example.com/repos/o/r/releases?limit=50&page=2>; rel="next",` +
				`<https://api.example.com/repos/o/r/releases?limit=50&page=4>; rel="last"`,
			want: 2,
		},
		{
			name: "prev, next, last and first",
			link: `<https://api.example.com/repos/o/r/releases?page=2>; rel="prev",` +
				`<https://api.example.com/repos/o/r/releases?page=4>; rel="next",` +
				`<https://api.example.com/repos/o/r/releases?page=9>; rel="last",` +
				`<https://api.example.com/repos/o/r/releases?page=1>; rel="first"`,
			want: 4,
		},
		{
			name: "page before per_page",
			link: `<https://api.example.com/repos/o/r/releases?page=1&per_page=100>; rel="prev",` +
				`<https://api.example.com/repos/o/r/releases?page=3&per_page=100>; rel="next"`,
			want: 3,
		},
		{
			name: "last page, no next",
			link: `<https://api.example.com/repos/o/r/releases?page=1>; rel="first",` +
				`<https://api.example.com/repos/o/r/releases?page=3>; rel="prev"`,
			want: 0,
		},
		{
			name: "no header",
			link: "",
			want: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NextPage(tc.link)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %d\nwant: %d", got, tc.want)
			}
		})
	}
}
