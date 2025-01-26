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

package filter

import (
	"testing"

	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestRequire_RegexCheckVersion(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require  *Require
		errRegex string
	}{
		"nil require": {
			require:  nil,
			errRegex: `^$`},
		"empty regex_version": {
			require:  &Require{},
			errRegex: `^$`},
		"match": {
			require:  &Require{RegexVersion: "^[0-9.]+-beta$"},
			errRegex: `^$`},
		"no match": {
			require:  &Require{RegexVersion: "^[0-9.]+$"},
			errRegex: `regex "[^"]+" not matched on version`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.require != nil {
				tc.require.Status = &status.Status{}
			}

			// WHEN RegexCheckVersion is called on it
			err := tc.require.RegexCheckVersion("0.1.1-beta", logutil.LogFrom{})

			// THEN the err is what we expect
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestRequire_RegexCheckContent(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require  *Require
		body     string
		errRegex string
	}{
		"nil require": {
			require:  nil,
			errRegex: `^$`,
		},
		"empty regex_content": {
			require:  &Require{},
			errRegex: `^$`,
		},
		"match": {
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`},
			errRegex: `^$`,
			body:     `darwin amd64 - argus-1.2.3.darwin-amd64, linux amd64 - argus-1.2.3.linux-amd64, windows amd64 - argus-1.2.3.windows-amd64,`,
		},
		"no match": {
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`},
			errRegex: `regex .* not matched on content`,
			body:     `darwin amd64 - argus-1.2.3.darwin-amd64, linux arm64 - argus-1.2.3.linux-arm64, windows amd64 - argus-1.2.3.windows-amd64,`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.require != nil {
				tc.require.Status = &status.Status{}
			}

			// WHEN RegexCheckContent is called on it
			err := tc.require.RegexCheckContent("0.1.1-beta", tc.body, logutil.LogFrom{})

			// THEN the err is what we expect
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch%q\ngot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestRequire_RegexCheckContentGitHub(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require         *Require
		body            []github_types.Asset
		wantReleaseDate string
		errRegex        string
	}{
		"nil require": {
			require:  nil,
			errRegex: `^$`,
		},
		"empty regex_content": {
			require:  &Require{},
			errRegex: `^$`,
		},
		"github api body match": {
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`},
			errRegex: `^$`,
			body: []github_types.Asset{
				{Name: "argus-1.2.3.darwin-amd64", CreatedAt: "2020-01-01T00:00:00Z"},
				{Name: "argus-1.2.3.linux-amd64", CreatedAt: "2021-01-01T00:00:00Z"},
				{Name: "argus-1.2.3.windows-amd64", CreatedAt: "2022-01-01T00:00:00Z"}},
			wantReleaseDate: "2021-01-01T00:00:00Z",
		},
		"github api body no match": {
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`},
			errRegex: `regex .* not matched on content`,
			body: []github_types.Asset{
				{Name: "argus-1.2.3.darwin-amd64"},
				{Name: "argus-1.2.3.linux-arm64"},
				{Name: "argus-1.2.3.windows-amd64"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			if tc.require != nil {
				tc.require.Status = &status.Status{}
			}

			// WHEN RegexCheckContent is called on it
			releaseDate, err := tc.require.RegexCheckContentGitHub("0.1.1-beta", tc.body, logutil.LogFrom{})

			// THEN the err is what we expect
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}
			// AND the release date is what we expect
			if releaseDate != tc.wantReleaseDate {
				t.Errorf("Release date mismatch\nwant: %q\ngot:  %q",
					tc.wantReleaseDate, releaseDate)
			}
		})
	}
}
