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

package web

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestCheckValues(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		lookup   *Lookup
		errRegex string
	}{
		"Valid Lookup": {
			lookup: testLookup(false),
		},
		"Empty URL": {
			lookup:   &Lookup{},
			errRegex: `^url: <required>[^\n]+$`,
		},
		"Invalid Require": {
			lookup: &Lookup{
				Lookup: base.Lookup{
					Require: &filter.Require{
						RegexVersion: "[0a",
					}},
			},
			errRegex: test.TrimYAML(`
				^url: <required>.*
				require:
					regex_version: "[^"]+" <invalid>.*$`),
		},
		"Invalid URL Commands": {
			lookup: &Lookup{
				Lookup: base.Lookup{
					URLCommands: filter.URLCommandSlice{
						filter.URLCommand{
							Type: "regex", Regex: `[0-9]+`},
						filter.URLCommand{
							Type: "regex", Regex: `[0-9]+`},
						filter.URLCommand{Type: "foo"}}}},
			errRegex: test.TrimYAML(`
				^url: <required>.*
				url_commands:
					- item_2:
						type: "foo" <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called on it.
			err := tc.lookup.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName,
					wantLines, tc.errRegex,
					len(lines), lines,
					e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}
