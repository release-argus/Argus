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

package web

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name     string
		input    *Lookup
		errRegex string
	}{
		{
			name:  "Valid Lookup",
			input: testLookup(t, false),
		},
		{
			name:     "Empty",
			input:    &Lookup{},
			errRegex: `^url: <required>[^\n]+$`,
		},
		{
			name: "Invalid Require",
			input: &Lookup{
				Lookup: base.Lookup{
					Require: &filter.Require{
						RegexVersion: "[0a",
					},
				},
			},
			errRegex: test.TrimYAML(`
				^url: <required>.*
				require:
					regex_version: "[^"]+" <invalid>.*$`,
			),
		},
		{
			name: "Invalid URL Commands",
			input: &Lookup{
				Lookup: base.Lookup{
					URLCommands: filter.URLCommands{
						{Type: "regex", Regex: `[0-9]+`},
						{Type: "regex", Regex: `[0-9]+`},
						{Type: "foo"},
					},
				},
			},
			errRegex: test.TrimYAML(`
				^url: <required>.*
				url_commands:
					- item_2:
						type: "foo" <invalid>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}
