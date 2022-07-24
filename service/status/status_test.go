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

package service_status

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestInitWithNil(t *testing.T) {
	// GIVEN we have a Status and no shoutrrrs or webhooks
	shoutrrrs := 0
	webhooks := 0
	commands := 0
	var status Status

	// WHEN Init is called
	status.Init(shoutrrrs, webhooks, commands, stringPtr("test"), nil)

	// THEN Fails will be empty
	if status.Fails.Shoutrrr != nil || status.Fails.WebHook != nil {
		t.Errorf("Init with %d shoutrrrs and %d webhooks should have nil Fails respectively, not %v",
			shoutrrrs, webhooks, status.Fails)
	}
}

func TestInitWithShoutrrs(t *testing.T) {
	// GIVEN we have a Status and some shoutrrs
	shoutrrrs := 4
	webhooks := 0
	commands := 0
	var status Status

	// WHEN Init is called
	status.Init(shoutrrrs, webhooks, commands, stringPtr("test"), nil)

	// THEN Fails will be empty
	got := len(status.Fails.Shoutrrr)
	if got != shoutrrrs {
		t.Errorf("Init with %d shoutrrrs should have made %d Fails, not %d",
			shoutrrrs, shoutrrrs, got)
	}
}

func TestInitWithWebHooks(t *testing.T) {
	// GIVEN we have a Status and some webhooks
	shoutrrrs := 0
	webhooks := 4
	commands := 0
	var status Status

	// WHEN Init is called
	status.Init(shoutrrrs, webhooks, commands, stringPtr("test"), nil)

	// THEN Fails will be empty
	got := len(status.Fails.WebHook)
	if got != webhooks {
		t.Errorf("Init with %d webhooks should have made %d Fails, not %d",
			webhooks, webhooks, got)
	}
}

func TestInitWithCommands(t *testing.T) {
	// GIVEN we have a Status and some commands
	shoutrrrs := 0
	webhooks := 0
	commands := 4
	var status Status

	// WHEN Init is called
	status.Init(shoutrrrs, webhooks, commands, stringPtr("test"), nil)

	// THEN Fails will be empty
	got := len(status.Fails.Command)
	if got != webhooks {
		t.Errorf("Init with %d commands should have made %d Fails, not %d",
			commands, commands, got)
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
	// GIVEN a Service with LatestVersion == ApprovedVersion
	status := testStatus()
	status.ApprovedVersion = status.LatestVersion

	// WHEN SetDeployedVersion is called on it
	status.SetDeployedVersion(status.LatestVersion)

	// THEN DeployedVersion is set to this version
	got := status.DeployedVersion
	want := status.LatestVersion
	if got != want {
		t.Errorf("Expected DeployedVersion to be set to %q, not %q",
			want, got)
	}
}

func TestSetDeployedVersionDidSetDeployedVersionTimestamp(t *testing.T) {
	// GIVEN a Service with LatestVersion == ApprovedVersion
	status := testStatus()
	status.ApprovedVersion = status.LatestVersion

	// WHEN SetDeployedVersion is called on it
	start := time.Now().UTC()
	status.SetDeployedVersion(status.LatestVersion)

	// THEN DeployedVersionTimestamp is set to now in time
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("DeployedVersionTimestamp was %v ago, not recent enough!",
			since)
	}
}

func TestSetDeployedVersionDidResetApprovedWhenMatch(t *testing.T) {
	// GIVEN a Service with LatestVersion == ApprovedVersion
	status := testStatus()
	status.ApprovedVersion = status.LatestVersion

	// WHEN SetDeployedVersion is called on it with this ApprovedVersion
	status.SetDeployedVersion(status.ApprovedVersion)

	// THEN ApprovedVersion is reset
	got := status.ApprovedVersion
	want := ""
	if got != want {
		t.Errorf("Expected ApprovedVersion to be reset to %q, not %q",
			want, got)
	}
}

func TestSetDeployedVersionDidntResetApprovedWhenMatch(t *testing.T) {
	// GIVEN a Service with LatestVersion != ApprovedVersion
	status := testStatus()
	status.ApprovedVersion = status.LatestVersion + "-beta"

	// WHEN SetDeployedVersion is called on it with the LatestVersion
	want := status.ApprovedVersion
	status.SetDeployedVersion(status.LatestVersion)

	// THEN ApprovedVersion is not reset
	got := status.ApprovedVersion
	if got != want {
		t.Errorf("ApprovedVersion shouldn't have changed and should still be %q, not %q",
			want, got)
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
		t.Run(name, func(t *testing.T) {
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
	// GIVEN a Service and a new version
	status := testStatus()
	version := "new"

	// WHEN SetLatestVersion is called on it
	status.SetLatestVersion(version)

	// THEN LatestVersion is set to this version
	got := status.LatestVersion
	if got != version {
		t.Errorf("Expected LatestVersion to be set to %q, not %q",
			version, got)
	}
}

func TestSetLatestVersionDidSetLatestVersionTimestamp(t *testing.T) {
	// GIVEN a Service and a new version
	status := testStatus()
	version := "new"

	// WHEN SetLatestVersion is called on it
	start := time.Now().UTC()
	status.SetLatestVersion(version)

	// THEN LatestVersionTimestamp is set to now in time
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("LatestVersionTimestamp was %v ago, not recent enough!",
			since)
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
	out, _ := ioutil.ReadAll(r)
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
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 0
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}
