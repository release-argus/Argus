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
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN: a Lookup.
	ghRepoParts := strings.Split(test.ArgusGitHubRepo, "/")
	t.Setenv("ARGUS_TEST_LV_GITHUB_OWNER", ghRepoParts[0])
	t.Setenv("ARGUS_TEST_LV_GITHUB_REPO", ghRepoParts[1])
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
				url: new(""),
			},
		},
		{
			name:     "corrects github url",
			errRegex: `^$`,
			wantURL:  &test.ArgusGitHubRepo,
			args: args{
				url: new("https://github.com/" + test.ArgusGitHubRepo),
			},
		},
		{
			name:     "invalid/url escapes the repos prefix with dot segments",
			errRegex: `^url: "\.\./\.\." <invalid> \(e\.g\. release-argus/Argus\)$`,
			wantURL:  new("../.."),
			args: args{
				url: new("../.."),
			},
		},
		{
			name:     "invalid/url owner is a dot segment",
			errRegex: `^url: "\.\./admin" <invalid> \(e\.g\. release-argus/Argus\)$`,
			wantURL:  new("../admin"),
			args: args{
				url: new("../admin"),
			},
		},
		{
			name:     "invalid/url carries a query, which would redirect the request",
			errRegex: `^url: "owner/repo\?ref=main" <invalid> \(e\.g\. release-argus/Argus\)$`,
			wantURL:  new("owner/repo?ref=main"),
			args: args{
				url: new("owner/repo?ref=main"),
			},
		},
		{
			name:     "invalid/a full url with a trailing slash leaves no repo",
			errRegex: `^url: "Argus/" <invalid> \(e\.g\. release-argus/Argus\)$`,
			wantURL:  new("Argus/"),
			args: args{
				url: new("https://github.com/" + test.ArgusGitHubRepo + "/"),
			},
		},
		{
			name:     "valid/url with dots, dashes and underscores",
			errRegex: `^$`,
			wantURL:  new("my-org.io/some_repo.v2"),
			args: args{
				url: new("my-org.io/some_repo.v2"),
			},
		},
		{
			name:     "valid/url from env vars",
			errRegex: `^$`,
			wantURL:  new("${ARGUS_TEST_LV_GITHUB_OWNER}/${ARGUS_TEST_LV_GITHUB_REPO}"),
			args: args{
				url: new("${ARGUS_TEST_LV_GITHUB_OWNER}/${ARGUS_TEST_LV_GITHUB_REPO}"),
			},
		},
		{
			name:     "invalid/url from an env var that expands to a bad repo",
			errRegex: `^url: "\${ARGUS_TEST_LV_GITHUB_OWNER}/a b" <invalid> \(e\.g\. release-argus/Argus\)$`,
			wantURL:  new("${ARGUS_TEST_LV_GITHUB_OWNER}/a b"),
			args: args{
				url: new("${ARGUS_TEST_LV_GITHUB_OWNER}/a b"),
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
				url: new(""),
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
