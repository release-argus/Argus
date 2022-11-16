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

//go:build unit

package svcstatus

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
)

func TestInit(t *testing.T) {
	// GIVEN we have a Status
	tests := map[string]struct {
		shoutrrrs int
		commands  int
		webhooks  int
		serviceID string
		webURL    string
	}{
		"ServiceID": {serviceID: "test"},
		"WebURL":    {webURL: "https://example.com"},
		"shoutrrrs": {shoutrrrs: 2},
		"commands":  {commands: 3},
		"webhooks":  {webhooks: 4},
		"all": {serviceID: "argus", webURL: "https://release-argus.io",
			shoutrrrs: 5, commands: 5, webhooks: 5},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var status Status

			// WHEN Init is called
			status.Init(tc.shoutrrrs, tc.commands, tc.webhooks, &tc.serviceID, &tc.webURL)

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
			got := len(status.Fails.Shoutrrr)
			if got != 0 {
				t.Errorf("Fails.Shoutrrr was initialised to %d. Want %d",
					got, 0)
			} else {
				for i := 0; i < tc.shoutrrrs; i++ {
					status.Fails.Shoutrrr[fmt.Sprint(i)] = boolPtr(false)
				}
				got := len(status.Fails.Shoutrrr)
				if got != tc.shoutrrrs {
					t.Errorf("Fails.Shoutrrr wanted capacity for %d, but only got to %d",
						tc.shoutrrrs, got)
				}
			}
			// Command
			got = len(status.Fails.Command)
			if got != tc.commands {
				t.Errorf("Fails.Command was initialised to %d. Want %d",
					got, tc.commands)
			}
			// WebHook
			got = len(status.Fails.WebHook)
			if got != 0 {
				t.Errorf("Fails.WebHook was initialised to %d. Want %d",
					got, 0)
			} else {
				for i := 0; i < tc.webhooks; i++ {
					status.Fails.WebHook[fmt.Sprint(i)] = boolPtr(false)
				}
				got := len(status.Fails.WebHook)
				if got != tc.webhooks {
					t.Errorf("Fails.WebHook wanted capacity for %d, but only got to %d",
						tc.webhooks, got)
				}
			}
		})
	}
}
func TestGetWebURL(t *testing.T) {
	// GIVEN we have a Status
	latestVersion := "1.2.3"
	tests := map[string]struct {
		webURL *string
		want   string
	}{
		"nil string":                {webURL: stringPtr(""), want: ""},
		"empty string":              {webURL: stringPtr(""), want: ""},
		"string without templating": {webURL: stringPtr("https://something.com/somewhere"), want: "https://something.com/somewhere"},
		"string with templating":    {webURL: stringPtr("https://something.com/somewhere/{{ version }}"), want: "https://something.com/somewhere/" + latestVersion},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			status := Status{LatestVersion: latestVersion, WebURL: tc.webURL}

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

func TestSetLastQueried(t *testing.T) {
	// GIVEN we have a Status and some webhooks
	var status Status

	// WHEN we SetLastQueried
	start := time.Now().UTC()
	status.SetLastQueried()

	// THEN LastQueried will have been set to the current timestamp
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("LastQueried was %v ago, not recent enough!",
			since)
	}
}

func TestSetDeployedVersion(t *testing.T) {
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
		"Deploying ApprovedVersion - DeployedVersion becomes ApprovedVersion and resets ApprovedVersion": {deploying: approvedVersion,
			approvedVersion: "", deployedVersion: approvedVersion, latestVersion: latestVersion},
		"Deploying unknown Version - DeployedVersion becomes this version": {deploying: "0.0.4",
			approvedVersion: approvedVersion, deployedVersion: "0.0.4", latestVersion: latestVersion},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dbChannel := make(chan dbtype.Message, 4)
			status := Status{ApprovedVersion: approvedVersion, DeployedVersion: deployedVersion, LatestVersion: latestVersion,
				DatabaseChannel: &dbChannel, ServiceID: stringPtr("test")}

			// WHEN SetDeployedVersion is called on it
			status.SetDeployedVersion(tc.deploying)

			// THEN DeployedVersion is set to this version
			if status.DeployedVersion != tc.deployedVersion {
				t.Errorf("Expected DeployedVersion to be set to %q, not %q",
					tc.deployedVersion, status.DeployedVersion)
			}
			if status.ApprovedVersion != tc.approvedVersion {
				t.Errorf("Expected ApprovedVersion to be set to %q, not %q",
					tc.approvedVersion, status.ApprovedVersion)
			}
			if status.LatestVersion != tc.latestVersion {
				t.Errorf("Expected LatestVersion to be set to %q, not %q",
					tc.latestVersion, status.LatestVersion)
			}
			// and the current time
			d, _ := time.Parse(time.RFC3339, status.DeployedVersionTimestamp)
			since := time.Since(d)
			if since > time.Second {
				t.Errorf("DeployedVersionTimestamp was %v ago, not recent enough!",
					since)
			}
		})
	}
}

func TestResetFails(t *testing.T) {
	// GIVEN a Fails struct
	tests := map[string]struct {
		fails Fails
	}{
		"all empty":     {fails: Fails{}},
		"only notifies": {fails: Fails{Shoutrrr: map[string]*bool{"0": nil, "1": boolPtr(false), "3": boolPtr(true)}}},
		"only commands": {fails: Fails{Command: []*bool{nil, boolPtr(false), boolPtr(true)}}},
		"only webhooks": {fails: Fails{WebHook: map[string]*bool{"0": nil, "1": boolPtr(false), "3": boolPtr(true)}}},
		"all filled": {fails: Fails{Shoutrrr: map[string]*bool{"0": nil, "1": boolPtr(false), "3": boolPtr(true)},
			Command: []*bool{nil, boolPtr(false), boolPtr(true)},
			WebHook: map[string]*bool{"0": nil, "1": boolPtr(false), "3": boolPtr(true)}}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// WHEN resetFails is called on it
			tc.fails.resetFails()

			// THEN all the fails become nil
			for i := range tc.fails.Shoutrrr {
				if tc.fails.Shoutrrr[i] != nil {
					t.Errorf("Shoutrrr.Failed[%s] should have been reset to nil and not be %t",
						i, *tc.fails.Shoutrrr[i])
				}
			}
			for i := range tc.fails.Command {
				if tc.fails.Command[i] != nil {
					t.Errorf("Command.Failed[%d] should have been reset to nil and not be %t",
						i, *tc.fails.Command[i])
				}
			}
			for i := range tc.fails.WebHook {
				if tc.fails.WebHook[i] != nil {
					t.Errorf("WebHook.Failed[%s] should have been reset to nil and not be %t",
						i, *tc.fails.WebHook[i])
				}
			}
		})
	}
}

func TestSetLatestVersion(t *testing.T) {
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
		"Sets LatestVersion and LatestVersionTimestamp": {deploying: "0.0.4",
			approvedVersion: approvedVersion, deployedVersion: deployedVersion, latestVersion: "0.0.4"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dbChannel := make(chan dbtype.Message, 4)
			status := Status{ApprovedVersion: approvedVersion, DeployedVersion: deployedVersion, LatestVersion: latestVersion,
				DatabaseChannel: &dbChannel, ServiceID: stringPtr("test")}

			// WHEN SetLatestVersion is called on it
			status.SetLatestVersion(tc.deploying)

			// THEN LatestVersion is set to this version
			if status.LatestVersion != tc.latestVersion {
				t.Errorf("Expected LatestVersion to be set to %q, not %q",
					tc.latestVersion, status.LatestVersion)
			}
			if status.DeployedVersion != tc.deployedVersion {
				t.Errorf("Expected DeployedVersion to be set to %q, not %q",
					tc.deployedVersion, status.DeployedVersion)
			}
			if status.ApprovedVersion != tc.approvedVersion {
				t.Errorf("Expected ApprovedVersion to be set to %q, not %q",
					tc.approvedVersion, status.ApprovedVersion)
			}
			// and the current time
			if status.LatestVersionTimestamp != status.LastQueried {
				t.Errorf("LatestVersionTimestamp should've been set to LastQueried \n%q, not \n%q",
					status.LastQueried, status.LatestVersionTimestamp)
			}
		})
	}
}

func TestPrintFull(t *testing.T) {
	// GIVEN we have a Status with everything defined
	status := Status{
		ApprovedVersion:          "1.2.4",
		DeployedVersion:          "1.2.3",
		DeployedVersionTimestamp: "2022-01-01T01:01:01Z",
		LatestVersion:            "1.2.4",
		LatestVersionTimestamp:   "2022-01-01T01:01:01Z",
	}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we SetLastQueried
	status.Print("")

	// THEN a line would have been printed for each var
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = stdout
	want := 5
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}

func TestPrintEmpty(t *testing.T) {
	// GIVEN we have a Status with nothing defined
	status := Status{}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we SetLastQueried
	status.Print("")

	// THEN no lines would have been printed
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = stdout
	want := 0
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}
