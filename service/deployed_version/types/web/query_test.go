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
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/yaml.v3"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	metric "github.com/release-argus/Argus/web/metric"
)

func TestLookup_Track(t *testing.T) {
	plainStableVersion := "1.2.1"
	plainNonSemanticVersionAsSemantic := "1.2.2"
	plainNonSemanticVersion := "ver" + plainNonSemanticVersionAsSemantic
	jsonBarVersion := "1.2.2"
	// GIVEN a Lookup.
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
		"get semantic version with regex": {
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain["url_valid"],
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get semantic version from JSON": {
			startLatestVersion:  jsonBarVersion,
			wantLatestVersion:   jsonBarVersion,
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  test.LookupJSON["url_valid"],
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1, wantAnnounces: 1,
		},
		"get semantic version from multi-level JSON": {
			startLatestVersion:  "3.2.1",
			wantLatestVersion:   "3.2.1",
			wantDeployedVersion: "3.2.1",
			lookup: &Lookup{
				URL:  test.LookupJSON["url_valid"],
				JSON: "foo.bar.version"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"reject non-semantic versions": {
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   test.LookupPlain["url_valid"],
				Regex: `non-semantic: ("[^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 0,
			wantAnnounces:        0,
		},
		"allow non-semantic version": {
			startLatestVersion:  plainNonSemanticVersion,
			wantLatestVersion:   plainNonSemanticVersion,
			wantDeployedVersion: plainNonSemanticVersion,
			lookup: &Lookup{
				URL:   test.LookupPlain["url_valid"],
				Regex: `non-semantic: "([^"]+)`},
			semanticVersioning:   false,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get version behind basic auth": {
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			basicAuth: &BasicAuth{
				Username: "test",
				Password: "123"},
			lookup: &Lookup{
				URL:   test.LookupBasicAuth["url_valid"],
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
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			basicAuth: &BasicAuth{
				Username: "${TEST_LOOKUP__DV_TRACK_ONE}t",
				Password: "1${TEST_LOOKUP__DV_TRACK_TWO}"},
			lookup: &Lookup{
				URL:   test.LookupBasicAuth["url_valid"],
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get version behind an invalid cert": {
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain["url_invalid"],
				Regex: `non-semantic: "ver([^"]+)`},
			allowInvalidCerts:    true,
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"fail due to an unallowed invalid cert": {
			startLatestVersion:  "",
			wantLatestVersion:   "",
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   test.LookupPlain["url_invalid"],
				Regex: `non-semantic: "ver([^"]+)`},
			allowInvalidCerts:    false,
			semanticVersioning:   true,
			wantDatabaseMessages: 0,
			wantAnnounces:        0,
		},
		"update to a newer version": {
			startLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantLatestVersion:    plainNonSemanticVersionAsSemantic,
			startDeployedVersion: plainStableVersion,
			wantDeployedVersion:  plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain["url_valid"],
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"update to an older version": {
			startLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantLatestVersion:    plainNonSemanticVersionAsSemantic,
			startDeployedVersion: "1.2.3",
			wantDeployedVersion:  plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain["url_valid"],
				Regex: `non-semantic: "ver([^"]+)`},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get a newer deployed version than latest version": {
			startLatestVersion:  plainStableVersion,
			wantLatestVersion:   plainStableVersion,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:  test.LookupJSON["url_valid"],
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get an older deployed version than latest version only updates deployed": {
			startLatestVersion:  "1.2.3",
			wantLatestVersion:   "1.2.3",
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  test.LookupJSON["url_valid"],
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"get a deployed version with no latest version": {
			startLatestVersion:  "",
			wantLatestVersion:   "",
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  test.LookupJSON["url_valid"],
				JSON: "bar"},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		"deleting service stops track": {
			deleting: true,
			lookup: &Lookup{
				URL:  test.LookupJSON["url_valid"],
				JSON: "bar"},
			startLatestVersion:   "",
			wantLatestVersion:    "",
			startDeployedVersion: "",
			wantDeployedVersion:  "",
			wantAnnounces:        0,
			wantDatabaseMessages: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}
			if tc.lookup != nil {
				// Marshal and Unmarshal to set Type.
				data, _ := json.Marshal(tc.lookup)
				json.Unmarshal(data, tc.lookup)

				tc.lookup.AllowInvalidCerts = test.BoolPtr(tc.allowInvalidCerts)
				tc.lookup.BasicAuth = tc.basicAuth
				tc.lookup.Defaults = &base.Defaults{}
				tc.lookup.HardDefaults = &base.Defaults{}
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

				tc.lookup.InitMetrics(tc.lookup)
				t.Cleanup(func() { tc.lookup.DeleteMetrics(tc.lookup) })
			}
			didFinish := make(chan bool, 1)

			// WHEN CheckValues is called on it.
			go func() {
				tc.lookup.Track()
				didFinish <- true
			}()

			// THEN the function exits straight away.
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
				passQ := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
					*tc.lookup.Status.ServiceID, "SUCCESS", tc.lookup.Type))
				failQ := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
					*tc.lookup.Status.ServiceID, "FAIL", tc.lookup.Type))
				if passQ != float64(0) && failQ != float64(0) {
					haveQueried = true
				}
				time.Sleep(time.Second)
			}
			time.Sleep(5 * time.Second)
			stdout := releaseStdout()
			t.Log(stdout)
			if gotDeployedVersion := tc.lookup.Status.DeployedVersion(); tc.wantDeployedVersion != gotDeployedVersion {
				t.Errorf("expected DeployedVersion to be %q after query, not %q",
					tc.wantDeployedVersion, gotDeployedVersion)
			}
			if gotLatestVersion := tc.lookup.Status.LatestVersion(); tc.wantLatestVersion != gotLatestVersion {
				t.Errorf("expected LatestVersion to be %q after query, not %q",
					tc.wantLatestVersion, gotLatestVersion)
			}
			if gotAnnounces := len(*tc.lookup.Status.AnnounceChannel); tc.wantAnnounces != gotAnnounces {
				for i := 0; i < gotAnnounces; i++ {
					t.Logf("%s\n", <-(*tc.lookup.Status.AnnounceChannel))
				}
				t.Errorf("expected AnnounceChannel to have %d messages in queue, not %d",
					tc.wantAnnounces, gotAnnounces)
			}
			if gotDatabaseMessages := len(*tc.lookup.Status.DatabaseChannel); tc.wantDatabaseMessages != gotDatabaseMessages {
				t.Errorf("expected DatabaseChannel to have %d messages in queue, not %d",
					tc.wantDatabaseMessages, gotDatabaseMessages)
			}

			// Set Deleting to stop the Track.
			tc.lookup.Status.SetDeleting()
		})
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		env                         map[string]string
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
	}{
		"JSON lookup value that doesn't exist": {
			overrides: test.TrimYAML(`
				url:  ` + test.LookupJSON["url_valid"] + `
				json: something
			`),
			errRegex: `failed to find value for \"[^"]+\" in `,
		},
		"URL that doesn't resolve to JSON": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				json: something
			`),
			errRegex: `failed to unmarshal`,
		},
		"POST - success": {
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupPlainPOST["url_valid"] + `
				body: '` + test.LookupPlainPOST["data_pass"] + `'
				regex: ver([0-9.]+)
			`),
			wantVersion: "[0-9.]+",
			errRegex:    `^$`,
		},
		"POST - fail, invalid body": {
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupPlainPOST["url_valid"] + `
				body: '` + test.LookupPlainPOST["data_fail"] + `'
			`),
			errRegex: `non-2XX response code`,
		},
		"passing regex": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: 'version: "([^"]+)'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: `\d\.\d\.\d`,
			errRegex:    `^$`,
		},
		"url from env": {
			env: map[string]string{
				"TEST_LOOKUP__DV_QUERY_ONE": test.LookupPlain["url_valid"]},
			overrides: test.TrimYAML(`
				url: ${TEST_LOOKUP__DV_QUERY_ONE}
				regex: 'version: "([^"]+)'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: `\d\.\d\.\d`,
			errRegex:    `^$`,
		},
		"url from env partial": {
			env: map[string]string{"TEST_LOOKUP__DV_QUERY_TWO": "valid.release-argus"},
			overrides: test.TrimYAML(`
				url: https://${TEST_LOOKUP__DV_QUERY_TWO}.io/json
				json: foo.bar.version
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: `\d\.\d\.\d`,
			errRegex:    `^$`,
		},
		"passing regex with no capture group": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '[0-9.]+'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: "[0-9.]+",
			errRegex:    `^$`,
		},
		"regex with template": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '(stable).*(version).*"([\d.]+).*(and)'
				regex_template: '$2 $1 $4, $3'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			wantVersion: "version stable and, 1.2.1",
			errRegex:    `^$`,
		},
		"failing regex": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '^bishBashBosh$'
			`),
			errRegex: `regex .* didn't return any matches on`,
		},
		"handle non-semantic (only major) version": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '(\d+)'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
		},
		"want semantic versioning but get non-semantic version": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: 'non-semantic: "([^"]+)'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: true
			`),
			errRegex: `failed converting .* to a semantic version`,
		},
		"allow non-semantic version": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: 'non-semantic: "([^"]+)'
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			errRegex: `^$`,
		},
		"valid semantic version": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupJSON["url_valid"] + `
				json: bar
			`),
			wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`,
			errRegex:    `^$`,
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
				url: ` + test.ValidCertHTTPS + `/foo/bar
			`),
			errRegex: `non-2XX response code: 404`,
		},
		"version from header - pass, exact casing": {
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupHeader["url_valid"] + `
				target_header: ` + test.LookupHeader["header_key_pass"]),
			wantVersion: `^\d+\.\d+\.\d+$`,
			errRegex:    `^$`,
		},
		"version from header - pass, mixed casing": {
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupHeader["url_valid"] + `
				target_header: ` + test.LookupHeader["header_key_pass_mixed_case"]),
			wantVersion: `^\d+\.\d+\.\d+$`,
			errRegex:    `^$`,
		},
		"version from header - fail": {
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupHeader["url_valid"] + `
				target_header: ` + test.LookupHeader["header_key_fail"]),
			errRegex: `^target header "[^"]+" not found$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}
			dvl := testLookup(false)
			dvl.JSON = ""
			err := yaml.Unmarshal([]byte(tc.overrides), dvl)
			if err != nil {
				t.Fatalf("failed to unmarshal overrides: %s", err)
			}
			err = yaml.Unmarshal([]byte(tc.optionsOverrides), dvl.Options)
			if err != nil {
				t.Fatalf("failed to unmarshal options overrides: %s", err)
			}

			// WHEN Query is called on it.
			err = dvl.Query(true, logutil.LogFrom{})

			// THEN any err is expected.
			if tc.wantVersion != "" {
				version := dvl.Status.DeployedVersion()
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

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		env       map[string]string
		overrides string
		bodyRegex string
		errRegex  string
	}{
		"url - invalid": {
			overrides: `
				url: "invalid://	test"`,
			errRegex: `invalid control character in URL`},
		"url - unknown": {
			overrides: `
				url: https://release-argus.invalid-tld`,
			errRegex: `no such host`},
		"url - valid": {
			overrides: `
				url: ` + test.LookupPlain["url_valid"],
			errRegex: `^$`},
		"url - from env": {
			env: map[string]string{
				"TEST_LOOKUP__DV_HTTP_REQUEST_ONE": test.LookupPlain["url_valid"]},
			overrides: `
				url: ${TEST_LOOKUP__DV_HTTP_REQUEST_ONE}`,
			errRegex: `^$`},
		"url - from env partial": {
			env: map[string]string{
				"TEST_LOOKUP__DV_HTTP_REQUEST_TWO": strings.TrimSuffix(
					strings.TrimPrefix(test.ValidCertHTTPS, "https://"),
					".io")},
			overrides: `
				url: https://${TEST_LOOKUP__DV_HTTP_REQUEST_TWO}.io/plain`,
			errRegex: `^$`},
		"404": {
			overrides: `
				url: ` + test.ValidCertHTTPS + `/foo/bar`,
			errRegex: `non-2XX response code: 404`,
		},
		"headers - pass": {
			overrides: `
				method: POST
				url: ` + test.LookupWithHeaderAuth["url_valid"] + `
				headers:
					- key: ` + test.LookupWithHeaderAuth["header_key"] + `
						value: ` + test.LookupWithHeaderAuth["header_value_pass"],
			bodyRegex: `^$`,
			errRegex:  `^$`,
		},
		"headers - fail": {
			overrides: `
				method: POST
				url: ` + test.LookupWithHeaderAuth["url_valid"] + `
				headers:
					- key: ` + test.LookupWithHeaderAuth["header_key"] + `
						value: ` + test.LookupWithHeaderAuth["header_value_fail"],
			bodyRegex: `Hook rules were not satisfied\.`,
			errRegex:  `^$`,
		},
		"basic auth - pass": {
			overrides: `
				url: ` + test.LookupBasicAuth["url_valid"] + `
				basic_auth:
					username: ` + test.LookupBasicAuth["username"] + `
					password: ` + test.LookupBasicAuth["password"],
			errRegex: `^$`,
		},
		"basic auth - fail": {
			overrides: `
				url: ` + test.LookupBasicAuth["url_valid"] + `
				basic_auth:
					username: ` + test.LookupBasicAuth["username"] + `
					password: ` + test.LookupBasicAuth["password"] + "-",
			errRegex: `non-2XX response code: 401`,
		},
		"self-signed cert - pass": {
			overrides: `
				url: ` + test.LookupPlain["url_invalid"] + `
				allow_invalid_certs: true`,
			errRegex: `^$`,
		},
		"self-signed cert - fail": {
			overrides: `
				url: ` + test.LookupPlain["url_invalid"] + `
				allow_invalid_certs: false`,
			errRegex: `x509 \(certificate invalid\)`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				previousValue := os.Getenv(k)
				os.Setenv(k, v)
				t.Cleanup(func() {
					if previousValue == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, previousValue)
					}
				})
			}
			lookup := testLookup(false)
			// Apply overrides.
			tc.overrides = test.TrimYAML(tc.overrides)
			err := yaml.Unmarshal([]byte(tc.overrides), lookup)
			if err != nil {
				t.Fatalf("failed to unmarshal overrides: %s", err)
			}

			// WHEN httpRequest is called on it.
			body, err := lookup.httpRequest(logutil.LogFrom{})

			// THEN any err is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			// AND the body matches the expected regex.
			if tc.bodyRegex != "" {
				if !util.RegexCheck(tc.bodyRegex, string(body)) {
					t.Errorf("body mismatch\n%q\ngot:\n%q",
						tc.bodyRegex, string(body))
				}
			}
		})
	}
}
