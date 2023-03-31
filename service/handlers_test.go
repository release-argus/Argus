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

package service

import (
	"fmt"
	"testing"
	"time"

	command "github.com/release-argus/Argus/commands"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/webhook"
)

func TestService_UpdateLatestApproved(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		latestVersion        string
		startApprovedVersion string
		wantAnnounces        int
	}{
		"empty ApprovedVersion does announce": {
			startApprovedVersion: "",
			latestVersion:        "1.2.3",
			wantAnnounces:        1},
		"same ApprovedVersion doesn't announce": {
			startApprovedVersion: "1.2.3",
			latestVersion:        "1.2.3",
			wantAnnounces:        0},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc.Status.SetApprovedVersion(tc.startApprovedVersion)
			svc.Status.SetLatestVersion(tc.latestVersion, false)

			// WHEN UpdateLatestApproved is called on it
			want := svc.Status.GetLatestVersion()
			svc.UpdateLatestApproved()

			// THEN ApprovedVersion becomes LatestVersion
			got := svc.Status.GetApprovedVersion()
			if got != want {
				t.Errorf("LatestVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
			}
		})
	}
}

func TestService_UpdatedVersion(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		commands              command.Slice
		webhooks              webhook.Slice
		fails                 svcstatus.Fails
		latestIsDeployed      bool
		deployedVersion       *deployedver.Lookup
		approvedBecomesLatest bool
		deployedBecomesLatest bool
		wantAnnounces         int
	}{
		"doesn't do anything if DeployedVersion == LatestVersion": {
			latestIsDeployed:      true,
			wantAnnounces:         0,
			deployedBecomesLatest: true,
		},
		"no webhooks/command/deployedVersionLookup does announce and update deployed_version": {
			wantAnnounces:         1,
			deployedBecomesLatest: true,
		},
		"commands that have no fails does announce and update deployed_version": {
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			commands: command.Slice{
				{"true"}, {"false"}},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), boolPtr(false)}},
		},
		"commands that haven't run fails doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), nil}},
		},
		"commands that have failed doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), boolPtr(true)}},
		},
		"webhooks that have no fails does announce and update deployed_version": {
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			fails: svcstatus.Fails{
				WebHook: map[string]*bool{
					"0": boolPtr(false),
					"1": boolPtr(false)}},
		},
		"webhooks that haven't run fails doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			fails: svcstatus.Fails{
				WebHook: map[string]*bool{
					"0": boolPtr(false),
					"1": nil}},
		},
		"webhooks that have failed doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			fails: svcstatus.Fails{
				WebHook: map[string]*bool{
					"0": boolPtr(false),
					"1": boolPtr(true)}},
		},
		"commands and webhooks that have no fails does announce and update deployed_version": {
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			commands: command.Slice{
				{"true"}, {"false"}},
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), boolPtr(false)},
				WebHook: map[string]*bool{
					"0": boolPtr(false),
					"1": boolPtr(false)}},
		},
		"commands and webhooks that have no fails with deployedVersionLookup does announce and only update approved_version": {
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			approvedBecomesLatest: true,
			commands: command.Slice{
				{"true"}, {"false"}},
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			deployedVersion: &deployedver.Lookup{},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), boolPtr(false)},
				WebHook: map[string]*bool{
					"0": boolPtr(false),
					"1": boolPtr(false)}},
		},
		"deployedVersionLookup with no commands/webhooks doesn't announce or update deployed_version/approved_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			deployedVersion:       &deployedver.Lookup{},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), boolPtr(false)},
				WebHook: map[string]*bool{
					"0": boolPtr(false),
					"1": boolPtr(false)}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc.Command = tc.commands
			svc.WebHook = tc.webhooks
			svc.DeployedVersionLookup = tc.deployedVersion
			svc.Status.Fails = tc.fails
			if tc.latestIsDeployed {
				svc.Status.SetDeployedVersion(svc.Status.GetLatestVersion(), false)
			}

			// WHEN UpdatedVersion is called on it
			want := svc.Status.GetLatestVersion()
			svc.UpdatedVersion()

			// THEN ApprovedVersion becomes LatestVersion if there's a dvl and commands/webhooks
			got := svc.Status.GetApprovedVersion()
			if (tc.approvedBecomesLatest && got != want) || (!tc.approvedBecomesLatest && got == want) {
				t.Errorf("ApprovedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN DeployedVersion becomes LatestVersion if there's no dvl
			got = svc.Status.GetDeployedVersion()
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
			}
		})
	}
}

func TestService_HandleUpdateActions(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		commands              command.Slice
		webhooks              webhook.Slice
		autoApprove           bool
		deployedBecomesLatest bool
		wantAnnounces         int
	}{
		"no auto_approve and no webhooks/command does announce and update deployed_version": {
			autoApprove:           false,
			wantAnnounces:         1,
			deployedBecomesLatest: true,
		},
		"no auto_approve but do have webhooks only announces the new version": {
			autoApprove:           false,
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			webhooks: webhook.Slice{
				"fail": testWebHook(true)},
		},
		"auto_approve and webhook that fails only announces the fail and doesn't update deployed_version": {
			autoApprove:           true,
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			webhooks: webhook.Slice{
				"fail": testWebHook(true)},
		},
		"auto_approve and webhook that passes announces the pass and version change and updates deployed_version": {
			autoApprove:           true,
			wantAnnounces:         2,
			deployedBecomesLatest: true,
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
		},
		"auto_approve and command that fails only announces the fail and doesn't update deployed_version": {
			autoApprove:           true,
			wantAnnounces:         2,
			deployedBecomesLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}},
		},
		"auto_approve and command that passes announces the pass and version change and updates deployed_version": {
			autoApprove:           true,
			wantAnnounces:         3,
			deployedBecomesLatest: true,
			commands: command.Slice{
				{"true"}, {"ls"}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)
		svc.Command = tc.commands
		svc.WebHook = tc.webhooks
		svc.Status.Init(
			len(svc.Notify), len(svc.Command), len(svc.WebHook),
			&svc.ID,
			&svc.Dashboard.WebURL)
		if len(tc.commands) != 0 {
			svc.CommandController = &command.Controller{}
		}
		svc.CommandController.Init(
			&svc.Status,
			&svc.Command,
			nil,
			&svc.Options.Interval)
		svc.WebHook.Init(
			&svc.Status,
			&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
			nil,
			&svc.Options.Interval)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc.Dashboard.AutoApprove = &tc.autoApprove
			svc.DeployedVersionLookup = nil

			// WHEN HandleUpdateActions is called on it
			want := svc.Status.GetLatestVersion()
			svc.HandleUpdateActions()
			// wait until all commands/webhooks have run
			if tc.deployedBecomesLatest {
				time.Sleep(2 * time.Second)
			}
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.Command != nil {
					for j := range svc.Command {
						if (tc.deployedBecomesLatest && svc.Status.Fails.Command[j] != nil) ||
							(!tc.deployedBecomesLatest && svc.Status.Fails.Command[j] == nil) {
							actionsRan = false
							break
						}
					}
				}
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						if (tc.deployedBecomesLatest && svc.Status.Fails.WebHook[j] != nil) ||
							(!tc.deployedBecomesLatest && svc.Status.Fails.WebHook[j] == nil) {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("finished running after %v",
						time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !tc.autoApprove {
				if actionsRan && (len(svc.Status.Fails.Command) != 0 || len(svc.Status.Fails.WebHook) != 0) {
					t.Fatalf("no actions should have run as auto_approve is %t\n%#v",
						tc.autoApprove, svc.Status.Fails)
				}
			} else if !actionsRan {
				t.Fatal("actions didn't finish running")
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := svc.Status.GetDeployedVersion()
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				fails := ""
				if len(svc.Status.Fails.Command) != 0 {
					for i := range svc.Status.Fails.Command {
						fails += fmt.Sprintf("%d=%t, ", i, *svc.Status.Fails.Command[i])
					}
					t.Logf("commandFails: {%s}", fails[:len(fails)-2])
				}
				fails = ""
				if len(svc.Status.Fails.WebHook) != 0 {
					for i := range svc.Status.Fails.WebHook {
						fails += fmt.Sprintf("%s=%t, ", i, *svc.Status.Fails.WebHook[i])
					}
					t.Logf("webhookFails: {%s}", fails[:len(fails)-2])
				}
				for len(*svc.Status.AnnounceChannel) != 0 {
					msg := <-*svc.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
		})
	}
}

func TestService_HandleFailedActions(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		commands              command.Slice
		commandNextRunnables  []time.Time
		webhooks              webhook.Slice
		webhookNextRunnables  map[string]time.Time
		fails                 svcstatus.Fails
		wantFails             svcstatus.Fails
		deployedBecomesLatest bool
		deployedLatest        bool
		wantAnnounces         int
	}{
		"no command or webhooks fails retries all": {
			wantAnnounces: 2, // 2 = 1 command fail, 1 webhook fail
			commands:      command.Slice{{"false"}},
			webhooks: webhook.Slice{
				"will_fail": testWebHook(true)},
			fails: svcstatus.Fails{
				Command: []*bool{boolPtr(false)},
				WebHook: map[string]*bool{
					"will_fail": boolPtr(false)}},
			wantFails: svcstatus.Fails{
				Command: []*bool{boolPtr(true)},
				WebHook: map[string]*bool{
					"will_fail": boolPtr(true)}},
		},
		"have command fails and no webhook fails retries only the failed commands": {
			wantAnnounces:  3, // 2 = 2 command runs
			deployedLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}, {"true"}, {"false"}},
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(true), boolPtr(false), boolPtr(true), boolPtr(true)},
				WebHook: map[string]*bool{
					"pass": boolPtr(false)}},
			wantFails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), boolPtr(false), boolPtr(false), boolPtr(true)},
				WebHook: map[string]*bool{
					"pass": boolPtr(false)}},
		},
		"command fails before their next_runnable don't run": {
			wantAnnounces:  1, // 0 = no runs
			deployedLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}, {"true"}, {"false"}},
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(true), boolPtr(false), boolPtr(true), boolPtr(true)},
				WebHook: map[string]*bool{
					"pass": boolPtr(false)}},
			wantFails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false), boolPtr(false), boolPtr(true), boolPtr(true)},
				WebHook: map[string]*bool{
					"pass": boolPtr(false)}},
			commandNextRunnables: []time.Time{
				time.Now().UTC(),
				time.Now().UTC(),
				time.Now().UTC().Add(time.Minute),
				time.Now().UTC().Add(time.Minute)},
		},
		"have command fails no webhook fails and retries only the failed commands and updates deployed_version": {
			wantAnnounces:         2, // 2 = 1 command, 1 deployed
			deployedBecomesLatest: true,
			commands:              command.Slice{{"true"}, {"false"}},
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			fails: svcstatus.Fails{
				Command: []*bool{boolPtr(true), boolPtr(false)},
				WebHook: map[string]*bool{
					"pass": boolPtr(false)}},
			wantFails: svcstatus.Fails{
				Command: []*bool{nil, nil},
				WebHook: map[string]*bool{
					"pass": nil}},
		},
		"have webhook fails and no command fails retries only the failed commands": {
			wantAnnounces:  2, // 2 = 2 webhook runs
			deployedLatest: false,
			commands:       command.Slice{{"false"}},
			webhooks: webhook.Slice{
				"will_fail":  testWebHook(true),
				"will_pass":  testWebHook(false),
				"would_fail": testWebHook(true)},
			fails: svcstatus.Fails{
				Command: []*bool{boolPtr(false)},
				WebHook: map[string]*bool{
					"will_fail":  boolPtr(true),
					"will_pass":  boolPtr(true),
					"would_fail": boolPtr(false)}},
			wantFails: svcstatus.Fails{
				Command: []*bool{boolPtr(false)},
				WebHook: map[string]*bool{
					"will_fail":  boolPtr(true),
					"will_pass":  boolPtr(false),
					"would_fail": boolPtr(false)}},
		},
		"webhook fails before their next_runnable don't run": {
			wantAnnounces:  1, // 0 runs
			deployedLatest: false,
			commands: command.Slice{
				{"false"}},
			webhooks: webhook.Slice{
				"is_runnable":  testWebHook(false),
				"not_runnable": testWebHook(true),
				"would_fail":   testWebHook(true)},
			fails: svcstatus.Fails{
				Command: []*bool{boolPtr(false)},
				WebHook: map[string]*bool{
					"is_runnable":  boolPtr(true),
					"not_runnable": boolPtr(true),
					"would_fail":   boolPtr(false)}},
			wantFails: svcstatus.Fails{
				Command: []*bool{boolPtr(false)},
				WebHook: map[string]*bool{
					"is_runnable":  boolPtr(false),
					"not_runnable": boolPtr(true),
					"would_fail":   boolPtr(false)}},
			webhookNextRunnables: map[string]time.Time{"is_runnable": time.Now().UTC(), "not_runnable": time.Now().UTC().Add(time.Minute)},
		},
		"have webhook fails and no command fails retries only the failed commands and updates deployed_version": {
			wantAnnounces:         3, // 2 webhook runs
			deployedBecomesLatest: true,
			commands: command.Slice{
				{"false"}},
			webhooks: webhook.Slice{
				"will_pass0": testWebHook(false),
				"will_pass1": testWebHook(false),
				"would_fail": testWebHook(true)},
			fails: svcstatus.Fails{
				Command: []*bool{
					boolPtr(false)},
				WebHook: map[string]*bool{
					"will_pass0": boolPtr(true),
					"will_pass1": boolPtr(true),
					"would_fail": boolPtr(false)}},
			wantFails: svcstatus.Fails{
				Command: []*bool{
					nil},
				WebHook: map[string]*bool{
					"will_pass0": nil,
					"will_pass1": nil,
					"would_fail": nil}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc.Status.Init(
				len(svc.Notify), len(tc.commands), len(tc.webhooks),
				&svc.ID,
				&svc.Dashboard.WebURL)
			svc.Status.Fails = tc.fails
			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.GetLatestVersion(), false)
			}
			svc.Command = tc.commands
			if len(tc.commands) != 0 {
				svc.CommandController = &command.Controller{}
			}
			svc.CommandController.Init(
				&svc.Status,
				&svc.Command,
				nil,
				&svc.Options.Interval)
			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
				nil,
				&svc.Options.Interval)
			svc.DeployedVersionLookup = nil
			for i := range tc.commandNextRunnables {
				svc.CommandController.NextRunnable[i] = tc.commandNextRunnables[i]
			}
			for i := range tc.webhookNextRunnables {
				svc.WebHook[i].NextRunnable = tc.webhookNextRunnables[i]
			}

			// WHEN HandleFailedActions is called on it
			want := svc.Status.GetLatestVersion()
			svc.HandleFailedActions()
			// wait until all commands/webhooks have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.Command != nil {
					for j := range svc.Command {
						if stringifyPointer(svc.Status.Fails.Command[j]) != stringifyPointer(tc.wantFails.Command[j]) {
							actionsRan = false
							break
						}
					}
				}
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						if stringifyPointer(svc.Status.Fails.WebHook[j]) != stringifyPointer(tc.wantFails.WebHook[j]) {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("finished running after %v",
						time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !actionsRan {
				t.Error("actions didn't finish running or gave unexpected results")
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := svc.Status.GetDeployedVersion()
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				fails := ""
				if len(svc.Status.Fails.Command) != 0 {
					for i := range svc.Status.Fails.Command {
						fails += fmt.Sprintf("%d=%s, ", i, stringifyPointer(svc.Status.Fails.Command[i]))
					}
					t.Logf("commandFails: {%s}", fails[:len(fails)-2])
				}
				fails = ""
				if len(svc.Status.Fails.WebHook) != 0 {
					for i := range svc.Status.Fails.WebHook {
						fails += fmt.Sprintf("%s=%t, ", i, *svc.Status.Fails.WebHook[i])
					}
					t.Logf("webhookFails: {%s}", fails[:len(fails)-2])
				}
				for len(*svc.Status.AnnounceChannel) != 0 {
					msg := <-*svc.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the Command fails are as expected
			for i := range tc.wantFails.Command {
				if stringifyPointer(svc.Status.Fails.Command[i]) != stringifyPointer(tc.wantFails.Command[i]) {
					t.Errorf("got, command[%d]=%s, want %s",
						i, stringifyPointer(svc.Status.Fails.Command[i]), stringifyPointer(tc.wantFails.Command[i]))
				}
			}
			// THEN the WebHook fails are as expected
			for i := range tc.wantFails.WebHook {
				if stringifyPointer(svc.Status.Fails.WebHook[i]) != stringifyPointer(tc.wantFails.WebHook[i]) {
					t.Errorf("got, webhook[%s]=%s, want %s",
						i, stringifyPointer(svc.Status.Fails.WebHook[i]), stringifyPointer(tc.wantFails.WebHook[i]))
				}
			}
		})
	}
}

func TestService_HandleCommand(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		command               string
		commands              command.Slice
		nextRunnable          time.Time
		fails                 []*bool
		wantFails             []*bool
		deployedBecomesLatest bool
		deployedLatest        bool
		wantAnnounces         int
	}{
		"empty Command slice does nothing": {
			commands:              command.Slice{},
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
		},
		"Command that failed passes": {
			commands: command.Slice{
				{"ls", "-lah"}},
			command:               "ls -lah",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				boolPtr(true)},
			wantFails: []*bool{
				boolPtr(false)},
		},
		"Command that passed fails": {
			commands:              command.Slice{{"false"}},
			command:               "false",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				boolPtr(false)},
			wantFails: []*bool{
				boolPtr(true)},
		},
		"Command that's not runnable doesn't run": {
			commands:              command.Slice{{"false"}},
			command:               "false",
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				boolPtr(false)},
			wantFails: []*bool{
				boolPtr(false)},
			nextRunnable: time.Now().UTC().Add(time.Minute),
		},
		"Command that's runnable does run": {
			commands:              command.Slice{{"false"}},
			command:               "false",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				boolPtr(false)},
			wantFails: []*bool{
				boolPtr(true)},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.GetLatestVersion(), false)
			}
			svc.Status.Fails.Command = tc.fails
			svc.Command = tc.commands
			if len(tc.commands) != 0 {
				svc.CommandController = &command.Controller{}
			}
			svc.CommandController.Init(
				&svc.Status,
				&svc.Command,
				nil,
				&svc.Options.Interval)
			svc.DeployedVersionLookup = nil
			for i := range svc.Command {
				svc.CommandController.NextRunnable[i] = tc.nextRunnable
			}

			// WHEN HandleCommand is called on it
			want := svc.Status.GetLatestVersion()
			svc.HandleCommand(tc.command)
			// wait until all commands have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.Command != nil {
					for j := range svc.Command {
						if stringifyPointer(svc.Status.Fails.Command[j]) != stringifyPointer(tc.wantFails[j]) {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("finished running after %v",
						time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !actionsRan {
				t.Error("actions didn't finish running or gave unexpected results")
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := svc.Status.GetDeployedVersion()
			if !tc.deployedLatest && ((tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want)) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				fails := ""
				if len(svc.Status.Fails.Command) != 0 {
					for i := range svc.Status.Fails.Command {
						fails += fmt.Sprintf("%d=%t, ", i, *svc.Status.Fails.Command[i])
					}
					t.Logf("commandFails: {%s}", fails[:len(fails)-2])
				}
				for len(*svc.Status.AnnounceChannel) != 0 {
					msg := <-*svc.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the Command fails are as expected
			for i := range tc.wantFails {
				if stringifyPointer(svc.Status.Fails.Command[i]) != stringifyPointer(tc.wantFails[i]) {
					t.Errorf("got, command[%d]=%s, want %s",
						i, stringifyPointer(svc.Status.Fails.Command[i]), stringifyPointer(tc.wantFails[i]))
				}
			}
		})
	}
}

func TestService_HandleWebHook(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		webhook               string
		webhooks              webhook.Slice
		nextRunnable          time.Time
		fails                 map[string]*bool
		wantFails             map[string]*bool
		deployedBecomesLatest bool
		deployedLatest        bool
		wantAnnounces         int
	}{
		"empty WebHook slice does nothing": {
			webhooks:              webhook.Slice{},
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
		},
		"WebHook that failed passes": {
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			webhook:               "pass",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": boolPtr(true)},
			wantFails: map[string]*bool{
				"pass": boolPtr(false)},
		},
		"WebHook that passed fails": {
			webhooks: webhook.Slice{
				"fail": testWebHook(true)},
			webhook:               "fail",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"fail": boolPtr(false)},
			wantFails: map[string]*bool{
				"fail": boolPtr(true)},
		},
		"WebHook that's not runnable doesn't run": {
			webhooks: webhook.Slice{
				"pass": testWebHook(true)},
			webhook:               "pass",
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": boolPtr(false)},
			wantFails: map[string]*bool{
				"pass": boolPtr(false)},
			nextRunnable: time.Now().UTC().Add(time.Minute),
		},
		"WebHook that's runnable does run": {
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			webhook:               "pass",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": boolPtr(true)},
			wantFails: map[string]*bool{
				"pass": boolPtr(false)},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)
		svc.Status.Init(
			len(svc.Notify), len(tc.webhooks), 0,
			&svc.ID,
			&svc.Dashboard.WebURL)
		svc.WebHook = tc.webhooks
		svc.WebHook.Init(
			&svc.Status,
			&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
			nil,
			&svc.Options.Interval)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.GetLatestVersion(), false)
			}
			svc.Status.Fails.WebHook = tc.fails
			svc.DeployedVersionLookup = nil
			for i := range svc.WebHook {
				svc.WebHook[i].NextRunnable = tc.nextRunnable
			}

			// WHEN HandleWebHook is called on it
			want := svc.Status.GetLatestVersion()
			svc.HandleWebHook(tc.webhook)
			// wait until all webhooks have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						if stringifyPointer(svc.Status.Fails.WebHook[j]) != stringifyPointer(tc.wantFails[j]) {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("finished running after %v",
						time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !actionsRan {
				t.Error("actions didn't finish running or gave unexpected results")
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := svc.Status.GetDeployedVersion()
			if !tc.deployedLatest && ((tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want)) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				fails := ""
				if len(svc.Status.Fails.WebHook) != 0 {
					for i := range svc.Status.Fails.WebHook {
						fails += fmt.Sprintf("%s=%t, ", i, *svc.Status.Fails.WebHook[i])
					}
					t.Logf("webhookFails: {%s}", fails[:len(fails)-2])
				}
				for len(*svc.Status.AnnounceChannel) != 0 {
					msg := <-*svc.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the WebHook fails are as expected
			for i := range tc.wantFails {
				if stringifyPointer(svc.Status.Fails.WebHook[i]) != stringifyPointer(tc.wantFails[i]) {
					t.Errorf("got, webhook[%s]=%s, want %s",
						i, stringifyPointer(svc.Status.Fails.WebHook[i]), stringifyPointer(tc.wantFails[i]))
				}
			}
		})
	}
}

func TestService_HandleSkip(t *testing.T) {
	// GIVEN a Service
	testLogging()
	latestVersion := "1.2.3"
	tests := map[string]struct {
		skipVersion          string
		approvedVersion      string
		wantAnnounces        int
		wantDatabaseMessages int
		prepDelete           bool
	}{
		"skip of not latest version does nothing": {
			skipVersion: latestVersion + "-beta"},
		"skip of latest version skips version and announces to announce and database channels": {
			skipVersion:          latestVersion,
			approvedVersion:      "SKIP_" + latestVersion,
			wantAnnounces:        1,
			wantDatabaseMessages: 1},
		"skip of latest version but PrepDelete has nil'd the database channel": {
			skipVersion:          latestVersion,
			approvedVersion:      "SKIP_" + latestVersion,
			prepDelete:           true,
			wantAnnounces:        0,
			wantDatabaseMessages: 0},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			// t.Parallel()
			svc.Status.SetApprovedVersion("")
			svc.Status.SetLatestVersion(latestVersion, false)
			if tc.prepDelete {
				svc.PrepDelete()
			}

			// WHEN HandleSkip is called on it
			svc.HandleSkip(tc.skipVersion)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			if tc.approvedVersion != svc.Status.GetApprovedVersion() {
				t.Errorf("ApprovedVersion should have changed to %q not %q",
					tc.approvedVersion, svc.Status.GetApprovedVersion())
			}
			// AND the correct number of changes are announced to the announce channel
			if tc.prepDelete {
				if svc.Status.AnnounceChannel != nil || svc.Status.DatabaseChannel != nil {
					t.Errorf("AnnounceChannel and DatabaseChannel should be nil but are not")
				}
				return
			}
			// AND the correct number of changes are announced to the announce channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
			}
			// AND the correct number of messages are announced to the database channel
			if len(*svc.Status.DatabaseChannel) != tc.wantDatabaseMessages {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantDatabaseMessages, len(*svc.Status.DatabaseChannel))
			}
		})
	}
}

func TestService_ShouldRetryAll(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		command []*bool
		webhook map[string]*bool
		want    bool
	}{
		"no commands or webhooks": {
			want: true,
		},
		"commands that haven't run": {
			command: []*bool{
				nil,
				nil},
			want: false,
		},
		"commands that have failed": {
			command: []*bool{
				boolPtr(true),
				boolPtr(true)},
			want: false,
		},
		"commands that have failed/haven't run": {
			command: []*bool{
				boolPtr(true),
				nil},
			want: false,
		},
		"commands that haven't failed": {
			command: []*bool{
				boolPtr(false),
				boolPtr(false)},
			want: true,
		},
		"mix of all command fail states": {
			command: []*bool{
				boolPtr(true),
				boolPtr(false),
				nil},
			want: false,
		},
		"webhooks that haven't run": {
			webhook: map[string]*bool{
				"1": nil,
				"2": nil},
			want: false,
		},
		"webhooks that have failed": {
			webhook: map[string]*bool{
				"1": boolPtr(true),
				"2": boolPtr(true)},
			want: false,
		},
		"webhooks that have failed/haven't run": {
			webhook: map[string]*bool{
				"1": boolPtr(true),
				"2": nil},
			want: false,
		},
		"webhooks that haven't failed": {
			webhook: map[string]*bool{
				"1": boolPtr(false),
				"2": boolPtr(false)},
			want: true,
		},
		"mix of all webhook fail states": {
			webhook: map[string]*bool{
				"1": boolPtr(true),
				"2": boolPtr(false),
				"3": nil},
			want: false,
		},
		"mix of all webhook and command fail states": {
			command: []*bool{
				boolPtr(true), boolPtr(false), nil},
			webhook: map[string]*bool{
				"1": boolPtr(true),
				"2": boolPtr(false),
				"3": nil},
			want: false,
		},
		"mix of all webhook and command no fails": {
			command: []*bool{
				boolPtr(false), boolPtr(false), boolPtr(false)},
			webhook: map[string]*bool{
				"1": boolPtr(false),
				"2": boolPtr(false),
				"3": boolPtr(false)},
			want: true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			commands := len(tc.command)
			svc.Command = command.Slice{}
			for commands != 0 {
				svc.Command = append(svc.Command, command.Command{})
				commands--
			}
			webhooks := len(tc.webhook)
			svc.WebHook = webhook.Slice{}
			for webhooks != 0 {
				svc.WebHook[fmt.Sprint(webhooks)] = &webhook.WebHook{}
				webhooks--
			}
			svc.Status.Fails.Command = tc.command
			svc.Status.Fails.WebHook = tc.webhook

			// WHEN shouldRetryAll is called on it
			got := svc.shouldRetryAll()

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			if tc.want != got {
				t.Errorf("want %t not %t",
					tc.want, got)
			}
		})
	}
}
