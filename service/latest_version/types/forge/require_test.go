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
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestReleaseMeetsRequirements(t *testing.T) {
	type wants struct {
		version     string
		releaseDate string
		errRegex    string
	}

	defaultRelease := testBodyObject[0]

	// GIVEN: a release and the requirements it is checked against.
	tests := []struct {
		name            string
		requireYAML     string
		releaseOverride *forgetypes.Release
		want            wants
	}{
		{
			name: "no requirements/first release is returned",
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.PublishedAt,
				errRegex:    `^$`,
			},
		},
		{
			name: "no requirements/semantic version preferred over the tag",
			releaseOverride: &forgetypes.Release{
				TagName:         "v1.0.0",
				SemanticVersion: semver.MustParse("v1.0.0"),
				PublishedAt:     "2021-01-01T00:00:00Z",
			},
			want: wants{
				version:     "1.0.0",
				releaseDate: "2021-01-01T00:00:00Z",
				errRegex:    `^$`,
			},
		},
		{
			name: "no requirements/release date not RFC3339 is dropped",
			releaseOverride: &forgetypes.Release{
				TagName:     "v1.0.0",
				PublishedAt: "yesterday",
			},
			want: wants{
				version:     "v1.0.0",
				releaseDate: "",
				errRegex:    `^$`,
			},
		},
		{
			name: "no requirements/absent release date",
			releaseOverride: &forgetypes.Release{
				TagName: "v1.0.0",
			},
			want: wants{
				version:     "v1.0.0",
				releaseDate: "",
				errRegex:    `^$`,
			},
		},
		{
			name:        "require.regex_version/match",
			requireYAML: `regex_version: "[0-9.]+"`,
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.PublishedAt,
				errRegex:    `^$`,
			},
		},
		{
			name:        "require.regex_version/no match",
			requireYAML: `regex_version: "x[0-9.]+"`,
			want: wants{
				errRegex: `^regex "[^"]+" not matched on version "[^"]+"$`,
			},
		},
		{
			name:        "require.regex_content/match takes the asset's date",
			requireYAML: `regex_content: "(?i)argus.*arm64"`,
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[1].CreatedAt,
				errRegex:    `^$`,
			},
		},
		{
			name:        "require.regex_content/no match",
			requireYAML: `regex_content: "aArgus"`,
			want: wants{
				errRegex: `^regex "[^"]+" not matched on content for version "[^"]+"$`,
			},
		},
		{
			name:        "require.regex_content/match, but the asset's date is not RFC3339",
			requireYAML: `regex_content: "(?i)argus.*amd64"`,
			releaseOverride: &forgetypes.Release{
				TagName:     "0.18.0",
				PublishedAt: "2020-01-02T03:04:05Z",
				Assets: []forgetypes.Asset{
					{Name: "Argus-0.18.0.linux-amd64", CreatedAt: "tomorrow"},
				},
			},
			want: wants{
				version:     "0.18.0",
				releaseDate: "2020-01-02T03:04:05Z",
				errRegex:    `^$`,
			},
		},
		{
			name:        "require.command/pass",
			requireYAML: `command: ["true"]`,
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.PublishedAt,
				errRegex:    `^$`,
			},
		},
		{
			name:        "require.command/fail",
			requireYAML: `command: ["false"]`,
			want: wants{
				errRegex: test.TrimYAML(`
					^command failed:
						exit status 1$`,
				),
			},
		},
		{
			name: "all requirements/pass together",
			requireYAML: test.TrimYAML(`
				regex_version: "[0-9.]+"
				regex_content: "(?i)argus.*arm64"
				command: ["true"]
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[1].CreatedAt,
				errRegex:    `^$`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			release := util.FirstNonNilPtr(
				tc.releaseOverride,
				&defaultRelease,
			)
			require := testRequire(t, tc.requireYAML)
			logFrom := logx.LogFrom{Primary: t.Name()}

			// WHEN: ReleaseMeetsRequirements is called on it.
			version, releaseDate, err := ReleaseMeetsRequirements(
				*release,
				require,
				"forge-test",
				logFrom,
			)

			prefix := fmt.Sprintf(
				"%s\nReleaseMeetsRequirements(%+v)",
				packageName, release,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.want.errRegex,
				)
			}

			// AND: the version is as expected.
			if version != tc.want.version {
				t.Errorf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, version, tc.want.version,
				)
			}

			// AND: the release date is as expected.
			if releaseDate != tc.want.releaseDate {
				t.Errorf(
					"%s release date mismatch\ngot:  %q\nwant: %q",
					prefix, releaseDate, tc.want.releaseDate,
				)
			}
		})
	}
}
