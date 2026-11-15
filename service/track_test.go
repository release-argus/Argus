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

package service

import (
	"regexp"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/web/metric"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestServices_Track(t *testing.T) {
	tests := []struct {
		name     string
		ordering []string
		services []string
		active   []bool
	}{
		{
			name:     "empty Ordering does no queries",
			ordering: []string{},
			services: []string{"github", "url"},
		},
		{
			name:     "only tracks active Services",
			ordering: []string{"github", "url"},
			services: []string{"github", "url"},
			active:   []bool{false, true},
		},
	}

	for _, tc := range tests {
		var services *Services
		if len(tc.services) != 0 {
			services = &Services{}
			i := 0
			for _, j := range tc.services {
				(*services)[j] = testService(t, tc.name, j, "url")
				if len(tc.active) != 0 {
					(*services)[j].Options.Active = new(tc.active[i])
				}
				(*services)[j].Status.SetLatestVersion("", "", false)
				(*services)[j].Status.SetDeployedVersion("", "", false)
				i++
			}
		}

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			servicesBefore := decode.ToYAMLString(services, "")
			t.Cleanup(func() {
				// Set Deleting to stop the Track.
				for _, s := range *services {
					s.Status.SetDeleting()
				}
			})

			services.Track(&tc.ordering, &sync.RWMutex{})

			time.Sleep(2 * time.Second)
			for i := range *services {
				switch {
				case !slices.Contains(tc.ordering, i):
					if got, want := (*services)[i].Status.LatestVersion(), ""; got != want {
						t.Fatalf(
							"not in ordering: Services[%q].Status.LatestVersion() mismatch\n"+
								"got:  %q\nwant: %q",
							i, got, want,
						)
					}
				case (*services)[i].Options.GetActive():
					if got, want := (*services)[i].Status.LatestVersion(), ""; got == want {
						t.Fatalf(
							"active: Services[%q].Status.LatestVersion() mismatch\n"+
								"got:  %q\nwant: NOT %q",
							i, got, want,
						)
					}
				default:
					if got, want := (*services)[i].Status.LatestVersion(), ""; got != want {
						t.Fatalf(
							"inactive: Services[%q].Status.LatestVersion() mismatch\n"+
								"got:  %q\nwant: %q",
							i, got, want,
						)
					}
				}
			}

			// Clear 'options.active: true' as that's nil'd on Track.
			servicesBefore = regexp.MustCompile(`\n +options:\n +active: true`).ReplaceAllString(servicesBefore, "")
			servicesAfter := decode.ToYAMLString(services, "")
			if servicesAfter != servicesBefore {
				t.Fatalf("Services mismatch\ngot:  %q\nwant: %q", servicesAfter, servicesBefore)
			}
		})
	}
}

func TestService_Track(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	testURLService := testService(t, t.Name(), "url", "url")
	if _, err := testURLService.LatestVersion.Query(false, logx.LogFrom{}); err != nil {
		t.Fatalf("Lookup.Query() failed: %v", err)
	}
	testURLLatestVersion := testURLService.Status.LatestVersion()

	type versions struct {
		startLatestVersion, wantLatestVersion     string
		startDeployedVersion, wantDeployedVersion string
	}

	tests := []struct {
		name                 string
		latestVersionType    string
		overrides            []byte
		wantQueryIn          string
		keepDeployedLookup   bool
		livenessMetric       metric.LatestVersionQueryResult
		ignoreLivenessMetric bool
		versions             versions
		wantAnnounces        int
		wantDatabaseMessages int
		deleting             bool
	}{
		{
			name:              "first query updates LatestVersion and DeployedVersion",
			latestVersionType: "url",
			livenessMetric:    metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: testURLLatestVersion,
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		{
			name:              "first query updates LatestVersion and DeployedVersion - unchanged by active=true",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					active: true
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: testURLLatestVersion,
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		{
			name:              "query finds a newer version/updates LatestVersion and DeployedVersion - no commands or webhooks",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare.URLValid + `/1.2.2
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
				wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for latest, 1 for deployed.
		},
		{
			name:              "query finds a newer version/updates LatestVersion and not DeployedVersion - has webhook",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare.URLValid + `/1.2.2
				webhook:
					test:
						` + test.Indent(whtest.Webhook(t, false, false, false).String(""), 4) + `
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
				wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.1",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 1, // DB: 1 for latest.
		},
		{
			name:              "query finds a newer version/does send webhooks if autoApprove enabled",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare.URLValid + `/1.2.2
				webhook:
					test:
						` + test.Indent(whtest.Webhook(t, false, false, false).String(""), 4) + `
				dashboard:
					auto_approve: true
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "1.2.1", startDeployedVersion: "1.2.1",
				wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed.
			wantDatabaseMessages: 2, // DB: 1 for latest, 1 for deployed.
		},
		{
			name:              "query does update versions if it finds one that's older semantically (version deleted?)",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					semantic_versioning: true
				latest_version:
					url: ` + test.LookupBare.URLValid + `/1.2.0
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "1.2.3", startDeployedVersion: "1.2.3",
				wantLatestVersion: "1.2.0", wantDeployedVersion: "1.2.0",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for latest, 1 for deployed.
		},
		{
			name:              "cert/invalid disallowed",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare.URLInvalid + `/1.2.1
					allow_invalid_certs: false
					require: null
			`)),
			livenessMetric: metric.LatestVersionQueryResultFailed,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces:        0, // Query fail.
			wantDatabaseMessages: 0, // Query fail.
		},
		{
			name:              "cert/invalid allowed",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare.URLInvalid + `/1.2.1
					allow_invalid_certs: true
					require: null
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "1.2.1", wantDeployedVersion: "1.2.1",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		{
			name:              "cert/signed allowed",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare.URLValid + `/1.2.1
					allow_invalid_certs: false
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "1.2.1", wantDeployedVersion: "1.2.1",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		{
			name:              "github/urlCommand, regex fail",
			latestVersionType: "github",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)foo'
			`)),
			livenessMetric: metric.LatestVersionQueryResultNoMatch,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "url/urlCommand, regex fail",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url_commands:
						- type: regex
							regex: 'ver([0-9.]+)foo'
			`)),
			livenessMetric: metric.LatestVersionQueryResultNoMatch,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "github/urlCommand, split fail",
			latestVersionType: "github",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url_commands:
						- type: split
							text: '_-_'
			`)),
			livenessMetric: metric.LatestVersionQueryResultNoMatch,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "url/urlCommand, split fail",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url_commands:
						- type: split
							text: '_-_'
			`)),
			livenessMetric: metric.LatestVersionQueryResultNoMatch,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "handle leading v's/semantic",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare.URLValid + `/v1.2.2
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "v1.2.2", wantDeployedVersion: "v1.2.2",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		{
			name:              "handle leading v's/non-semantic",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					semantic_versioning: false
				latest_version:
					url: ` + test.LookupBare.URLValid + `/v1.2.2
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "v1.2.2", wantDeployedVersion: "v1.2.2",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		{
			name:              "github/non-semantic version fail",
			latestVersionType: "github",
			overrides: []byte(test.TrimYAML(`
				options:
					semantic_versioning: true
				latest_version:
					url_commands:
						- type: regex
							regex: '[0-9.]+'
							template: ver$1
					require: null
			`)),
			livenessMetric: metric.LatestVersionQueryResultNoMatch,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "url/non-semantic version fail",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					semantic_versioning: true
				latest_version:
					url: ` + test.LookupBare.URLValid + `/ver1.2.3
					url_commands:
						- type: regex
							regex: 'ver[0-9.]+'
					require: null
			`)),
			livenessMetric: metric.LatestVersionQueryResultSemanticVersionFail,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "progressive versioning (get older version)",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					semantic_versioning: true
				latest_version:
					url: ` + test.LookupBare.URLValid + `/1.2.2
			`)),
			livenessMetric: metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: "1.2.3", startDeployedVersion: "1.2.3",
				wantLatestVersion: "1.2.2", wantDeployedVersion: "1.2.2",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 2, // DB: 1 for deployed, 1 for latest.
		},
		{
			name:              "track gets DeployedVersion/unchanged",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				deployed_version:
					url: ` + test.LookupBare.URLValid + `/1.2.4
			`)),
			keepDeployedLookup: true,
			livenessMetric:     metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: testURLLatestVersion, startDeployedVersion: "1.2.4",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: "1.2.4",
			},
			wantAnnounces:        1, // Announce: 1 for latest query.
			wantDatabaseMessages: 0,
		},
		{
			name:              "track gets DeployedVersion/newer does not change LatestVersion",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				deployed_version:
					url: ` + test.LookupBare.URLValid + `/3.2.1
			`)),
			keepDeployedLookup:   true,
			ignoreLivenessMetric: true, // Ignore as DeployedVersionLookup may be done before.
			versions: versions{
				startLatestVersion: testURLLatestVersion, startDeployedVersion: "0.0.0",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: "3.2.1",
			},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed change.
			wantDatabaseMessages: 1, // DB: 1 for deployed change.
		},
		{
			name:              "track that last did a Query less than interval ago waits until interval",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				deployed_version:
					url: ` + test.LookupBare.URLValid + `/` + testURLLatestVersion + `
			`)),
			wantQueryIn:        "5s",
			keepDeployedLookup: true,
			livenessMetric:     metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion:  testURLLatestVersion,
				wantLatestVersion:   testURLLatestVersion,
				wantDeployedVersion: testURLLatestVersion,
			},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed change.
			wantDatabaseMessages: 1, // DB: 1 for deployed change.
		},
		{
			name:              "inactive service doesn't track",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					active: false
			`)),
			// Never queried, so the gauge keeps its zero value, Failed.
			livenessMetric: metric.LatestVersionQueryResultFailed,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "deleting service stops track",
			latestVersionType: "url",
			// Never queried, so the gauge keeps its zero value, Failed.
			livenessMetric: metric.LatestVersionQueryResultFailed,
			deleting:       true,
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
		{
			name:              "nil latest_version doesn't track",
			latestVersionType: "url",
			overrides:         []byte("latest_version: null"),
			versions: versions{
				startLatestVersion: "", startDeployedVersion: "",
				wantLatestVersion: "", wantDeployedVersion: "",
			},
			wantAnnounces: 0, wantDatabaseMessages: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, tc.latestVersionType, "url")
			svcStatus := svc.Status.Copy(true)
			t.Cleanup(func() {
				// Set Deleting to stop the Track.
				svc.Status.SetDeleting()
			})

			if tc.overrides != nil {
				var err error
				svc, err = ApplyOverrides(
					"yaml", tc.overrides,
					svc,
					tc.name,
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					t.Fatalf("ApplyOverrides() failed: %v", err)
				}
				svc.Status.AnnounceChannel = svcStatus.AnnounceChannel
				svc.Status.DatabaseChannel = svcStatus.DatabaseChannel
				svc.Status.SaveChannel = svcStatus.SaveChannel
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
			// Neither Lookup may query before this.
			var queryNotBefore time.Time
			if tc.wantQueryIn != "" {
				wantQueryIn, err := time.ParseDuration(tc.wantQueryIn)
				if err != nil {
					t.Fatalf("time.ParseDuration() failed: %v", err)
				}
				lastQueried := time.Now().Add(-interval + wantQueryIn).UTC().Truncate(time.Second)
				svc.Status.SetLastQueried(lastQueried.Format(time.RFC3339))
				queryNotBefore = lastQueried.Add(interval)
			}
			didFinish := make(chan bool, 1)
			svcBefore := svc.String("")

			go func() {
				svc.Track()
				didFinish <- true
			}()
			for range 200 {
				var passQ, failQ float64
				if svc.LatestVersion != nil {
					passQ = testutil.ToFloat64(
						metric.LatestVersionQueryResultTotal.WithLabelValues(
							svc.ID, svc.LatestVersion.GetType(), metric.ActionResultSuccess,
						),
					)
					failQ = testutil.ToFloat64(
						metric.LatestVersionQueryResultTotal.WithLabelValues(
							svc.ID, svc.LatestVersion.GetType(), metric.ActionResultFail,
						),
					)
				}
				if passQ != float64(0) || failQ != float64(0) {
					if tc.keepDeployedLookup {
						dvPassQ := testutil.ToFloat64(
							metric.DeployedVersionQueryResultTotal.WithLabelValues(
								svc.ID, svc.DeployedVersionLookup.GetType(), metric.ActionResultSuccess,
							),
						)
						dvFailQ := testutil.ToFloat64(
							metric.DeployedVersionQueryResultTotal.WithLabelValues(
								svc.ID, svc.DeployedVersionLookup.GetType(), metric.ActionResultFail,
							),
						)
						if dvPassQ != float64(0) || dvFailQ != float64(0) {
							break
						}
					} else {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
			// Wait for the expected messages.
			for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
				if len(svc.Status.AnnounceChannel) >= tc.wantAnnounces &&
					len(svc.Status.DatabaseChannel) >= tc.wantDatabaseMessages {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
			// Catch any unexpected extra message.
			time.Sleep(1000 * time.Millisecond)

			if tc.wantQueryIn != "" {
				want := queryNotBefore.Format(time.RFC3339)
				lvQueried := svc.Status.LastQueried()
				if got, err := time.Parse(time.RFC3339, lvQueried); err != nil || got.Before(queryNotBefore) {
					t.Errorf(
						"Status.LastQueried() mismatch\ngot:  %q (err=%v)\nwant: %q or later",
						lvQueried, err, want,
					)
				}
				// The DeployedVersion changes on this query, so its timestamp is the query time.
				dvQueried := svc.Status.DeployedVersionTimestamp()
				if got, err := time.Parse(time.RFC3339, dvQueried); err != nil || got.Before(queryNotBefore) {
					t.Errorf(
						"Status.DeployedVersionTimestamp() mismatch\n"+
							"got:  %q (err=%v)\nwant: %q or later",
						dvQueried, err, want,
					)
				}
			}

			if got := svc.Status.LatestVersion(); got != tc.versions.wantLatestVersion {
				t.Fatalf(
					"Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.versions.wantLatestVersion,
				)
			}
			if got := svc.Status.DeployedVersion(); got != tc.versions.wantDeployedVersion {
				t.Fatalf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.versions.wantDeployedVersion,
				)
			}
			if svc.LatestVersion != nil {
				gotMetric := testutil.ToFloat64(
					metric.LatestVersionQueryResultLast.WithLabelValues(
						svc.ID, svc.LatestVersion.GetType(),
					),
				)
				if gotMetric != float64(tc.livenessMetric) && !tc.ignoreLivenessMetric {
					t.Errorf(
						"metric.LatestVersionQueryResultLast mismatch\ngot:  %d\nwant: %d",
						int(gotMetric), tc.livenessMetric,
					)
				}
			}
			if gotAnnounceMessages := len(svc.Status.AnnounceChannel); gotAnnounceMessages != tc.wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotAnnounceMessages, tc.wantAnnounces,
				)
				test.LogChannel(t, "Status.AnnounceChannel", svc.Status.AnnounceChannel)
			}
			if gotDatabaseMessages := len(svc.Status.DatabaseChannel); gotDatabaseMessages != tc.wantDatabaseMessages {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					gotDatabaseMessages, tc.wantDatabaseMessages,
				)
				for len(svc.Status.DatabaseChannel) != 0 {
					msg := <-svc.Status.DatabaseChannel
					cells := make([]string, len(msg.Cells))
					for i, cell := range msg.Cells {
						cells[i] = cell.Column + "=" + cell.Value
					}
					t.Logf(
						"database message: ServiceID=%q, Delete=%t, Cells=%q",
						msg.ServiceID, msg.Delete, cells,
					)
				}
			}
			// Track returns only when inactive, deleting, or LatestVersion is nil.
			shouldFinish := !svc.Options.GetActive() || tc.deleting || svc.LatestVersion == nil
			if gotFinished := len(didFinish) != 0; gotFinished != shouldFinish {
				t.Fatalf(
					"Service.Track() finished mismatch\ngot:  %t\nwant: %t",
					gotFinished, shouldFinish,
				)
			}

			if got := svc.String(""); got != svcBefore {
				t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", got, svcBefore)
			}
		})
	}
}
