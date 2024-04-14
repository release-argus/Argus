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

package deployedver

import (
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	dbtype "github.com/release-argus/Argus/db/types"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
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
			errRegex: "invalid control character in URL"},
		"unknown url": {
			url:      "https://release-argus.invalid-tld",
			errRegex: "no such host"},
		"valid url": {
			url:      "https://release-argus.io",
			errRegex: "^$"},
		"url from env": {
			env:      map[string]string{"TESTLOOKUP_DV_HTTPREQUEST_ONE": "https://release-argus.io"},
			url:      "${TESTLOOKUP_DV_HTTPREQUEST_ONE}",
			errRegex: "^$"},
		"url from env partial": {
			env:      map[string]string{"TESTLOOKUP_DV_HTTPREQUEST_TWO": "release-argus"},
			url:      "https://${TESTLOOKUP_DV_HTTPREQUEST_TWO}.io",
			errRegex: "^$"},
		"404": {
			errRegex: "non-2XX response code: 404",
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
	// GIVEN a Lookup()
	tests := map[string]struct {
		env                  map[string]string
		url                  string
		allowInvalidCerts    bool
		noSemanticVersioning bool
		basicAuth            *BasicAuth
		headers              []Header
		json                 string
		regex                string
		regexTemplate        *string
		errRegex             string
		wantVersion          string
	}{
		"JSON lookup that doesn't resolve": {
			errRegex: `failed to find value for \"[^"]+\" in `,
			url:      "https://api.github.com/repos/release-argus/argus/releases/latest",
			json:     "something",
		},
		"URL that doesn't resolve to JSON": {
			errRegex: "failed to unmarshal",
			url:      "https://release-argus.io",
			json:     "something",
		},
		"passing regex": {
			noSemanticVersioning: true,
			wantVersion:          "[0-9]{4}",
			errRegex:             "^$",
			url:                  "https://release-argus.io",
			regex:                `([0-9]+)\s<[^>]+>The Argus Developers`,
		},
		"url from env": {
			env:                  map[string]string{"TESTLOOKUP_DV_QUERY_ONE": "https://release-argus.io"},
			noSemanticVersioning: true,
			wantVersion:          "[0-9]{4}",
			errRegex:             "^$",
			url:                  "${TESTLOOKUP_DV_QUERY_ONE}",
			regex:                `([0-9]+)\s<[^>]+>The Argus Developers`,
		},
		"url from env partial": {
			env:                  map[string]string{"TESTLOOKUP_DV_QUERY_TWO": "release-argus"},
			noSemanticVersioning: true,
			wantVersion:          "[0-9]{4}",
			errRegex:             "^$",
			url:                  "https://${TESTLOOKUP_DV_QUERY_TWO}.io",
			regex:                `([0-9]+)\s<[^>]+>The Argus Developers`,
		},
		"passing regex with no capture group": {
			noSemanticVersioning: true,
			wantVersion:          "[0-9]{4}",
			errRegex:             "^$",
			url:                  "https://release-argus.io",
			regex:                "[0-9]{4}",
		},
		"regex with template": {
			noSemanticVersioning: true,
			errRegex:             "^$",
			url:                  "https://release-argus.io",
			regex:                `([0-9]+)\s<[^>]+>(The) (Argus) (Developers)`,
			regexTemplate:        stringPtr("$2 $1 $4, $3"),
			wantVersion:          "The [0-9]+ Developers, Argus",
		},
		"failing regex": {
			errRegex: "regex .* didn't find a match on",
			url:      "https://release-argus.io",
			regex:    "^bishbashbosh$",
		},
		"handle non-semantic (only major) version": {
			noSemanticVersioning: false,
			url:                  "https://release-argus.io",
			regex:                `([0-9]+)\s<[^>]+>The Argus Developers`,
		},
		"want semantic versioning but get non-semantic version": {
			noSemanticVersioning: false,
			errRegex:             "failed converting .* to a semantic version",
			url:                  "https://release-argus.io",
			regex:                `([0-9]+\s)<[^>]+>The Argus Developers`,
		},
		"allow non-semantic versioning and get non-semantic version": {
			noSemanticVersioning: true,
			errRegex:             "^$",
			url:                  "https://release-argus.io",
			regex:                `([0-9]+\s)<[^>]+>The Argus Developers`,
		},
		"valid semantic version": {
			errRegex:    "^$",
			wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`,
			regex:       "argus-([0-9.]+)\\.",
			url:         "https://release-argus.io/docs/getting-started/",
		},
		"headers fail": {
			errRegex: "non-2XX response code: 401",
			headers: []Header{
				{Key: "Authorization", Value: "token ghp_FAIL"}},
			url:  "https://api.github.com/repos/release-argus/argus/releases/latest",
			json: "something",
		},
		"404": {
			errRegex: "non-2XX response code: 404",
			url:      "https://release-argus.io/foo/bar",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			dvl := testLookup()
			dvl.URL = tc.url
			dvl.AllowInvalidCerts = &tc.allowInvalidCerts
			dvl.BasicAuth = tc.basicAuth
			dvl.Headers = tc.headers
			dvl.JSON = tc.json
			dvl.Regex = tc.regex
			dvl.RegexTemplate = tc.regexTemplate
			*dvl.Options.SemanticVersioning = !tc.noSemanticVersioning

			// WHEN Query is called on it
			version, err := dvl.Query(true, &util.LogFrom{})

			// THEN any err is expected
			if tc.wantVersion != "" {
				re := regexp.MustCompile(tc.wantVersion)
				match := re.MatchString(version)
				if !match {
					t.Errorf("want version=%q\ngot  version=%q",
						tc.wantVersion, version)
				}
			}
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestLookup_Track(t *testing.T) {
	// GIVEN a Lookup()
	tests := map[string]struct {
		env                  map[string]string
		lookup               *Lookup
		allowInvalidCerts    bool
		semanticVersioning   bool
		basicAuth            *BasicAuth
		expectFinish         bool
		wait                 time.Duration
		errRegex             string
		startDeployedVersion string
		wantDeployedVersion  string
		startLatestVersion   string
		wantLatestVersion    string
		wantAnnounces        int
		wantDatabaseMesages  int
		deleting             bool
	}{
		"nil Lookup exits immediately": {
			lookup:       nil,
			wait:         10 * time.Millisecond,
			expectFinish: true,
		},
		"get semantic version with regex": {
			startLatestVersion:  "1.2.2",
			wantDeployedVersion: "1.2.2",
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "v([^"]+)`},
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"get semantic version from json": {
			startLatestVersion:  "1.2.2",
			wantDeployedVersion: "1.2.2",
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:  true,
			wantDatabaseMesages: 1, wantAnnounces: 1,
		},
		"get semantic version from multi-level json": {
			startLatestVersion:  "3.2.1",
			wantDeployedVersion: "3.2.1",
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "foo.bar.version"},
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"reject non-semantic versions": {
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: ("[^"]+)`},
			semanticVersioning:  true,
			wantDatabaseMesages: 0,
			wantAnnounces:       0,
		},
		"allow non-semantic version": {
			startLatestVersion:  "v1.2.2",
			wantDeployedVersion: "v1.2.2",
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "([^"]+)`},
			semanticVersioning:  false,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"allow non-semantic version (leading v's)": {
			startLatestVersion:  "v1.2.2",
			wantDeployedVersion: "v1.2.2",
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "([^"]+)`},
			semanticVersioning:  false,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"get version behind basic auth": {
			startLatestVersion:  "1.2.2",
			wantDeployedVersion: "1.2.2",
			basicAuth: &BasicAuth{
				Username: "test",
				Password: "123"},
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/basic-auth",
				Regex: `non-semantic: "v([^"]+)`},
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"env vars in basic auth": {
			env: map[string]string{
				"TESTLOOKUP_DV_TRACK_ONE": "tes",
				"TESTLOOKUP_DV_TRACK_TWO": "23"},
			startLatestVersion:  "1.2.2",
			wantDeployedVersion: "1.2.2",
			basicAuth: &BasicAuth{
				Username: "${TESTLOOKUP_DV_TRACK_ONE}t",
				Password: "1${TESTLOOKUP_DV_TRACK_TWO}"},
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/basic-auth",
				Regex: `non-semantic: "v([^"]+)`},
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"get version behind an invalid cert": {
			startLatestVersion:  "1.2.2",
			wantDeployedVersion: "1.2.2",
			lookup: &Lookup{
				URL:   "https://invalid.release-argus.io/plain",
				Regex: `non-semantic: "v([^"]+)`},
			allowInvalidCerts:   true,
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"fail due to an unallowed invalid cert": {
			startLatestVersion:  "",
			wantDeployedVersion: "",
			lookup: &Lookup{
				URL:   "https://invalid.release-argus.io/plain",
				Regex: `non-semantic: "v([^"]+)`},
			allowInvalidCerts:   false,
			semanticVersioning:  true,
			wantDatabaseMesages: 0,
			wantAnnounces:       0,
		},
		"update from an older version": {
			startLatestVersion:   "1.2.2",
			startDeployedVersion: "1.2.1",
			wantDeployedVersion:  "1.2.2",
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "v([^"]+)`},
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"update from a newer version": {
			startLatestVersion:   "1.2.2",
			startDeployedVersion: "1.2.3",
			wantDeployedVersion:  "1.2.2",
			lookup: &Lookup{
				URL:   "https://valid.release-argus.io/plain",
				Regex: `non-semantic: "v([^"]+)`},
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"get a newer deployed version than latest version updates both": {
			startLatestVersion:  "1.2.1",
			wantLatestVersion:   "1.2.2",
			wantDeployedVersion: "1.2.2",
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:  true,
			wantDatabaseMesages: 2,
			wantAnnounces:       2,
		},
		"get an older deployed version than latest version only updates deployed": {
			startLatestVersion:  "1.2.3",
			wantLatestVersion:   "1.2.3",
			wantDeployedVersion: "1.2.2",
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:  true,
			wantDatabaseMesages: 1,
			wantAnnounces:       1,
		},
		"get a deployed version with no latest version updates both": {
			startLatestVersion:  "",
			wantLatestVersion:   "1.2.2",
			wantDeployedVersion: "1.2.2",
			lookup: &Lookup{
				URL:  "https://valid.release-argus.io/json",
				JSON: "bar"},
			semanticVersioning:  true,
			wantDatabaseMesages: 2,
			wantAnnounces:       2,
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
			wantDatabaseMesages:  0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - cannot run in parallel because of stdout

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() {
				os.Stdout = stdout
			}()
			if tc.lookup != nil {
				tc.lookup.AllowInvalidCerts = boolPtr(tc.allowInvalidCerts)
				tc.lookup.BasicAuth = tc.basicAuth
				tc.lookup.Defaults = &LookupDefaults{}
				tc.lookup.HardDefaults = &LookupDefaults{}
				tc.lookup.Options = opt.New(
					nil, "2s", &tc.semanticVersioning,
					&opt.OptionsDefaults{}, &opt.OptionsDefaults{})
				dbChannel := make(chan dbtype.Message, 4)
				announceChannel := make(chan []byte, 4)
				svcStatus := svcstatus.New(
					&announceChannel, &dbChannel, nil,
					"", "", "", "", "", "")
				tc.lookup.Status = svcStatus
				tc.lookup.Status.ServiceID = stringPtr(name)
				tc.lookup.Status.WebURL = &tc.lookup.URL
				if tc.deleting {
					tc.lookup.Status.SetDeleting()
				}
				tc.lookup.Status.SetDeployedVersion(tc.startDeployedVersion, false)
				tc.lookup.Status.SetLatestVersion(tc.startLatestVersion, false)

				metric.InitPrometheusCounter(metric.DeployedVersionQueryMetric,
					*tc.lookup.Status.ServiceID,
					"",
					"",
					"SUCCESS")
				metric.InitPrometheusCounter(metric.DeployedVersionQueryMetric,
					*tc.lookup.Status.ServiceID,
					"",
					"",
					"FAIL")
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
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			t.Log(string(out))
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
			if tc.wantDatabaseMesages != len(*tc.lookup.Status.DatabaseChannel) {
				t.Errorf("expected DatabaseChannel to have %d messages in queue, not %d",
					tc.wantDatabaseMesages, len(*tc.lookup.Status.DatabaseChannel))
			}
			tc.lookup.Status.SetDeleting()
		})
	}
}
