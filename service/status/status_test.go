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

package status

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/yaml.v3"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

func TestStatus_Unmarshal(t *testing.T) {
	// GIVEN a Status.
	tests := map[string]struct {
		format string
	}{
		"YAML": {
			format: "YAML",
		},
		"JSON": {
			format: "JSON",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN UnmarshalX is called on the Status.
			var status Status
			if tc.format == "YAML" {
				var node yaml.Node
				status.UnmarshalYAML(&node)
			} else if tc.format == "JSON" {
				status.UnmarshalJSON([]byte(""))
			}

			// THEN the mutex is correctly handed to the ServiceInfo.
			_ = status.GetServiceInfo()
			// AND when the mutex is locked, GetServiceInfo is held.
			unlockedAtChan := make(chan time.Time, 1)
			gotServiceInfoAtChan := make(chan time.Time, 1)
			status.mutex.Lock()
			go func() {
				status.GetServiceInfo()
				gotServiceInfoAtChan <- time.Now()
			}()
			go func() {
				time.Sleep(100 * time.Millisecond)
				status.mutex.Unlock()
				unlockedAtChan <- time.Now()
			}()
			unlockedAt := <-unlockedAtChan
			gotServiceInfoAt := <-gotServiceInfoAtChan
			if gotServiceInfoAt.Before(unlockedAt) {
				t.Errorf("%s\nGetServiceInfo was not held while mutex was locked!\n"+
					"want: %v\ngot:  %v",
					packageName, gotServiceInfoAt, unlockedAt)
			}
		})
	}
}

func TestStatus_Init(t *testing.T) {
	// GIVEN we have a Status.
	tests := map[string]struct {
		shoutrrrs, commands, webhooks int
		serviceID                     string
		webURL                        string
	}{
		"ServiceID": {
			serviceID: "test"},
		"WebURL": {
			webURL: "https://example.com"},
		"shoutrrrs": {
			shoutrrrs: 2},
		"commands": {
			commands: 3},
		"webhooks": {
			webhooks: 4},
		"all": {
			serviceID: "argus",
			webURL:    "https://release-argus.io",
			shoutrrrs: 5,
			commands:  5,
			webhooks:  5},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var status Status

			// WHEN Init is called.
			status.Init(
				tc.shoutrrrs, tc.commands, tc.webhooks,
				tc.serviceID, name, "",
				&dashboard.Options{
					WebURL: tc.webURL})

			// THEN the Status is initialised as expected:
			// 	ServiceID:
			if status.ServiceInfo.ID != tc.serviceID {
				t.Errorf("%s\nServiceID address mismatch\n\nwant: %v\ngot:  %v",
					packageName, &tc.serviceID, status.ServiceInfo.ID)
			}
			// 	Shoutrrr:
			want := 0
			got := status.Fails.Shoutrrr.Length()
			if got != want {
				t.Errorf("%s\nFails.Shoutrrr initial length mismatch\nwant: %d\ngot:  %d",
					packageName, want, got)
			} else {
				for i := 0; i < tc.shoutrrrs; i++ {
					failed := false
					status.Fails.Shoutrrr.Set(fmt.Sprint(i), &failed)
				}
				got := status.Fails.Shoutrrr.Length()
				if got != tc.shoutrrrs {
					t.Errorf("%s\nFails.Shoutrrr capacity mismatch\nwant: %d\ngot:  %d",
						packageName, tc.shoutrrrs, got)
				}
			}
			// 	Command:
			got = status.Fails.Command.Length()
			if got != tc.commands {
				t.Errorf("%s\nFails.Command length mismatch\nwant: %d\ngot:  %d",
					packageName, tc.commands, got)
			}
			// 	WebHook:
			want = 0
			got = status.Fails.WebHook.Length()
			if got != want {
				t.Errorf("%s\nFails.WebHook initial length mismatch\nwant: %d\ngot:  %d",
					packageName, want, got)
			} else {
				for i := 0; i < tc.webhooks; i++ {
					failed := false
					status.Fails.WebHook.Set(fmt.Sprint(i), &failed)
				}
				got := status.Fails.WebHook.Length()
				if got != tc.webhooks {
					t.Errorf("%s\nFails.WebHook capacity mismatch\nwant:  %d\ngot:  %d",
						packageName, tc.webhooks, got)
				}
			}
		})
	}
}

func TestService_ServiceInfo(t *testing.T) {
	// GIVEN a Status.
	status := testStatus()

	id := "test_id"
	status.ServiceInfo.ID = id
	name := "test_name"
	status.ServiceInfo.Name = name
	url := "https://example.com"
	status.ServiceInfo.URL = url

	icon := "https://example.com/icon"
	status.Dashboard.Icon = icon
	iconLinkTo := "https://example.com/icon_link"
	status.Dashboard.IconLinkTo = iconLinkTo
	webURL := "https://example.com/web"
	status.Dashboard.WebURL = webURL

	approvedVersion := "approved.version"
	status.SetApprovedVersion(approvedVersion, false)
	deployedVersion := "deployed.version"
	status.SetDeployedVersion(deployedVersion, "", false)
	latestVersion := "latest.version"
	status.SetLatestVersion(latestVersion, "", false)

	time.Sleep(10 * time.Millisecond)
	time.Sleep(time.Second)

	// When ServiceInfo is called on it.
	got := status.GetServiceInfo()
	want := serviceinfo.ServiceInfo{
		ID:   id,
		Name: name,
		URL:  url,

		Icon:       icon,
		IconLinkTo: iconLinkTo,
		WebURL:     webURL,

		DeployedVersion: deployedVersion,
		ApprovedVersion: approvedVersion,
		LatestVersion:   latestVersion,
	}

	// THEN we get the correct ServiceInfo.
	gotStr := util.ToJSONString(got)
	wantStr := util.ToJSONString(want)
	if gotStr != wantStr {
		t.Errorf("%s\nwant: %#v\ngot:  %#v",
			packageName, wantStr, gotStr)
	}
}

func TestStatus_GetWebURL(t *testing.T) {
	// GIVEN we have a Status.
	latestVersion := "1.2.3"
	tests := map[string]struct {
		webURL string
		want   string
	}{
		"empty string": {
			webURL: "",
			want:   ""},
		"string without templating": {
			webURL: "https://example.com/somewhere",
			want:   "https://example.com/somewhere"},
		"string with templating": {
			webURL: "https://example.com/somewhere/{{ version }}",
			want:   "https://example.com/somewhere/" + latestVersion},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := Status{}
			status.Init(
				0, 0, 0,
				name, name, "",
				&dashboard.Options{
					WebURL: tc.webURL})
			status.SetLatestVersion(latestVersion, "", false)

			// WHEN GetWebURL is called.
			got := status.GetWebURL()

			// THEN the returned WebURL is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestStatus_SetLastQueried(t *testing.T) {
	// GIVEN we have a Status and some webhooks.
	var status Status

	// WHEN we SetLastQueried.
	start := time.Now().UTC()
	status.SetLastQueried("")

	// THEN LastQueried will have been set to the current timestamp.
	lastQueried, _ := time.Parse(time.RFC3339, status.LastQueried())
	since := lastQueried.Sub(start)
	if since > time.Second {
		t.Errorf("%s\nLastQueried was %v ago, not recent enough!",
			packageName, since)
	}
}

func TestStatus_ApprovedVersion(t *testing.T) {
	deployedVersion := "0.0.1"
	latestVersion := "0.0.3"
	// GIVEN a Status.
	tests := map[string]struct {
		hadApprovedVersion            string
		approving                     string
		latestVersionIsDeployedMetric float64
		wantMessages                  int
	}{
		"Same version": {
			hadApprovedVersion:            "0.0.0",
			approving:                     "0.0.0",
			latestVersionIsDeployedMetric: 0,
			wantMessages:                  0,
		},
		"Approving LatestVersion": {
			approving:                     latestVersion,
			latestVersionIsDeployedMetric: 2,
			wantMessages:                  1,
		},
		"Skipping LatestVersion": {
			approving:                     "SKIP_" + latestVersion,
			latestVersionIsDeployedMetric: 3,
			wantMessages:                  1,
		},
		"Approving non-latest version": {
			approving:                     "0.0.2a",
			latestVersionIsDeployedMetric: 0,
			wantMessages:                  1,
		},
	}

	// Changing UpdatesCurrent.
	metricsMutex.RLock()
	t.Cleanup(func() { metricsMutex.RUnlock() })

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			announceChannel := make(chan []byte, 4)
			databaseChannel := make(chan dbtype.Message, 4)
			status := New(
				&announceChannel, &databaseChannel, nil,
				tc.hadApprovedVersion,
				"", "",
				"", "",
				"",
				&dashboard.Options{
					WebURL: "https://example.com"})
			status.Init(
				0, 0, 0,
				"TestStatus_SetApprovedVersion_"+name, "TestStatus_SetApprovedVersion_"+name, "",
				status.Dashboard)
			status.SetLatestVersion(latestVersion, "", false)
			status.SetDeployedVersion(deployedVersion, "", false)

			// WHEN SetApprovedVersion is called.
			status.SetApprovedVersion(tc.approving, true)

			// THEN the Status is as expected:
			// 	ApprovedVersion:
			got := status.ApprovedVersion()
			if got != tc.approving {
				t.Errorf("%s\nApprovedVersion mismatch\nwant: %s\ngot:  %s",
					packageName, tc.approving, got)
			}
			// 	LatestVersion:
			got = status.LatestVersion()
			if got != latestVersion {
				t.Errorf("%s\nLatestVersion mismatch\nwant: %s\ngot:  %s",
					packageName, latestVersion, got)
			}
			// 	DeployedVersion:
			got = status.DeployedVersion()
			if got != deployedVersion {
				t.Errorf("%s\nDeployedVersion mismatch\nwant: %s\ngot:  %s",
					packageName, deployedVersion, got)
			}
			// 	AnnounceChannel:
			if len(*status.AnnounceChannel) != tc.wantMessages {
				t.Errorf("%s\nAnnounceChannel length mismatch\nwant: %d messages\ngot:  %d",
					packageName, tc.wantMessages, len(*status.AnnounceChannel))
			}
			// 	DatabaseChannel:
			if len(*status.DatabaseChannel) != tc.wantMessages {
				t.Errorf("%s\nDatabaseChannel length mismatch\nwant: %d messages\ngot:  %d",
					packageName, tc.wantMessages, len(*status.DatabaseChannel))
			}
			// AND LatestVersionIsDeployedVersion metric is updated.
			gotMetric := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID))
			if gotMetric != tc.latestVersionIsDeployedMetric {
				t.Errorf("%s\nLatestVersionIsDeployedVersion metric mismatch\nwant: %f\ngot:  %f",
					packageName, tc.latestVersionIsDeployedMetric, gotMetric)
			}
		})
	}
}

func TestStatus_DeployedVersion(t *testing.T) {
	type values struct {
		approvedVersion, deployedVersion, latestVersion string
		deployedVersionTimestamp                        string
	}
	type args struct {
		version, timestamp string
	}
	// GIVEN a Status.
	tests := map[string]struct {
		had  values
		args args
		want values
	}{
		"Same version": {
			had: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.1",
				latestVersion:   "0.0.3",
			},
			args: args{
				version:   "0.0.1",
				timestamp: "2020-01-01T00:00:00Z",
			},
			want: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.1",
				latestVersion:   "0.0.3",
			},
		},
		"Deploying ApprovedVersion - DeployedVersion becomes ApprovedVersion and resets ApprovedVersion": {
			had: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.1",
				latestVersion:   "0.0.3",
			},
			args: args{
				version: "0.0.2",
			},
			want: values{
				approvedVersion: "",
				deployedVersion: "0.0.2",
				latestVersion:   "0.0.3",
			},
		},
		"Deploying unknown Version - DeployedVersion becomes this version": {
			had: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.1",
				latestVersion:   "0.0.3",
			},
			args: args{
				version: "0.0.4-dev",
			},
			want: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.4-dev",
				latestVersion:   "0.0.3",
			},
		},
		"Deploying LatestVersion - DeployedVersion becomes this version": {
			had: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.1",
				latestVersion:   "0.0.3",
			},
			args: args{
				version: "0.0.3",
			},
			want: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.3",
				latestVersion:   "0.0.3",
			},
		},
		"Deploying X with timestamp - DeployedVersion and DeployedVersionTimestamp are set": {
			had: values{
				approvedVersion: "0.0.2",
				deployedVersion: "0.0.1",
				latestVersion:   "0.0.3",
			},
			args: args{
				version:   "0.0.4",
				timestamp: "2020-01-01T00:00:00Z",
			},
			want: values{
				approvedVersion:          "0.0.2",
				deployedVersion:          "0.0.4",
				latestVersion:            "0.0.3",
				deployedVersionTimestamp: "2020-01-01T00:00:00Z",
			},
		},
	}

	// Changing UpdatesCurrent.
	metricsMutex.RLock()
	t.Cleanup(func() { metricsMutex.RUnlock() })

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, haveDB := range []bool{false, true} {
				dbChannel := make(chan dbtype.Message, 4)
				status := New(
					nil, &dbChannel, nil,
					tc.had.approvedVersion,
					tc.had.deployedVersion, tc.had.deployedVersionTimestamp,
					tc.had.latestVersion, "",
					"",
					&dashboard.Options{
						WebURL: "https://example.com"})
				if !haveDB {
					status.DatabaseChannel = nil
				}
				status.Init(
					0, 0, 0,
					name, name, "",
					status.Dashboard)

				// WHEN SetDeployedVersion is called on it.
				status.SetDeployedVersion(tc.args.version, tc.args.timestamp,
					haveDB)

				// THEN DeployedVersion is set to this version.
				if status.DeployedVersion() != tc.want.deployedVersion {
					t.Errorf("%s\nDeployedVersion mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want.deployedVersion, status.DeployedVersion())
				}
				if status.ApprovedVersion() != tc.want.approvedVersion {
					t.Errorf("%s\nApprovedVersion mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want.approvedVersion, status.ApprovedVersion())
				}
				if status.LatestVersion() != tc.want.latestVersion {
					t.Errorf("%s\nLatestVersion mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want.latestVersion, status.LatestVersion())
				}
				// AND the DeployedVersionTimestamp is unchanged when DeployedVersion is unchanged.
				if tc.had.deployedVersion == tc.args.version {
					if timestamp := status.DeployedVersionTimestamp(); timestamp != tc.had.deployedVersionTimestamp {
						t.Errorf("%s\nDeployedVersionTimestamp mismatch\nwant: %s (unchanged)\ngot:  %s",
							packageName, tc.had.deployedVersionTimestamp, timestamp)
					}
				} else {
					// AND the DeployedVersionTimestamp is set to the provided date (when provided).
					if tc.want.deployedVersionTimestamp != "" {
						if timestamp := status.DeployedVersionTimestamp(); timestamp != tc.want.deployedVersionTimestamp {
							t.Errorf("%s\nDeployedVersionTimestamp mismatch\nwant: %s\ngot:  %s",
								packageName, tc.want.deployedVersionTimestamp, timestamp)
						}
					} else {
						// AND the DeployedVersionTimestamp is set to current time (when no releaseDate was given).
						d, _ := time.Parse(time.RFC3339, status.DeployedVersionTimestamp())
						since := time.Since(d)
						if since > time.Second {
							t.Errorf("%s\nDeployedVersionTimestamp was %v ago, not recent enough!",
								packageName, since)
						}
					}
				}
				// AND the LatestVersionIsDeployedVersion metric is updated.
				got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(name))
				want := float64(0)
				if haveDB && status.LatestVersion() == status.DeployedVersion() {
					want = 1
				}
				if got != want {
					t.Errorf("%s\nhaveDB=%t LatestVersionIsDeployedVersion metric mismatch\nwant: %f\ngot:  %f",
						packageName, haveDB, want, got)
				}
			}
		})
	}
}

func TestStatus_LatestVersion(t *testing.T) {
	type values struct {
		version, timestamp string
	}
	// GIVEN a Status.
	lastQueried := "2021-01-01T00:00:00Z"
	tests := map[string]struct {
		had, args values
		want      *values // Default to args.
	}{
		"same version": {
			had: values{
				version: "1.2.3", timestamp: "2020-01-01T00:00:00Z",
			},
			args: values{
				version: "1.2.3", timestamp: "2020-01-01T00:00:00Z",
			},
		},
		"timestamp - Empty == Set to lastQueried": {
			had: values{
				version: "0.0.0", timestamp: "2021-01-01T00:00:00Z",
			},
			args: values{
				version: "0.0.1", timestamp: "",
			},
			want: &values{
				version: "0.0.1", timestamp: lastQueried,
			},
		},
		"Timestamp - Given == Set to value given": {
			had: values{
				version: "0.0.0", timestamp: "2022-01-01T00:00:00Z",
			},
			args: values{
				version: "0.0.1", timestamp: "2022-01-01T00:00:00Z",
			},
		},
	}

	// Changing UpdatesCurrent.
	metricsMutex.RLock()
	t.Cleanup(func() { metricsMutex.RUnlock() })

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, haveDB := range []bool{false, true} {
				dbChannel := make(chan dbtype.Message, 8)
				status := New(
					nil, &dbChannel, nil,
					"",
					"", "",
					tc.had.version, tc.had.timestamp,
					lastQueried,
					&dashboard.Options{
						WebURL: "https://example.com"})
				if !haveDB {
					status.DatabaseChannel = nil
				}
				status.Init(
					0, 0, 0,
					name, name, "",
					status.Dashboard)
				if tc.want == nil {
					tc.want = &tc.args
				}

				// WHEN SetLatestVersion is called on it.
				status.SetLatestVersion(tc.args.version, tc.args.timestamp,
					haveDB)

				// THEN LatestVersion is set to this version.
				if version := status.LatestVersion(); version != tc.want.version {
					t.Errorf("%s\nLatestVersion mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want.version, version)
				}
				// AND the LatestVersionTimestamp is set to the current time.
				if timestamp := status.LatestVersionTimestamp(); timestamp != tc.want.timestamp {
					t.Errorf("%s\nhaveDB=%t LatestVersionTimestamp mismatch\nwant: %q\ngot:  %q",
						packageName, haveDB, tc.want.timestamp, timestamp)
				}
				// AND the LatestVersionIsDeployedVersion metric is updated.
				got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(name))
				want := float64(0)
				if haveDB && status.LatestVersion() == status.DeployedVersion() {
					want = 1
				}
				// LatestVersionIsDeployedVersion metric.
				if got != want {
					t.Errorf("%s\nhaveDB=%t LatestVersionIsDeployedVersion metric mismatch\nwant: %f\ngot:  %f",
						packageName, haveDB, want, got)
				}
			}
		})
	}
}

func TestStatus_RegexMissesContent(t *testing.T) {
	// GIVEN a Status.
	status := Status{}

	// WHEN RegexMissContent is called on it.
	status.RegexMissContent()

	// THEN RegexMisses is incremented.
	var want uint = 1
	got := status.RegexMissesContent()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (1 RegexMissContent())\nwant: %d\ngot:  %d",
			packageName, want, got)
	}

	// WHEN RegexMissContent is called on it again.
	status.RegexMissContent()

	// THEN RegexMisses is incremented again.
	want++
	got = status.RegexMissesContent()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (2 RegexMissContent()\nwant: %d\ngot:  %d",
			packageName, want, got)
	}

	// WHEN RegexMissContent is called on it again.
	status.RegexMissContent()

	// THEN RegexMisses is incremented again.
	want++
	got = status.RegexMissesContent()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (3 RegexMissContent())\nwant: %d\ngot:  %d",
			packageName, want, got)
	}

	// WHEN ResetRegexMisses is called on it.
	status.ResetRegexMisses()

	// THEN RegexMisses is reset.
	want = 0
	got = status.RegexMissesContent()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (ResetRegexMisses())\nwant: %d\ngot:  %d",
			packageName, want, got)
	}
}

func TestStatus_RegexMissesVersion(t *testing.T) {
	// GIVEN a Status.
	status := Status{}

	// WHEN RegexMissVersion is called on it.
	status.RegexMissVersion()

	// THEN RegexMisses is incremented.
	var want uint = 1
	got := status.RegexMissesVersion()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (1 RegexMissVersion())\nwant: %d\ngot:  %d",
			packageName, want, got)
	}

	// WHEN RegexMissVersion is called on it again.
	status.RegexMissVersion()

	// THEN RegexMisses is incremented again.
	want = 2
	got = status.RegexMissesVersion()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (2 RegexMissVersion())\nwant: %d\ngot:  %d",
			packageName, want, got)
	}

	// WHEN RegexMissVersion is called on it again.
	status.RegexMissVersion()

	// THEN RegexMisses is incremented again.
	want = 3
	got = status.RegexMissesVersion()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (3 RegexMissVersion())\nwant: %d\ngot:  %d",
			packageName, want, got)
	}

	// WHEN ResetRegexMisses is called on it.
	status.ResetRegexMisses()

	// THEN RegexMisses is reset.
	want = 0
	got = status.RegexMissesVersion()
	if got != want {
		t.Errorf("%s\nRegexMisses mismatch (ResetRegexMisses())\nwant: %d\ngot:  %d",
			packageName, want, got)
	}
}

func TestStatus_SendAnnounce(t *testing.T) {
	// GIVEN a Status with channels.
	tests := map[string]struct {
		deleting   bool
		nilChannel bool
	}{
		"not deleting, or nil channel": {},
		"deleting":                     {deleting: true},
		"nil channel":                  {nilChannel: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			announceChannel := make(chan []byte, 4)
			status := New(
				&announceChannel, nil, nil,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{})
			if tc.nilChannel {
				status.AnnounceChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			// WHEN SendAnnounce is called on it.
			status.SendAnnounce(&[]byte{})

			// THEN the AnnounceChannel is sent a message if not deleting or nil.
			got := 0
			if status.AnnounceChannel != nil {
				got = len(*status.AnnounceChannel)
			}
			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			if got != want {
				t.Errorf("%s\nAnnounceChannel length mismatch\nwant: %d messages\ngot:  %d",
					packageName, want, got)
			}
		})
	}
}

func TestStatus_sendDatabase(t *testing.T) {
	// GIVEN a Status with channels.
	tests := map[string]struct {
		deleting   bool
		nilChannel bool
	}{
		"not deleting, or nil channel": {},
		"deleting":                     {deleting: true},
		"nil channel":                  {nilChannel: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			databaseChannel := make(chan dbtype.Message, 4)
			status := New(
				nil, &databaseChannel, nil,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{})
			if tc.nilChannel {
				status.DatabaseChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			// WHEN sendDatabase is called on it.
			status.sendDatabase(&dbtype.Message{})

			// THEN the DatabaseChannel is sent a message if not deleting or nil.
			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			got := 0
			if status.DatabaseChannel != nil {
				got = len(*status.DatabaseChannel)
			}
			if got != want {
				t.Errorf("%s\nDatabaseChannel length mismatch\nwant: %d messages\ngot:  %d",
					packageName, want, got)
			}
		})
	}
}

func TestStatus_SendSave(t *testing.T) {
	// GIVEN a Status with channels.
	tests := map[string]struct {
		deleting   bool
		nilChannel bool
	}{
		"not deleting, or nil channel": {},
		"deleting":                     {deleting: true},
		"nil channel":                  {nilChannel: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			saveChannel := make(chan bool, 4)
			status := New(
				nil, nil, &saveChannel,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{})
			if tc.nilChannel {
				status.SaveChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			// WHEN SendSave is called on it.
			status.SendSave()

			// THEN the SaveChannel is sent a message if not deleting or nil.
			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			got := 0
			if status.SaveChannel != nil {
				got = len(*status.SaveChannel)
			}
			if got != want {
				t.Errorf("%s\nSaveChannel length mismatch\nwant: %d messages\ngot:  %d",
					packageName, want, got)
			}
		})
	}
}

func TestFails_ResetFails(t *testing.T) {
	// GIVEN a Fails struct.
	tests := map[string]struct {
		commandFails                *[]*bool
		shoutrrrFails, webhookFails *map[string]*bool
	}{
		"all default": {},
		"all empty": {
			commandFails:  &[]*bool{},
			shoutrrrFails: &map[string]*bool{},
			webhookFails:  &map[string]*bool{},
		},
		"only notifies": {
			shoutrrrFails: &map[string]*bool{
				"0": nil,
				"1": test.BoolPtr(false),
				"3": test.BoolPtr(true)},
		},
		"only commands": {
			commandFails: &[]*bool{
				nil,
				test.BoolPtr(false),
				test.BoolPtr(true)},
		},
		"only webhooks": {
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": test.BoolPtr(false),
				"3": test.BoolPtr(true)},
		},
		"all filled": {
			shoutrrrFails: &map[string]*bool{
				"0": nil,
				"1": test.BoolPtr(false),
				"3": test.BoolPtr(true)},
			commandFails: &[]*bool{nil, test.BoolPtr(false), test.BoolPtr(true)},
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": test.BoolPtr(false),
				"3": test.BoolPtr(true)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fails := Fails{}
			if tc.commandFails != nil {
				fails.Command.Init(len(*tc.commandFails))
			}
			if tc.shoutrrrFails != nil {
				fails.Shoutrrr.Init(len(*tc.shoutrrrFails))
			}
			if tc.webhookFails != nil {
				fails.WebHook.Init(len(*tc.webhookFails))
			}

			// WHEN resetFails is called on it.
			fails.resetFails()

			// THEN all the fails become nil.
			if tc.shoutrrrFails != nil {
				for i := range *tc.shoutrrrFails {
					if fails.Shoutrrr.Get(i) != nil {
						t.Errorf("%s\nShoutrrr.Failed[%s] not reset\nwant: nil\ngot:  %t",
							packageName, i, *fails.Shoutrrr.Get(i))
					}
				}
			}
			if tc.commandFails != nil {
				for i := range *tc.commandFails {
					if fails.Command.Get(i) != nil {
						t.Errorf("%s\nCommand.Failed[%d] not reset\nwant: nil\ngot:  %t",
							packageName, i, *fails.Command.Get(i))
					}
				}
			}
			if tc.webhookFails != nil {
				for i := range *tc.webhookFails {
					if fails.WebHook.Get(i) != nil {
						t.Errorf("%s\nWebHook.Failed[%s] not reset\nwant: nil\ngot:  %t",
							packageName, i, *fails.WebHook.Get(i))
					}
				}
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	// GIVEN a Status.
	tests := map[string]struct {
		status                                 *Status
		regexMissesContent, regexMissesVersion int
		want                                   string
	}{
		"empty status": {
			status: &Status{},
			want:   "",
		},
		"only fails": {
			status: &Status{
				Fails: Fails{
					Shoutrrr: FailsShoutrrr{
						failsBase: failsBase{
							fails: map[string]*bool{
								"bash": test.BoolPtr(false),
								"bish": nil,
								"bosh": test.BoolPtr(true)}}},
					Command: FailsCommand{
						fails: []*bool{
							nil,
							test.BoolPtr(false),
							test.BoolPtr(true)}},
					WebHook: FailsWebHook{
						failsBase: failsBase{
							fails: map[string]*bool{
								"bar": nil,
								"foo": test.BoolPtr(false)}}}},
			},
			want: test.TrimYAML(`
				fails:
					shoutrrr:
						bash: false
						bish: nil
						bosh: true
					command:
						- 0: nil
						- 1: false
						- 2: true
					webhook:
						bar: nil
						foo: false
			`),
		},
		"all fields": {
			regexMissesContent: 1,
			regexMissesVersion: 2,
			status: &Status{
				ServiceInfo: serviceinfo.ServiceInfo{
					ApprovedVersion: "1.2.4",
					DeployedVersion: "1.2.3",
					LatestVersion:   "1.2.4"},
				deployedVersionTimestamp: "2022-01-01T01:01:02Z",
				latestVersionTimestamp:   "2022-01-01T01:01:01Z",
				lastQueried:              "2022-01-01T01:01:01Z",
				Fails: Fails{
					Shoutrrr: FailsShoutrrr{
						failsBase: failsBase{
							fails: map[string]*bool{
								"bish": nil,
								"bash": test.BoolPtr(false),
								"bosh": test.BoolPtr(true)}}},
					Command: FailsCommand{
						fails: []*bool{
							nil,
							test.BoolPtr(false),
							test.BoolPtr(true)}},
					WebHook: FailsWebHook{
						failsBase: failsBase{
							fails: map[string]*bool{
								"foo": test.BoolPtr(false),
								"bar": nil}}}},
			},
			want: test.TrimYAML(`
				approved_version: 1.2.4
				deployed_version: 1.2.3
				deployed_version_timestamp: 2022-01-01T01:01:02Z
				latest_version: 1.2.4
				latest_version_timestamp: 2022-01-01T01:01:01Z
				last_queried: 2022-01-01T01:01:01Z
				regex_misses_content: 1
				regex_misses_version: 2
				fails:
					shoutrrr:
						bash: false
						bish: nil
						bosh: true
					command:
						- 0: nil
						- 1: false
						- 2: true
					webhook:
						bar: nil
						foo: false
			`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			{ // RegEz misses.
				for i := 0; i < tc.regexMissesContent; i++ {
					tc.status.RegexMissContent()
				}
				for i := 0; i < tc.regexMissesVersion; i++ {
					tc.status.RegexMissVersion()
				}
			}

			// WHEN the Status is stringified with String.
			got := tc.status.String()

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestStatus_SetLatestVersionIsDeployedMetric(t *testing.T) {
	// GIVEN a Status.
	tests := map[string]struct {
		serviceID                                       string
		approvedVersion, latestVersion, deployedVersion string
		isSkipped                                       bool
		had, want                                       float64
	}{
		"latest version is deployed": {
			latestVersion:   "1.2.3",
			deployedVersion: "1.2.3",
			want:            1,
		},
		"latest version is not deployed": {
			latestVersion:   "1.2.3",
			deployedVersion: "1.2.4",
			want:            0,
		},
		"latest version is not deployed, but is approved": {
			approvedVersion: "1.2.3",
			latestVersion:   "1.2.3",
			deployedVersion: "1.2.4",
			want:            2,
		},
		"latest version is not deployed, but is skipped": {
			approvedVersion: "SKIP_1.2.3",
			latestVersion:   "1.2.3",
			deployedVersion: "1.2.4",
			want:            3,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := New(
				nil, nil, nil,
				tc.approvedVersion,
				tc.deployedVersion, testStatus().deployedVersionTimestamp,
				tc.latestVersion, "",
				"",
				&dashboard.Options{
					WebURL: "https://example.com"})
			status.Init(
				0, 0, 0,
				name, name, "",
				status.Dashboard)

			// WHEN setLatestVersion is called on it.
			status.setLatestVersionIsDeployedMetric()
			got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(name))

			// THEN the metric is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %f\ngot:  %f",
					packageName, tc.want, got)
			}
		})
	}
}

func TestStatus_InitMetrics_DeleteMetrics(t *testing.T) {
	type versions struct {
		latestVersion, deployedVersion, approvedVersion string
	}
	// GIVEN a Status.
	tests := map[string]struct {
		versions                versions
		latestVersionIsDeployed float64
	}{
		"latest version is not deployed": {
			versions: versions{
				latestVersion:   "0.0.1",
				deployedVersion: "0.0.0",
			},
			latestVersionIsDeployed: 0,
		},
		"latest version is deployed": {
			versions: versions{
				latestVersion:   "0.0.0",
				deployedVersion: "0.0.0",
			},
			latestVersionIsDeployed: 1,
		},
		"latest version is approved": {
			versions: versions{
				approvedVersion: "0.0.1",
				latestVersion:   "0.0.1",
				deployedVersion: "0.0.0",
			},
			latestVersionIsDeployed: 2,
		},
		"latest version is skipped": {
			versions: versions{
				approvedVersion: "SKIP_0.0.1",
				latestVersion:   "0.0.1",
				deployedVersion: "0.0.0",
			},
			latestVersionIsDeployed: 3,
		},
	}

	// Changing and reading UpdatesCurrent.
	metricsMutex.Lock()
	t.Cleanup(func() { metricsMutex.Unlock() })

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			// Reset global metrics.
			metric.InitMetrics()

			status := New(
				nil, nil, nil,
				tc.versions.approvedVersion,
				tc.versions.deployedVersion, "",
				tc.versions.latestVersion, "",
				"",
				&dashboard.Options{
					WebURL: "https://example.com"})
			status.Init(
				0, 0, 0,
				name, name, "",
				status.Dashboard)
			var updatesCurrentAvailableDelta float64
			var updatesCurrentSkippedDelta float64
			switch tc.latestVersionIsDeployed {
			case 0, 2:
				updatesCurrentAvailableDelta = 1
			case 3:
				updatesCurrentAvailableDelta = 1
				updatesCurrentSkippedDelta = 1
			}
			hadUpdatesCurrentAvailable := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			hadUpdatesCurrentSkipped := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))

			// WHEN InitMetrics is called on it.
			status.InitMetrics()

			// THEN the metrics are created.
			want := tc.latestVersionIsDeployed
			got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID))
			if got != want {
				t.Errorf("%s\nLatestVersionIsDeployed mismatch after InitMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = hadUpdatesCurrentAvailable + updatesCurrentAvailableDelta
			got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			if got != want {
				t.Logf("%s - hadUpdatesCurrentAvailable=%f, updatesCurrentAvailableDelta=%f",
					packageName, hadUpdatesCurrentAvailable, updatesCurrentAvailableDelta)
				t.Errorf("%s\nUpdatesCurrent('AVAILABLE') mismatch after InitMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = hadUpdatesCurrentSkipped + updatesCurrentSkippedDelta
			got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))
			if got != want {
				t.Logf("%s\nhadUpdatesCurrentSkipped=%f, updatesCurrentSkippedDelta=%f",
					packageName, hadUpdatesCurrentSkipped, updatesCurrentSkippedDelta)
				t.Errorf("%s\nUpdatesCurrent('SKIPPED') mismatch after InitMetrics()\nwant: %f\ngot:  %f",
					packageName, updatesCurrentSkippedDelta, got)
			}

			// WHEN DeleteMetrics is called on it.
			status.DeleteMetrics()

			// THEN the metrics are deleted.
			want = 0
			got = testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(status.ServiceInfo.ID))
			if got != want {
				t.Errorf("%s\nLatestVersionIsDeployed mismatch after DeleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = hadUpdatesCurrentAvailable
			got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			if got != want {
				t.Errorf("%s\nUpdatesCurrent('AVAILABLE') mismatch after DeleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = hadUpdatesCurrentSkipped
			got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))
			if got != want {
				t.Errorf("%s\nUpdatesCurrent('SKIPPED') mismatch after DeleteMetrics()\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
		})
	}
}

func TestNewDefaults(t *testing.T) {
	// GIVEN we have channels.
	announceChannel := make(chan []byte, 4)
	databaseChannel := make(chan dbtype.Message, 4)
	saveChannel := make(chan bool, 4)

	// WHEN NewDefaults is called.
	statusDefaults := NewDefaults(&announceChannel, &databaseChannel, &saveChannel)

	// THEN the AnnounceChannel is set to the given channel.
	if &announceChannel != statusDefaults.AnnounceChannel {
		t.Errorf("%s\nAnnounceChannel not initialised correctly.\nwant: %v\ngot:  %v",
			packageName, &announceChannel, statusDefaults.AnnounceChannel)
	}
	// AND the DatabaseChannel is set to the given channel.
	if &databaseChannel != statusDefaults.DatabaseChannel {
		t.Errorf("%s\nDatabaseChannel not initialised correctly.\nwant: %v\ngot:  %v",
			packageName, &databaseChannel, statusDefaults.DatabaseChannel)
	}
	// AND the SaveChannel is set to the given channel.
	if &saveChannel != statusDefaults.SaveChannel {
		t.Errorf("%s\nSaveChannel not initialised correctly.\nwant: %v\ngot:  %v",
			packageName, &saveChannel, statusDefaults.SaveChannel)
	}
}

func TestStatus_Copy(t *testing.T) {
	// GIVEN a Status.
	announceChannel := make(chan []byte, 4)
	databaseChannel := make(chan dbtype.Message, 4)
	saveChannel := make(chan bool, 4)
	approvedVersion := "1.0.0"
	deployedVersion := "0.9.0"
	deployedVersionTimestamp := "2023-01-01T00:00:00Z"
	latestVersion := "1.0.0"
	latestVersionTimestamp := "2023-01-02T00:00:00Z"
	lastQueried := "2023-01-03T00:00:00Z"
	status := New(
		&announceChannel, &databaseChannel, &saveChannel,
		approvedVersion,
		deployedVersion, deployedVersionTimestamp,
		latestVersion, latestVersionTimestamp,
		lastQueried,
		&dashboard.Options{})

	tests := map[string]struct {
		copyChannels bool
	}{
		"copy channels":       {copyChannels: true},
		"don't copy channels": {copyChannels: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN Copy is called on it.
			copiedStatus := status.Copy(tc.copyChannels)

			// THEN the copied Status should have the same values as the original.
			if copiedStatus.ServiceInfo.ApprovedVersion != status.ServiceInfo.ApprovedVersion {
				t.Errorf("%s\nApprovedVersion not copied correctly\nwant: %v\ngot:  %v",
					packageName, status.ServiceInfo.ApprovedVersion, copiedStatus.ServiceInfo.ApprovedVersion)
			}
			if copiedStatus.ServiceInfo.DeployedVersion != status.ServiceInfo.DeployedVersion {
				t.Errorf("%s\nSeployedVersion not copied correctly\nwant: %v\ngot:  %v",
					packageName, status.ServiceInfo.DeployedVersion, copiedStatus.ServiceInfo.DeployedVersion)
			}
			if copiedStatus.deployedVersionTimestamp != status.deployedVersionTimestamp {
				t.Errorf("%s\ndeployedVersionTimestamp not copied correctly\nwant: %v\ngot:  %v",
					packageName, status.deployedVersionTimestamp, copiedStatus.deployedVersionTimestamp)
			}
			if copiedStatus.ServiceInfo.LatestVersion != status.ServiceInfo.LatestVersion {
				t.Errorf("%s\nLatestVersion not copied correctly\nwant: %v\ngot:  %v",
					packageName, status.ServiceInfo.LatestVersion, copiedStatus.ServiceInfo.LatestVersion)
			}
			if copiedStatus.latestVersionTimestamp != status.latestVersionTimestamp {
				t.Errorf("%s\nlatestVersionTimestamp not copied correctly\nwant: %v\ngot:  %v",
					packageName, status.latestVersionTimestamp, copiedStatus.latestVersionTimestamp)
			}
			if copiedStatus.lastQueried != status.lastQueried {
				t.Errorf("%s\nlastQueried not copied correctly\nwant: %v\ngot:  %v",
					packageName, status.lastQueried, copiedStatus.lastQueried)
			}

			// AND the channels are only copied over if copyChannels is true.
			if tc.copyChannels {
				if copiedStatus.AnnounceChannel != status.AnnounceChannel {
					t.Errorf("%s\nAnnounceChannel not copied correctly\nwant: %v\ngot:  %v",
						packageName, status.AnnounceChannel, copiedStatus.AnnounceChannel)
				}
				if copiedStatus.DatabaseChannel != status.DatabaseChannel {
					t.Errorf("%s\nDatabaseChannel not copied correctly\nwant: %v\ngot:  %v",
						packageName, status.DatabaseChannel, copiedStatus.DatabaseChannel)
				}
				if copiedStatus.SaveChannel != status.SaveChannel {
					t.Errorf("%s\nSaveChannel not copied correctly\nwant: %v\ngot:  %v",
						packageName, status.SaveChannel, copiedStatus.SaveChannel)
				}
			} else {
				if copiedStatus.AnnounceChannel != nil {
					t.Errorf("%s\nAnnounceChannel should not be copied\nwant: nil\ngot:  %v",
						packageName, copiedStatus.AnnounceChannel)
				}
				if copiedStatus.DatabaseChannel != nil {
					t.Errorf("%s\nDatabaseChannel should not be copied\nwant: nil\ngot:  %v",
						packageName, copiedStatus.DatabaseChannel)
				}
				if copiedStatus.SaveChannel != nil {
					t.Errorf("%s\nSaveChannel should not be copied\nwant: nil\ngot:  %v",
						packageName, copiedStatus.SaveChannel)
				}
			}
		})
	}
}

func TestStatus_SetAnnounceChannel(t *testing.T) {
	// GIVEN a Status with an initial AnnounceChannel.
	initialChannel := make(chan []byte, 4)
	status := New(
		&initialChannel, nil, nil,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{})

	// WHEN SetAnnounceChannel is called with a new channel.
	newChannel := make(chan []byte, 4)
	status.SetAnnounceChannel(&newChannel)

	// THEN the AnnounceChannel should be updated to the new channel.
	if &newChannel != status.AnnounceChannel {
		t.Errorf("%s\nAnnounceChannel not set correctly.\nwant: %v\ngot:  %v",
			packageName, &newChannel, status.AnnounceChannel)
	}

	// AND the initial channel should no longer be the AnnounceChannel.
	if &initialChannel == status.AnnounceChannel {
		t.Errorf("%s\nAnnounceChannel shouldn't have been reset to be the initial channel.\nwant: %v\ngot:  %v",
			packageName, &newChannel, status.AnnounceChannel)
	}
}

func TestStatus_SetDeleting(t *testing.T) {
	// GIVEN a Status.
	status := Status{}

	// WHEN SetDeleting is called on it.
	status.SetDeleting()

	// THEN the deleting flag should be set to true.
	if !status.Deleting() {
		t.Errorf("%s\ndeleting flag mismatch\nwant: true\ngot:  false",
			packageName)
	}

	// WHEN SetDeleting is called on it again.
	status.SetDeleting()

	// THEN the deleting flag should still be true.
	if !status.Deleting() {
		t.Errorf("%s\ndeleting flag mismatch after SetDeleting()\nwant: true\ngot:  false",
			packageName)
	}
}

func TestStatus_SameVersions(t *testing.T) {
	type versions struct {
		approvedVersion, deployedVersion, latestVersion string
	}
	// GIVEN different Status version combinations.
	tests := []struct {
		name             string
		status1, status2 versions
		expected         bool
	}{
		{
			name: "identical versions",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			expected: true,
		},
		{
			name: "different approved version",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.1.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			expected: false,
		},
		{
			name: "different deployed version",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "1.0.0",
				latestVersion:   "1.0.0",
			},
			expected: false,
		},
		{
			name: "different latest version",
			status1: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.0.0",
			},
			status2: versions{
				approvedVersion: "1.0.0",
				deployedVersion: "0.9.0",
				latestVersion:   "1.1.0",
			},
			expected: false,
		},
		{
			name: "all empty versions match",
			status1: versions{
				approvedVersion: "",
				deployedVersion: "",
				latestVersion:   "",
			},
			status2: versions{
				approvedVersion: "",
				deployedVersion: "",
				latestVersion:   "",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status1 := New(
				nil, nil, nil,
				tc.status1.approvedVersion,
				tc.status1.deployedVersion, "",
				tc.status1.latestVersion, "",
				"",
				&dashboard.Options{})

			status2 := New(
				nil, nil, nil,
				tc.status2.approvedVersion,
				tc.status2.deployedVersion, "",
				tc.status2.latestVersion, "",
				"",
				&dashboard.Options{})

			// WHEN comparing versions.
			result := status1.SameVersions(status2)

			// THEN the result matches expected.
			if result != tc.expected {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, tc.expected, result)
			}
		})
	}
}

func TestUpdateUpdatesCurrent(t *testing.T) {
	type versions struct {
		approvedVersion, deployedVersion, latestVersion string
	}
	// GIVEN a Status that has just had a version change.
	tests := map[string]struct {
		previousVersions     versions
		newVersions          versions
		updateCountAvailable float64
		updateCountSkipped   float64
	}{
		"0 to 1 - Latest version not deployed/approved/skipped -> Latest version deployed": {
			previousVersions: versions{
				approvedVersion: "",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				deployedVersion: "1.2.0",
			},
			updateCountAvailable: -1,
			updateCountSkipped:   0,
		},
		"0 to 2 - Latest version not deployed/approved/skipped -> Latest version approved": {
			previousVersions: versions{
				approvedVersion: "",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				approvedVersion: "1.2.0",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   0,
		},
		"0 to 3 - Latest version not deployed/approved/skipped -> Latest version skipped": {
			previousVersions: versions{
				approvedVersion: "",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				approvedVersion: "SKIP_1.2.0",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   1,
		},
		"1 to 0 - Latest version deployed -> Latest version not deployed/approved/skipped": {
			previousVersions: versions{
				approvedVersion: "",
				deployedVersion: "1.2.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				latestVersion: "1.3.0",
			},
			updateCountAvailable: 1,
			updateCountSkipped:   0,
		},
		// Cannot go from deployed to approved/skipped without first being available.
		// "1 to 2 - Latest version deployed -> Latest version approved": {}.
		// "1 to 3 - Latest version deployed -> Latest version skipped": {}.
		"2 to 0 - Latest version approved -> Latest version not deployed/approved/skipped": {
			previousVersions: versions{
				approvedVersion: "1.2.0",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				approvedVersion: "",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   0,
		},
		"2 to 1 - Latest version approved -> Latest version deployed": {
			previousVersions: versions{
				approvedVersion: "1.2.0",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				deployedVersion: "1.2.0",
			},
			updateCountAvailable: -1,
			updateCountSkipped:   0,
		},
		"2 to 3 - Latest version approved -> Latest version skipped": {
			previousVersions: versions{
				approvedVersion: "1.2.0",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				approvedVersion: "SKIP_1.2.0",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   1,
		},
		"3 to 0 - Latest version skipped -> Latest version not deployed/approved/skipped": {
			previousVersions: versions{
				approvedVersion: "SKIP_1.2.0",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				approvedVersion: "",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   -1,
		},
		"3 to 1 - Latest version skipped -> Latest version deployed": {
			previousVersions: versions{
				approvedVersion: "SKIP_1.2.0",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				deployedVersion: "1.2.0",
			},
			updateCountAvailable: -1,
			updateCountSkipped:   -1,
		},
		"3 to 2 - Latest version skipped -> Latest version approved": {
			previousVersions: versions{
				approvedVersion: "SKIP_1.2.0",
				deployedVersion: "1.1.0",
				latestVersion:   "1.2.0",
			},
			newVersions: versions{
				approvedVersion: "1.2.0",
			},
			updateCountAvailable: 0,
			updateCountSkipped:   -1,
		},
	}

	// Changing and reading UpdatesCurrent.
	metricsMutex.Lock()
	t.Cleanup(func() { metricsMutex.Unlock() })

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			tc.newVersions.deployedVersion = util.ValueOrValue(tc.newVersions.deployedVersion, tc.previousVersions.deployedVersion)
			tc.newVersions.latestVersion = util.ValueOrValue(tc.newVersions.latestVersion, tc.previousVersions.latestVersion)
			status := New(nil, nil, nil,
				tc.newVersions.approvedVersion,
				tc.newVersions.deployedVersion, "",
				tc.newVersions.latestVersion, "",
				"",
				&dashboard.Options{})
			hadUpdatesCurrentAvailable := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			hadUpdatesCurrentSkipped := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))

			// WHEN updateUpdatesCurrent is called with the previous and new versions.
			status.updateUpdatesCurrent(
				tc.previousVersions.approvedVersion,
				tc.previousVersions.latestVersion,
				tc.previousVersions.deployedVersion)

			// Validate the update counts for both the approved and skipped metrics.
			// For this, we assume that `SetUpdatesCurrent` has been correctly implemented,
			// and the metrics have been updated accordingly.
			want := hadUpdatesCurrentAvailable + tc.updateCountAvailable
			got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE"))
			if got != want {
				t.Errorf("%s\navailable count mismatch\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
			want = hadUpdatesCurrentSkipped + tc.updateCountSkipped
			got = testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED"))
			if got != want {
				t.Errorf("%s\nskipped count mismatch\nwant: %f\ngot:  %f",
					packageName, want, got)
			}
		})
	}
}
