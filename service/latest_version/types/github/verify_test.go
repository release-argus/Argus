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

// Package github provides a github-based lookup type.
package github

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN: a Lookup.
	type args struct {
		url         *string
		require     *filter.Require
		urlCommands *filter.URLCommands
	}
	tests := []struct {
		name     string
		errRegex string
		wantURL  *string
		args     args
	}{
		{
			name:     "valid",
			errRegex: "^$",
		},
		{
			name:     "no url",
			errRegex: `^url: <required>.*$`,
			args: args{
				url: test.Ptr(""),
			},
		},
		{
			name:     "corrects github url",
			errRegex: `^$`,
			wantURL:  &test.ArgusGitHubRepo,
			args: args{
				url: test.Ptr("https://github.com/" + test.ArgusGitHubRepo),
			},
		},
		{
			name: "invalid require",
			errRegex: test.TrimYAML(`
				^require:
					regex_content: "[^"]+" <invalid>.*$`,
			),
			args: args{
				require: &filter.Require{
					RegexContent: "[0-",
				},
			},
		},
		{
			name: "invalid urlCommands",
			errRegex: test.TrimYAML(`
				^url_commands:
					- item_0:
						type: "[^"]+" <invalid>.*$`,
			),
			args: args{
				urlCommands: &filter.URLCommands{
					{Type: "foo"},
				},
			},
		},
		{
			name: "all decode",
			errRegex: test.TrimYAML(`
				^url: <required>.*
				url_commands:
					- item_0:
						type: "[^"]+" <invalid>.*
				require:
					regex_content: "[^"]+" <invalid>.*$`,
			),
			args: args{
				url: test.Ptr(""),
				require: &filter.Require{
					RegexContent: "[0-",
				},
				urlCommands: &filter.URLCommands{
					{Type: "foo"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testLookup(t, false)
			if tc.args.url != nil {
				input.URL = *tc.args.url
			}
			if tc.args.require != nil {
				input.Require = tc.args.require
			}
			if tc.args.urlCommands != nil {
				input.URLCommands = *tc.args.urlCommands
			}

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.CheckValues,
			)
		})
	}
}
