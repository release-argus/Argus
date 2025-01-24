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
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

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
				&tc.serviceID, &name,
				&tc.webURL)

			// THEN the Status is initialised as expected:
			// 	ServiceID
			if status.ServiceID != &tc.serviceID {
				t.Errorf("ServiceID not initialised to address of %s (%v). Got %v",
					tc.serviceID, &tc.serviceID, status.ServiceID)
			}
			// 	WebURL
			if status.WebURL != &tc.webURL {
				t.Errorf("WebURL not initialised to address of %s (%v). Got %v",
					tc.webURL, &tc.webURL, status.WebURL)
			}
			// 	Shoutrrr
			got := status.Fails.Shoutrrr.Length()
			if got != 0 {
				t.Errorf("Fails.Shoutrrr was initialised to %d. Want %d",
					got, 0)
			} else {
				for i := 0; i < tc.shoutrrrs; i++ {
					failed := false
					status.Fails.Shoutrrr.Set(fmt.Sprint(i), &failed)
				}
				got := status.Fails.Shoutrrr.Length()
				if got != tc.shoutrrrs {
					t.Errorf("Fails.Shoutrrr wanted capacity for %d, but only got to %d",
						tc.shoutrrrs, got)
				}
			}
			// 	Command
			got = status.Fails.Command.Length()
			if got != tc.commands {
				t.Errorf("Fails.Command was initialised to %d. Want %d",
					got, tc.commands)
			}
			// 	WebHook
			got = status.Fails.WebHook.Length()
			if got != 0 {
				t.Errorf("Fails.WebHook was initialised to %d. Want %d",
					got, 0)
			} else {
				for i := 0; i < tc.webhooks; i++ {
					failed := false
					status.Fails.WebHook.Set(fmt.Sprint(i), &failed)
				}
				got := status.Fails.WebHook.Length()
				if got != tc.webhooks {
					t.Errorf("Fails.WebHook wanted capacity for %d, but only got to %d",
						tc.webhooks, got)
				}
			}
		})
	}
}

func TestStatus_GetWebURL(t *testing.T) {
	nilValue := "<nil>"
	// GIVEN we have a Status.
	latestVersion := "1.2.3"
	tests := map[string]struct {
		webURL *string
		want   string
	}{
		"nil string": {
			webURL: nil,
			want:   nilValue},
		"empty string": {
			webURL: test.StringPtr(""),
			want:   nilValue},
		"string without templating": {
			webURL: test.StringPtr("https://something.com/somewhere"),
			want:   "https://something.com/somewhere"},
		"string with templating": {
			webURL: test.StringPtr("https://something.com/somewhere/{{ version }}"),
			want:   "https://something.com/somewhere/" + latestVersion},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := Status{}
			status.Init(
				0, 0, 0,
				&name, &name,
				tc.webURL)
			status.SetLatestVersion(latestVersion, "", false)

			// WHEN GetWebURL is called.
			got := status.GetWebURL()

			// THEN the returned WebURL is as expected.
			gotStr := util.DereferenceOrNilValue(got, nilValue)
			if gotStr != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, gotStr)
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
		t.Errorf("LastQueried was %v ago, not recent enough!",
			since)
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
				"")
			status.Init(
				0, 0, 0,
				test.StringPtr("TestStatus_SetApprovedVersion_"+name), test.StringPtr("TestStatus_SetApprovedVersion_"+name),
				test.StringPtr("https://example.com"))
			status.SetLatestVersion(latestVersion, "", false)
			status.SetDeployedVersion(deployedVersion, "", false)

			// WHEN SetApprovedVersion is called.
			status.SetApprovedVersion(tc.approving, true)

			// THEN the Status is as expected:
			// 	ApprovedVersion
			got := status.ApprovedVersion()
			if got != tc.approving {
				t.Errorf("ApprovedVersion not set to %s. Got %s",
					tc.approving, got)
			}
			// 	LatestVersion
			got = status.LatestVersion()
			if got != latestVersion {
				t.Errorf("LatestVersion not set to %s. Got %s",
					latestVersion, got)
			}
			// 	DeployedVersion
			got = status.DeployedVersion()
			if got != deployedVersion {
				t.Errorf("DeployedVersion not set to %s. Got %s",
					deployedVersion, got)
			}
			// 	AnnounceChannel
			if len(*status.AnnounceChannel) != tc.wantMessages {
				t.Errorf("AnnounceChannel should have %d message(s), but has %d",
					tc.wantMessages, len(*status.AnnounceChannel))
			}
			// 	DatabaseChannel
			if len(*status.DatabaseChannel) != tc.wantMessages {
				t.Errorf("DatabaseChannel should have %d message(s), but has %d",
					tc.wantMessages, len(*status.DatabaseChannel))
			}
			// AND LatestVersionIsDeployedVersion metric is updated.
			gotMetric := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(*status.ServiceID))
			if gotMetric != tc.latestVersionIsDeployedMetric {
				t.Errorf("LatestVersionIsDeployedVersion metric should be %f, not %f",
					tc.latestVersionIsDeployedMetric, gotMetric)
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
					"")
				if !haveDB {
					status.DatabaseChannel = nil
				}
				status.Init(
					0, 0, 0,
					&name, &name,
					test.StringPtr("https://example.com"))

				// WHEN SetDeployedVersion is called on it.
				status.SetDeployedVersion(tc.args.version, tc.args.timestamp,
					haveDB)

				// THEN DeployedVersion is set to this version.
				if status.DeployedVersion() != tc.want.deployedVersion {
					t.Errorf("Expected DeployedVersion to be set to %q, not %q",
						tc.want.deployedVersion, status.DeployedVersion())
				}
				if status.ApprovedVersion() != tc.want.approvedVersion {
					t.Errorf("Expected ApprovedVersion to be set to %q, not %q",
						tc.want.approvedVersion, status.ApprovedVersion())
				}
				if status.LatestVersion() != tc.want.latestVersion {
					t.Errorf("Expected LatestVersion to be set to %q, not %q",
						tc.want.latestVersion, status.LatestVersion())
				}
				// AND the DeployedVersionTimestamp is unchanged when DeployedVersion is unchanged.
				if tc.had.deployedVersion == tc.args.version {
					if timestamp := status.DeployedVersionTimestamp(); timestamp != tc.had.deployedVersionTimestamp {
						t.Errorf("Expected DeployedVersionTimestamp to be unchanged\n%s\ngot:\n%s",
							tc.had.deployedVersionTimestamp, timestamp)
					}
				} else {
					// AND the DeployedVersionTimestamp is set to the provided date (when provided).
					if tc.want.deployedVersionTimestamp != "" {
						if timestamp := status.DeployedVersionTimestamp(); timestamp != tc.want.deployedVersionTimestamp {
							t.Errorf("Expected DeployedVersionTimestamp to be set to\n%s\ngot:\n%s",
								tc.want.deployedVersionTimestamp, timestamp)
						}
					} else {
						// AND the DeployedVersionTimestamp is set to current time (when no releaseDate was given).
						d, _ := time.Parse(time.RFC3339, status.DeployedVersionTimestamp())
						since := time.Since(d)
						if since > time.Second {
							t.Errorf("DeployedVersionTimestamp was %v ago, not recent enough!",
								since)
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
					t.Errorf("haveDB=%t LatestVersionIsDeployedVersion metric should be %f, not %f",
						haveDB, want, got)
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
					lastQueried)
				if !haveDB {
					status.DatabaseChannel = nil
				}
				status.Init(
					0, 0, 0,
					&name, &name,
					test.StringPtr("https://example.com"))
				if tc.want == nil {
					tc.want = &tc.args
				}

				// WHEN SetLatestVersion is called on it.
				status.SetLatestVersion(tc.args.version, tc.args.timestamp,
					haveDB)

				// THEN LatestVersion is set to this version.
				if version := status.LatestVersion(); version != tc.want.version {
					t.Errorf("Expected LatestVersion to be set to %q, not %q",
						version, status.LatestVersion())
				}
				// AND the LatestVersionTimestamp is set to the current time.
				if timestamp := status.LatestVersionTimestamp(); timestamp != tc.want.timestamp {
					t.Errorf("haveDB=%t LatestVersionTimestamp should have been set to \n%q, not \n%q",
						haveDB, timestamp, status.LatestVersionTimestamp())
				}
				// AND the LatestVersionIsDeployedVersion metric is updated.
				got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(name))
				want := float64(0)
				if haveDB && status.LatestVersion() == status.DeployedVersion() {
					want = 1
				}
				// LatestVersionIsDeployedVersion incorrect?
				if got != want {
					t.Errorf("haveDB=%t LatestVersionIsDeployedVersion metric should be %f, not %f",
						haveDB, want, got)
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
	got := status.RegexMissesContent()
	if got != 1 {
		t.Errorf("Expected RegexMisses to be 1, not %d",
			got)
	}

	// WHEN RegexMissContent is called on it again.
	status.RegexMissContent()

	// THEN RegexMisses is incremented again.
	got = status.RegexMissesContent()
	if got != 2 {
		t.Errorf("Expected RegexMisses to be 2, not %d",
			got)
	}

	// WHEN RegexMissContent is called on it again.
	status.RegexMissContent()

	// THEN RegexMisses is incremented again.
	got = status.RegexMissesContent()
	if got != 3 {
		t.Errorf("Expected RegexMisses to be 3, not %d",
			got)
	}

	// WHEN ResetRegexMisses is called on it.
	status.ResetRegexMisses()

	// THEN RegexMisses is reset.
	got = status.RegexMissesContent()
	if got != 0 {
		t.Errorf("Expected RegexMisses to be 0 after ResetRegexMisses, not %d",
			got)
	}
}

func TestStatus_RegexMissesVersion(t *testing.T) {
	// GIVEN a Status.
	status := Status{}

	// WHEN RegexMissVersion is called on it.
	status.RegexMissVersion()

	// THEN RegexMisses is incremented.
	got := status.RegexMissesVersion()
	if got != 1 {
		t.Errorf("Expected RegexMisses to be 1, not %d",
			got)
	}

	// WHEN RegexMissVersion is called on it again.
	status.RegexMissVersion()

	// THEN RegexMisses is incremented again.
	got = status.RegexMissesVersion()
	if got != 2 {
		t.Errorf("Expected RegexMisses to be 2, not %d",
			got)
	}

	// WHEN RegexMissVersion is called on it again.
	status.RegexMissVersion()

	// THEN RegexMisses is incremented again.
	got = status.RegexMissesVersion()
	if got != 3 {
		t.Errorf("Expected RegexMisses to be 3, not %d",
			got)
	}

	// WHEN ResetRegexMisses is called on it.
	status.ResetRegexMisses()

	// THEN RegexMisses is reset.
	got = status.RegexMissesVersion()
	if got != 0 {
		t.Errorf("Expected RegexMisses to be 0 after ResetRegexMisses, not %d",
			got)
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
				"", "", "", "", "", "")
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
				t.Errorf("Expected %d messages on AnnounceChannel, not %d",
					want, got)
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
				"", "", "", "", "", "")
			if tc.nilChannel {
				status.DatabaseChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			// WHEN sendDatabase is called on it.
			status.sendDatabase(&dbtype.Message{})

			// THEN the DatabaseChannel is sent a message if not deleting or nil.
			got := 0
			if status.DatabaseChannel != nil {
				got = len(*status.DatabaseChannel)
			}
			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			if got != want {
				t.Errorf("Expected %d messages on DatabaseChannel, not %d",
					want, got)
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
				"", "", "", "", "", "")
			if tc.nilChannel {
				status.SaveChannel = nil
			}
			if tc.deleting {
				status.SetDeleting()
			}

			// WHEN SendSave is called on it.
			status.SendSave()

			// THEN the SaveChannel is sent a message if not deleting or nil.
			got := 0
			if status.SaveChannel != nil {
				got = len(*status.SaveChannel)
			}
			want := 1
			if tc.deleting || tc.nilChannel {
				want = 0
			}
			if got != want {
				t.Errorf("Expected %d messages on SaveChannel, not %d",
					want, got)
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
						t.Errorf("Shoutrrr.Failed[%s] should have been reset to nil and not be %t",
							i, *fails.Shoutrrr.Get(i))
					}
				}
			}
			if tc.commandFails != nil {
				for i := range *tc.commandFails {
					if fails.Command.Get(i) != nil {
						t.Errorf("Command.Failed[%d] should have been reset to nil and not be %t",
							i, *fails.Command.Get(i))
					}
				}
			}
			if tc.webhookFails != nil {
				for i := range *tc.webhookFails {
					if fails.WebHook.Get(i) != nil {
						t.Errorf("WebHook.Failed[%s] should have been reset to nil and not be %t",
							i, *fails.WebHook.Get(i))
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
				approvedVersion:          "1.2.4",
				deployedVersion:          "1.2.3",
				deployedVersionTimestamp: "2022-01-01T01:01:02Z",
				latestVersion:            "1.2.4",
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
				t.Errorf("Status.String() mismatch\n%q\ngot:\n%q",
					tc.want, got)
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
				"")
			status.Init(
				0, 0, 0,
				&name, &name,
				test.StringPtr("https://example.com"))

			// WHEN setLatestVersion is called on it.
			status.setLatestVersionIsDeployedMetric()
			got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(name))

			// THEN the metric is as expected.
			if got != tc.want {
				t.Errorf("Expected SetLatestVersionIsDeployedMetric to be %f, not %f",
					tc.want, got)
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
				"")
			status.Init(
				0, 0, 0,
				&name, &name,
				test.StringPtr("https://example.com"))
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
			if got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(*status.ServiceID)); got != tc.latestVersionIsDeployed {
				t.Errorf("InitMetrics, Expected LatestVersionIsDeployed to be %f, not %f",
					tc.latestVersionIsDeployed, got)
			}
			want := hadUpdatesCurrentAvailable + updatesCurrentAvailableDelta
			if got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE")); got != want {
				t.Logf("hadUpdatesCurrentAvailable=%f, updatesCurrentAvailableDelta=%f", hadUpdatesCurrentAvailable, updatesCurrentAvailableDelta)
				t.Errorf("InitMetrics, Expected UpdatesCurrent('AVAILABLE') to be %f, not %f",
					want, got)
			}
			want = hadUpdatesCurrentSkipped + updatesCurrentSkippedDelta
			if got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED")); got != want {
				t.Logf("hadUpdatesCurrentSkipped=%f, updatesCurrentSkippedDelta=%f", hadUpdatesCurrentSkipped, updatesCurrentSkippedDelta)
				t.Errorf("InitMetrics, Expected UpdatesCurrent('SKIPPED') to be %f, not %f",
					updatesCurrentSkippedDelta, got)
			}

			// WHEN DeleteMetrics is called on it.
			status.DeleteMetrics()

			// THEN the metrics are deleted.
			if got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(*status.ServiceID)); got != 0 {
				t.Errorf("DeleteMetrics, Expected LatestVersionIsDeployed to be 0, not %f",
					got)
			}
			want = hadUpdatesCurrentAvailable
			if got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE")); got != want {
				t.Errorf("DeleteMetrics, Expected UpdatesCurrent('AVAILABLE') to be %f, not %f",
					want, got)
			}
			want = hadUpdatesCurrentSkipped
			if got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED")); got != want {
				t.Errorf("DeleteMetrics, Expected UpdatesCurrent('SKIPPED') to be %f, not %f",
					want, got)
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
	if statusDefaults.AnnounceChannel != &announceChannel {
		t.Errorf("status.NewDefaults() AnnounceChannel not initialised correctly.\nwant: %v\ngot:  %v",
			&announceChannel, statusDefaults.AnnounceChannel)
	}
	// AND the DatabaseChannel is set to the given channel.
	if statusDefaults.DatabaseChannel != &databaseChannel {
		t.Errorf("status.NewDefaults()DatabaseChannel not initialised correctly.\nwant: %v\ngot:  %v",
			&databaseChannel, statusDefaults.DatabaseChannel)
	}
	// AND the SaveChannel is set to the given channel.
	if statusDefaults.SaveChannel != &saveChannel {
		t.Errorf("status.NewDefaults()SaveChannel not initialised correctly.\nwant: %v\ngot:  %v",
			&saveChannel, statusDefaults.SaveChannel)
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
		approvedVersion, deployedVersion, deployedVersionTimestamp,
		latestVersion, latestVersionTimestamp, lastQueried)

	// WHEN Copy is called on it.
	copiedStatus := status.Copy()

	// THEN the copied Status should have the same values as the original.
	if copiedStatus.AnnounceChannel != status.AnnounceChannel {
		t.Errorf("status.Status.Copy() AnnounceChannel not copied correctly.\nwant: %v\ngot:  %v",
			status.AnnounceChannel, copiedStatus.AnnounceChannel)
	}
	if copiedStatus.DatabaseChannel != status.DatabaseChannel {
		t.Errorf("status.Status.Copy() DatabaseChannel not copied correctly.\nwant: %v\ngot:  %v",
			status.DatabaseChannel, copiedStatus.DatabaseChannel)
	}
	if copiedStatus.SaveChannel != status.SaveChannel {
		t.Errorf("status.Status.Copy() SaveChannel not copied correctly.\nwant: %v\ngot:  %v",
			status.SaveChannel, copiedStatus.SaveChannel)
	}
	if copiedStatus.approvedVersion != status.approvedVersion {
		t.Errorf("status.Status.Copy() approvedVersion not copied correctly.\nwant: %v\ngot:  %v",
			status.approvedVersion, copiedStatus.approvedVersion)
	}
	if copiedStatus.deployedVersion != status.deployedVersion {
		t.Errorf("status.Status.Copy() deployedVersion not copied correctly.\nwant: %v\ngot:  %v",
			status.deployedVersion, copiedStatus.deployedVersion)
	}
	if copiedStatus.deployedVersionTimestamp != status.deployedVersionTimestamp {
		t.Errorf("status.Status.Copy() deployedVersionTimestamp not copied correctly.\nwant: %v\ngot:  %v",
			status.deployedVersionTimestamp, copiedStatus.deployedVersionTimestamp)
	}
	if copiedStatus.latestVersion != status.latestVersion {
		t.Errorf("status.Status.Copy() latestVersion not copied correctly.\nwant: %v\ngot:  %v",
			status.latestVersion, copiedStatus.latestVersion)
	}
	if copiedStatus.latestVersionTimestamp != status.latestVersionTimestamp {
		t.Errorf("status.Status.Copy() latestVersionTimestamp not copied correctly.\nwant: %v\ngot:  %v",
			status.latestVersionTimestamp, copiedStatus.latestVersionTimestamp)
	}
	if copiedStatus.lastQueried != status.lastQueried {
		t.Errorf("status.Status.Copy() lastQueried not copied correctly.\nwant: %v\ngot:  %v",
			status.lastQueried, copiedStatus.lastQueried)
	}
}

func TestStatus_SetAnnounceChannel(t *testing.T) {
	// GIVEN a Status with an initial AnnounceChannel.
	initialChannel := make(chan []byte, 4)
	status := New(
		&initialChannel, nil, nil,
		"", "", "", "", "", "")

	// WHEN SetAnnounceChannel is called with a new channel.
	newChannel := make(chan []byte, 4)
	status.SetAnnounceChannel(&newChannel)

	// THEN the AnnounceChannel should be updated to the new channel.
	if status.AnnounceChannel != &newChannel {
		t.Errorf("AnnounceChannel not set correctly.\nwant: %v\ngot:  %v",
			&newChannel, status.AnnounceChannel)
	}

	// AND the initial channel should no longer be the AnnounceChannel.
	if status.AnnounceChannel == &initialChannel {
		t.Errorf("AnnounceChannel shouldn't have been reset to be the initial channel.\nwant: %v\ngot:  %v",
			&newChannel, status.AnnounceChannel)
	}
}

func TestStatus_SetDeleting(t *testing.T) {
	// GIVEN a Status.
	status := Status{}

	// WHEN SetDeleting is called on it.
	status.SetDeleting()

	// THEN the deleting flag should be set to true.
	if !status.Deleting() {
		t.Errorf("Expected deleting to be true, but got false")
	}

	// WHEN SetDeleting is called on it again.
	status.SetDeleting()

	// THEN the deleting flag should still be true.
	if !status.Deleting() {
		t.Errorf("Expected deleting to be true on second call, but got false")
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
				"")

			status2 := New(
				nil, nil, nil,
				tc.status2.approvedVersion,
				tc.status2.deployedVersion, "",
				tc.status2.latestVersion, "",
				"")

			// WHEN comparing versions.
			result := status1.SameVersions(status2)

			// THEN the result matches expected.
			if result != tc.expected {
				t.Errorf("got %v, want %v", result, tc.expected)
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
		// "1 to 2 - Latest version deployed -> Latest version approved": {},
		// "1 to 3 - Latest version deployed -> Latest version skipped": {},
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
				"")
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
			if got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("AVAILABLE")); got != want {
				t.Errorf("Expected available count %v, got %v",
					want, got)
			}
			want = hadUpdatesCurrentSkipped + tc.updateCountSkipped
			if got := testutil.ToFloat64(metric.UpdatesCurrent.WithLabelValues("SKIPPED")); got != want {
				t.Errorf("Expected skipped count %v, got %v",
					want, got)
			}
		})
	}
}
