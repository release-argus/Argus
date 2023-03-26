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

package latestver

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/util"
)

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN a Lookup
	testLogging("WARN")
	tests := map[string]struct {
		url         string
		githubType  bool
		accessToken string
		errRegex    string
	}{
		"invalid url": {
			url:      "invalid://	test",
			errRegex: "invalid control character in URL"},
		"unknown url": {
			url:      "https://release-argus.invalid-tld",
			errRegex: "no such host"},
		"valid url": {
			url:      "https://release-argus.io",
			errRegex: "^$"},
		"github token": {
			url:         "release-argus/Argus",
			accessToken: "foo", githubType: true,
			errRegex: "^$"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lookup := testLookup(!tc.githubType, false)
			if tc.githubType && util.DefaultIfNil(lookup.AccessToken) == "" {
				lookup.AccessToken = &tc.accessToken
			}
			lookup.URL = tc.url

			// WHEN httpRequest is called on it
			_, err := lookup.httpRequest(&util.LogFrom{})

			// THEN any err is expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN a Lookup
	testLogging("WARN")
	tests := map[string]struct {
		githubService         bool
		noAccessToken         bool
		url                   string
		regex                 *string
		latestVersion         string
		deployedVersion       string
		nonSemanticVersioning bool
		allowInvalidCerts     bool
		wantLatestVersion     *string
		requireRegexContent   string
		requireRegexVersion   string
		requireCommand        []string
		requireDockerCheck    *filter.DockerCheck
		errRegex              string
	}{
		"invalid url": {
			url:      "invalid://	test",
			errRegex: "invalid control character in URL",
		},
		"query that gets a non-semantic version": {
			url:      "https://valid.release-argus.io/plain",
			regex:    stringPtr("v[0-9.]+"),
			errRegex: "failed converting .* to a semantic version",
		},
		"query on self-signed https works when allowed": {
			url:               "https://invalid.release-argus.io/plain",
			regex:             stringPtr("v[0-9.]+"),
			allowInvalidCerts: true,
			errRegex:          "failed converting .* to a semantic version",
		},
		"query on self-signed https fails when not allowed": {
			url:               "https://invalid.release-argus.io/plain",
			regex:             stringPtr("v[0-9.]+"),
			allowInvalidCerts: false,
			errRegex:          "x509",
		},
		"changed to semantic_versioning but had a non-semantic deployed_version": {
			deployedVersion: "1.2.3.4",
			errRegex:        `^$`,
		},
		"regex content mismatch": {
			requireRegexContent: "argus[0-9]+.exe",
			errRegex:            "regex .* not matched on content for version",
		},
		"regex content match": {
			requireRegexContent: "v{{ version }}",
			errRegex:            "^$",
		},
		"regex version mismatch": {
			requireRegexVersion: "^v[0-9]+$",
			errRegex:            "regex not matched on version",
		},
		"command fail": {
			requireCommand: []string{"false"},
			errRegex:       "exit status 1",
		},
		"command pass": {
			requireCommand: []string{"true"},
			errRegex:       "^$",
		},
		"docker tag mismatch": {
			requireDockerCheck: &filter.DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Tag:   "0.9.0-beta",
				Token: os.Getenv("GH_TOKEN")},
			errRegex: "manifest unknown",
		},
		"docker tag match": {
			requireDockerCheck: &filter.DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Tag:   "0.9.0",
				Token: os.Getenv("GH_TOKEN")},
			errRegex: "^$",
		},
		"regex version match": {
			requireRegexVersion: "v([0-9.]+)",
			errRegex:            "regex not matched on version",
		},
		"urlCommand regex mismatch": {
			regex:    stringPtr("^[0-9]+$"),
			errRegex: "regex .* didn't return any matches",
		},
		"valid semantic version query": {
			regex:    stringPtr("v([0-9.]+)"),
			errRegex: "^$",
		},
		"older version found": {
			regex:           stringPtr("([0-9.]+)"),
			latestVersion:   "0.0.0",
			deployedVersion: "9.9.9",
			errRegex:        "queried version .* is less than the deployed version",
		},
		"newer version found": {
			regex:           stringPtr("([0-9.]+)"),
			deployedVersion: "0.0.0",
			errRegex:        "^$",
		},
		"same version found": {
			regex:           stringPtr("([0-9.]+)"),
			deployedVersion: "1.2.1",
			errRegex:        "^$",
		},
		"no deployed version lookup": {
			regex:             stringPtr("([0-9.]+)-beta"),
			errRegex:          "^$",
			wantLatestVersion: stringPtr("1.2.2"),
		},
		"non-semantic version lookup": {
			regex:                 stringPtr("v[0-9.]+"),
			errRegex:              "^$",
			wantLatestVersion:     stringPtr("v1.2.2"),
			nonSemanticVersioning: true,
		},
		"github lookup": {
			githubService: true,
			errRegex:      "^$",
		},
		"github lookup with no access token": {
			githubService: true,
			noAccessToken: true,
			errRegex:      "^$",
		},
		"github lookup with failing urlCommand match": {
			githubService: true,
			regex:         stringPtr("x([0-9.]+)"),
			errRegex:      "no releases were found matching the url_commands",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			try := 0
			temporaryFailureInNameResolution := true
			for temporaryFailureInNameResolution != false {
				try++
				temporaryFailureInNameResolution = false
				lookup := testLookup(!tc.githubService, tc.allowInvalidCerts)
				// switch to LetsEncrypt cert
				lookup.URL = strings.Replace(lookup.URL, "://invalid", "://valid", 1)
				if tc.githubService && tc.noAccessToken {
					lookup.AccessToken = nil
				}
				lookup.Status.ServiceID = &name
				if tc.url != "" {
					lookup.URL = tc.url
				}
				if tc.regex != nil {
					if lookup.URLCommands == nil {
						lookup.URLCommands = filter.URLCommandSlice{{Type: "regex"}}
					}
					lookup.URLCommands[0].Regex = tc.regex
				}
				*lookup.Options.SemanticVersioning = !tc.nonSemanticVersioning
				lookup.Status.LatestVersion = tc.latestVersion
				lookup.Status.DeployedVersion = tc.deployedVersion
				lookup.Require.RegexContent = tc.requireRegexContent
				lookup.Require.RegexVersion = tc.requireRegexVersion
				lookup.Require.Command = tc.requireCommand
				lookup.Require.Docker = tc.requireDockerCheck

				// WHEN Query is called on it
				_, err := lookup.Query()

				// THEN any err is expected
				e := util.ErrorToString(err)
				re := regexp.MustCompile(tc.errRegex)
				match := re.MatchString(e)
				if !match {
					if strings.Contains(e, "context deadline exceeded") {
						temporaryFailureInNameResolution = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
				if tc.wantLatestVersion != nil && *tc.wantLatestVersion != lookup.Status.LatestVersion {
					t.Fatalf("wanted LatestVersion to become %q, not %q",
						*tc.wantLatestVersion, lookup.Status.LatestVersion)
				}
			}
		})
	}
}

func TestLookup_QueryGitHubETag(t *testing.T) {
	// GIVEN a Lookup
	testLogging("VERBOSE")
	tests := map[string]struct {
		attempts                   int
		eTagChanged                int
		eTagUnchangedUseCache      int
		initialRequireRegexVersion string
		errRegex                   string
	}{
		// Keeps .Releases incase filters change
		"three requests only uses 1 api limit": {
			attempts:              3,
			eTagChanged:           1,
			eTagUnchangedUseCache: 2,
			errRegex:              `^$`},
		"if initial request fails filter, cached results will be used": {
			attempts:                   3,
			eTagChanged:                1,
			eTagUnchangedUseCache:      2,
			initialRequireRegexVersion: `^FOO$`,
			errRegex:                   `regex not matched on version`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testLookup(false, false)
			lookup.GitHubData.ETag = "foo"
			lookup.Status.ServiceID = &name
			lookup.Require.RegexVersion = tc.initialRequireRegexVersion

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			attempt := 0
			// WHEN Query is called on it attempts number of times
			var errors string = ""
			for tc.attempts != attempt {
				attempt++
				if attempt == 2 {
					lookup.Require = &filter.Require{}
				}

				_, err := lookup.Query()
				if err != nil {
					errors += "--" + err.Error()
				}
			}

			// THEN any err is expected
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(errors)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, errors)
			}
			gotETagChanged := strings.Count(string(out), "ETag changed")
			if gotETagChanged != tc.eTagChanged {
				t.Errorf("ETag changed - got=%d, want=%d\n%s", gotETagChanged, tc.eTagChanged, out)
			}
			gotETagUnchangedUseCache := strings.Count(string(out), "Using cached releases")
			if gotETagUnchangedUseCache != tc.eTagUnchangedUseCache {
				t.Errorf("ETag unchanged use cache - got=%d, want=%d\n%s", gotETagUnchangedUseCache, tc.eTagUnchangedUseCache, out)
			}
		})
	}
}
