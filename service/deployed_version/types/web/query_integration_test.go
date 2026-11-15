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

//go:build integration

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/config/decode"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Track(t *testing.T) {
	dvCfg := plainDefaultsConfig()

	plainStableVersion := "1.2.1"
	plainNonSemanticVersionAsSemantic := "1.2.2"
	plainNonSemanticVersion := "ver" + plainNonSemanticVersionAsSemantic
	jsonBarVersion := "1.2.2"

	tests := []struct {
		name                                      string
		env                                       map[string]string
		lookup                                    *Lookup
		allowInvalidCerts, semanticVersioning     bool
		basicAuth                                 *BasicAuth
		startDeployedVersion, wantDeployedVersion string
		startLatestVersion, wantLatestVersion     string
		wantAnnounces, wantDatabaseMessages       int
		deleting                                  bool
	}{
		{
			name:                "get semantic version with regex",
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain.URLValid,
				Regex: `non-semantic: "ver([^"]+)`,
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "get semantic version from JSON",
			startLatestVersion:  jsonBarVersion,
			wantLatestVersion:   jsonBarVersion,
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  test.LookupJSON.URLValid,
				JSON: "bar",
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1, wantAnnounces: 1,
		},
		{
			name:                "get semantic version from multi-level JSON",
			startLatestVersion:  "3.2.1",
			wantLatestVersion:   "3.2.1",
			wantDeployedVersion: "3.2.1",
			lookup: &Lookup{
				URL:  test.LookupJSON.URLValid,
				JSON: "foo.bar.version",
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "reject non-semantic versions",
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   test.LookupPlain.URLValid,
				Regex: `non-semantic: ("[^"]+)`,
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 0,
			wantAnnounces:        0,
		},
		{
			name:                "allow non-semantic version",
			startLatestVersion:  plainNonSemanticVersion,
			wantLatestVersion:   plainNonSemanticVersion,
			wantDeployedVersion: plainNonSemanticVersion,
			lookup: &Lookup{
				URL:   test.LookupPlain.URLValid,
				Regex: `non-semantic: "([^"]+)`,
			},
			semanticVersioning:   false,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "get version behind basic auth",
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			basicAuth: &BasicAuth{
				Username: "test",
				Password: "123",
			},
			lookup: &Lookup{
				URL:   test.LookupWithBasicAuth.URLValid,
				Regex: `non-semantic: "ver([^"]+)`,
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name: "env vars in basic auth",
			env: map[string]string{
				"TEST_LOOKUP__DV_TRACK_ONE": "tes",
				"TEST_LOOKUP__DV_TRACK_TWO": "23",
			},
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			basicAuth: &BasicAuth{
				Username: "${TEST_LOOKUP__DV_TRACK_ONE}t",
				Password: "1${TEST_LOOKUP__DV_TRACK_TWO}",
			},
			lookup: &Lookup{
				URL:   test.LookupWithBasicAuth.URLValid,
				Regex: `non-semantic: "ver([^"]+)`,
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "get version behind an invalid cert",
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain.URLInvalid,
				Regex: `non-semantic: "ver([^"]+)`,
			},
			allowInvalidCerts:    true,
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "fail due to disallowed invalid cert",
			startLatestVersion:  "",
			wantLatestVersion:   "",
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   test.LookupPlain.URLInvalid,
				Regex: `non-semantic: "ver([^"]+)`,
			},
			allowInvalidCerts:    false,
			semanticVersioning:   true,
			wantDatabaseMessages: 0,
			wantAnnounces:        0,
		},
		{
			name:                 "update to a newer version",
			startLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantLatestVersion:    plainNonSemanticVersionAsSemantic,
			startDeployedVersion: plainStableVersion,
			wantDeployedVersion:  plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain.URLValid,
				Regex: `non-semantic: "ver([^"]+)`,
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                 "update to an older version",
			startLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantLatestVersion:    plainNonSemanticVersionAsSemantic,
			startDeployedVersion: "1.2.3",
			wantDeployedVersion:  plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain.URLValid,
				Regex: `non-semantic: "ver([^"]+)`,
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "get a deployed version newer than latest version",
			startLatestVersion:  plainStableVersion,
			wantLatestVersion:   plainStableVersion,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:  test.LookupJSON.URLValid,
				JSON: "bar",
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "get a deployed version older than latest version only updates deployed",
			startLatestVersion:  "1.2.3",
			wantLatestVersion:   "1.2.3",
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  test.LookupJSON.URLValid,
				JSON: "bar",
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "get a deployed version with no latest version",
			startLatestVersion:  "",
			wantLatestVersion:   "",
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  test.LookupJSON.URLValid,
				JSON: "bar",
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:     "deleting service stops track",
			deleting: true,
			lookup: &Lookup{
				URL:  test.LookupJSON.URLValid,
				JSON: "bar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars and captures stdout.

			releaseLog := test.CaptureLog(t, logx.Default())

			test.IsolateEnv(t, tc.env)
			if tc.lookup != nil {
				// Marshal and Unmarshal to set Type.
				data, err := decode.Marshal("json", tc.lookup)
				if err != nil {
					t.Fatalf("decode.Marshal() failed: %v", err)
				}
				if err := decode.Unmarshal("json", data, tc.lookup); err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}

				tc.lookup.AllowInvalidCerts = &tc.allowInvalidCerts
				tc.lookup.BasicAuth = tc.basicAuth
				tc.lookup.Defaults = dvCfg.Soft
				tc.lookup.HardDefaults = dvCfg.Hard
				optCfg := opttest.PlainDefaultsConfig()
				tc.lookup.Options, err = opt.Decode(
					"yaml", []byte("interval: 2s"),
					optCfg,
				)
				if err != nil {
					t.Fatalf("opt.Decode() failed: %v", err)
				}
				tc.lookup.Options.SemanticVersioning = &tc.semanticVersioning
				dbChannel := make(chan dbtype.Message, 4)
				announceChannel := make(chan []byte, 4)
				svcStatus := status.New(
					announceChannel, dbChannel, nil,
					"",
					tc.startDeployedVersion, "",
					tc.startLatestVersion, "",
					"",
					&dashboard.Options{},
				)
				tc.lookup.Status = svcStatus
				tc.lookup.Status.ServiceInfo.ID = tc.name
				tc.lookup.Status.ServiceInfo.WebURL = tc.lookup.URL
				if tc.deleting {
					tc.lookup.Status.SetDeleting()
				}

				tc.lookup.InitMetrics(tc.lookup)
				t.Cleanup(func() {
					tc.lookup.Status.SetDeleting()
					tc.lookup.DeleteMetrics(tc.lookup)
				})
			}
			didFinish := make(chan bool, 1)

			go func() {
				tc.lookup.Track()
				didFinish <- true
			}()

			if tc.deleting {
				select {
				case <-didFinish:
				case <-time.After(time.Second):
					t.Fatal("Lookup.Track() didn't return within 1s")
				}
				if got := len(tc.lookup.Status.AnnounceChannel); got != 0 {
					t.Errorf(
						"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: 0",
						got,
					)
				}
				if got := len(tc.lookup.Status.DatabaseChannel); got != 0 {
					t.Errorf(
						"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: 0",
						got,
					)
				}
				return
			}

			serviceID := tc.lookup.GetServiceID()
			queries := func() float64 {
				return testutil.ToFloat64(
					metric.DeployedVersionQueryResultTotal.WithLabelValues(
						serviceID, tc.lookup.GetType(), metric.ActionResultSuccess,
					),
				) + testutil.ToFloat64(
					metric.DeployedVersionQueryResultTotal.WithLabelValues(
						serviceID, tc.lookup.GetType(), metric.ActionResultFail,
					),
				)
			}
			deadline := time.Now().Add(10 * time.Second)
			for queries() == 0 {
				if time.Now().After(deadline) {
					t.Fatal("Lookup.Track() didn't query within 10s")
				}
				time.Sleep(50 * time.Millisecond)
			}
			_ = releaseLog()

			if gotDeployedVersion := tc.lookup.Status.DeployedVersion(); gotDeployedVersion != tc.wantDeployedVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					gotDeployedVersion, tc.wantDeployedVersion,
				)
			}

			if gotLatestVersion := tc.lookup.Status.LatestVersion(); gotLatestVersion != tc.wantLatestVersion {
				t.Errorf(
					"Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
					gotLatestVersion, tc.wantLatestVersion,
				)
			}

			if gotAnnounces := len(tc.lookup.Status.AnnounceChannel); gotAnnounces != tc.wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotAnnounces, tc.wantAnnounces,
				)
				test.LogChannel(t, "Status.AnnounceChannel", tc.lookup.Status.AnnounceChannel)
			}
			if gotDatabaseMessages := len(tc.lookup.Status.DatabaseChannel); gotDatabaseMessages != tc.wantDatabaseMessages {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					gotDatabaseMessages, tc.wantDatabaseMessages,
				)
			}
		})
	}
}

func TestLookup_Query(t *testing.T) {
	tests := []struct {
		name                        string
		env                         map[string]string
		serverSetup                 func() *httptest.Server
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
	}{
		{
			name: "JSON/lookup value that doesn't exist",
			overrides: test.TrimYAML(`
				url:  ` + test.LookupJSON.URLValid + `
				json: something
			`),
			errRegex: `failed to find value for \"[^"]+\" in `,
		},
		{
			name: "URL that doesn't resolve to JSON",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				json: something
			`),
			errRegex: `failed to unmarshal`,
		},
		{
			name: "POST/success",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupPlainPOST.URLValid + `
				body: '` + test.LookupPlainPOST.DataPass + `'
				regex: ver([0-9.]+)
			`),
			wantVersion: `^[0-9.]+$`,
			errRegex:    `^$`,
		},
		{
			name: "POST/fail, invalid body",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupPlainPOST.URLValid + `
				body: '` + test.LookupPlainPOST.DataFail + `'
			`),
			errRegex: `non-2XX response code`,
		},
		{
			name: "passing regex",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				regex: 'version: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^\d+\.\d+\.\d+$`,
			errRegex:         `^$`,
		},
		{
			name: "url from env",
			env: map[string]string{
				"TEST_LOOKUP__DV_QUERY_ONE": test.LookupPlain.URLValid,
			},
			overrides: test.TrimYAML(`
				url: ${TEST_LOOKUP__DV_QUERY_ONE}
				regex: 'version: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^\d+\.\d+\.\d+$`,
			errRegex:         `^$`,
		},
		{
			name: "url from env partial",
			env:  map[string]string{"TEST_LOOKUP__DV_QUERY_TWO": "valid.release-argus"},
			overrides: test.TrimYAML(`
				url: https://${TEST_LOOKUP__DV_QUERY_TWO}.io/json
				json: foo.bar.version
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^\d+\.\d+\.\d+$`,
			errRegex:         `^$`,
		},
		{
			name: "passing regex with no capture group",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				regex: '[0-9.]+'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^[0-9.]+$`,
			errRegex:         `^$`,
		},
		{
			name: "regex with template",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				regex: '(stable).*(version).*"([\d.]+).*(and)'
				regex_template: '$2 $1 $4, $3'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^version stable and, 1\.2\.1$`,
			errRegex:         `^$`,
		},
		{
			name: "failing regex",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				regex: '^bishBashBosh$'
			`),
			errRegex: `regex .* didn't return any matches on`,
		},
		{
			name: "handle non-semantic (only major) version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				regex: '(\d+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^\d+$`,
			errRegex:         `^$`,
		},
		{
			name: "want semantic versioning but get non-semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				regex: 'non-semantic: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: true`,
			errRegex:         `failed to convert "[^"]+" to a semantic version`,
		},
		{
			name: "allow non-semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				regex: 'non-semantic: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^ver[0-9.]+$`,
			errRegex:         `^$`,
		},
		{
			name: "valid semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupJSON.URLValid + `
				json: bar
			`),
			wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`,
			errRegex:    `^$`,
		},
		{
			name: "headers fail",
			overrides: test.TrimYAML(`
				url: https://api.github.com/repos/` + test.ArgusGitHubRepo + `/releases/latest
				json: something
				headers:
					- key: Authorization
						value: token ghp_FAIL
			`),
			errRegex: `non-2XX response code: 401`,
		},
		{
			name:      "404",
			overrides: `url: ` + test.ValidCertHTTPS + `/foo/bar`,
			errRegex:  `non-2XX response code: 404`,
		},
		{
			name:        "body/over the size cap is reported, not truncated",
			serverSetup: oversizedBodyServer,
			errRegex:    `^response body is larger than the \d+ MiB limit$`,
		},
		{
			name: "version from header/pass, exact casing",
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupResponseHeader.URLValid + `
				target_header: ` + test.LookupResponseHeader.HeaderKeyPass + `
			`),
			wantVersion: `^\d+\.\d+\.\d+$`,
			errRegex:    `^$`,
		},
		{
			name: "version from header/pass, mixed casing",
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupResponseHeader.URLValid + `
				target_header: ` + test.LookupResponseHeader.HeaderKeyPassMixedCase + `
			`),
			wantVersion: `^\d+\.\d+\.\d+$`,
			errRegex:    `^$`,
		},
		{
			name: "version from header/fail",
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupResponseHeader.URLValid + `
				target_header: ` + test.LookupResponseHeader.HeaderKeyFail + `
			`),
			errRegex: `^target header "[^"]+" not found$`,
		},
		{
			name:        "version from header/non-2XX, header present",
			serverSetup: versionHeaderServer("1.2.3", http.StatusForbidden),
			overrides:   `target_header: X-Version`,
			wantVersion: `^1\.2\.3$`, // target_header is read before the status check.
			errRegex:    `^$`,
		},
		{
			name:        "version from header/non-2XX, header absent",
			serverSetup: versionHeaderServer("", http.StatusForbidden),
			overrides:   `target_header: X-Version`,
			errRegex:    `^target header "X-Version" not found \(status: 403\)$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.IsolateEnv(t, tc.env)
			dvl := testLookup(t)
			dvl.JSON = ""
			overrides := tc.overrides
			if tc.serverSetup != nil {
				server := tc.serverSetup()
				t.Cleanup(server.Close)
				overrides = "url: " + server.URL + "\n" + overrides
			}
			if err := dvl.ApplyOverrides("yaml", []byte(overrides)); err != nil {
				t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
			}
			if tc.optionsOverrides != "" {
				if err := decode.Unmarshal("yaml", []byte(tc.optionsOverrides), dvl.Options); err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}
			}

			err := dvl.Query(true, logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if tc.wantVersion != "" {
				if version := dvl.Status.DeployedVersion(); !util.RegexCheck(tc.wantVersion, version) {
					t.Errorf(
						"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
						version, tc.wantVersion,
					)
				}
			}
		})
	}
}

func TestLookup_HTTPRequest(t *testing.T) {
	headerAuthURL := test.HeaderAuthServer(t, "1.2.3")
	tests := []struct {
		name        string
		env         map[string]string
		serverSetup func() *httptest.Server
		overrides   string
		bodyRegex   string
		errRegex    string
	}{
		{
			name:      "url/invalid",
			overrides: `url: "https://	test"`,
			errRegex:  `invalid control character in URL`,
		},
		{
			name:      "url/unknown",
			overrides: `url: https://release-argus.invalid-tld`,
			errRegex:  `no such host`,
		},
		{
			name:      "url/valid",
			overrides: `url: ` + test.LookupPlain.URLValid,
			errRegex:  `^$`,
		},
		{
			name: "url/from env",
			env: map[string]string{
				"TEST_LOOKUP__DV_HTTP_REQUEST_ONE": test.LookupPlain.URLValid,
			},
			overrides: `url: ${TEST_LOOKUP__DV_HTTP_REQUEST_ONE}`,
			errRegex:  `^$`,
		},
		{
			name: "url/from env partial",
			env: map[string]string{
				"TEST_LOOKUP__DV_HTTP_REQUEST_TWO": strings.TrimSuffix(
					strings.TrimPrefix(test.ValidCertHTTPS, "https://"),
					".io",
				),
			},
			overrides: `url: https://${TEST_LOOKUP__DV_HTTP_REQUEST_TWO}.io/plain`,
			errRegex:  `^$`,
		},
		{
			name:      "404",
			overrides: `url: ` + test.ValidCertHTTPS + `/foo/bar`,
			errRegex:  `non-2XX response code: 404`,
		},
		{
			name:        "body/over the size cap is reported, not truncated",
			serverSetup: oversizedBodyServer,
			errRegex:    `^response body is larger than the \d+ MiB limit$`,
		},
		{
			name: "headers/pass",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + headerAuthURL + `
				headers:
					- key: ` + test.LookupWithHeaderAuth.HeaderKey + `
						value: ` + test.LookupWithHeaderAuth.HeaderValuePass + `
			`),
			bodyRegex: `^[\d.]+$`,
			errRegex:  `^$`,
		},
		{
			name: "headers/fail",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + headerAuthURL + `
				headers:
					- key: ` + test.LookupWithHeaderAuth.HeaderKey + `
						value: ` + test.LookupWithHeaderAuth.HeaderValueFail + `
			`),
			bodyRegex: `Hook rules were not satisfied\.`,
			errRegex:  `^$`,
		},
		{
			name: "basic auth/pass",
			overrides: test.TrimYAML(`
				url: ` + test.LookupWithBasicAuth.URLValid + `
				basic_auth:
					username: ` + test.LookupWithBasicAuth.Username + `
					password: ` + test.LookupWithBasicAuth.Password + `
			`),
			errRegex: `^$`,
		},
		{
			name: "basic auth/fail",
			overrides: test.TrimYAML(`
				url: ` + test.LookupWithBasicAuth.URLValid + `
				basic_auth:
					username: ` + test.LookupWithBasicAuth.Username + `
					password: ` + test.LookupWithBasicAuth.Password + `-
			`),
			errRegex: `non-2XX response code: 401`,
		},
		{
			name: "self-signed cert/pass",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLInvalid + `
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name: "self-signed cert/fail",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLInvalid + `
				allow_invalid_certs: false
			`),
			errRegex: `x509 \(certificate invalid\)`,
		},
		{
			name:        "target_header/2XX, header found",
			serverSetup: versionHeaderServer("1.2.3", http.StatusOK),
			overrides:   `target_header: X-Version`,
			bodyRegex:   `^1\.2\.3$`,
			errRegex:    `^$`,
		},
		{
			name:        "target_header/2XX, header not found",
			serverSetup: versionHeaderServer("", http.StatusOK),
			overrides:   `target_header: X-Version`,
			errRegex:    `^target header "X-Version" not found$`,
		},
		{
			name:        "target_header/non-2XX, header found",
			serverSetup: versionHeaderServer("1.2.3", http.StatusForbidden),
			overrides:   `target_header: X-Version`,
			bodyRegex:   `^1\.2\.3$`, // target_header is read before the status check.
			errRegex:    `^$`,
		},
		{
			name:        "target_header/non-2XX, header not found",
			serverSetup: versionHeaderServer("", http.StatusInternalServerError),
			overrides:   `target_header: X-Version`,
			errRegex:    `^target header "X-Version" not found \(status: 500\)$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.IsolateEnv(t, tc.env)
			lookup := testLookup(t)
			overrides := tc.overrides
			if tc.serverSetup != nil {
				server := tc.serverSetup()
				t.Cleanup(server.Close)
				overrides = "url: " + server.URL + "\n" + overrides
			}
			if err := lookup.UnmarshalYAML([]byte(overrides)); err != nil {
				t.Fatalf("Lookup.UnmarshalYAML() failed: %v", err)
			}

			body, err := lookup.httpRequest(logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if tc.bodyRegex != "" {
				if b := string(body); !util.RegexCheck(tc.bodyRegex, b) {
					t.Errorf("body mismatch\ngot:  %q\nwant: %q", b, tc.bodyRegex)
				}
			}
		})
	}
}

// oversizedBodyServer serves a body larger than httpx.MaxBodyBytes.
func oversizedBodyServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			filler := bytes.Repeat([]byte("a"), 1<<20)
			for written := 0; written <= httpx.MaxBodyBytes; written += len(filler) {
				if _, err := w.Write(filler); err != nil {
					return
				}
			}
		}))
}

// versionHeaderServer responds with statusCode, and with version in the X-Version header when set.
func versionHeaderServer(version string, statusCode int) func() *httptest.Server {
	return func() *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if version != "" {
				w.Header().Set("X-Version", version)
			}
			w.WriteHeader(statusCode)
		}))
	}
}
