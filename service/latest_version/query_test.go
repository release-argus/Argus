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
		url                 string
		regex               *string
		latestVersion       string
		wantlatestVersion   *string
		requireRegexContent *string
		requireRegexVersion *string
		errRegex            string
	}{
		"invalid url": {url: "invalid://	test", errRegex: "invalid control character in URL"},
		"query that gets a non-semantic version": {url: "https://release-argus.io/docs/config/service/",
			errRegex: "failed converting .* to a semantic version"},
		"regex content mismatch": {requireRegexContent: stringPtr("argus[0-9]+.exe"),
			errRegex: "regex .* not matched on content for version"},
		"regex version mismatch": {requireRegexVersion: stringPtr("^[0-9]+$"),
			errRegex: "regex not matched on version"},
		"valid semantic version query": {regex: stringPtr("([0-9.]+)test"), errRegex: "^$"},
		"older version found": {regex: stringPtr("([0-9.]+)test"), latestVersion: "9.9.9",
			errRegex: "queried version .* is less than the deployed version"},
		"newer version found": {regex: stringPtr("([0-9.]+)test"), latestVersion: "0.0.0",
			errRegex: "^$"},
		"same version found": {regex: stringPtr("([0-9.]+)test"), latestVersion: "1.2.3",
			errRegex: "^$"},
		"no deployed version lookup": {errRegex: "^%", wantLatestVersion: stringPtr("1.2.3")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testLookupURL()
			if tc.url != "" {
				lookup.URL = tc.url
			}
			if tc.regex != nil {
				lookup.URLCommands[0].Regex = tc.regex
			}
			lookup.Status.LatestVersion = tc.latestVersion
			lookup.Require.RegexContent = tc.requireRegexContent
			lookup.Require.RegexVersion = tc.requireRegexVersion

			// WHEN httpRequest is called on it
			_, err := lookup.Query()

			// THEN any err is expected
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("%s:\nwant match for %q\nnot: %q",
					name, tc.errRegex, e)
			}
			if tc.wantLatestVersion != nil && *tc.wantLatestVersion != lookup.Status.LatestVersion {
				t.Fatalf("%s:\nwanted LatestVersion to become %q, not %q",
					name, tc.wantLatestVersion, lookup.Status.LatestVersion)
			}
		})
	}
}
