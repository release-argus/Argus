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
			url: "https://release-argus.io"},
		"github token": {
			url:         "release-argus/Argus",
			accessToken: "foo", githubType: true},
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
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
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
		outputRegex           string
		errRegex              string
	}{
		"invalid url": {
			url:      "invalid://	test",
			errRegex: "invalid control character in URL",
		},
		"query that gets a non-semantic version": {
			url:      "https://valid.release-argus.io/plain",
			regex:    stringPtr(`"v[0-9.]+`),
			errRegex: "failed converting .* to a semantic version",
		},
		"query on self-signed https works when allowed": {
			url:               "https://invalid.release-argus.io/plain",
			regex:             stringPtr("v[0-9.]+"),
			allowInvalidCerts: true,
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
		},
		"command fail": {
			requireCommand: []string{"false"},
			errRegex:       "exit status 1",
		},
		"command pass": {
			requireCommand: []string{"true"},
		},
		"docker tag mismatch": {
			requireDockerCheck: filter.NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"0.9.0-beta",
				"",
				os.Getenv("GH_TOKEN"),
				"", time.Now(), nil),
			errRegex: "manifest unknown",
		},
		"docker tag match": {
			requireDockerCheck: filter.NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"0.9.0",
				"",
				os.Getenv("GH_TOKEN"),
				"", time.Now(), nil),
		},
		"regex version mismatch": {
			requireRegexVersion: "v([0-9.]+)",
			errRegex:            "regex not matched on version",
		},
		"urlCommand regex mismatch": {
			regex:    stringPtr("^[0-9]+$"),
			errRegex: "regex .* didn't return any matches",
		},
		"valid semantic version query": {
			regex: stringPtr("v([0-9.]+)"),
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
		},
		"same version found": {
			regex:           stringPtr("([0-9.]+)"),
			deployedVersion: "1.2.1",
		},
		"no deployed version lookup": {
			regex:             stringPtr("([0-9.]+)-beta"),
			wantLatestVersion: stringPtr("1.2.2"),
		},
		"non-semantic version lookup": {
			regex:                 stringPtr("v[0-9.]+"),
			wantLatestVersion:     stringPtr("v1.2.2"),
			nonSemanticVersioning: true,
		},
		"github lookup": {
			githubService: true,
		},
		"github lookup on repo that uses tags, not releases": {
			githubService: true,
			url:           "go-vikunja/api",
			regex:         stringPtr("v([0-9.]+)"),
			outputRegex:   `no tags found on /releases, trying /tags`,
		},
		"github lookup with no access token": {
			githubService: true,
			noAccessToken: true,
		},
		"github lookup with failing urlCommand match": {
			githubService: true,
			regex:         stringPtr("x([0-9.]+)"),
			errRegex:      "no releases were found matching the url_commands",
		},
		"url_command makes all versions non-semmantic": {
			githubService: true,
			regex:         stringPtr(`([0-9.]+\.)`),
			errRegex:      "no releases were found matching the url_commands",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			try := 0
			temporaryFailureInNameResolution := true
			for temporaryFailureInNameResolution != false {
				stdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
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
				lookup.Status.SetLatestVersion(tc.latestVersion, false)
				lookup.Status.SetDeployedVersion(tc.deployedVersion, false)
				lookup.Require.RegexContent = tc.requireRegexContent
				lookup.Require.RegexVersion = tc.requireRegexVersion
				lookup.Require.Command = tc.requireCommand
				lookup.Require.Docker = tc.requireDockerCheck

				// WHEN Query is called on it
				_, err := lookup.Query(true, &util.LogFrom{})

				// THEN any err is expected
				w.Close()
				out, _ := io.ReadAll(r)
				os.Stdout = stdout
				e := util.ErrorToString(err)
				if tc.errRegex == "" {
					tc.errRegex = "^$"
				}
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
				// AND the output contains the expected strings
				output := string(out)
				re = regexp.MustCompile(tc.outputRegex)
				match = re.MatchString(output)
				if !match {
					t.Fatalf("match for %q not found in:\n%q",
						tc.outputRegex, output)
				}
				// AND the LatestVersion is as expected
				if tc.wantLatestVersion != nil &&
					*tc.wantLatestVersion != lookup.Status.LatestVersion() {
					t.Fatalf("wanted LatestVersion to become %q, not %q",
						*tc.wantLatestVersion, lookup.Status.LatestVersion())
				}
			}
		})
	}
}

func TestLookup_Query__EmptyListETagChanged(t *testing.T) {
	// Lock so that default empty list ETag isn't changed by other tests
	emptyListETagTestMutex.Lock()
	defer emptyListETagTestMutex.Unlock()
	invalidETag := "123"

	// GIVEN a Lookup
	try := 0
	temporaryFailureInNameResolution := true
	for temporaryFailureInNameResolution != false {
		stdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		try++
		setEmptyListETag(invalidETag)
		temporaryFailureInNameResolution = false
		lookup := testLookup(false, false)
		lookup.URL = "go-vikunja/api"
		lookup.URLCommands[0].Regex = stringPtr("v([0-9.]+)")

		// WHEN Query is called on it
		_, err := lookup.Query(true, &util.LogFrom{})

		// THEN any err is expected
		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = stdout
		e := util.ErrorToString(err)
		errRegex := "^$"
		re := regexp.MustCompile(errRegex)
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
				errRegex, e)
		}
		// AND the output contains the expected strings
		output := string(out)
		wantOutputRegex := `/releases gave \[\], trying /tags`
		re = regexp.MustCompile(wantOutputRegex)
		match = re.MatchString(output)
		if !match {
			t.Fatalf("match for %q not found in:\n%q",
				wantOutputRegex, output)
		}
	}
}

func TestLookup_QueryGitHubETag(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		attempts                   int
		eTagChanged                int
		eTagUnchangedUseCache      int
		initialRequireRegexVersion string
		urlCommands                filter.URLCommandSlice
		errRegex                   string
	}{
		// Keeps .Releases incase filters change
		"three requests only uses 1 api limit": {
			attempts:              3,
			eTagChanged:           1,
			eTagUnchangedUseCache: 3, // 2 attempts + 1 recheck
			errRegex:              `^$`},
		"if initial request fails filter, cached results will be used": {
			attempts:                   3,
			eTagChanged:                1,
			eTagUnchangedUseCache:      3, // 2 attempts + 1 recheck
			initialRequireRegexVersion: `^FOO$`,
			errRegex:                   `regex not matched on version`},
		"invalid url_commands will catch no versions": {
			attempts:              2,
			eTagChanged:           1,
			eTagUnchangedUseCache: 1,
			urlCommands: filter.URLCommandSlice{
				{Type: "regex", Regex: stringPtr(`^FOO$`)}},
			errRegex: `no releases were found matching the url_commands
no releases were found matching the url_commands and/or require`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			lookup := testLookup(false, false)
			lookup.GitHubData.SetETag("foo")
			lookup.Status.ServiceID = &name
			lookup.Require.RegexVersion = tc.initialRequireRegexVersion
			lookup.URLCommands = tc.urlCommands

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

				_, err := lookup.Query(true, &util.LogFrom{})
				if err != nil {
					errors += "--" + err.Error()
				}
				t.Logf("attempt %d, ETag: %s",
					attempt, lookup.GitHubData.ETag())
			}

			// THEN any err is expected
			w.Close()
			o, _ := io.ReadAll(r)
			out := string(o)
			os.Stdout = stdout
			tc.errRegex = strings.ReplaceAll(tc.errRegex, "\n", "--")
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(errors)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, errors)
			}
			gotETagChanged := strings.Count(out, "new ETag")
			if gotETagChanged != tc.eTagChanged {
				t.Errorf("new ETag - got=%d, want=%d\n%s",
					gotETagChanged, tc.eTagChanged, out)
			}
			gotETagUnchangedUseCache := strings.Count(out, "Using cached releases")
			if gotETagUnchangedUseCache != tc.eTagUnchangedUseCache {
				t.Errorf("ETag unchanged use cache - got=%d, want=%d\n%s",
					gotETagUnchangedUseCache, tc.eTagUnchangedUseCache, out)
			}
		})
	}
}
