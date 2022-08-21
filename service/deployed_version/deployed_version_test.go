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

package deployed_version

import (
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

func TestGetAllowInvalidCerts(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		allowInvalidCertsRoot        *bool
		allowInvalidCertsDefault     *bool
		allowInvalidCertsHardDefault *bool
		wantBool                     bool
	}{
		"root overrides all": {wantBool: true, allowInvalidCertsRoot: boolPtr(true),
			allowInvalidCertsDefault: boolPtr(false), allowInvalidCertsHardDefault: boolPtr(false)},
		"default overrides hardDefault": {wantBool: true, allowInvalidCertsRoot: nil,
			allowInvalidCertsDefault: boolPtr(true), allowInvalidCertsHardDefault: boolPtr(false)},
		"hardDefault is last resort": {wantBool: true, allowInvalidCertsRoot: nil, allowInvalidCertsDefault: nil,
			allowInvalidCertsHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testDeployedVersion()
			lookup.AllowInvalidCerts = tc.allowInvalidCertsRoot
			lookup.Defaults.AllowInvalidCerts = tc.allowInvalidCertsDefault
			lookup.HardDefaults.AllowInvalidCerts = tc.allowInvalidCertsHardDefault

			// WHEN GetAllowInvalidCerts is called
			got := lookup.GetAllowInvalidCerts()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("want: %t\ngot:  %t",
					tc.wantBool, got)
			}
		})
	}
}

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
			lookup := testDeployedVersion()
			lookup.URL = tc.url

			// WHEN httpRequest is called on it
			_, err := lookup.httpRequest(utils.LogFrom{})

			// THEN any err is expected
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	// GIVEN a Lookup
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		url                  string
		allowInvalidCerts    bool
		noSemanticVersioning bool
		basicAuth            *BasicAuth
		headers              []Header
		json                 string
		regex                string
		errRegex             string
		wantVersion          string
	}{
		"JSON lookup that doesn't resolve": {errRegex: "could not be found in the following JSON",
			url: "https://api.github.com/repos/release-argus/argus/releases/latest", json: "something"},
		"URL that doesn't resolve to JSON":    {errRegex: "failed to unmarshal", url: "https://release-argus.io", json: "something"},
		"passing regex":                       {noSemanticVersioning: true, wantVersion: "[0-9]{4}", errRegex: "^$", url: "https://release-argus.io", regex: "([0-9]+) The Argus Developers"},
		"passing regex with no capture group": {noSemanticVersioning: true, wantVersion: "[0-9]{4}", errRegex: "^$", url: "https://release-argus.io", regex: "[0-9]{4}"},
		"failing regex":                       {errRegex: "regex .* didn't return any matches on", url: "https://release-argus.io", regex: "^bishbashbosh$"},
		"want semantic versioning but get non-semantic version": {noSemanticVersioning: false, errRegex: "failed converting .* to a semantic version", url: "https://release-argus.io",
			regex: "([0-9]+) The Argus Developers"},
		"allow non-semantic versioning and get non-semantic version": {noSemanticVersioning: true, errRegex: "^$", url: "https://release-argus.io",
			regex: "([0-9]+) The Argus Developers"},
		"valid semantic version": {errRegex: "^$", wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`, regex: "argus-([0-9.]+)\\.",
			url: "https://release-argus.io/docs/getting-started/"},
		"headers fail": {errRegex: "Bad credentials", headers: []Header{{Key: "Authorization", Value: "token ghp_FAIL"}},
			url: "https://api.github.com/repos/release-argus/argus/releases/latest", json: "something"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dvl := testDeployedVersion()
			dvl.URL = tc.url
			dvl.AllowInvalidCerts = &tc.allowInvalidCerts
			dvl.BasicAuth = tc.basicAuth
			dvl.Headers = tc.headers
			dvl.JSON = tc.json
			dvl.Regex = tc.regex
			*dvl.Options.SemanticVersioning = !tc.noSemanticVersioning

			// WHEN Query is called on it
			version, err := dvl.Query(utils.LogFrom{})

			// THEN any err is expected
			if tc.wantVersion != "" {
				re := regexp.MustCompile(tc.wantVersion)
				match := re.MatchString(version)
				if !match {
					t.Errorf("want version=%q\ngot  version=%q",
						tc.wantVersion, version)
				}
			}
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestTrack(t *testing.T) {
	// GIVEN a Lookup
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
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
	}{
		"nil Lookup exits immediately": {lookup: nil, wait: 10 * time.Millisecond, expectFinish: true},
		"get semantic version with regex": {startLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/plain", Regex: `non-semantic: "v([^"]+)`}, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"get semantic version from json": {startLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/json", JSON: "bar"}, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"get semantic version from multi-level json": {startLatestVersion: "3.2.1", wantDeployedVersion: "3.2.1",
			lookup: &Lookup{URL: "https://valid.release-argus.io/json", JSON: "foo.bar.version"}, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"reject non-semantic versions": {wantDeployedVersion: "",
			lookup: &Lookup{URL: "https://valid.release-argus.io/plain", Regex: `non-semantic: "([^"]+)`}, semanticVersioning: true, wantDatabaseMesages: 0, wantAnnounces: 0},
		"allow non-semantic version": {startLatestVersion: "v1.2.2", wantDeployedVersion: "v1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/plain", Regex: `non-semantic: "([^"]+)`}, semanticVersioning: false, wantDatabaseMesages: 1, wantAnnounces: 1},
		"get version behind basic auth": {startLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2", basicAuth: &BasicAuth{Username: "test", Password: "123"},
			lookup: &Lookup{URL: "https://valid.release-argus.io/basic-auth", Regex: `non-semantic: "v([^"]+)`}, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"get version behind an invalid cert": {startLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://invalid.release-argus.io/plain", Regex: `non-semantic: "v([^"]+)`}, allowInvalidCerts: true, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"fail due to an unallowed invalid cert": {startLatestVersion: "", wantDeployedVersion: "",
			lookup: &Lookup{URL: "https://invalid.release-argus.io/plain", Regex: `non-semantic: "v([^"]+)`}, allowInvalidCerts: false, semanticVersioning: true, wantDatabaseMesages: 0, wantAnnounces: 0},
		"update from an older version": {startLatestVersion: "1.2.2", startDeployedVersion: "1.2.1", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/plain", Regex: `non-semantic: "v([^"]+)`}, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"update from a newer version": {startLatestVersion: "1.2.2", startDeployedVersion: "1.2.3", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/plain", Regex: `non-semantic: "v([^"]+)`}, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"get a newer deployed version than latest version updates both": {startLatestVersion: "1.2.1", wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/json", JSON: "bar"}, semanticVersioning: true, wantDatabaseMesages: 2, wantAnnounces: 2},
		"get a older deployed version than latest version only updates deployed": {startLatestVersion: "1.2.3", wantLatestVersion: "1.2.3", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/json", JSON: "bar"}, semanticVersioning: true, wantDatabaseMesages: 1, wantAnnounces: 1},
		"get a deployed version with no latest version updates both": {startLatestVersion: "", wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			lookup: &Lookup{URL: "https://valid.release-argus.io/json", JSON: "bar"}, semanticVersioning: true, wantDatabaseMesages: 2, wantAnnounces: 2},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			if tc.lookup != nil {
				defaults := &Lookup{}
				dbChannel := make(chan db_types.Message, 4)
				announceChannel := make(chan []byte, 4)
				webURL := &tc.lookup.URL
				tc.lookup.AllowInvalidCerts = boolPtr(tc.allowInvalidCerts)
				tc.lookup.BasicAuth = tc.basicAuth
				tc.lookup.Defaults = defaults
				tc.lookup.HardDefaults = defaults
				tc.lookup.Options = &options.Options{
					Interval:           "2s",
					SemanticVersioning: boolPtr(tc.semanticVersioning),
					Defaults:           &options.Options{},
					HardDefaults:       &options.Options{},
				}
				tc.lookup.Status = &service_status.Status{
					ServiceID:       stringPtr(name),
					DeployedVersion: tc.startDeployedVersion,
					LatestVersion:   tc.startLatestVersion,
					AnnounceChannel: &announceChannel,
					DatabaseChannel: &dbChannel,
					WebURL:          webURL,
				}
				metrics.InitPrometheusCounterWithIDAndResult(metrics.DeployedVersionQueryMetric, *tc.lookup.Status.ServiceID, "SUCCESS")
				metrics.InitPrometheusCounterWithIDAndResult(metrics.DeployedVersionQueryMetric, *tc.lookup.Status.ServiceID, "FAIL")
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
				passQ := testutil.ToFloat64(metrics.DeployedVersionQueryMetric.WithLabelValues(*tc.lookup.Status.ServiceID, "SUCCESS"))
				failQ := testutil.ToFloat64(metrics.DeployedVersionQueryMetric.WithLabelValues(*tc.lookup.Status.ServiceID, "FAIL"))
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
			if tc.wantDeployedVersion != tc.lookup.Status.DeployedVersion {
				t.Errorf("expected DeployedVersion to be %q after query, not %q",
					tc.wantDeployedVersion, tc.lookup.Status.DeployedVersion)
			}
			if tc.wantLatestVersion == "" {
				tc.wantLatestVersion = tc.wantDeployedVersion
			}
			if tc.wantLatestVersion != tc.lookup.Status.LatestVersion {
				t.Errorf("expected LatestVersion to be %q after query, not %q",
					tc.wantLatestVersion, tc.lookup.Status.LatestVersion)
			}
			if tc.wantAnnounces != len(*tc.lookup.Status.AnnounceChannel) {
				t.Errorf("expected AnnounceChannel to have %d messages in queue, not %d",
					tc.wantAnnounces, len(*tc.lookup.Status.AnnounceChannel))
			}
			if tc.wantDatabaseMesages != len(*tc.lookup.Status.DatabaseChannel) {
				t.Errorf("expected DatabaseChannel to have %d messages in queue, not %d",
					tc.wantDatabaseMesages, len(*tc.lookup.Status.DatabaseChannel))
			}
		})
	}
}
