// Copyright [2022] [Argus]
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

package latest_version

import (
	"os"
	"regexp"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/utils"
)

func TestHTTPRequest(t *testing.T) {
	// GIVEN a Lookup
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		url      string
		errRegex string
	}{
		"invalid url": {url: "invalid://	test", errRegex: "invalid control character in URL"},
		"unknown url": {url: "https://release-argus.invalid-tld", errRegex: "no such host"},
		"valid url":   {url: "https://release-argus.io", errRegex: "^$"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testLookupURL()
			lookup.URL = tc.url

			// WHEN httpRequest is called on it
			_, err := lookup.httpRequest(utils.LogFrom{})

			// THEN any err is expected
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("%s:\nwant match for %q\nnot: %q",
					name, tc.errRegex, e)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	// GIVEN a Lookup
	jLog = utils.NewJLog("WARN", false)
	var logURLCommand *filters.URLCommandSlice
	logURLCommand.Init(jLog)
	tests := map[string]struct {
		githubService         bool
		noAccessToken         bool
		url                   string
		regex                 *string
		latestVersion         string
		nonSemanticVersioning bool
		allowInvalidCerts     bool
		wantLatestVersion     *string
		requireRegexContent   string
		requireRegexVersion   string
		errRegex              string
	}{
		"invalid url": {url: "invalid://	test", errRegex: "invalid control character in URL"},
		"query that gets a non-semantic version": {url: "https://valid.release-argus.io/plain", regex: stringPtr("v[0-9.]+"),
			errRegex: "failed converting .* to a semantic version"},
		"query on self-signed https works when allowed":     {url: "https://invalid.release-argus.io/plain", regex: stringPtr("v[0-9.]+"), errRegex: "failed converting .* to a semantic version", allowInvalidCerts: true},
		"query on self-signed https fails when not allowed": {url: "https://invalid.release-argus.io/plain", regex: stringPtr("v[0-9.]+"), errRegex: "x509", allowInvalidCerts: false},
		"changed to semantic_versioning but had a non-semantic latest_version": {latestVersion: "1.2.3.4",
			errRegex: "failed converting .* to a semantic version .* old version"}, "regex content mismatch": {requireRegexContent: "argus[0-9]+.exe", errRegex: "regex .* not matched on content for version"},
		"regex content match":                         {requireRegexContent: "v{{ version }}", errRegex: "^$"},
		"regex version mismatch":                      {requireRegexVersion: "^v[0-9]+$", errRegex: "regex not matched on version"},
		"regex version match":                         {requireRegexVersion: "v([0-9.]+)", errRegex: "regex not matched on version"},
		"urlCommand regex mismatch":                   {regex: stringPtr("^[0-9]+$"), errRegex: "regex .* didn't return any matches"},
		"valid semantic version query":                {regex: stringPtr("v([0-9.]+)"), errRegex: "^$"},
		"older version found":                         {regex: stringPtr("([0-9.]+)"), latestVersion: "9.9.9", errRegex: "queried version .* is less than the deployed version"},
		"newer version found":                         {regex: stringPtr("([0-9.]+)"), latestVersion: "0.0.0", errRegex: "^$"},
		"same version found":                          {regex: stringPtr("([0-9.]+)"), latestVersion: "1.2.1", errRegex: "^$"},
		"no deployed version lookup":                  {regex: stringPtr("([0-9.]+)-beta"), errRegex: "^$", wantLatestVersion: stringPtr("1.2.2")},
		"non-semantic version lookup":                 {regex: stringPtr("v[0-9.]+"), errRegex: "^$", wantLatestVersion: stringPtr("v1.2.2"), nonSemanticVersioning: true},
		"github lookup":                               {githubService: true, errRegex: "^$"},
		"github lookup with no access token":          {githubService: true, noAccessToken: true, errRegex: "^$"},
		"github lookup with failing urlCommand match": {githubService: true, regex: stringPtr("x([0-9.]+)"), errRegex: "no releases were found matching the url_commands"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var lookup Lookup
			if tc.githubService {
				lookup = testLookupGitHub()
				if !tc.noAccessToken {
					accessToken := os.Getenv("GITHUB_TOKEN")
					lookup.AccessToken = &accessToken
				}
			} else {
				lookup = testLookupURL()
			}
			lookup.AllowInvalidCerts = &tc.allowInvalidCerts
			lookup.status.ServiceID = &name
			if tc.url != "" {
				lookup.URL = tc.url
			}
			if tc.regex != nil {
				if lookup.URLCommands == nil {
					lookup.URLCommands = filters.URLCommandSlice{{Type: "regex"}}
				}
				lookup.URLCommands[0].Regex = tc.regex
			}
			*lookup.options.SemanticVersioning = !tc.nonSemanticVersioning
			lookup.status.LatestVersion = tc.latestVersion
			lookup.Require.RegexContent = tc.requireRegexContent
			lookup.Require.RegexVersion = tc.requireRegexVersion

			// WHEN Query is called on it
			_, err := lookup.Query()

			// THEN any err is expected
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("%s:\nwant match for %q\nnot: %q",
					name, tc.errRegex, e)
			}
			if tc.wantLatestVersion != nil && *tc.wantLatestVersion != lookup.status.LatestVersion {
				t.Fatalf("%s:\nwanted LatestVersion to become %q, not %q",
					name, *tc.wantLatestVersion, lookup.status.LatestVersion)
			}
		})
	}
}
