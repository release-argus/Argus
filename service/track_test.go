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

//go:build integration

package service

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
	"github.com/release-argus/Argus/webhook"
)

func TestSliceTrack(t *testing.T) {
	// GIVEN a Slice
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		ordering     []string
		slice        []string
		active       []bool
		expectFinish bool
	}{
		"empty Ordering does no queries": {ordering: []string{}, slice: []string{"github", "url"}},
		"only tracks active Services":    {ordering: []string{"github", "url"}, slice: []string{"github", "url"}, active: []bool{false, true}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var slice *Slice
			if len(tc.slice) != 0 {
				slice = &Slice{}
				i := 0
				for _, j := range tc.slice {
					switch j {
					case "github":
						(*slice)[j] = testServiceGitHub()
					case "url":
						(*slice)[j] = testServiceURL()
					}
					if len(tc.active) != 0 {
						(*slice)[j].Options.Active = boolPtr(tc.active[i])
					}
					(*slice)[j].Status.LatestVersion = ""
					(*slice)[j].Status.DeployedVersion = ""
					i++
				}
			}

			// WHEN Track is called on it
			slice.Track(&tc.ordering)

			// THEN the function exits straight away
			time.Sleep(time.Second)
			for i := range *slice {
				if !utils.Contains(tc.ordering, i) {
					if (*slice)[i].Status.LatestVersion != "" {
						t.Fatalf("didn't expect Query to have done anything for %s as it's not in the ordering %v\n%#v",
							i, tc.ordering, (*slice)[i].Status)
					}
				} else if utils.EvalNilPtr((*slice)[i].Options.Active, true) {
					if (*slice)[i].Status.LatestVersion == "" {
						t.Fatalf("expected Query to have found a LatestVersion\n%#v",
							(*slice)[i].Status)
					}
				} else if (*slice)[i].Status.LatestVersion != "" {
					t.Fatalf("didn't expect Query to have done anything for %s\n%#v",
						i, (*slice)[i].Status)
				}
			}
		})
	}
}

func TestServiceTrack(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		url                  string
		urlRegex             string
		allowInvalidCerts    bool
		keepDeployedLookup   bool
		deployedVersionJSON  string
		nilRequire           bool
		autoApprove          bool
		webhook              *webhook.WebHook
		expectFinish         bool
		livenessMetric       int
		ignoreLivenessMetric bool
		wait                 time.Duration
		startLatestVersion   string
		startDeployedVersion string
		wantLatestVersion    string
		wantDeployedVersion  string
		wantAnnounces        int
		wantDatabaseMesages  int
	}{
		"first query updates LatestVersion and DeployedVersion": {livenessMetric: 1,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2", // db: 1 for deployed, 1 for latest
			wantAnnounces: 1, wantDatabaseMesages: 2}, // announce: 1 for latest query
		"query finds a newer version and updates LatestVersion and DeployedVersion as no commands/webhooks": {urlRegex: "v([0-9.]+)", livenessMetric: 1,
			startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2", // db: 1 for latest, 1 for deployed
			wantAnnounces: 1, wantDatabaseMesages: 2}, // announce: 1 for latest query
		"query finds a newer version and updates LatestVersion and not DeployedVersion": {urlRegex: "v([0-9.]+)", livenessMetric: 1,
			webhook:            testWebHook(false),
			startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.1", // db: 1 for latest
			wantAnnounces: 1, wantDatabaseMesages: 1}, // announce: 1 for latest query
		"query finds a newer version does send webhooks if autoApprove enabled": {urlRegex: "v([0-9.]+)", livenessMetric: 1,
			webhook: testWebHook(false), autoApprove: true,
			startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2", // db: 1 for latest, 1 for deployed
			wantAnnounces: 2, wantDatabaseMesages: 2}, // announce: 1 for latest query, 1 for deployed
		"query doesn't update versions if it finds one that's older semantically": {urlRegex: `"([0-9.]+)"`, livenessMetric: 4, nilRequire: true,
			startLatestVersion: "1.2.2", startDeployedVersion: "1.2.2",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces: 0, wantDatabaseMesages: 0},
		"track on invalid cert allowed": {urlRegex: `"([0-9.]+)"`, livenessMetric: 1, nilRequire: true,
			url: "https://invalid.release-argus.io/plain", allowInvalidCerts: true,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "1.2.1", wantDeployedVersion: "1.2.1", // db: 1 for deployed, 1 for latest
			wantAnnounces: 1, wantDatabaseMesages: 2}, // announce: 1 for latest query
		"url regex fail": {urlRegex: "v[0-9.]foo", livenessMetric: 2,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0},
		"non-semantic version fail": {urlRegex: "v[0-9.]+", livenessMetric: 3, nilRequire: true,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0},
		"progressive version fail (got older version)": {urlRegex: "v([0-9.]+)", livenessMetric: 4,
			startLatestVersion: "1.2.3", startDeployedVersion: "1.2.3",
			wantLatestVersion: "1.2.3", wantDeployedVersion: "1.2.3",
			wantAnnounces: 0, wantDatabaseMesages: 0},
		"other fail (invalid cert)": {urlRegex: `"([0-9.]+)"`, livenessMetric: 0, nilRequire: true,
			url: "https://invalid.release-argus.io/plain", allowInvalidCerts: false,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0},
		"track gets DeployedVersion": {keepDeployedLookup: true, deployedVersionJSON: "bar", livenessMetric: 1,
			startLatestVersion: "1.2.2", startDeployedVersion: "1.2.0",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2", // db: 1 for deployed change
			wantAnnounces: 2, wantDatabaseMesages: 1}, // announce: 1 for latest query, 1 for deployed change
		"track gets DeployedVersion that's newer updates LatestVersion too": {keepDeployedLookup: true, deployedVersionJSON: "foo.bar.version", ignoreLivenessMetric: true, // ignore as deployed lookup may be done before
			startLatestVersion: "1.2.2", startDeployedVersion: "0.0.0",
			wantLatestVersion: "3.2.1", wantDeployedVersion: "3.2.1", // db: 1 for latest change, 1 for deployed change
			wantAnnounces: 3, wantDatabaseMesages: 2}, // announce: 1 for latest query (as <latestVersion), 1 for latest change, 1 for deployed change
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			didFinish := make(chan bool, 1)
			service := testServiceURL()
			service.ID = name
			service.Status.LatestVersion = tc.startLatestVersion
			service.Status.DeployedVersion = tc.startDeployedVersion
			if tc.keepDeployedLookup {
				service.DeployedVersionLookup.JSON = tc.deployedVersionJSON
			} else {
				service.DeployedVersionLookup = nil
			}
			if tc.nilRequire {
				service.LatestVersion.Require = nil
			}
			if tc.urlRegex != "" {
				service.LatestVersion.URLCommands[0].Regex = &tc.urlRegex
			}
			if tc.url != "" {
				service.LatestVersion.URL = tc.url
			}
			if tc.webhook != nil {
				service.WebHook = make(webhook.Slice, 1)
				service.WebHook["test"] = tc.webhook
			}
			*service.Dashboard.AutoApprove = tc.autoApprove
			service.LatestVersion.AllowInvalidCerts = &tc.allowInvalidCerts
			service.Init(jLog, service.Defaults, service.HardDefaults, &shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{}, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{})

			// WHEN CheckValues is called on it
			go func() {
				service.Track()
				didFinish <- true
			}()
			haveQueried := false
			for haveQueried != false {
				passQ := testutil.ToFloat64(metrics.LatestVersionQueryMetric.WithLabelValues(service.ID, "SUCCESS"))
				failQ := testutil.ToFloat64(metrics.LatestVersionQueryMetric.WithLabelValues(service.ID, "FAIL"))
				if passQ != float64(0) && failQ != float64(0) {
					haveQueried = true
					if tc.keepDeployedLookup {
						passQ := testutil.ToFloat64(metrics.LatestVersionQueryMetric.WithLabelValues(service.ID, "SUCCESS"))
						failQ := testutil.ToFloat64(metrics.LatestVersionQueryMetric.WithLabelValues(service.ID, "FAIL"))
						// if deployedVersionLookup hasn't queried, reset haveQueried
						if passQ == float64(0) && failQ == float64(0) {
							haveQueried = false
						}
					}
				}
				time.Sleep(time.Second)
			}
			time.Sleep(5 * time.Second)

			// THEN the scrape updates the Status correctly
			if tc.wantLatestVersion != service.Status.LatestVersion || tc.wantDeployedVersion != service.Status.DeployedVersion {
				t.Fatalf("LatestVersion, want %q, got %q\nDeployedVersion, want %q, got %q\n",
					tc.wantLatestVersion, service.Status.LatestVersion, tc.wantDeployedVersion, service.Status.DeployedVersion)
			}
			// LatestVersionQueryMetric
			gotMetric := testutil.ToFloat64(metrics.LatestVersionQueryLiveness.WithLabelValues(service.ID))
			if !tc.ignoreLivenessMetric && gotMetric != float64(tc.livenessMetric) {
				t.Errorf("LatestVersionQueryLiveness should be %d, not %f",
					tc.livenessMetric, gotMetric)
			}
			// AnnounceChannel
			if tc.wantAnnounces != len(*service.Status.AnnounceChannel) {
				t.Errorf("expected AnnounceChannel to have %d messages in queue, not %d",
					tc.wantAnnounces, len(*service.Status.AnnounceChannel))
			}
			// DatabaseChannel
			if tc.wantDatabaseMesages != len(*service.Status.DatabaseChannel) {
				t.Errorf("expected DatabaseChannel to have %d messages in queue, not %d",
					tc.wantDatabaseMesages, len(*service.Status.DatabaseChannel))
			}
			// Track should never finish
			if len(didFinish) != 0 {
				t.Fatal("didn't expect Track to finish")
			}
		})
	}
}
