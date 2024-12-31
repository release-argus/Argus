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
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN a Lookup
	type args struct {
		url         *string
		require     *filter.Require
		urlCommands *filter.URLCommandSlice
	}
	tests := map[string]struct {
		errRegex string
		wantURL  *string
		args     args
	}{
		"valid": {
			errRegex: "^$",
		},
		"no url": {
			errRegex: `^url: <required>.*$`,
			args: args{
				url: test.StringPtr("")},
		},
		"corrects github url": {
			errRegex: `^$`,
			wantURL:  test.StringPtr("release-argus/Argus"),
			args: args{
				url: test.StringPtr("https://github.com/release-argus/Argus")},
		},
		"invalid require": {
			errRegex: test.TrimYAML(`
				^require:
					regex_content: "[^"]+" <invalid>.*$`),
			args: args{
				require: &filter.Require{RegexContent: "[0-"}},
		},
		"invalid urlCommands": {
			errRegex: test.TrimYAML(`
				^url_commands:
					- item_0:
						type: "[^"]+" <invalid>.*$`),
			args: args{
				urlCommands: &filter.URLCommandSlice{{Type: "foo"}}},
		},
		"all errs": {
			errRegex: test.TrimYAML(`
				^url: <required>.*
				url_commands:
					- item_0:
						type: "[^"]+" <invalid>.*
				require:
					regex_content: "[^"]+" <invalid>.*$`),
			args: args{
				url:         test.StringPtr(""),
				require:     &filter.Require{RegexContent: "[0-"},
				urlCommands: &filter.URLCommandSlice{{Type: "foo"}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			if tc.args.url != nil {
				lookup.URL = *tc.args.url
			}
			if tc.args.require != nil {
				lookup.Require = tc.args.require
			}
			if tc.args.urlCommands != nil {
				lookup.URLCommands = *tc.args.urlCommands
			}

			// WHEN CheckValues is called
			err := lookup.CheckValues("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("github.Lookup.CheckValues() want %d lines of error:\n%q\ngot %d lines:\n%v\nstdout: %q",
					wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("github.Lookup.CheckValues() error mismatch\nwant match for:\n%q\ngot:\n%q",
					tc.errRegex, e)
				return
			}
		})
	}
}
