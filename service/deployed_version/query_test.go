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

package deployedver

import (
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	dbtype "github.com/release-argus/Argus/db/types"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metric"
	"gopkg.in/yaml.v3"
)

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN a Lookup()
	tests := map[string]struct {
		env      map[string]string
		url      string
		errRegex string
	}{
		"invalid url": {
			url:      "invalid://	test",
			errRegex: `invalid control character in URL`},
		"unknown url": {
			url:      "https://release-argus.invalid-tld",
			errRegex: `no such host`},
		"valid url": {
			url:      "https://release-argus.io",
			errRegex: `^$`},
		"url from env": {
			env:      map[string]string{"TEST_LOOKUP__DV_HTTP_REQUEST_ONE": "https://release-argus.io"},
			url:      "${TEST_LOOKUP__DV_HTTP_REQUEST_ONE}",
			errRegex: `^$`},
		"url from env partial": {
			env:      map[string]string{"TEST_LOOKUP__DV_HTTP_REQUEST_TWO": "release-argus"},
			url:      "https://${TEST_LOOKUP__DV_HTTP_REQUEST_TWO}.io",
			errRegex: `^$`},
		"404": {
			errRegex: `non-2XX response code: 404`,
			url:      "https://release-argus.io/foo/bar",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
			}
			lookup := testLookup()
			lookup.URL = tc.url

			// WHEN httpRequest is called on it
			_, err := lookup.httpRequest(util.LogFrom{})

			// THEN any err is expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN a Lookup()
	tests := map[string]struct {
		env                         map[string]string
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
	}{
		"JSON lookup value that doesn't exist": {
			overrides: test.TrimYAML(`
				url: https://valid.release-argus.io/json
				json: something
			`),
			errRegex: `failed to find value for \"[^"]+\" in `,
		},
		"URL that doesn't resolve to JSON": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				json: something
			`),
			errRegex: `failed to unmarshal`,
		},
		"POST - success": {
			overrides: test.TrimYAML(`
				method: POST
				url: https://valid.release-argus.io/plain_post
				body: '{"argus":"test"}'
				regex: ver([0-9.]+)
			`),
			wantVersion: "[0-9.]+",
			errRegex:    `^$`,
		},
		"POST - fail, invalid body": {
			overrides: test.TrimYAML(`
				method: POST
				url: https://valid.release-argus.io/plain_post
				body: '{"argus":"fail"}'
			`),
			errRegex: `non-2XX response code`,
		},
		"passing regex": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				regex: '([0-9]+)\s+[^>]+>The Argus Developers'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: "[0-9]{4}",
			errRegex:    `^$`,
		},
		"url from env": {
			env: map[string]string{"TEST_LOOKUP__DV_QUERY_ONE": "https://release-argus.io"},
			overrides: test.TrimYAML(`
				url: ${TEST_LOOKUP__DV_QUERY_ONE}
				regex: '([0-9]+)\s+[^>]+>The Argus Developers'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: "[0-9]{4}",
			errRegex:    `^$`,
		},
		"url from env partial": {
			env: map[string]string{"TEST_LOOKUP__DV_QUERY_TWO": "release-argus"},
			overrides: test.TrimYAML(`
				url: https://${TEST_LOOKUP__DV_QUERY_TWO}.io
				regex: '([0-9]+)\s+[^>]+>The Argus Developers'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: "[0-9]{4}",
			errRegex:    `^$`,
		},
		"passing regex with no capture group": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				regex: '[0-9]{4}'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: "[0-9]{4}",
			errRegex:    `^$`,
		},
		"regex with template": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				regex: '([0-9]+)\s+<[^>]+>(The) (Argus) (Developers)'
				regex_template: '$2 $1 $4, $3'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			errRegex:    `^$`,
			wantVersion: "The [0-9]+ Developers, Argus",
		},
		"failing regex": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				regex: '^bishBashBosh$'
			`),
			errRegex: `regex .* didn't return any matches on`,
		},
		"handle non-semantic (only major) version": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				regex: '([0-9]+)\s+<[^>]+>The Argus Developers'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
		},
		"want semantic versioning but get non-semantic version": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				regex: '([0-9]+\s+)<[^>]+>The Argus Developers'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: true
			`),
			errRegex: `failed converting .* to a semantic version`,
		},
		"allow non-semantic versioning and get non-semantic version": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io
				regex: '([0-9]+\s+)<[^>]+>The Argus Developers'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			errRegex: `^$`,
		},
		"valid semantic version": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io/docs/getting-started/
				regex: argus-([0-9.]+)\.
			`),
			errRegex:    `^$`,
			wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`,
		},
		"headers fail": {
			overrides: test.TrimYAML(`
				url: https://api.github.com/repos/release-argus/argus/releases/latest
				json: something
				headers:
					- key: Authorization
						value: token ghp_FAIL
			`),
			errRegex: `non-2XX response code: 401`,
		},
		"404": {
			overrides: test.TrimYAML(`
				url: https://release-argus.io/foo/bar
			`),
			errRegex: `non-2XX response code: 404`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}
			dvl := testLookup()
			dvl.JSON = ""
			err := yaml.Unmarshal([]byte(tc.overrides), dvl)
			if err != nil {
				t.Fatalf("failed to unmarshal overrides: %s", err)
			}
			err = yaml.Unmarshal([]byte(tc.optionsOverrides), dvl.Options)
			if err != nil {
				t.Fatalf("failed to unmarshal options overrides: %s", err)
			}

			// WHEN Query is called on it
			version, err := dvl.Query(true, util.LogFrom{})

			// THEN any err is expected
			if tc.wantVersion != "" {
				if !util.RegexCheck(tc.wantVersion, version) {
					t.Errorf("want version=%q\ngot  version=%q",
						tc.wantVersion, version)
				}
			}
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestLookup_Track(t *testing.T) {
	plainStableVersion := "1.2.1"
	plainNonSemanticVersionAsSemantic := "1.2.2"
	plainNonSemanticVersion := "ver" + plainNonSemanticVersionAsSemantic
	jsonBarVersion := "1.2.2"
	// GIVEN a Lookup()
	tests := map[string]struct {
		env                                       map[string]string
		lookup                                    *Lookup
		allowInvalidCerts, semanticVersioning     bool
		basicAuth                                 *BasicAuth
		expectFinish                              bool
		wait                                      time.Duration
		errRegex                                  string
		startDeployedVersion, wantDeployedVersion string
		startLatestVersion, wantLatestVersion     string
		wantAnnounces, wantDatabaseMessages       int
		deleting                                  bool
	}{
		"nil Lookup exits immediately": {
			lookup:       nil,
			wait:         10 * time.Millisecond,
			expectFinish: true,
		},
		"get semantic version with regex": {
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get semantic version from json": {
			startLatestVersion:  jsonBarVersion,
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1, wantAnnounces: 1,
		},
		"get semantic version from multi-level json": {
			startLatestVersion:  "3.2.1",
			wantDeployedVersion: "3.2.1",
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "foo.bar.version"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"reject non-semantic versions": {
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: ("[^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 0,
			wantAnnounces:        0,
		},
		"allow non-semantic version": {
			startLatestVersion:  plainNonSemanticVersion,
			wantDeployedVersion: plainNonSemanticVersion,
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "([^"]+)`},
			semanticVersioning:   false,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get version behind basic auth": {
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			basicAuth: &BasicAuth{
				Username: "test",
				Password: "123"},
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/basic-auth",
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"env vars in basic auth": {
			env: map[string]string{
				"TEST_LOOKUP__DV_TRACK_ONE": "tes",
				"TEST_LOOKUP__DV_TRACK_TWO": "23"},
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			basicAuth: &BasicAuth{
				Username: "${TEST_LOOKUP__DV_TRACK_ONE}t",
				Password: "1${TEST_LOOKUP__DV_TRACK_TWO}"},
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/basic-auth",
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get version behind an invalid cert": {
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   "https://invalid.release-argus.io/plain",
				Regex: `non-semantic: "ver([^"]+)`},
			allowInvalidCerts:    true,
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"fail due to an unallowed invalid cert": {
			startLatestVersion:  "",
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   "https://invalid.release-argus.io/plain",
				Regex: `non-semantic: "ver([^"]+)`},
			allowInvalidCerts:    false,
			semanticVersioning:   true,
			wantDatabaseMessages: 0,
			wantAnnounces:        0,
		},
		"update to a newer version": {
			startLatestVersion:   plainNonSemanticVersionAsSemantic,
			startDeployedVersion: plainStableVersion,
			wantDeployedVersion:  plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"update to an older version": {
			startLatestVersion:   plainNonSemanticVersionAsSemantic,
			startDeployedVersion: "1.2.3",
			wantDeployedVersion:  plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get a newer deployed version than latest version updates both": {
			startLatestVersion:  plainStableVersion,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 2,
			wantAnnounces:        2,
		},
		"get an older deployed version than latest version only updates deployed": {
			startLatestVersion:  "1.2.3",
			wantLatestVersion:   "1.2.3",
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get a deployed version with no latest version updates both": {
			startLatestVersion:  "",
			wantLatestVersion:   jsonBarVersion,
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 2,
			wantAnnounces:        2,
		},
		"deleting service stops track": {
			deleting: true,
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			startLatestVersion:   "",
			startDeployedVersion: "",
			wantLatestVersion:    "",
			wantDeployedVersion:  "",
			wantAnnounces:        0,
			wantDatabaseMessages: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}
			if tc.lookup != nil {
				tc.lookup.AllowInvalidCerts = test.BoolPtr(tc.allowInvalidCerts)
				tc.lookup.BasicAuth = tc.basicAuth
				tc.lookup.Defaults = &Defaults{}
				tc.lookup.HardDefaults = &Defaults{}
				tc.lookup.HardDefaults.Default()
				tc.lookup.Options = opt.New(
					nil, "2s", &tc.semanticVersioning,
					&opt.Defaults{}, &opt.Defaults{})
				dbChannel := make(chan dbtype.Message, 4)
				announceChannel := make(chan []byte, 4)
				svcStatus := status.New(
					&announceChannel, &dbChannel, nil,
					"",
					tc.startDeployedVersion, "",
					tc.startLatestVersion, "",
					"")
				tc.lookup.Status = svcStatus
				tc.lookup.Status.ServiceID = test.StringPtr(name)
				tc.lookup.Status.WebURL = &tc.lookup.URL
				if tc.deleting {
					tc.lookup.Status.SetDeleting()
				}

				tc.lookup.InitMetrics()
				t.Cleanup(func() { tc.lookup.DeleteMetrics() })
			}
			didFinish := make(chan bool, 1)

			// WHEN CheckValues is called on it
			go func() {
				tc.lookup.Track()
				didFinish <- true
			}()

			// THEN the function exits straight away
			time.Sleep(tc.wait)
			if tc.expectFinish {
				if len(didFinish) == 0 {
					t.Fatalf("expected Track to finish in %s, but it didn't",
						tc.wait)
				}
				releaseStdout()
				return
			}
			haveQueried := false
			for haveQueried != false {
				passQ := testutil.ToFloat64(metric.DeployedVersionQueryMetric.WithLabelValues(*tc.lookup.Status.ServiceID, "SUCCESS"))
				failQ := testutil.ToFloat64(metric.DeployedVersionQueryMetric.WithLabelValues(*tc.lookup.Status.ServiceID, "FAIL"))
				if passQ != float64(0) && failQ != float64(0) {
					haveQueried = true
				}
				time.Sleep(time.Second)
			}
			time.Sleep(5 * time.Second)
			stdout := releaseStdout()
			t.Log(stdout)
			if tc.wantDeployedVersion != tc.lookup.Status.DeployedVersion() {
				t.Errorf("expected DeployedVersion to be %q after query, not %q",
					tc.wantDeployedVersion, tc.lookup.Status.DeployedVersion())
			}
			if tc.wantLatestVersion == "" {
				tc.wantLatestVersion = tc.wantDeployedVersion
			}
			if tc.wantLatestVersion != tc.lookup.Status.LatestVersion() {
				t.Errorf("expected LatestVersion to be %q after query, not %q",
					tc.wantLatestVersion, tc.lookup.Status.LatestVersion())
			}
			if tc.wantAnnounces != len(*tc.lookup.Status.AnnounceChannel) {
				t.Errorf("expected AnnounceChannel to have %d messages in queue, not %d",
					tc.wantAnnounces, len(*tc.lookup.Status.AnnounceChannel))
			}
			if tc.wantDatabaseMessages != len(*tc.lookup.Status.DatabaseChannel) {
				t.Errorf("expected DatabaseChannel to have %d messages in queue, not %d",
					tc.wantDatabaseMessages, len(*tc.lookup.Status.DatabaseChannel))
			}

			// Set Deleting to stop the Track
			tc.lookup.Status.SetDeleting()
		})
	}
}
