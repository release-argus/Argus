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
	metric "github.com/release-argus/Argus/web/metric"
)

func TestStatus_Init(t *testing.T) {
	// GIVEN we have a Status
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

			// WHEN Init is called
			status.Init(
				tc.shoutrrrs, tc.commands, tc.webhooks,
				&tc.serviceID, &name,
				&tc.webURL)

			// THEN the Status is initialised as expected
			// ServiceID
			if status.ServiceID != &tc.serviceID {
				t.Errorf("ServiceID not initialised to address of %s (%v). Got %v",
					tc.serviceID, &tc.serviceID, status.ServiceID)
			}
			// WebURL
			if status.WebURL != &tc.webURL {
				t.Errorf("WebURL not initialised to address of %s (%v). Got %v",
					tc.webURL, &tc.webURL, status.WebURL)
			}
			// Shoutrrr
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
			// Command
			got = status.Fails.Command.Length()
			if got != tc.commands {
				t.Errorf("Fails.Command was initialised to %d. Want %d",
					got, tc.commands)
			}
			// WebHook
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
	// GIVEN we have a Status
	latestVersion := "1.2.3"
	tests := map[string]struct {
		webURL *string
		want   string
	}{
		"nil string": {
			webURL: test.StringPtr(""),
			want:   ""},
		"empty string": {
			webURL: test.StringPtr(""),
			want:   ""},
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

			// WHEN GetWebURL is called
			got := status.GetWebURL()

			// THEN the returned WebURL is as expected
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestStatus_SetLastQueried(t *testing.T) {
	// GIVEN we have a Status and some webhooks
	var status Status

	// WHEN we SetLastQueried
	start := time.Now().UTC()
	status.SetLastQueried("")

	// THEN LastQueried will have been set to the current timestamp
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("LastQueried was %v ago, not recent enough!",
			since)
	}
}

func TestStatus_ApprovedVersion(t *testing.T) {
	deployedVersion := "0.0.1"
	latestVersion := "0.0.3"
	// GIVEN a Status
	tests := map[string]struct {
		approving                     string
		latestVersionIsDeployedMetric float64
	}{
		"Approving LatestVersion": {
			approving:                     latestVersion,
			latestVersionIsDeployedMetric: 2,
		},
		"Skipping LatestVersion": {
			approving:                     "SKIP_" + latestVersion,
			latestVersionIsDeployedMetric: 3,
		},
		"Approving non-latest version": {
			approving:                     "0.0.2a",
			latestVersionIsDeployedMetric: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			announceChannel := make(chan []byte, 4)
			databaseChannel := make(chan dbtype.Message, 4)
			status := New(
				&announceChannel, &databaseChannel, nil,
				"", "", "", "", "", "")
			status.Init(
				0, 0, 0,
				test.StringPtr("TestStatus_SetApprovedVersion_"+name), test.StringPtr("TestStatus_SetApprovedVersion_"+name),
				test.StringPtr("https://example.com"))
			status.SetLatestVersion(latestVersion, "", false)
			status.SetDeployedVersion(deployedVersion, "", false)

			// WHEN SetApprovedVersion is called
			status.SetApprovedVersion(tc.approving, true)

			// THEN the Status is as expected
			// ApprovedVersion
			got := status.ApprovedVersion()
			if got != tc.approving {
				t.Errorf("ApprovedVersion not set to %s. Got %s",
					tc.approving, got)
			}
			// LatestVersion
			got = status.LatestVersion()
			if got != latestVersion {
				t.Errorf("LatestVersion not set to %s. Got %s",
					latestVersion, got)
			}
			// DeployedVersion
			got = status.DeployedVersion()
			if got != deployedVersion {
				t.Errorf("DeployedVersion not set to %s. Got %s",
					deployedVersion, got)
			}
			// AnnounceChannel
			if len(*status.AnnounceChannel) != 1 {
				t.Errorf("AnnounceChannel should have 1 message, but has %d",
					len(*status.AnnounceChannel))
			}
			// DatabaseChannel
			if len(*status.DatabaseChannel) != 1 {
				t.Errorf("DatabaseChannel should have 1 message, but has %d",
					len(*status.DatabaseChannel))
			}
			// AND LatestVersionIsDeployedVersion metric is updated
			gotMetric := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(*status.ServiceID))
			if gotMetric != tc.latestVersionIsDeployedMetric {
				t.Errorf("LatestVersionIsDeployedVersion metric should be %f, not %f",
					tc.latestVersionIsDeployedMetric, gotMetric)
			}
		})
	}
}

func TestStatus_DeployedVersion(t *testing.T) {
	// GIVEN a Status
	approvedVersion := "0.0.2"
	deployedVersion := "0.0.1"
	latestVersion := "0.0.3"
	tests := map[string]struct {
		deploying                                       string
		approvedVersion, deployedVersion, latestVersion string
	}{
		"Deploying ApprovedVersion - DeployedVersion becomes ApprovedVersion and resets ApprovedVersion": {
			deploying:       approvedVersion,
			approvedVersion: "",
			deployedVersion: approvedVersion,
			latestVersion:   latestVersion,
		},
		"Deploying unknown Version - DeployedVersion becomes this version": {
			deploying:       "0.0.4",
			approvedVersion: approvedVersion,
			deployedVersion: "0.0.4",
			latestVersion:   latestVersion,
		},
		"Deploying LatestVersion - DeployedVersion becomes this version": {
			deploying:       latestVersion,
			approvedVersion: approvedVersion,
			deployedVersion: latestVersion,
			latestVersion:   latestVersion,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, haveDB := range []bool{false, true} {
				dbChannel := make(chan dbtype.Message, 4)
				status := New(
					nil, &dbChannel, nil,
					"", "", "", "", "", "")
				if !haveDB {
					status.DatabaseChannel = nil
				}
				status.Init(
					0, 0, 0,
					&name, &name,
					test.StringPtr("http://example.com"))
				status.SetApprovedVersion(approvedVersion, haveDB)
				status.SetDeployedVersion(deployedVersion, "", haveDB)
				status.SetLatestVersion(latestVersion, "", haveDB)

				// WHEN SetDeployedVersion is called on it
				status.SetDeployedVersion(tc.deploying, "", haveDB)

				// THEN DeployedVersion is set to this version
				if status.DeployedVersion() != tc.deployedVersion {
					t.Errorf("Expected DeployedVersion to be set to %q, not %q",
						tc.deployedVersion, status.DeployedVersion())
				}
				if status.ApprovedVersion() != tc.approvedVersion {
					t.Errorf("Expected ApprovedVersion to be set to %q, not %q",
						tc.approvedVersion, status.ApprovedVersion())
				}
				if status.LatestVersion() != tc.latestVersion {
					t.Errorf("Expected LatestVersion to be set to %q, not %q",
						tc.latestVersion, status.LatestVersion())
				}
				// AND the DeployedVersionTimestamp is set to current time
				d, _ := time.Parse(time.RFC3339, status.DeployedVersionTimestamp())
				since := time.Since(d)
				if since > time.Second {
					t.Errorf("DeployedVersionTimestamp was %v ago, not recent enough!",
						since)
				}
				// AND the LatestVersionIsDeployedVersion metric is updated
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
	// GIVEN a Status
	lastQueried := "2021-01-01T00:00:00Z"
	tests := map[string]struct {
		latestVersionTimestamp, wantLatestVersionTimestamp string
	}{
		"LatestVersionTimestamp - Empty == Set to lastQueried": {
			latestVersionTimestamp:     "",
			wantLatestVersionTimestamp: lastQueried,
		},
		"LatestVersionTimestamp - Given == Set to value given": {
			latestVersionTimestamp:     "2020-01-01T00:00:00Z",
			wantLatestVersionTimestamp: "2020-01-01T00:00:00Z",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, haveDB := range []bool{false, true} {
				dbChannel := make(chan dbtype.Message, 8)
				status := New(
					nil, &dbChannel, nil,
					"",
					"", "",
					"0.0.0", "",
					lastQueried)
				if !haveDB {
					status.DatabaseChannel = nil
				}
				status.Init(
					0, 0, 0,
					&name, &name,
					test.StringPtr("http://example.com"))
				versions := []string{"0.0.1", "0.0.2", "0.0.2-dev", "something-else"}
				for _, version := range versions {

					// WHEN SetLatestVersion is called on it
					status.SetLatestVersion(version, tc.latestVersionTimestamp, haveDB)

					// THEN LatestVersion is set to this version
					if status.LatestVersion() != version {
						t.Errorf("Expected LatestVersion to be set to %q, not %q",
							version, status.LatestVersion())
					}
					// AND the LatestVersionTimestamp is set to the current time
					if status.LatestVersionTimestamp() != tc.wantLatestVersionTimestamp {
						t.Errorf("haveDB=%t LatestVersionTimestamp should've been set to LastQueried \n%q, not \n%q",
							haveDB, tc.wantLatestVersionTimestamp, status.LatestVersionTimestamp())
					}
					// AND the LatestVersionIsDeployedVersion metric is updated
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
			}
		})
	}
}

func TestStatus_RegexMissesContent(t *testing.T) {
	// GIVEN a Status
	status := Status{}

	// WHEN RegexMissContent is called on it
	status.RegexMissContent()

	// THEN RegexMisses is incremented
	got := status.RegexMissesContent()
	if got != 1 {
		t.Errorf("Expected RegexMisses to be 1, not %d",
			got)
	}

	// WHEN RegexMissContent is called on it again
	status.RegexMissContent()

	// THEN RegexMisses is incremented again
	got = status.RegexMissesContent()
	if got != 2 {
		t.Errorf("Expected RegexMisses to be 2, not %d",
			got)
	}

	// WHEN RegexMissContent is called on it again
	status.RegexMissContent()

	// THEN RegexMisses is incremented again
	got = status.RegexMissesContent()
	if got != 3 {
		t.Errorf("Expected RegexMisses to be 3, not %d",
			got)
	}

	// WHEN ResetRegexMisses is called on it
	status.ResetRegexMisses()

	// THEN RegexMisses is reset
	got = status.RegexMissesContent()
	if got != 0 {
		t.Errorf("Expected RegexMisses to be 0 after ResetRegexMisses, not %d",
			got)
	}
}

func TestStatus_RegexMissesVersion(t *testing.T) {
	// GIVEN a Status
	status := Status{}

	// WHEN RegexMissVersion is called on it
	status.RegexMissVersion()

	// THEN RegexMisses is incremented
	got := status.RegexMissesVersion()
	if got != 1 {
		t.Errorf("Expected RegexMisses to be 1, not %d",
			got)
	}

	// WHEN RegexMissVersion is called on it again
	status.RegexMissVersion()

	// THEN RegexMisses is incremented again
	got = status.RegexMissesVersion()
	if got != 2 {
		t.Errorf("Expected RegexMisses to be 2, not %d",
			got)
	}

	// WHEN RegexMissVersion is called on it again
	status.RegexMissVersion()

	// THEN RegexMisses is incremented again
	got = status.RegexMissesVersion()
	if got != 3 {
		t.Errorf("Expected RegexMisses to be 3, not %d",
			got)
	}

	// WHEN ResetRegexMisses is called on it
	status.ResetRegexMisses()

	// THEN RegexMisses is reset
	got = status.RegexMissesVersion()
	if got != 0 {
		t.Errorf("Expected RegexMisses to be 0 after ResetRegexMisses, not %d",
			got)
	}
}

func TestStatus_SendAnnounce(t *testing.T) {
	// GIVEN a Status with channels
	tests := map[string]struct {
		deleting   bool
		nilChannel bool
	}{
		"not deleting or nil channel": {},
		"deleting":                    {deleting: true},
		"nil channel":                 {nilChannel: true},
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

			// WHEN SendAnnounce is called on it
			status.SendAnnounce(&[]byte{})

			// THEN the AnnounceChannel is sent a message if not deleting or nil
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
	// GIVEN a Status with channels
	tests := map[string]struct {
		deleting   bool
		nilChannel bool
	}{
		"not deleting or nil channel": {},
		"deleting":                    {deleting: true},
		"nil channel":                 {nilChannel: true},
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

			// WHEN sendDatabase is called on it
			status.sendDatabase(&dbtype.Message{})

			// THEN the DatabaseChannel is sent a message if not deleting or nil
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
	// GIVEN a Status with channels
	tests := map[string]struct {
		deleting   bool
		nilChannel bool
	}{
		"not deleting or nil channel": {},
		"deleting":                    {deleting: true},
		"nil channel":                 {nilChannel: true},
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

			// WHEN SendSave is called on it
			status.SendSave()

			// THEN the SaveChannel is sent a message if not deleting or nil
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
	// GIVEN a Fails struct
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

			// WHEN resetFails is called on it
			fails.resetFails()

			// THEN all the fails become nil
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
	// GIVEN a Status
	tests := map[string]struct {
		status                                    *Status
		approvedVersion                           string
		deployedVersion, deployedVersionTimestamp string
		latestVersion, latestVersionTimestamp     string
		lastQueried                               *string
		regexMissesContent, regexMissesVersion    int
		commandFails                              []*bool
		shoutrrrFails, webhookFails               map[string]*bool
		want                                      string
	}{
		"empty status": {
			status: &Status{},
			want:   "",
		},
		"only fails": {
			commandFails: []*bool{
				nil,
				test.BoolPtr(false),
				test.BoolPtr(true)},
			shoutrrrFails: map[string]*bool{
				"bash": test.BoolPtr(false),
				"bish": nil,
				"bosh": test.BoolPtr(true)},
			webhookFails: map[string]*bool{
				"bar": nil,
				"foo": test.BoolPtr(false)},
			status: &Status{},
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
			status:             &Status{},
			shoutrrrFails: map[string]*bool{
				"bish": nil,
				"bash": test.BoolPtr(false),
				"bosh": test.BoolPtr(true)},
			commandFails: []*bool{
				nil,
				test.BoolPtr(false),
				test.BoolPtr(true)},
			webhookFails: map[string]*bool{
				"foo": test.BoolPtr(false),
				"bar": nil},
			approvedVersion:          "1.2.4",
			deployedVersion:          "1.2.3",
			deployedVersionTimestamp: "2022-01-01T01:01:02Z",
			latestVersion:            "1.2.4",
			latestVersionTimestamp:   "2022-01-01T01:01:01Z",
			lastQueried:              test.StringPtr("2022-01-01T01:01:01Z"),
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

			tc.status.SetApprovedVersion(tc.approvedVersion, false)
			tc.status.SetDeployedVersion(tc.deployedVersion, tc.deployedVersionTimestamp, false)
			if tc.deployedVersionTimestamp == "" {
				tc.status.deployedVersionTimestamp = tc.deployedVersionTimestamp
			}
			tc.status.SetLatestVersion(tc.latestVersion, tc.latestVersionTimestamp, false)
			if tc.lastQueried != nil {
				tc.status.SetLastQueried(*tc.lastQueried)
			}
			{ // RegEz misses
				for i := 0; i < tc.regexMissesContent; i++ {
					tc.status.RegexMissContent()
				}
				for i := 0; i < tc.regexMissesVersion; i++ {
					tc.status.RegexMissVersion()
				}
			}
			{ // Fails
				tc.status.Init(
					len(tc.shoutrrrFails), len(tc.commandFails), len(tc.webhookFails),
					tc.status.ServiceID, &name,
					&name)
				for k, v := range tc.commandFails {
					if v != nil {
						tc.status.Fails.Command.Set(k, *v)
					}
				}
				for k, v := range tc.shoutrrrFails {
					tc.status.Fails.Shoutrrr.Set(k, v)
				}
				for k, v := range tc.webhookFails {
					tc.status.Fails.WebHook.Set(k, v)
				}
			}

			// WHEN the Status is stringified with String
			got := tc.status.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("Status.String() mismatch\n%q\ngot:\n%q",
					tc.want, got)
			}
		})
	}
}

func TestStatus_SetLatestVersionIsDeployedMetric(t *testing.T) {
	// GIVEN a Status
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
		"nil ServiceID - initial state": {
			serviceID: "<nil>",
			want:      0,
		},
		"nil ServiceID - ignore latest version is deployed": {
			serviceID:       "<nil>",
			latestVersion:   "1.2.3",
			deployedVersion: "1.2.3",
			want:            0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := Status{}
			status.Init(
				0, 0, 0,
				&name, &name,
				test.StringPtr("http://example.com"))
			status.SetApprovedVersion(tc.approvedVersion, false)
			status.SetLatestVersion(tc.latestVersion, "", false)
			status.SetDeployedVersion(tc.deployedVersion, "", false)
			if tc.serviceID == "<nil>" {
				status.ServiceID = nil
			}

			// WHEN setLatestVersion is called on it
			status.setLatestVersionIsDeployedMetric()
			got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(name))

			// THEN the metric is as expected
			if got != tc.want {
				t.Errorf("Expected SetLatestVersionIsDeployedMetric to be %f, not %f",
					tc.want, got)
			}
		})
	}
}

func TestStatus_InitMetrics_DeleteMetrics(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		nilStatus, nilServiceID bool
		wantCreate              float64
	}{
		"nil status": {
			nilStatus: true,
		},
		"nil serviceID": {
			nilServiceID: true,
		},
		"non-nil serviceID": {
			wantCreate: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := &Status{}
			status.Init(
				0, 0, 0,
				&name, &name,
				test.StringPtr("http://example.com"))
			status.SetLatestVersion("0.0.2", "", false)
			status.SetDeployedVersion("0.0.2", "", false)
			if tc.nilServiceID {
				status.ServiceID = nil
			}
			if tc.nilStatus {
				status = nil
			}

			// WHEN InitMetrics is called on it
			status.InitMetrics()

			// THEN the metrics are created
			got := float64(0)
			if status != nil && status.ServiceID != nil {
				got = testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(*status.ServiceID))
			}
			if got != tc.wantCreate {
				t.Errorf("Expected LatestVersionIsDeployed to be %f, not %f",
					tc.wantCreate, got)
			}

			// WHEN DeleteMetrics is called on it
			status.DeleteMetrics()

			// THEN the metrics are deleted
			got = float64(0)
			if status != nil && status.ServiceID != nil {
				got = testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(*status.ServiceID))
			}
			if got != 0 {
				t.Errorf("Expected LatestVersionIsDeployed to be 0, not %f",
					got)
			}
		})
	}
}

func TestNewDefaults(t *testing.T) {
	// GIVEN we have channels
	announceChannel := make(chan []byte, 4)
	databaseChannel := make(chan dbtype.Message, 4)
	saveChannel := make(chan bool, 4)

	// WHEN NewDefaults is called
	statusDefaults := NewDefaults(&announceChannel, &databaseChannel, &saveChannel)

	// THEN the AnnounceChannel is set to the given channel
	if statusDefaults.AnnounceChannel != &announceChannel {
		t.Errorf("status.NewDefaults() AnnounceChannel not initialised correctly.\nwant: %v\ngot:  %v",
			&announceChannel, statusDefaults.AnnounceChannel)
	}
	// AND the DatabaseChannel is set to the given channel
	if statusDefaults.DatabaseChannel != &databaseChannel {
		t.Errorf("status.NewDefaults()DatabaseChannel not initialised correctly.\nwant: %v\ngot:  %v",
			&databaseChannel, statusDefaults.DatabaseChannel)
	}
	// AND the SaveChannel is set to the given channel
	if statusDefaults.SaveChannel != &saveChannel {
		t.Errorf("status.NewDefaults()SaveChannel not initialised correctly.\nwant: %v\ngot:  %v",
			&saveChannel, statusDefaults.SaveChannel)
	}
}

func TestStatus_Copy(t *testing.T) {
	// GIVEN a Status
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

	// WHEN Copy is called on it
	copiedStatus := status.Copy()

	// THEN the copied Status should have the same values as the original
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
	// GIVEN a Status with an initial AnnounceChannel
	initialChannel := make(chan []byte, 4)
	status := New(
		&initialChannel, nil, nil,
		"", "", "", "", "", "")

	// WHEN SetAnnounceChannel is called with a new channel
	newChannel := make(chan []byte, 4)
	status.SetAnnounceChannel(&newChannel)

	// THEN the AnnounceChannel should be updated to the new channel
	if status.AnnounceChannel != &newChannel {
		t.Errorf("AnnounceChannel not set correctly.\nwant: %v\ngot:  %v",
			&newChannel, status.AnnounceChannel)
	}

	// AND the initial channel should no longer be the AnnounceChannel
	if status.AnnounceChannel == &initialChannel {
		t.Errorf("AnnounceChannel shouldn't have been reset to be the initial channel.\nwant: %v\ngot:  %v",
			&newChannel, status.AnnounceChannel)
	}
}

func TestStatus_SetDeleting(t *testing.T) {
	// GIVEN a Status
	status := Status{}

	// WHEN SetDeleting is called on it
	status.SetDeleting()

	// THEN the deleting flag should be set to true
	if !status.Deleting() {
		t.Errorf("Expected deleting to be true, but got false")
	}

	// WHEN SetDeleting is called on it again
	status.SetDeleting()

	// THEN the deleting flag should still be true
	if !status.Deleting() {
		t.Errorf("Expected deleting to be true on second call, but got false")
	}
}

func TestStatus_SameVersions(t *testing.T) {
	type versions struct {
		approvedVersion, deployedVersion, latestVersion string
	}
	// GIVEN different Status version combinations
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

			status1 := Status{}
			status1.SetApprovedVersion(tc.status1.approvedVersion, false)
			status1.SetDeployedVersion(tc.status1.deployedVersion, "", false)
			status1.SetLatestVersion(tc.status1.latestVersion, "", false)

			status2 := Status{}
			status2.SetApprovedVersion(tc.status2.approvedVersion, false)
			status2.SetDeployedVersion(tc.status2.deployedVersion, "", false)
			status2.SetLatestVersion(tc.status2.latestVersion, "", false)

			// WHEN comparing versions
			result := status1.SameVersions(&status2)

			// THEN the result matches expected
			if result != tc.expected {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}
