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
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/config/decode"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	whtest "github.com/release-argus/Argus/webhook/test"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/metric"
)

func TestServices_Track(t *testing.T) {
	// GIVEN: a Services.
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
					(*services)[j].Options.Active = test.Ptr(tc.active[i])
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

			// WHEN: Track is called on it.
			services.Track(&tc.ordering, &sync.RWMutex{})

			prefix := fmt.Sprintf("%s\nServices.Track()", packageName)

			// THEN: the function exits straight away.
			time.Sleep(2 * time.Second)
			for i := range *services {
				if !util.Contains(tc.ordering, i) {
					if wantLatestVersion := ""; (*services)[i].Status.LatestVersion() != wantLatestVersion {
						t.Fatalf(
							"%s query on Services[%q] shouldn't have happened as not in ordering\ngot:  latest_version=%q\nwant: latest_version=%q\norder: %v",
							prefix, i,
							wantLatestVersion, (*services)[i].Status.String(),
							tc.ordering,
						)
					}
				} else if (*services)[i].Options.GetActive() {
					if (*services)[i].Status.LatestVersion() == "" {
						t.Fatalf(
							"%s query on Services[%q] didn't find LatestVersion\ngot:  ''\nwant: %q",
							prefix, i,
							(*services)[i].Status.String(),
						)
					}
				} else if (*services)[i].Status.LatestVersion() != "" {
					t.Fatalf(
						"%s query on Services[%q] shouldn't have updated LatestVersion\ngot:  %q\nwant: ''",
						prefix, i,
						(*services)[i].Status.String(),
					)
				}
			}

			// AND: the Services haven't changed.
			// Clear 'options.active: true' as that's nil'd on Track.
			servicesBefore = regexp.MustCompile(`\n +options:\n +active: true`).ReplaceAllString(servicesBefore, "")
			servicesAfter := decode.ToYAMLString(services, "")
			if servicesAfter != servicesBefore {
				t.Fatalf(
					"%s Services shouldn't have changed\ngot:  %q\nwant: %q",
					prefix, servicesAfter, servicesBefore,
				)
			}
		})
	}
}

func TestService_Track(t *testing.T) {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	testURLService := testService(t, "TestService_Track", "url", "url")
	_, _ = testURLService.LatestVersion.Query(false, logx.LogFrom{})
	testURLLatestVersion := testURLService.Status.LatestVersion()

	type versions struct {
		startLatestVersion, wantLatestVersion     string
		startDeployedVersion, wantDeployedVersion string
	}
	// GIVEN: a Service.
	tests := []struct {
		name                 string
		latestVersionType    string
		overrides            []byte
		wantQueryIn          string
		keepDeployedLookup   bool
		livenessMetric       metric.LatestVersionQueryResult
		ignoreLivenessMetric bool
		takesAtLeast         time.Duration
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
			name:              "query finds a newer version and updates LatestVersion and DeployedVersion - no commands/webhooks",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare["url_valid"] + `/1.2.2
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
			name:              "query finds a newer version and updates LatestVersion and not DeployedVersion - has webhook",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare["url_valid"] + `/1.2.2
				webhook:
					test:
						` + test.Indent(whtest.WebHook(t, false, false, false).String(""), 4) + `
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
			name:              "query finds a newer version does send webhooks if autoApprove enabled",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare["url_valid"] + `/1.2.2
				webhook:
					test:
						` + test.Indent(whtest.WebHook(t, false, false, false).String(""), 4) + `
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
					url: ` + test.LookupBare["url_valid"] + `/1.2.0
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
			name:              "track on invalid cert disallowed",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare["url_invalid"] + `/1.2.1
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
			name:              "track on invalid cert allowed",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare["url_invalid"] + `/1.2.1
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
			name:              "track on signed cert allowed",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare["url_valid"] + `/1.2.1
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
			name:              "github - urlCommand, regex fail",
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
			name:              "url - urlCommand, regex fail",
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
			name:              "github - urlCommand, split fail",
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
			name:              "url - urlCommand, split fail",
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
			name:              "handle leading v's - semantic",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				latest_version:
					url: ` + test.LookupBare["url_valid"] + `/v1.2.2
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
			name:              "handle leading v's - non-semantic",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					semantic_versioning: false
				latest_version:
					url: ` + test.LookupBare["url_valid"] + `/v1.2.2
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
			name:              "github - non-semantic version fail",
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
			name:              "url - non-semantic version fail",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					semantic_versioning: true
				latest_version:
					url: ` + test.LookupBare["url_valid"] + `/ver1.2.3
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
					url: ` + test.LookupBare["url_valid"] + `/1.2.2
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
			name:              "track gets DeployedVersion",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				deployed_version:
					url: ` + test.LookupBare["url_valid"] + `/1.2.4
			`)),
			keepDeployedLookup: true,
			livenessMetric:     metric.LatestVersionQueryResultSuccess,
			versions: versions{
				startLatestVersion: testURLLatestVersion, startDeployedVersion: testURLLatestVersion,
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: "1.2.4",
			},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed change.
			wantDatabaseMessages: 1, // DB: 1 for deployed change.
		},
		{
			name:              "track gets DeployedVersion that is newer and does not change LatestVersion",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				deployed_version:
					url: ` + test.LookupBare["url_valid"] + `/3.2.1
			`)),
			keepDeployedLookup:   true,
			ignoreLivenessMetric: true, // Ignore as DeployedVersionLookup may be done before.
			versions: versions{
				startLatestVersion: testURLLatestVersion, startDeployedVersion: "0.0.0",
				wantLatestVersion: testURLLatestVersion, wantDeployedVersion: "3.2.1",
			},
			wantAnnounces:        2, // Announce: 1 for latest query, 1 for deployed change.
			wantDatabaseMessages: 1, // db: 1 for deployed change.
		},
		{
			name:              "track that last did a Query less than interval ago waits until interval",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				deployed_version:
					json: bar
			`)),
			wantQueryIn:        "5s",
			keepDeployedLookup: false,
			livenessMetric:     metric.LatestVersionQueryResultSuccess,
			takesAtLeast:       5 * time.Second,
			versions: versions{
				startLatestVersion: testURLLatestVersion,
				wantLatestVersion:  testURLLatestVersion,
			},
			wantAnnounces:        1, // announce: 1 for latest query.
			wantDatabaseMessages: 0, // db: 0 for nothing changing.
		},
		{
			name:              "inactive service doesn't track",
			latestVersionType: "url",
			overrides: []byte(test.TrimYAML(`
				options:
					active: false
			`)),
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
			livenessMetric:    metric.LatestVersionQueryResultFailed,
			deleting:          true,
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
			livenessMetric:    metric.LatestVersionQueryResultFailed,
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

			// Overrides.
			if tc.overrides != nil {
				var err error
				svc, err = ApplyOverrides(
					"yaml", tc.overrides,
					svc,
					tc.name,
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					t.Fatalf(
						"%s\nfailed to unmarshal overrides: %s",
						packageName, err,
					)
				}
				svc.Status.AnnounceChannel = svcStatus.AnnounceChannel
				svc.Status.DatabaseChannel = svcStatus.DatabaseChannel
				svc.Status.SaveChannel = svcStatus.SaveChannel
			}

			svc.Status.SetLatestVersion(tc.versions.startLatestVersion, "", false)
			latestVersionTimestamp := svc.Status.LatestVersionTimestamp()
			svc.Status.SetDeployedVersion(tc.versions.startDeployedVersion, "", false)
			deployedVersionTimestamp := svc.Status.DeployedVersionTimestamp()
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
					time.Now().Add(-interval + wantQueryIn).UTC().Format(time.RFC3339),
				)
			}
			didFinish := make(chan bool, 1)
			svcBefore := svc.String("")

			// WHEN: Track is called on it.
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
						passQ := testutil.ToFloat64(
							metric.DeployedVersionQueryResultTotal.WithLabelValues(
								svc.ID, svc.DeployedVersionLookup.GetType(), metric.ActionResultSuccess,
							),
						)
						failQ := testutil.ToFloat64(
							metric.DeployedVersionQueryResultTotal.WithLabelValues(
								svc.ID, svc.DeployedVersionLookup.GetType(), metric.ActionResultFail,
							),
						)
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
			time.Sleep(1000 * time.Millisecond)

			prefix := fmt.Sprintf("%s\nService.Track()", packageName)

			// Check that we waited until interval had passed since the last latestVersionLookup.
			if tc.wantQueryIn != "" {
				// When we'd expect the query to be done after.
				timeUntilInterval, _ := time.ParseDuration(tc.wantQueryIn)
				lvPreviousTimestamp, _ := time.Parse(time.RFC3339, latestVersionTimestamp)
				lvExpectedAfter := lvPreviousTimestamp.Add(timeUntilInterval)
				dvPreviousTimestamp, _ := time.Parse(time.RFC3339, deployedVersionTimestamp)
				dvExpectedAfter := dvPreviousTimestamp.Add(timeUntilInterval)

				// When we actually did the query.
				didAt, _ := time.Parse(time.RFC3339, svc.Status.LastQueried())
				if didAt.Before(lvExpectedAfter) {
					t.Errorf(
						"%s LatestVersionLookup happened too early\ngot:  %s\nwant: %s or later\nnow:  %s",
						prefix,
						svc.Status.LastQueried(), lvExpectedAfter,
						time.Now().UTC(),
					)
				}
				if didAt.Before(dvExpectedAfter) {
					t.Errorf(
						"%s DeployedVersionLookup happened too early\ngot  %s\nwant: %s or later\nnow:  %s",
						prefix,
						svc.Status.LastQueried(),
						dvExpectedAfter,
						time.Now().UTC(),
					)
				}
			}

			// THEN: the scrape updates the Status correctly.
			if got := svc.Status.LatestVersion(); got != tc.versions.wantLatestVersion {
				t.Fatalf(
					"%s LatestVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.versions.wantLatestVersion,
				)
			}
			if got := svc.Status.DeployedVersion(); got != tc.versions.wantDeployedVersion {
				t.Fatalf(
					"%s DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.versions.wantDeployedVersion,
				)
			}
			// LatestVersionQueryResultTotal.
			if svc.LatestVersion != nil {
				gotMetric := testutil.ToFloat64(
					metric.LatestVersionQueryResultLast.WithLabelValues(
						svc.ID, svc.LatestVersion.GetType(),
					),
				)
				if gotMetric != float64(tc.livenessMetric) && !tc.ignoreLivenessMetric {
					t.Errorf(
						"%s LatestVersionQueryResultLast metric mismatch\ngot:  %d\nwant: %d",
						prefix, int(gotMetric), tc.livenessMetric,
					)
				}
			}
			// AnnounceChannel.
			if gotAnnounceMessages := len(svc.Status.AnnounceChannel); gotAnnounceMessages != tc.wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotAnnounceMessages, tc.wantAnnounces,
				)
				for gotAnnounceMessages > 0 {
					var msg apitype.WebSocketMessage
					msgBytes := <-svc.Status.AnnounceChannel
					_ = decode.Unmarshal("json", msgBytes, &msg)
					t.Logf(
						"%s - AnnounceChannel message: {%+v}",
						prefix, msg,
					)
					gotAnnounceMessages = len(svc.Status.AnnounceChannel)
				}
			}
			// DatabaseChannel.
			if gotDatabaseMessages := len(svc.Status.DatabaseChannel); gotDatabaseMessages != tc.wantDatabaseMessages {
				t.Errorf(
					"%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotDatabaseMessages, tc.wantDatabaseMessages,
				)
				for gotDatabaseMessages > 0 {
					msg := <-svc.Status.DatabaseChannel
					t.Logf(
						"%s - DatabaseChannel message:\n{%v}\n",
						prefix, msg,
					)
					gotDatabaseMessages = len(svc.Status.DatabaseChannel)
				}
			}
			// Track should finish if it is not Active and is not being deleted.
			shouldFinish := !svc.Options.GetActive() || tc.deleting || svc.LatestVersion == nil
			// Didn't finish, but should have.
			if shouldFinish && len(didFinish) == 0 {
				t.Fatalf(
					"%s expected to finish when not active, deleting, or LatestVersion is nil",
					prefix,
				)
			}
			// Finished when it shouldn't have.
			if !shouldFinish && len(didFinish) != 0 {
				t.Fatalf("%s unexpected finish", prefix)
			}

			// AND: the service should marshal the same.
			if got := svc.String(""); got != svcBefore {
				t.Errorf(
					"%s String() mismatch\ngot:  %q\nwant: %q",
					prefix, got, svcBefore,
				)
			}
		})
	}
}
