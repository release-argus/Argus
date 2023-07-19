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

package svcstatus

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	dbtype "github.com/release-argus/Argus/db/types"
	metric "github.com/release-argus/Argus/web/metrics"
)

func TestStatus_Init(t *testing.T) {
	// GIVEN we have a Status
	tests := map[string]struct {
		shoutrrrs int
		commands  int
		webhooks  int
		serviceID string
		webURL    string
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var status Status

			// WHEN Init is called
			status.Init(
				tc.shoutrrrs, tc.commands, tc.webhooks,
				&tc.serviceID,
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
			webURL: stringPtr(""),
			want:   ""},
		"empty string": {
			webURL: stringPtr(""),
			want:   ""},
		"string without templating": {
			webURL: stringPtr("https://something.com/somewhere"),
			want:   "https://something.com/somewhere"},
		"string with templating": {
			webURL: stringPtr("https://something.com/somewhere/{{ version }}"),
			want:   "https://something.com/somewhere/" + latestVersion},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := Status{}
			status.Init(
				0, 0, 0,
				&name,
				tc.webURL)
			status.SetLatestVersion(latestVersion, false)

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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			announceChannel := make(chan []byte, 4)
			databaseChannel := make(chan dbtype.Message, 4)
			status := New(
				&announceChannel, &databaseChannel, nil,
				"", "", "", "", "", "")
			status.Init(
				0, 0, 0,
				stringPtr("TestStatus_SetApprovedVersion_"+name),
				stringPtr("https://example.com"))
			status.SetLatestVersion(latestVersion, false)
			status.SetDeployedVersion(deployedVersion, false)

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
		deploying       string
		approvedVersion string
		deployedVersion string
		latestVersion   string
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbChannel := make(chan dbtype.Message, 4)
			status := New(
				nil, &dbChannel, nil,
				"", "", "", "", "", "")
			status.Init(
				0, 0, 0,
				&name,
				stringPtr("http://example.com"))
			status.SetApprovedVersion(approvedVersion, false)
			status.SetDeployedVersion(deployedVersion, false)
			status.SetLatestVersion(latestVersion, false)

			// WHEN SetDeployedVersion is called on it
			status.SetDeployedVersion(tc.deploying, false)

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
			if status.LatestVersion() == status.DeployedVersion() {
				want = 1
			}
			if got != want {
				t.Errorf("LatestVersionIsDeployedVersion metric should be %f, not %f",
					want, got)
			}
		})
	}
}

func TestStatus_LatestVersion(t *testing.T) {
	// GIVEN a Status
	approvedVersion := "0.0.2"
	deployedVersion := "0.0.1"
	latestVersion := "0.0.3"
	tests := map[string]struct {
		deploying       string
		approvedVersion string
		deployedVersion string
		latestVersion   string
	}{
		"Sets LatestVersion and LatestVersionTimestamp": {
			deploying:       "0.0.4",
			approvedVersion: approvedVersion,
			deployedVersion: deployedVersion,
			latestVersion:   "0.0.4",
		},
		"Set LatestVersion to DeployedVersion": {
			deploying:       deployedVersion,
			approvedVersion: approvedVersion,
			deployedVersion: deployedVersion,
			latestVersion:   deployedVersion,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbChannel := make(chan dbtype.Message, 4)
			status := New(
				nil, &dbChannel, nil,
				"", "", "", "", "", "")
			status.Init(
				0, 0, 0,
				&name,
				stringPtr("http://example.com"))
			status.SetApprovedVersion(approvedVersion, false)
			status.SetDeployedVersion(deployedVersion, false)
			status.SetLatestVersion(latestVersion, false)

			// WHEN SetLatestVersion is called on it
			status.SetLatestVersion(tc.deploying, false)

			// THEN LatestVersion is set to this version
			if status.LatestVersion() != tc.latestVersion {
				t.Errorf("Expected LatestVersion to be set to %q, not %q",
					tc.latestVersion, status.LatestVersion())
			}
			if status.DeployedVersion() != tc.deployedVersion {
				t.Errorf("Expected DeployedVersion to be set to %q, not %q",
					tc.deployedVersion, status.DeployedVersion())
			}
			if status.ApprovedVersion() != tc.approvedVersion {
				t.Errorf("Expected ApprovedVersion to be set to %q, not %q",
					tc.approvedVersion, status.ApprovedVersion())
			}
			// AND the LatestVersionTimestamp is set to the current time
			if status.LatestVersionTimestamp() != status.LastQueried() {
				t.Errorf("LatestVersionTimestamp should've been set to LastQueried \n%q, not \n%q",
					status.LastQueried(), status.LatestVersionTimestamp())
			}
			// AND the LatestVersionIsDeployedVersion metric is updated
			got := testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(name))
			want := float64(0)
			if status.LatestVersion() == status.DeployedVersion() {
				want = 1
			}
			if got != want {
				t.Errorf("LatestVersionIsDeployedVersion metric should be %f, not %f",
					want, got)
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
		name, tc := name, tc
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

func TestStatus_SendDatabase(t *testing.T) {
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
		name, tc := name, tc
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

			// WHEN SendDatabase is called on it
			status.SendDatabase(&dbtype.Message{})

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
		name, tc := name, tc
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
		commandFails  *[]*bool
		shoutrrrFails *map[string]*bool
		webhookFails  *map[string]*bool
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
				"1": boolPtr(false),
				"3": boolPtr(true)},
		},
		"only commands": {
			commandFails: &[]*bool{
				nil,
				boolPtr(false),
				boolPtr(true)},
		},
		"only webhooks": {
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": boolPtr(false),
				"3": boolPtr(true)},
		},
		"all filled": {
			shoutrrrFails: &map[string]*bool{
				"0": nil,
				"1": boolPtr(false),
				"3": boolPtr(true)},
			commandFails: &[]*bool{nil, boolPtr(false), boolPtr(true)},
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": boolPtr(false),
				"3": boolPtr(true)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
		status                   *Status
		approvedVersion          string
		deployedVersion          string
		deployedVersionTimestamp string
		latestVersion            string
		latestVersionTimestamp   string
		lastQueried              *string
		regexMissesContent       int
		regexMissesVersion       int
		commandFails             []*bool
		shoutrrrFails            map[string]*bool
		webhookFails             map[string]*bool
		want                     string
	}{
		"empty status": {
			status: &Status{},
			want:   "",
		},
		"only fails": {
			commandFails: []*bool{
				nil,
				boolPtr(false),
				boolPtr(true)},
			shoutrrrFails: map[string]*bool{
				"bash": boolPtr(false),
				"bish": nil,
				"bosh": boolPtr(true)},
			webhookFails: map[string]*bool{
				"bar": nil,
				"foo": boolPtr(false)},
			status: &Status{},
			want: `
fails: {
shoutrrr: {bash: false, bish: nil, bosh: true},
 command: [0: nil, 1: false, 2: true],
 webhook: {bar: nil, foo: false}
}`,
		},
		"all fields": {
			regexMissesContent: 1,
			regexMissesVersion: 2,
			status:             &Status{},
			shoutrrrFails: map[string]*bool{
				"bish": nil,
				"bash": boolPtr(false),
				"bosh": boolPtr(true)},
			commandFails: []*bool{
				nil,
				boolPtr(false),
				boolPtr(true)},
			webhookFails: map[string]*bool{
				"foo": boolPtr(false),
				"bar": nil},
			approvedVersion:          "1.2.4",
			deployedVersion:          "1.2.3",
			deployedVersionTimestamp: "2022-01-01T01:01:02Z",
			latestVersion:            "1.2.4",
			latestVersionTimestamp:   "2022-01-01T01:01:01Z",
			lastQueried:              stringPtr("2022-01-01T01:01:01Z"),
			want: `
approved_version: 1.2.4,
 deployed_version: 1.2.3,
 deployed_version_timestamp: 2022-01-01T01:01:02Z,
 latest_version: 1.2.4,
 latest_version_timestamp: 2022-01-01T01:01:01Z,
 last_queried: 2022-01-01T01:01:01Z,
 regex_misses_content: 1,
 regex_misses_version: 2,
 fails: {
shoutrrr: {bash: false, bish: nil, bosh: true},
 command: [0: nil, 1: false, 2: true],
 webhook: {bar: nil, foo: false}
}`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.status.SetApprovedVersion(tc.approvedVersion, false)
			tc.status.SetDeployedVersion(tc.deployedVersion, false)
			tc.status.SetDeployedVersionTimestamp(tc.deployedVersionTimestamp)
			tc.status.SetLatestVersion(tc.latestVersion, false)
			tc.status.SetLatestVersionTimestamp(tc.latestVersionTimestamp)
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
					tc.status.ServiceID,
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
			tc.want = strings.ReplaceAll(tc.want, "\n", "")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestStatus_SetLatestVersionIsDeployedMetric(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		latestVersion   string
		deployedVersion string
		want            float64
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
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := Status{}
			status.Init(
				0, 0, 0,
				&name,
				stringPtr("http://example.com"))
			status.SetLatestVersion(tc.latestVersion, false)
			status.SetDeployedVersion(tc.deployedVersion, false)

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

func TestStatus_DeleteMetrics(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		serviceID *string
	}{
		"nil serviceID": {
			serviceID: nil,
		},
		"non-nil serviceID": {
			serviceID: stringPtr("TestStatus_DeleteMetrics"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status := Status{}
			status.Init(
				0, 0, 0,
				tc.serviceID,
				stringPtr("http://example.com"))

			// WHEN DeleteMetrics is called on it
			status.DeleteMetrics()

			// THEN the metrics are deleted
			got := float64(0)
			if tc.serviceID != nil {
				got = testutil.ToFloat64(metric.LatestVersionIsDeployed.WithLabelValues(*status.ServiceID))
			}
			if got != 0 {
				t.Errorf("Expected LatestVersionIsDeployed to be 0, not %f",
					got)
			}
		})
	}
}
