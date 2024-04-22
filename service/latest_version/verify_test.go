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
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestLookupDefaults_CheckValues(t *testing.T) {
	// GIVEN a LookupDefault
	tests := map[string]struct {
		require  filter.RequireDefaults
		errRegex []string
	}{
		"valid": {
			require: *filter.NewRequireDefaults(
				filter.NewDockerCheckDefaults(
					"ghcr", "", "", "", "", nil)),
			errRegex: []string{},
		},
		"invalid require": {
			errRegex: []string{
				`^latest_version:$`,
				`^  require:$`,
				`^    docker:$`,
				`^      type: "[^"]+" <invalid>`},
			require: *filter.NewRequireDefaults(
				filter.NewDockerCheckDefaults(
					"someType", "", "", "", "", nil)),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defaults := LookupDefaults{
				Require: tc.require}

			// WHEN CheckValues is called
			err := defaults.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			if len(tc.errRegex) > len(lines) {
				t.Fatalf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Fatalf("%q didn't match %q\ngot:  %q",
						lines[i], tc.errRegex[i], e)
				}
			}
		})
	}
}

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		lType       *string
		url         *string
		wantURL     *string
		require     *filter.Require
		urlCommands *filter.URLCommandSlice
		errRegex    []string
	}{
		"valid": {
			errRegex: []string{},
		},
		"no type": {
			errRegex: []string{
				`^latest_version:$`,
				`^  type: <required>`},
			lType: test.StringPtr(""),
		},
		"invalid type": {
			errRegex: []string{
				`^latest_version:$`,
				`^  type: "[^"]+" <invalid>`},
			lType: test.StringPtr("foo"),
		},
		"no url": {
			errRegex: []string{
				`^latest_version:$`,
				`^  url: <required>`},
			url: test.StringPtr(""),
		},
		"corrects github url": {
			errRegex: []string{},
			url:      test.StringPtr("https://github.com/release-argus/Argus"),
			wantURL:  test.StringPtr("release-argus/Argus"),
		},
		"invalid require": {
			errRegex: []string{
				`^latest_version:$`,
				`^  require:$`,
				`^    regex_content: "[^"]+" <invalid>`},
			require: &filter.Require{RegexContent: "[0-"},
		},
		"invalid urlCommands": {
			errRegex: []string{
				`^latest_version:$`,
				`^  url_commands:$`,
				`^    item_0:$`,
				`^      type: "[^"]+" <invalid>`},
			urlCommands: &filter.URLCommandSlice{{Type: "foo"}},
		},
		"all errs": {
			errRegex: []string{
				`^latest_version:$`,
				`^  url: <required>`,
				`^  require:$`,
				`^    regex_content: "[^"]+" <invalid>`,
				`^  url_commands:$`,
				`^    item_0:$`,
				`^      type: "[^"]+" <invalid>`},
			url:         test.StringPtr(""),
			require:     &filter.Require{RegexContent: "[0-"},
			urlCommands: &filter.URLCommandSlice{{Type: "foo"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false, false)
			if tc.lType != nil {
				lookup.Type = *tc.lType
			}
			if tc.url != nil {
				lookup.URL = *tc.url
			}
			if tc.require != nil {
				lookup.Require = tc.require
			}
			if tc.urlCommands != nil {
				lookup.URLCommands = *tc.urlCommands
			}

			// WHEN CheckValues is called
			err := lookup.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			if len(tc.errRegex) > len(lines) {
				t.Fatalf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Fatalf("%q didn't match %q\ngot:  %q",
						lines[i], tc.errRegex[i], e)
				}
			}
		})
	}
}
