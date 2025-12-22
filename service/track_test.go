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

//go:build integration

package service

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/notify/shoutrrr"
	dv_web "github.com/release-argus/Argus/service/deployed_version/types/web"
	lv_web "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/metric"
	"github.com/release-argus/Argus/webhook"
	webhook_test "github.com/release-argus/Argus/webhook/test"
)

func TestService_Track(t *testing.T) {
	testURLService := testService(t, "TestService_Track", "url")
	testURLService.LatestVersion.Query(false, logutil.LogFrom{})
	testURLLatestVersion := testURLService.Status.LatestVersion()

	type overrides struct {
		latestVersion    string
		nilLatestVersion bool
		deployedVersion  string
		other            string
	}
	type versions struct {
		startLatestVersion, wantLatestVersion     string
		startDeployedVersion, wantDeployedVersion string
	}
	// GIVEN a Service.
	tests := map[string]struct {
		latestVersionType    string
		overrides            overrides
		wantQueryIn          string
		keepDeployedLookup   bool
		livenessMetric       int
		ignoreLivenessMetric bool
		takesAtLeast         time.Duration
		versions             versions
		wantAnnounces        int
		wantDatabaseMessages int
		deleting             bool
	}{
		"first query updates LatestVersion and DeployedVersion": {
			latestVersionType: "url",
			livenessMetric:    1,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: testURLLatestVersion},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		"first query updates LatestVersion and DeployedVersion - unchanged by active=true": {
			latestVersionType: "url",
			overrides: overrides{
				other: test.TrimYAML(`
					options:
						active: true
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: testURLLatestVersion},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		"query finds a newer version and updates LatestVersion and DeployedVersion - no commands/webhooks": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)'
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: testURLLatestVersion},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for latest, 1 for deployed.
		},
		"query finds a newer version and updates LatestVersion and not DeployedVersion - has webhook": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)'
				`),
				other: test.TrimYAML(`
					webhook:
						test:
							` + test.Indent(
					webhook_test.WebHook(false, false, false).String(), 4) + `
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: "1.2.1"},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 1, // DB: 1 for latest.
		},
		"query finds a newer version does send webhooks if autoApprove enabled": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)'
				`),
				other: test.TrimYAML(`
					webhook:
						test:
							` + test.Indent(
					webhook_test.WebHook(false, false, false).String(), 4) + `
					dashboard:
						auto_approve: true
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: testURLLatestVersion},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed.
			wantDatabaseMessages: 2, // DB: 1 for latest, 1 for deployed.
		},
		"query doesn't update versions if it finds one that's older semantically": {
			latestVersionType: "url",
			// would get '1.2.1', but stay on '1.2.2'.
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: '"([0-9.]+)"'
					require: null
				`),
			},
			livenessMetric: 4,
			versions: versions{
				startLatestVersion: testURLLatestVersion, startDeployedVersion: testURLLatestVersion,
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: testURLLatestVersion},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"track on invalid cert allowed": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url: ` + test.LookupPlain["url_invalid"] + `
					allow_invalid_certs: true
					url_commands:
						- type: regex
							regex: '"([0-9.]+)"'
					require: null
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "1.2.1", wantDeployedVersion: "1.2.1"},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		"track on signed cert allowed": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url: ` + test.LookupPlain["url_valid"] + `
					allow_invalid_certs: false
					url_commands:
						- type: regex
							regex: '"([0-9.]+)"'
					require: null
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "1.2.1", wantDeployedVersion: "1.2.1"},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		"github - urlCommand, regex fail": {
			latestVersionType: "github",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)foo'
				`),
			},
			livenessMetric: 2,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"url - urlCommand, regex fail": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)foo'
				`),
			},
			livenessMetric: 2,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"github - urlCommand, split fail": {
			latestVersionType: "github",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: split
							text: '_-_'
				`),
			},
			livenessMetric: 2,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"url - urlCommand, split fail": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: split
							text: '_-_'
				`),
			},
			livenessMetric: 2,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"handle leading v's - semantic": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)'
							template: 'v$1'
					require: null
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "v1.2.2", wantDeployedVersion: "v1.2.2"},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		"handle leading v's - non-semantic": {
			latestVersionType: "url",
			overrides: overrides{
				other: test.TrimYAML(`
					options:
						semantic_versioning: false
				`),
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)'
							template: 'v$1'
					require: null
				`),
			},
			livenessMetric: 1,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "v1.2.2", wantDeployedVersion: "v1.2.2"},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		"non-semantic version fail": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver[0-9.]+'
					require: null
				`),
			},
			livenessMetric: 3,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"progressive version fail (got older version)": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)'
					regexp_template: 'v$1'
					require: null
				`),
			},
			livenessMetric: 4,
			versions: versions{
				startLatestVersion: "1.2.3", startDeployedVersion: "1.2.3",
				wantLatestVersion: "1.2.3", wantDeployedVersion: "1.2.3"},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"other fail (invalid cert)": {
			latestVersionType: "url",
			overrides: overrides{
				latestVersion: test.TrimYAML(`
					url: ` + test.LookupPlain["url_invalid"] + `
					allow_invalid_certs: false
					url_commands:
						- type: regex
							regex: '([0-9.]+)'
					require: null
				`),
			},
			livenessMetric: 0,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"track gets DeployedVersion": {
			latestVersionType: "url",
			overrides: overrides{
				deployedVersion: test.TrimYAML(`
					json: bar
					regex: ""
					regex_template: ""
				`),
			},
			keepDeployedLookup: true, livenessMetric: 1,
			versions: versions{
				startLatestVersion: testURLLatestVersion, startDeployedVersion: "1.2.0",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: "1.2.2"},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed change.
			wantDatabaseMessages: 1, // DB: 1 for deployed change.
		},
		"track gets DeployedVersion that is newer and does not change LatestVersion": {
			latestVersionType: "url",
			overrides: overrides{
				deployedVersion: test.TrimYAML(`
					json: foo.bar.version
					regex: ""
					regex_template: ""
				`),
			},
			keepDeployedLookup:   true,
			ignoreLivenessMetric: true, // Ignore as DeployedVersionLookup may be done before.
			versions: versions{
				startLatestVersion: testURLLatestVersion, startDeployedVersion: "0.0.0",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: "3.2.1"},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed change.
			wantDatabaseMessages: 1, // db: 1 for deployed change.
		},
		"track that last did a Query less than interval ago waits until interval": {
			latestVersionType: "url",
			overrides: overrides{
				other: test.TrimYAML(`
					deployed_version:
						json: bar
				`),
			},
			wantQueryIn:        "5s",
			keepDeployedLookup: false,
			livenessMetric:     1,
			takesAtLeast:       5 * time.Second,
			versions: versions{
				startLatestVersion: testURLLatestVersion,
				wantLatestVersion:  testURLLatestVersion},
			wantAnnounces:        1, // announce: 1 for latest query.
			wantDatabaseMessages: 0, // db: 0 for nothing changing.
		},
		"inactive service doesn't track": {
			latestVersionType: "url",
			overrides: overrides{
				other: test.TrimYAML(`
					options:
						active: false
				`),
			},
			livenessMetric: 0,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"deleting service stops track": {
			latestVersionType: "url",
			livenessMetric:    0, deleting: true,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		"nil latest_version doesn't track": {
			latestVersionType: "url",
			overrides: overrides{
				nilLatestVersion: true},
			livenessMetric: 0,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: ""},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, name, tc.latestVersionType)

			// overrides - other.
			err := yaml.Unmarshal([]byte(tc.overrides.other), &svc)
			if err != nil {
				t.Fatalf("%s\nfailed to unmarshal overrides: %s",
					packageName, err)
			}
			// overrides - latest_version.
			if lv, ok := svc.LatestVersion.(*lv_web.Lookup); ok {
				err = yaml.Unmarshal([]byte(tc.overrides.latestVersion), lv)
				if err != nil {
					t.Fatalf("%s\nfailed to unmarshal overrides: %s",
						packageName, err)
				}
			}
			if tc.overrides.nilLatestVersion {
				svc.LatestVersion = nil
			}
			// overrides - deployed_version.
			if dv, ok := svc.DeployedVersionLookup.(*dv_web.Lookup); ok {
				err = yaml.Unmarshal([]byte(tc.overrides.deployedVersion), dv)
				if err != nil {
					t.Fatalf("%s\nfailed to unmarshal overrides: %s",
						packageName, err)
				}
			}

			svc.Status.SetLatestVersion(tc.versions.startLatestVersion, "", false)
			svc.Status.SetDeployedVersion(tc.versions.startDeployedVersion, "", false)
			if tc.deleting {
				svc.Status.SetDeleting()
			}
			if !tc.keepDeployedLookup {
				svc.DeployedVersionLookup = nil
			}
			interval := svc.Options.GetIntervalDuration()
			// Subtract this from now to get the timestamp.
			if tc.wantQueryIn != "" {
				wantQueryIn, _ := time.ParseDuration(tc.wantQueryIn)
				svc.Status.SetLastQueried(
					time.Now().Add(-interval + wantQueryIn).UTC().Format(time.RFC3339))
			}
			latestVersionTimestamp := svc.Status.LatestVersionTimestamp()
			deployedVersionTimestamp := svc.Status.DeployedVersionTimestamp()
			shoutrrrHardDefaults := shoutrrr.ShoutrrrsDefaults{}
			shoutrrrHardDefaults.Default()
			webhookHardDefaults := webhook.Defaults{}
			webhookHardDefaults.Default()
			svc.Init(
				svc.Defaults, svc.HardDefaults,
				&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrrHardDefaults,
				&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhookHardDefaults)
			didFinish := make(chan bool, 1)

			// WHEN Track is called on it.
			go func() {
				svc.Track()
				didFinish <- true
			}()
			for range 200 {
				var passQ, failQ float64
				if !tc.overrides.nilLatestVersion {
					passQ = testutil.ToFloat64(metric.LatestVersionQueryResultTotal.WithLabelValues(
						svc.ID, svc.LatestVersion.GetType(), "SUCCESS"))
					failQ = testutil.ToFloat64(metric.LatestVersionQueryResultTotal.WithLabelValues(
						svc.ID, svc.LatestVersion.GetType(), "FAIL"))
				}
				if passQ != float64(0) || failQ != float64(0) {
					if tc.keepDeployedLookup {
						passQ := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
							svc.ID, svc.DeployedVersionLookup.GetType(), "SUCCESS"))
						failQ := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
							svc.ID, svc.DeployedVersionLookup.GetType(), "FAIL"))
						// if deployedVersionLookup hasn't queried, keep waiting.
						if passQ != float64(0) || failQ != float64(0) {
							break
						}
					} else {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
			time.Sleep(500 * time.Millisecond)
			// Check that we waited until interval had gone since the last latestVersionLookup.
			if tc.wantQueryIn != "" {
				// When we'd expect the query to be done after.
				timeUntilInterval, _ := time.ParseDuration(tc.wantQueryIn)
				lvPreviousTimestamp, _ := time.Parse(time.RFC3339, latestVersionTimestamp)
				lvExpectedAfter := lvPreviousTimestamp.Add(timeUntilInterval)
				dvPreviousTimestamp, _ := time.Parse(time.RFC3339, deployedVersionTimestamp)
				dvExpectedAfter := dvPreviousTimestamp.Add(timeUntilInterval)

				// WHen we actually did the query.
				didAt, _ := time.Parse(time.RFC3339, svc.Status.LastQueried())
				if didAt.Before(lvExpectedAfter) {
					t.Errorf("%s\nLatestVersionLookup happened too early\nwant: %s or later\ngot:  %s\nnow:  %s",
						packageName, lvExpectedAfter, svc.Status.LastQueried(), time.Now().UTC())
				}
				if didAt.Before(dvExpectedAfter) {
					t.Errorf("%s\nDeployedVersionLookup happened too early\nwant: %s or later\ngot  %s\nnow:  %s",
						packageName, dvExpectedAfter, svc.Status.LastQueried(), time.Now().UTC())
				}
			}

			// THEN the scrape updates the Status correctly.
			if svc.Status.LatestVersion() != tc.versions.wantLatestVersion ||
				svc.Status.DeployedVersion() != tc.versions.wantDeployedVersion {
				t.Fatalf("%s\nLatestVersion possible mismatch\nwant: %q\ngot  %q\nDeployedVersion possible mismatch\nwant: %q\ngot:  %q",
					packageName,
					tc.versions.wantLatestVersion, svc.Status.LatestVersion(),
					tc.versions.wantDeployedVersion, svc.Status.DeployedVersion())
			}
			// LatestVersionQueryResultTotal.
			if !tc.overrides.nilLatestVersion {
				gotMetric := testutil.ToFloat64(metric.LatestVersionQueryResultLast.WithLabelValues(svc.ID, svc.LatestVersion.GetType()))
				if !tc.ignoreLivenessMetric && gotMetric != float64(tc.livenessMetric) {
					t.Errorf("%s\nLatestVersionQueryResultLast mismatch\nwant: %d\ngot:  %f",
						packageName, tc.livenessMetric, gotMetric)
				}
			}
			// AnnounceChannel.
			if gotAnnounceMessages := len(svc.Status.AnnounceChannel); gotAnnounceMessages != tc.wantAnnounces {
				t.Errorf("%s\nAnnounceChannel length mismatch\nwant: %d messages\ngot:  %d",
					packageName, tc.wantAnnounces, gotAnnounceMessages)
				for gotAnnounceMessages > 0 {
					var msg apitype.WebSocketMessage
					msgBytes := <-svc.Status.AnnounceChannel
					_ = json.Unmarshal(msgBytes, &msg)
					t.Logf("%s - got message: {%+v}",
						packageName, msg)
					gotAnnounceMessages = len(svc.Status.AnnounceChannel)
				}
			}
			// DatabaseChannel.
			if gotDatabaseMessages := len(svc.Status.DatabaseChannel); gotDatabaseMessages != tc.wantDatabaseMessages {
				t.Errorf("%s\nDatabaseChannel length mismatch\nwant: %d messages\ngot:  %d",
					packageName, tc.wantDatabaseMessages, gotDatabaseMessages)
				for gotDatabaseMessages > 0 {
					var msg apitype.WebSocketMessage
					msgBytes := <-svc.Status.AnnounceChannel
					_ = json.Unmarshal(msgBytes, &msg)
					t.Logf("%s - got message:\n{%v}\n",
						packageName, msg)
					gotDatabaseMessages = len(svc.Status.DatabaseChannel)
				}
			}
			// Track should finish if it is not Active and is not being deleted.
			shouldFinish := !svc.Options.GetActive() || tc.deleting || tc.overrides.nilLatestVersion
			// Didn't finish, but should have.
			if shouldFinish && len(didFinish) == 0 {
				t.Fatalf("%s\nexpected Track to finish when not active, deleting, or LatestVersion is nil",
					packageName)
			}
			// Finished when it shouldn't have.
			if !shouldFinish && len(didFinish) != 0 {
				t.Fatalf("%s\ndidn't expect Track to finish",
					packageName)
			}

			// Set Deleting to stop the Track.
			svc.Status.SetDeleting()
		})
	}
}

func TestServices_Track(t *testing.T) {
	// GIVEN a Services.
	tests := map[string]struct {
		ordering []string
		services []string
		active   []bool
	}{
		"empty Ordering does no queries": {
			ordering: []string{},
			services: []string{"github", "url"}},
		"only tracks active Services": {
			ordering: []string{"github", "url"},
			services: []string{"github", "url"},
			active:   []bool{false, true}},
	}

	for name, tc := range tests {
		var services *Services
		if len(tc.services) != 0 {
			services = &Services{}
			i := 0
			for _, j := range tc.services {
				switch j {
				case "github":
					(*services)[j] = testService(t, name, "github")
				case "url":
					(*services)[j] = testService(t, name, "url")
				}
				if len(tc.active) != 0 {
					(*services)[j].Options.Active = test.BoolPtr(tc.active[i])
				}
				(*services)[j].Status.SetLatestVersion("", "", false)
				(*services)[j].Status.SetDeployedVersion("", "", false)
				i++
			}
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Track is called on it.
			services.Track(&tc.ordering, &sync.RWMutex{})

			// THEN the function exits straight away.
			time.Sleep(2 * time.Second)
			for i := range *services {
				if !util.Contains(tc.ordering, i) {
					if wantLatestVersion := ""; (*services)[i].Status.LatestVersion() != wantLatestVersion {
						t.Fatalf("%s\nQuery on Services[%q] shouldn't have updated LatestVersion as not in ordering\nwant: %q\ngot:  %q\norder: %v",
							packageName, i,
							wantLatestVersion, (*services)[i].Status.String(),
							tc.ordering)
					}
				} else if (*services)[i].Options.GetActive() {
					if (*services)[i].Status.LatestVersion() == "" {
						t.Fatalf("%s\nQuery on Services[%q] didn't find LatestVersion\nwant: %s\ngot:  %q",
							packageName, i, (*services)[i].Status.String(), "")
					}
				} else if (*services)[i].Status.LatestVersion() != "" {
					t.Fatalf("%s\nQuery on Services[%q] shouldn't have updated LatestVersion\nwant: %q\ngot:  %q",
						packageName, i, "", (*services)[i].Status.String())
				}

				// Set Deleting to stop the Track.
				(*services)[i].Status.SetDeleting()
			}
		})
	}
}
