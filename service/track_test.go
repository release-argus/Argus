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
	"encoding/json"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
	metric "github.com/release-argus/Argus/web/metrics"
	"github.com/release-argus/Argus/webhook"
)

func TestSlice_Track(t *testing.T) {
	// GIVEN a Slice
	jLog = util.NewJLog("WARN", false)
	tests := map[string]struct {
		ordering     []string
		slice        []string
		active       []bool
		expectFinish bool
	}{
		"empty Ordering does no queries": {
			ordering: []string{},
			slice:    []string{"github", "url"}},
		"only tracks active Services": {
			ordering: []string{"github", "url"},
			slice:    []string{"github", "url"},
			active:   []bool{false, true}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var slice *Slice
			if len(tc.slice) != 0 {
				slice = &Slice{}
				i := 0
				for _, j := range tc.slice {
					switch j {
					case "github":
						(*slice)[j] = testServiceGitHub(name)
					case "url":
						(*slice)[j] = testServiceURL(name)
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
				if !util.Contains(tc.ordering, i) {
					if (*slice)[i].Status.LatestVersion != "" {
						t.Fatalf("didn't expect Query to have done anything for %s as it's not in the ordering %v\n%#v",
							i, tc.ordering, (*slice)[i].Status)
					}
				} else if (*slice)[i].Options.GetActive() {
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

func TestService_Track(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		active                     *bool
		url                        string
		urlRegex                   string
		signedCerts                bool
		latestVersionTimestampIn   string
		keepDeployedLookup         bool
		deployedVersionJSON        string
		deployedVersionTimestampIn string
		nilRequire                 bool
		autoApprove                bool
		webhook                    *webhook.WebHook
		expectFinish               bool
		livenessMetric             int
		ignoreLivenessMetric       bool
		takesAtLeast               time.Duration
		startLatestVersion         string
		startDeployedVersion       string
		wantLatestVersion          string
		wantDeployedVersion        string
		wantAnnounces              int
		wantDatabaseMesages        int
		deleting                   bool
	}{
		"first query updates LatestVersion and DeployedVersion": {
			livenessMetric:     1,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces:       1, // announce: 1 for latest query
			wantDatabaseMesages: 2, // db: 1 for deployed, 1 for latest
		},
		"first query updates LatestVersion and DeployedVersion with active true": {
			livenessMetric: 1, active: boolPtr(true),
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces:       1, // announce: 1 for latest query
			wantDatabaseMesages: 2, // db: 1 for deployed, 1 for latest
		},
		"query finds a newer version and updates LatestVersion and DeployedVersion as no commands/webhooks": {
			urlRegex: "v([0-9.]+)", livenessMetric: 1,
			startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces:       1, // announce: 1 for latest query
			wantDatabaseMesages: 2, // db: 1 for latest, 1 for deployed
		},
		"query finds a newer version and updates LatestVersion and not DeployedVersion": {
			urlRegex: "v([0-9.]+)", livenessMetric: 1,
			webhook:            testWebHook(false),
			startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.1",
			wantAnnounces:       1, // announce: 1 for latest query
			wantDatabaseMesages: 1, // db: 1 for latest
		},
		"query finds a newer version does send webhooks if autoApprove enabled": {
			urlRegex: "v([0-9.]+)", livenessMetric: 1,
			webhook: testWebHook(false), autoApprove: true,
			startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces:       2, // announce: 1 for latest query, 1 for deployed
			wantDatabaseMesages: 2, // db: 1 for latest, 1 for deployed
		},
		"query doesn't update versions if it finds one that's older semantically": {
			urlRegex: `"([0-9.]+)"`, livenessMetric: 4, nilRequire: true,
			startLatestVersion: "1.2.2", startDeployedVersion: "1.2.2",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces: 0, wantDatabaseMesages: 0,
		},
		"track on invalid cert allowed": {
			urlRegex: `"([0-9.]+)"`, livenessMetric: 1, nilRequire: true,
			url: "https://invalid.release-argus.io/plain", signedCerts: false,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "1.2.1", wantDeployedVersion: "1.2.1",
			wantAnnounces:       1, // announce: 1 for latest query
			wantDatabaseMesages: 2, // db: 1 for deployed, 1 for latest
		},
		"track on signed cert allowed": {
			urlRegex: `"([0-9.]+)"`, livenessMetric: 1, nilRequire: true,
			url: "https://valid.release-argus.io/plain", signedCerts: true,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "1.2.1", wantDeployedVersion: "1.2.1",
			wantAnnounces:       1, // announce: 1 for latest query
			wantDatabaseMesages: 2, // db: 1 for deployed, 1 for latest
		},
		"url regex fail": {
			urlRegex: "v[0-9.]foo", livenessMetric: 2,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0,
		},
		"non-semantic version fail": {
			urlRegex: "v[0-9.]+", livenessMetric: 3, nilRequire: true,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0,
		},
		"progressive version fail (got older version)": {
			urlRegex: "v([0-9.]+)", livenessMetric: 4,
			startLatestVersion: "1.2.3", startDeployedVersion: "1.2.3",
			wantLatestVersion: "1.2.3", wantDeployedVersion: "1.2.3",
			wantAnnounces: 0, wantDatabaseMesages: 0,
		},
		"other fail (invalid cert)": {
			urlRegex: `"([0-9.]+)"`, livenessMetric: 0, nilRequire: true,
			url: "https://invalid.release-argus.io/plain", signedCerts: true,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0,
		},
		"track gets DeployedVersion": {
			keepDeployedLookup: true, deployedVersionJSON: "bar", livenessMetric: 1,
			startLatestVersion: "1.2.2", startDeployedVersion: "1.2.0",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces:       2, // announce: 1 for latest query, 1 for deployed change
			wantDatabaseMesages: 1, // db: 1 for deployed change
		},
		"track gets DeployedVersion that's newer updates LatestVersion too": {
			keepDeployedLookup: true, deployedVersionJSON: "foo.bar.version",
			ignoreLivenessMetric: true, // ignore as deployed lookup may be done before
			startLatestVersion:   "1.2.2", startDeployedVersion: "0.0.0",
			wantLatestVersion: "3.2.1", wantDeployedVersion: "3.2.1",
			wantAnnounces: 3, // announce: 1 for latest query (as it'll give <latestVersion, but be called before we have deployedVersion),
			// 1 for latest change,
			// 1 for deployed change
			wantDatabaseMesages: 2, // db: 1 for latest change, 1 for deployed change
		},
		"track that last did its LatestVersionLookup less than interval ago waits until interval": {
			latestVersionTimestampIn: "5s",
			keepDeployedLookup:       false,
			deployedVersionJSON:      "bar",
			livenessMetric:           1,
			takesAtLeast:             5 * time.Second,
			startLatestVersion:       "1.2.2",
			wantLatestVersion:        "1.2.2",
			wantAnnounces:            1, // announce: 1 for latest query
			wantDatabaseMesages:      0, // db: 0 for nothing changing
		},
		"track that last did its DeployedVersionLookup less than interval ago waits until interval": {
			deployedVersionTimestampIn: "3s",
			keepDeployedLookup:         true,
			deployedVersionJSON:        "bar",
			livenessMetric:             1,
			takesAtLeast:               5 * time.Second,
			startLatestVersion:         "1.2.2", startDeployedVersion: "1.2.0",
			wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			wantAnnounces:       2, // announce: 1 for latest query, 1 for deployed change
			wantDatabaseMesages: 1, // db: 1 for deployed change
		},
		"inactive service doesn't track": {
			livenessMetric: 0, active: boolPtr(false),
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0,
		},
		"deleting service stops track": {
			livenessMetric: 0, deleting: true,
			startLatestVersion: "", startDeployedVersion: "",
			wantLatestVersion: "", wantDeployedVersion: "",
			wantAnnounces: 0, wantDatabaseMesages: 0,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			didFinish := make(chan bool, 1)
			service := testServiceURL(name)
			service.Status.Deleting = tc.deleting
			service.Options.Active = tc.active
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
			interval := service.Options.GetIntervalDuration()
			// subtract this from now to get the timestamp
			if tc.latestVersionTimestampIn != "" {
				latestVersionTimestampIn, _ := time.ParseDuration(tc.latestVersionTimestampIn)
				service.Status.LatestVersionTimestamp = time.Now().Add(-interval + latestVersionTimestampIn).UTC().Format(time.RFC3339)
			}
			latestVersionTimestamp := service.Status.LatestVersionTimestamp
			if tc.deployedVersionTimestampIn != "" {
				deployedVersionTimestampIn, _ := time.ParseDuration(tc.deployedVersionTimestampIn)
				service.Status.DeployedVersionTimestamp = time.Now().Add(-interval + deployedVersionTimestampIn).UTC().Format(time.RFC3339)
			}
			deployedVersionTimestamp := service.Status.DeployedVersionTimestamp
			*service.Dashboard.AutoApprove = tc.autoApprove
			allowInvalidCerts := !tc.signedCerts
			service.LatestVersion.AllowInvalidCerts = &allowInvalidCerts
			service.Init(jLog, service.Defaults, service.HardDefaults, &shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{}, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{})

			// WHEN CheckValues is called on it
			go func() {
				service.Track()
				didFinish <- true
			}()
			for i := 0; i < 200; i++ {
				passQ := testutil.ToFloat64(metric.LatestVersionQueryMetric.WithLabelValues(service.ID, "SUCCESS"))
				failQ := testutil.ToFloat64(metric.LatestVersionQueryMetric.WithLabelValues(service.ID, "FAIL"))
				if passQ != float64(0) || failQ != float64(0) {
					if tc.keepDeployedLookup {
						passQ := testutil.ToFloat64(metric.DeployedVersionQueryMetric.WithLabelValues(service.ID, "SUCCESS"))
						failQ := testutil.ToFloat64(metric.DeployedVersionQueryMetric.WithLabelValues(service.ID, "FAIL"))
						// if deployedVersionLookup hasn't queried, keep waiting
						if passQ != float64(0) || failQ != float64(0) {
							break
						}
					} else {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
			time.Sleep(time.Second)
			// Check that we waited until interval had gone since the last deployedVersionLookup
			if tc.deployedVersionTimestampIn != "" {
				// When we'd expect the query to be done after
				timeUntilInterval, _ := time.ParseDuration(tc.deployedVersionTimestampIn)
				previousTimestamp, _ := time.Parse(time.RFC3339, deployedVersionTimestamp)
				expectedAfter := previousTimestamp.Add(timeUntilInterval)

				// WHen we actually did the query
				didAt, _ := time.Parse(time.RFC3339, service.Status.DeployedVersionTimestamp)
				if didAt.Before(expectedAfter) {
					t.Errorf("DeployedVersionLookup should have waited until\n%s, but did it at\n%s",
						expectedAfter, service.Status.DeployedVersionTimestamp)
				}
			}
			// Check that we waited until interval had gone since the last latestVersionLookup
			if tc.latestVersionTimestampIn != "" {
				// When we'd expect the query to be done after
				timeUntilInterval, _ := time.ParseDuration(tc.latestVersionTimestampIn)
				previousTimestamp, _ := time.Parse(time.RFC3339, latestVersionTimestamp)
				expectedAfter := previousTimestamp.Add(timeUntilInterval)

				// WHen we actually did the query
				didAt, _ := time.Parse(time.RFC3339, service.Status.LastQueried)
				if didAt.Before(expectedAfter) {
					t.Errorf("LatestVersionLookup should have waited until\n%s, but did it at\n%s\n%v",
						expectedAfter, service.Status.LastQueried, time.Now().UTC())
				}
			}

			// THEN the scrape updates the Status correctly
			if tc.wantLatestVersion != service.Status.LatestVersion || tc.wantDeployedVersion != service.Status.DeployedVersion {
				t.Fatalf("\nLatestVersion, want %q, got %q\nDeployedVersion, want %q, got %q\n",
					tc.wantLatestVersion, service.Status.LatestVersion, tc.wantDeployedVersion, service.Status.DeployedVersion)
			}
			// LatestVersionQueryMetric
			gotMetric := testutil.ToFloat64(metric.LatestVersionQueryLiveness.WithLabelValues(service.ID))
			if !tc.ignoreLivenessMetric && gotMetric != float64(tc.livenessMetric) {
				t.Errorf("LatestVersionQueryLiveness should be %d, not %f",
					tc.livenessMetric, gotMetric)
			}
			// AnnounceChannel
			gotAnnounceMessages := len(*service.Status.AnnounceChannel)
			if tc.wantAnnounces != len(*service.Status.AnnounceChannel) {
				t.Errorf("expected AnnounceChannel to have %d messages in queue, not %d",
					tc.wantAnnounces, len(*service.Status.AnnounceChannel))
				for gotAnnounceMessages > 0 {
					var msg api_type.WebSocketMessage
					msgBytes := <-*service.Status.AnnounceChannel
					json.Unmarshal(msgBytes, &msg)
					t.Logf("got message:\n{%v}\n", msg)
					gotAnnounceMessages = len(*service.Status.AnnounceChannel)
				}
			}
			// DatabaseChannel
			gotDatabaseMessages := len(*service.Status.DatabaseChannel)
			if tc.wantDatabaseMesages != gotDatabaseMessages {
				t.Errorf("expected DatabaseChannel to have %d messages in queue, not %d",
					tc.wantDatabaseMesages, gotDatabaseMessages)
				for gotDatabaseMessages > 0 {
					var msg api_type.WebSocketMessage
					msgBytes := <-*service.Status.AnnounceChannel
					json.Unmarshal(msgBytes, &msg)
					t.Logf("got message:\n{%v}\n", msg)
					gotDatabaseMessages = len(*service.Status.DatabaseChannel)
				}
			}
			// Track should finish if it's not active and is not being deleted
			shouldFinish := util.EvalNilPtr(tc.active, true) && !tc.deleting
			// Finished when it shouldn't have?
			if len(didFinish) != 0 && shouldFinish {
				t.Fatal("didn't expect Track to finish")
			}
			// Didn't finish but should have?
			if len(didFinish) == 0 && !shouldFinish {
				t.Fatal("expected Track to finish when not active, or is deleting")
			}
			service.Status.Deleting = true
		})
	}
}
