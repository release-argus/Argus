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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/config/decode"
	dbtype "github.com/release-argus/Argus/db/types"
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
	dvCfg := plainDefaultsConfig(t)

	plainStableVersion := "1.2.1"
	plainNonSemanticVersionAsSemantic := "1.2.2"
	plainNonSemanticVersion := "ver" + plainNonSemanticVersionAsSemantic
	jsonBarVersion := "1.2.2"

	// GIVEN: a Lookup.
	tests := []struct {
		name                                      string
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
		{
			name:                "get semantic version with regex",
			startLatestVersion:  plainNonSemanticVersionAsSemantic,
			wantLatestVersion:   plainNonSemanticVersionAsSemantic,
			wantDeployedVersion: plainNonSemanticVersionAsSemantic,
			lookup: &Lookup{
				URL:   test.LookupPlain["url_valid"],
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
				URL:  test.LookupJSON["url_valid"],
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
				URL:  test.LookupJSON["url_valid"],
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
				URL:   test.LookupPlain["url_valid"],
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
				URL:   test.LookupPlain["url_valid"],
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
				URL:   test.LookupBasicAuth["url_valid"],
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
				URL:   test.LookupBasicAuth["url_valid"],
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
				URL:   test.LookupPlain["url_invalid"],
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
				URL:   test.LookupPlain["url_invalid"],
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
				URL:   test.LookupPlain["url_valid"],
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
				URL:   test.LookupPlain["url_valid"],
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
				URL:  test.LookupJSON["url_valid"],
				JSON: "bar",
			},
			semanticVersioning:   true,
			wantDatabaseMessages: 1,
			wantAnnounces:        1,
		},
		{
			name:                "get an deployed version older than latest version only updates deployed",
			startLatestVersion:  "1.2.3",
			wantLatestVersion:   "1.2.3",
			wantDeployedVersion: jsonBarVersion,
			lookup: &Lookup{
				URL:  test.LookupJSON["url_valid"],
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
				URL:  test.LookupJSON["url_valid"],
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
				URL:  test.LookupJSON["url_valid"],
				JSON: "bar",
			},
			startLatestVersion:   "",
			wantLatestVersion:    "",
			startDeployedVersion: "",
			wantDeployedVersion:  "",
			wantAnnounces:        0,
			wantDatabaseMessages: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			test.SetEnv(t, tc.env)
			if tc.lookup != nil {
				// Marshal and Unmarshal to set Type.
				data, _ := decode.Marshal("json", tc.lookup)
				_ = decode.Unmarshal("json", data, tc.lookup)

				tc.lookup.AllowInvalidCerts = &tc.allowInvalidCerts
				tc.lookup.BasicAuth = tc.basicAuth
				tc.lookup.Defaults = dvCfg.Soft
				tc.lookup.HardDefaults = dvCfg.Hard
				optCfg := opttest.PlainDefaultsConfig(t)
				tc.lookup.Options, _ = opt.Decode(
					"yaml", []byte("interval: 2s"),
					optCfg,
				)
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
				t.Cleanup(func() { tc.lookup.DeleteMetrics(tc.lookup) })
			}
			didFinish := make(chan bool, 1)

			// WHEN: CheckValues is called on it.
			go func() {
				tc.lookup.Track()
				didFinish <- true
			}()

			prefix := fmt.Sprintf("%s\nLookup.Track()", packageName)

			// THEN: the function exits straight away.
			time.Sleep(tc.wait)
			if tc.expectFinish {
				if len(didFinish) == 0 {
					t.Fatalf(
						"%s didn't finish in <= %s",
						prefix, tc.wait,
					)
				}
				releaseStdout()
				return
			}
			haveQueried := false
			for haveQueried != false {
				serviceID := tc.lookup.GetServiceID()
				passQ := testutil.ToFloat64(
					metric.DeployedVersionQueryResultTotal.WithLabelValues(
						serviceID, metric.ActionResultSuccess, tc.lookup.Type,
					),
				)
				failQ := testutil.ToFloat64(
					metric.DeployedVersionQueryResultTotal.WithLabelValues(
						serviceID, metric.ActionResultFail, tc.lookup.Type,
					),
				)
				if passQ != float64(0) && failQ != float64(0) {
					haveQueried = true
				}
				time.Sleep(time.Second)
			}
			time.Sleep(5 * time.Second)
			stdout := releaseStdout()
			t.Log(stdout)
			if gotDeployedVersion := tc.lookup.Status.DeployedVersion(); gotDeployedVersion != tc.wantDeployedVersion {
				t.Errorf(
					"%s .DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, gotDeployedVersion, tc.wantDeployedVersion,
				)
			}
			if gotLatestVersion := tc.lookup.Status.LatestVersion(); gotLatestVersion != tc.wantLatestVersion {
				t.Errorf(
					"%s .LatestVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, gotLatestVersion, tc.wantLatestVersion,
				)
			}
			if gotAnnounces := len(tc.lookup.Status.AnnounceChannel); gotAnnounces != tc.wantAnnounces {
				for i := 0; i < gotAnnounces; i++ {
					t.Logf(
						"%s Announce message - %s\n",
						prefix, <-(tc.lookup.Status.AnnounceChannel),
					)
				}
				t.Errorf(
					"%s Lookup.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					packageName, gotAnnounces, tc.wantAnnounces,
				)
			}
			if gotDatabaseMessages := len(tc.lookup.Status.DatabaseChannel); gotDatabaseMessages != tc.wantDatabaseMessages {
				t.Errorf(
					"%s Lookup.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotDatabaseMessages, tc.wantDatabaseMessages,
				)
			}

			// Set Deleting to stop the Track.
			tc.lookup.Status.SetDeleting()
		})
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                        string
		env                         map[string]string
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
	}{
		{
			name: "JSON lookup value that doesn't exist",
			overrides: test.TrimYAML(`
				url:  ` + test.LookupJSON["url_valid"] + `
				json: something
			`),
			errRegex: `failed to find value for \"[^"]+\" in `,
		},
		{
			name: "URL that doesn't resolve to JSON",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				json: something
			`),
			errRegex: `failed to unmarshal`,
		},
		{
			name: "POST - success",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupPlainPOST["url_valid"] + `
				body: '` + test.LookupPlainPOST["data_pass"] + `'
				regex: ver([0-9.]+)
			`),
			wantVersion: "[0-9.]+",
			errRegex:    `^$`,
		},
		{
			name: "POST - fail, invalid body",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupPlainPOST["url_valid"] + `
				body: '` + test.LookupPlainPOST["data_fail"] + `'
			`),
			errRegex: `non-2XX response code`,
		},
		{
			name: "passing regex",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: 'version: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `\d\.\d\.\d`,
			errRegex:         `^$`,
		},
		{
			name: "url from env",
			env: map[string]string{
				"TEST_LOOKUP__DV_QUERY_ONE": test.LookupPlain["url_valid"],
			},
			overrides: test.TrimYAML(`
				url: ${TEST_LOOKUP__DV_QUERY_ONE}
				regex: 'version: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `\d\.\d\.\d`,
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
			wantVersion:      `\d\.\d\.\d`,
			errRegex:         `^$`,
		},
		{
			name: "passing regex with no capture group",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '[0-9.]+'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      "[0-9.]+",
			errRegex:         `^$`,
		},
		{
			name: "regex with template",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '(stable).*(version).*"([\d.]+).*(and)'
				regex_template: '$2 $1 $4, $3'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      "version stable and, 1.2.1",
			errRegex:         `^$`,
		},
		{
			name: "failing regex",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '^bishBashBosh$'
			`),
			errRegex: `regex .* didn't return any matches on`,
		},
		{
			name: "handle non-semantic (only major) version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: '(\d+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
		},
		{
			name: "want semantic versioning but get non-semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: 'non-semantic: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: true`,
			errRegex:         `failed to convert "[^"]+" to a semantic version`,
		},
		{
			name: "allow non-semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				regex: 'non-semantic: "([^"]+)'
			`),
			optionsOverrides: `semantic_versioning: false`,
			errRegex:         `^$`,
		},
		{
			name: "valid semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupJSON["url_valid"] + `
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
			name: "version from header - pass, exact casing",
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupResponseHeader["url_valid"] + `
				target_header: ` + test.LookupResponseHeader["header_key_pass"] + `
			`),
			wantVersion: `^\d+\.\d+\.\d+$`,
			errRegex:    `^$`,
		},
		{
			name: "version from header - pass, mixed casing",
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupResponseHeader["url_valid"] + `
				target_header: ` + test.LookupResponseHeader["header_key_pass_mixed_case"] + `
			`),
			wantVersion: `^\d+\.\d+\.\d+$`,
			errRegex:    `^$`,
		},
		{
			name: "version from header - fail",
			overrides: test.TrimYAML(`
				method: GET
				url: ` + test.LookupResponseHeader["url_valid"] + `
				target_header: ` + test.LookupResponseHeader["header_key_fail"] + `
			`),
			errRegex: `^target header "[^"]+" not found$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)
			dvl := testLookup(t, false)
			dvl.JSON = ""
			if err := dvl.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup overrides: %s",
					packageName, err,
				)
			}
			if tc.optionsOverrides != "" {
				if err := decode.Unmarshal("yaml", []byte(tc.optionsOverrides), dvl.Options); err != nil {
					t.Fatalf(
						"%s\nfailed to unmarshal Lookup.Options overrides: %s",
						packageName, err,
					)
				}
			}

			// WHEN: Query is called on it.
			err := dvl.Query(true, logx.LogFrom{})

			// THEN: any decode is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nLookup.Query() error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}

			// AND: the version matches the expected regex.
			if tc.wantVersion != "" {
				if version := dvl.Status.DeployedVersion(); !util.RegexCheck(tc.wantVersion, version) {
					t.Errorf(
						"%s\nLookup.Query() .DeployedVersion() mismatch\ngot:  %q\nwant %q",
						packageName, version, tc.wantVersion,
					)
				}
			}
		})
	}
}

func TestLookup_GetVersion(t *testing.T) {
	const (
		jsonBody  = `{"bar":"1.2.2","foo":{"bar":{"version":"3.2.1"}}}`
		plainBody = `version: "1.2.1"
non-semantic: "ver1.2.2"
`
	)

	tests := []struct {
		name                 string
		body                 []byte
		json                 string
		regex, regexTemplate string
		semVer               bool
		wantVersion          string
		errRegex             string
	}{
		{
			name:     "empty body",
			body:     nil,
			semVer:   true,
			errRegex: `^no version found in`,
		},
		{
			name:        "plain body without regex or JSON",
			body:        []byte("1.2.3"),
			semVer:      true,
			wantVersion: "1.2.3",
			errRegex:    `^$`,
		},
		{
			name:     "JSON empty string value",
			body:     []byte(`{"bar":""}`),
			json:     "bar",
			semVer:   true,
			errRegex: `^no version found in.*$`,
		},
		{
			name:        "JSON top-level key",
			body:        []byte(jsonBody),
			json:        "bar",
			semVer:      true,
			wantVersion: "1.2.2",
			errRegex:    `^$`,
		},
		{
			name:        "JSON nested key",
			body:        []byte(jsonBody),
			json:        "foo.bar.version",
			semVer:      true,
			wantVersion: "3.2.1",
			errRegex:    `^$`,
		},
		{
			name:     "JSON invalid body",
			body:     []byte(plainBody),
			json:     "bar",
			semVer:   true,
			errRegex: `^failed to unmarshal response from.*$`,
		},
		{
			name:   "JSON missing key",
			body:   []byte(`{}`),
			json:   "missing",
			semVer: true,
			errRegex: test.TrimYAML(`
				^failed to navigate JSON:
					failed to find value for .*$`,
			),
		},
		{
			name:        "regex with capture group",
			body:        []byte(plainBody),
			regex:       `version: "([^"]+)"`,
			semVer:      true,
			wantVersion: "1.2.1",
			errRegex:    `^$`,
		},
		{
			name:        "regex without capture group",
			body:        []byte(plainBody),
			regex:       `[0-9]+\.[0-9]+\.[0-9]+`,
			semVer:      true,
			wantVersion: "1.2.1",
			errRegex:    `^$`,
		},
		{
			name:     "regex no match",
			body:     []byte(plainBody),
			regex:    `^bishBashBosh$`,
			semVer:   true,
			errRegex: `^regex .* didn't return any matches on.*$`,
		},
		{
			name: "regex with template",
			body: []byte(
				`stable release version info "1.2.1" and more`,
			),
			regex:         `(stable).*(version).*"([\d.]+).*(and)`,
			regexTemplate: `$2 $1 $4, $3`,
			semVer:        false,
			wantVersion:   "version stable and, 1.2.1",
			errRegex:      `^$`,
		},
		{
			name:        "JSON then regex",
			body:        []byte(jsonBody),
			json:        "bar",
			regex:       `^([0-9.]+)$`,
			semVer:      true,
			wantVersion: "1.2.2",
			errRegex:    `^$`,
		},
		{
			name:     "semantic versioning rejects non-semantic version",
			body:     []byte(plainBody),
			regex:    `non-semantic: "([^"]+)"`,
			semVer:   true,
			errRegex: `^failed to convert .* to a semantic version.*$`,
		},
		{
			name:        "semantic versioning disabled allows non-semantic version",
			body:        []byte(plainBody),
			regex:       `non-semantic: "([^"]+)"`,
			semVer:      false,
			wantVersion: "ver1.2.2",
			errRegex:    `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.JSON = tc.json
			lookup.Regex = tc.regex
			lookup.RegexTemplate = tc.regexTemplate
			lookup.Options.SemanticVersioning = &tc.semVer

			version, err := lookup.getVersion(tc.body, logx.LogFrom{})

			prefix := fmt.Sprintf("%s\nLookup.getVersion()", packageName)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if got, want := version, tc.wantVersion; got != want {
				t.Errorf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name      string
		env       map[string]string
		overrides string
		bodyRegex string
		errRegex  string
	}{
		{
			name:      "url - invalid",
			overrides: `url: "https://	test"`,
			errRegex:  `invalid control character in URL`,
		},
		{
			name:      "url - unknown",
			overrides: `url: https://release-argus.invalid-tld`,
			errRegex:  `no such host`,
		},
		{
			name:      "url - valid",
			overrides: `url: ` + test.LookupPlain["url_valid"],
			errRegex:  `^$`,
		},
		{
			name: "url - from env",
			env: map[string]string{
				"TEST_LOOKUP__DV_HTTP_REQUEST_ONE": test.LookupPlain["url_valid"],
			},
			overrides: `url: ${TEST_LOOKUP__DV_HTTP_REQUEST_ONE}`,
			errRegex:  `^$`,
		},
		{
			name: "url - from env partial",
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
			name: "headers - pass",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupWithHeaderAuth["url_valid"] + `
				headers:
					- key: ` + test.LookupWithHeaderAuth["header_key"] + `
						value: ` + test.LookupWithHeaderAuth["header_value_pass"] + `
			`),
			bodyRegex: `^[\d.]+$`,
			errRegex:  `^$`,
		},
		{
			name: "headers - fail",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupWithHeaderAuth["url_valid"] + `
				headers:
					- key: ` + test.LookupWithHeaderAuth["header_key"] + `
						value: ` + test.LookupWithHeaderAuth["header_value_fail"] + `
			`),
			bodyRegex: `Hook rules were not satisfied\.`,
			errRegex:  `^$`,
		},
		{
			name: "basic auth - pass",
			overrides: test.TrimYAML(`
				url: ` + test.LookupBasicAuth["url_valid"] + `
				basic_auth:
					username: ` + test.LookupBasicAuth["username"] + `
					password: ` + test.LookupBasicAuth["password"] + `
			`),
			errRegex: `^$`,
		},
		{
			name: "basic auth - fail",
			overrides: test.TrimYAML(`
				url: ` + test.LookupBasicAuth["url_valid"] + `
				basic_auth:
					username: ` + test.LookupBasicAuth["username"] + `
					password: ` + test.LookupBasicAuth["password"] + `-
			`),
			errRegex: `non-2XX response code: 401`,
		},
		{
			name: "self-signed cert - pass",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_invalid"] + `
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name: "self-signed cert - fail",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_invalid"] + `
				allow_invalid_certs: false
			`),
			errRegex: `x509 \(certificate invalid\)`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)
			lookup := testLookup(t, false)
			// Apply overrides.
			if err := lookup.UnmarshalYAML([]byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup.overrides: %s",
					packageName, err,
				)
			}

			// WHEN: httpRequest is called on it.
			body, err := lookup.httpRequest(logx.LogFrom{})

			prefix := fmt.Sprintf("%s\nLookup.httpRequest()", packageName)

			// THEN: any decode is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the body matches the expected regex.
			if tc.bodyRegex != "" {
				if b := string(body); !util.RegexCheck(tc.bodyRegex, b) {
					t.Errorf(
						"%s body mismatch\ngot:  %q\nwant: %q",
						prefix, b, tc.bodyRegex,
					)
				}
			}
		})
	}
}
