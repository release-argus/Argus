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

package service

import (
	"fmt"
	"testing"
	"time"

	command "github.com/release-argus/Argus/commands"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/test"
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
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc.Status.SetApprovedVersion(tc.startApprovedVersion, false)
			svc.Status.SetLatestVersion(tc.latestVersion, false)

			// WHEN UpdateLatestApproved is called on it
			want := svc.Status.LatestVersion()
			svc.UpdateLatestApproved()

			// THEN ApprovedVersion becomes LatestVersion
			got := svc.Status.ApprovedVersion()
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
		commandFails          []*bool
		webhooks              webhook.Slice
		webhookFails          map[string]*bool
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
			commandFails: []*bool{
				test.BoolPtr(false), test.BoolPtr(false)},
		},
		"commands that haven't run fails doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}},
			commandFails: []*bool{
				test.BoolPtr(false), nil},
		},
		"commands that have failed doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}},
			commandFails: []*bool{
				test.BoolPtr(false), test.BoolPtr(true)},
		},
		"webhooks that have no fails does announce and update deployed_version": {
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			webhookFails: map[string]*bool{
				"0": test.BoolPtr(false),
				"1": test.BoolPtr(false)},
		},
		"webhooks that haven't run fails doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			webhookFails: map[string]*bool{
				"0": test.BoolPtr(false),
				"1": nil},
		},
		"webhooks that have failed doesn't announce or update deployed_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			webhookFails: map[string]*bool{
				"0": test.BoolPtr(false),
				"1": test.BoolPtr(true)},
		},
		"commands and webhooks that have no fails does announce and update deployed_version": {
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			commands: command.Slice{
				{"true"}, {"false"}},
			webhooks: webhook.Slice{
				"0": {},
				"1": {}},
			commandFails: []*bool{
				test.BoolPtr(false), test.BoolPtr(false)},
			webhookFails: map[string]*bool{
				"0": test.BoolPtr(false),
				"1": test.BoolPtr(false)},
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
			commandFails: []*bool{
				test.BoolPtr(false), test.BoolPtr(false)},
			webhookFails: map[string]*bool{
				"0": test.BoolPtr(false),
				"1": test.BoolPtr(false)},
		},
		"deployedVersionLookup with no commands/webhooks doesn't announce or update deployed_version/approved_version": {
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			deployedVersion:       &deployedver.Lookup{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := testServiceURL(name)
			svc.Command = tc.commands
			svc.WebHook = tc.webhooks
			svc.Status.Init(
				0, len(svc.Command), len(svc.WebHook),
				&svc.ID,
				&svc.Dashboard.WebURL)
			svc.DeployedVersionLookup = tc.deployedVersion
			for i := range tc.commandFails {
				if tc.commandFails[i] != nil {
					svc.Status.Fails.Command.Set(i, *tc.commandFails[i])
				}
			}
			for i := range tc.webhookFails {
				svc.Status.Fails.WebHook.Set(i, tc.webhookFails[i])
			}
			if tc.latestIsDeployed {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), false)
			}

			// WHEN UpdatedVersion is called on it
			startLV := svc.Status.LatestVersion()
			svc.UpdatedVersion(true)

			// THEN ApprovedVersion becomes LatestVersion if there's a dvl and commands/webhooks
			gotAV := svc.Status.ApprovedVersion()
			if (tc.approvedBecomesLatest && gotAV != startLV) ||
				(!tc.approvedBecomesLatest && gotAV == startLV) {
				t.Errorf("ApprovedVersion should have changed to %q not %q",
					startLV, gotAV)
			}
			// THEN DeployedVersion becomes LatestVersion if there's no dvl
			gotDV := svc.Status.DeployedVersion()
			if (tc.deployedBecomesLatest && gotDV != startLV) ||
				(!tc.deployedBecomesLatest && gotDV == startLV) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					startLV, gotDV)
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
			&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{},
			nil,
			&svc.Options.Interval)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc.Dashboard.AutoApprove = &tc.autoApprove
			svc.DeployedVersionLookup = nil

			// WHEN HandleUpdateActions is called on it
			want := svc.Status.LatestVersion()
			svc.HandleUpdateActions(true)
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
						commandFailed := svc.Status.Fails.Command.Get(j)
						if (tc.deployedBecomesLatest && commandFailed != nil) ||
							(!tc.deployedBecomesLatest && commandFailed == nil) {
							actionsRan = false
							break
						}
					}
				}
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						webhookFailed := svc.Status.Fails.WebHook.Get(j)
						if (tc.deployedBecomesLatest && webhookFailed != nil) ||
							(!tc.deployedBecomesLatest && webhookFailed == nil) {
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
				if actionsRan {
					ranCommand := false
					for i := range svc.Command {
						if svc.Status.Fails.Command.Get(i) != nil {
							ranCommand = true
							break
						}
					}
					for i := range svc.WebHook {
						if svc.Status.Fails.WebHook.Get(i) != nil {
							ranCommand = true
							break
						}
					}
					if ranCommand {
						t.Fatalf("no actions should have run as auto_approve is %t\n%#v",
							tc.autoApprove, svc.Status.Fails.String())
					}
				}
			} else if !actionsRan {
				t.Fatal("actions didn't finish running")
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := svc.Status.DeployedVersion()
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				t.Logf("Fails: %s", svc.Status.Fails.String())
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
	tests := map[string]struct {
		commands              command.Slice
		commandNextRunnables  []time.Time
		webhooks              webhook.Slice
		webhookNextRunnables  map[string]time.Time
		startFailsComand      []*bool
		wantFailsCommand      []*bool
		startFailsWebHook     map[string]*bool
		wantFailsWebHook      map[string]*bool
		deployedBecomesLatest bool
		deployedLatest        bool
		wantAnnounces         int
	}{
		"no command or webhooks fails retries all": {
			wantAnnounces: 3, // 3 = 2 command fail, 1 webhook fail
			commands: command.Slice{
				{"false"}, {"false"}},
			webhooks: webhook.Slice{
				"will_fail": testWebHook(true)},
			startFailsComand: []*bool{
				nil, nil},
			wantFailsCommand: []*bool{
				test.BoolPtr(true), test.BoolPtr(true)},
			startFailsWebHook: map[string]*bool{
				"will_fail": nil},
			wantFailsWebHook: map[string]*bool{
				"will_fail": test.BoolPtr(true)},
		},
		"have command fails and no webhook fails retries only the failed commands": {
			wantAnnounces:  3, // 3 = 2 command pass, 1 command fail
			deployedLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}, {"true"}, {"false"}},
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			startFailsComand: []*bool{
				test.BoolPtr(true), test.BoolPtr(false), test.BoolPtr(true), test.BoolPtr(true)},
			wantFailsCommand: []*bool{
				test.BoolPtr(false), test.BoolPtr(false), test.BoolPtr(false), test.BoolPtr(true)},
			startFailsWebHook: map[string]*bool{
				"pass": test.BoolPtr(false)},
			wantFailsWebHook: map[string]*bool{
				"pass": test.BoolPtr(false)},
		},
		"command fails before their next_runnable don't run": {
			wantAnnounces:  1, // 0 = no runs
			deployedLatest: false,
			commands: command.Slice{
				{"true"}, {"false"}, {"true"}, {"false"}},
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			startFailsComand: []*bool{
				test.BoolPtr(true), test.BoolPtr(false), test.BoolPtr(true), test.BoolPtr(true)},
			wantFailsCommand: []*bool{
				test.BoolPtr(false), test.BoolPtr(false), test.BoolPtr(true), test.BoolPtr(true)},
			startFailsWebHook: map[string]*bool{
				"pass": test.BoolPtr(false)},
			wantFailsWebHook: map[string]*bool{
				"pass": test.BoolPtr(false)},
			commandNextRunnables: []time.Time{
				time.Now().UTC(),
				time.Now().UTC(),
				time.Now().UTC().Add(time.Minute),
				time.Now().UTC().Add(time.Minute)},
		},
		"have command fails no webhook fails and retries only the failed commands and updates deployed_version": {
			wantAnnounces:         2, // 2 = 1 command, 1 deployed
			deployedBecomesLatest: true,
			commands: command.Slice{
				{"true"}, {"false"}},
			webhooks: webhook.Slice{
				"pass": testWebHook(false)},
			startFailsComand: []*bool{test.BoolPtr(true), test.BoolPtr(false)},
			wantFailsCommand: []*bool{
				nil, nil},
			startFailsWebHook: map[string]*bool{
				"pass": test.BoolPtr(false)},
			wantFailsWebHook: map[string]*bool{
				"pass": nil},
		},
		"have webhook fails and no command fails retries only the failed commands": {
			wantAnnounces:  2, // 2 = 2 webhook runs
			deployedLatest: false,
			commands:       command.Slice{{"false"}},
			webhooks: webhook.Slice{
				"will_fail":  testWebHook(true),
				"will_pass":  testWebHook(false),
				"would_fail": testWebHook(true)},
			startFailsComand: []*bool{
				test.BoolPtr(false)},
			wantFailsCommand: []*bool{
				test.BoolPtr(false)},
			startFailsWebHook: map[string]*bool{
				"will_fail":  test.BoolPtr(true),
				"will_pass":  test.BoolPtr(true),
				"would_fail": test.BoolPtr(false)},
			wantFailsWebHook: map[string]*bool{
				"will_fail":  test.BoolPtr(true),
				"will_pass":  test.BoolPtr(false),
				"would_fail": test.BoolPtr(false)},
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
			startFailsComand: []*bool{
				test.BoolPtr(false)},
			wantFailsCommand: []*bool{
				test.BoolPtr(false)},
			startFailsWebHook: map[string]*bool{
				"is_runnable":  test.BoolPtr(true),
				"not_runnable": test.BoolPtr(true),
				"would_fail":   test.BoolPtr(false)},
			wantFailsWebHook: map[string]*bool{
				"is_runnable":  test.BoolPtr(false),
				"not_runnable": test.BoolPtr(true),
				"would_fail":   test.BoolPtr(false)},
			webhookNextRunnables: map[string]time.Time{
				"is_runnable":  time.Now().UTC(),
				"not_runnable": time.Now().UTC().Add(time.Minute)},
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
			startFailsComand: []*bool{
				test.BoolPtr(false)},
			wantFailsCommand: []*bool{
				nil},
			startFailsWebHook: map[string]*bool{
				"will_pass0": test.BoolPtr(true),
				"will_pass1": test.BoolPtr(true),
				"would_fail": test.BoolPtr(false)},
			wantFailsWebHook: map[string]*bool{
				"will_pass0": nil,
				"will_pass1": nil,
				"would_fail": nil},
		},
	}

	for name, tc := range tests {
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc.Status.Init(
				len(svc.Notify), len(tc.commands), len(tc.webhooks),
				&svc.ID,
				&svc.Dashboard.WebURL)
			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), false)
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
			for k, v := range tc.startFailsComand {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{},
				nil,
				&svc.Options.Interval)
			for k, v := range tc.startFailsWebHook {
				svc.Status.Fails.WebHook.Set(k, v)
			}
			svc.DeployedVersionLookup = nil
			for i := range tc.commandNextRunnables {
				nextRunnable := tc.commandNextRunnables[i]
				svc.CommandController.SetNextRunnable(i, &nextRunnable)
			}
			for i := range tc.webhookNextRunnables {
				nextRunnable := tc.webhookNextRunnables[i]
				svc.WebHook[i].SetNextRunnable(&nextRunnable)
			}

			// WHEN HandleFailedActions is called on it
			want := svc.Status.LatestVersion()
			svc.HandleFailedActions()
			// wait until all commands/webhooks have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.Command != nil {
					for j := range svc.Command {
						if test.StringifyPtr(svc.Status.Fails.Command.Get(j)) != test.StringifyPtr(tc.wantFailsCommand[j]) {
							actionsRan = false
							break
						}
					}
				}
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						if test.StringifyPtr(svc.Status.Fails.WebHook.Get(j)) != test.StringifyPtr(tc.wantFailsWebHook[j]) {
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
			got := svc.Status.DeployedVersion()
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				t.Logf("Fails: {%s}", svc.Status.Fails.String())
				for len(*svc.Status.AnnounceChannel) != 0 {
					msg := <-*svc.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the Command fails are as expected
			for i := range tc.wantFailsCommand {
				if test.StringifyPtr(svc.Status.Fails.Command.Get(i)) != test.StringifyPtr(tc.wantFailsCommand[i]) {
					t.Errorf("got, command[%d]=%s, want %s",
						i, test.StringifyPtr(svc.Status.Fails.Command.Get(i)), test.StringifyPtr(tc.wantFailsCommand[i]))
				}
			}
			// THEN the WebHook fails are as expected
			for i := range tc.wantFailsWebHook {
				if test.StringifyPtr(svc.Status.Fails.WebHook.Get(i)) != test.StringifyPtr(tc.wantFailsWebHook[i]) {
					t.Errorf("got, webhook[%s]=%s, want %s",
						i, test.StringifyPtr(svc.Status.Fails.WebHook.Get(i)), test.StringifyPtr(tc.wantFailsWebHook[i]))
				}
			}
		})
	}
}

func TestService_HandleCommand(t *testing.T) {
	// GIVEN a Service
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
				test.BoolPtr(true)},
			wantFails: []*bool{
				test.BoolPtr(false)},
		},
		"Command that passed fails": {
			commands:              command.Slice{{"false"}},
			command:               "false",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				test.BoolPtr(false)},
			wantFails: []*bool{
				test.BoolPtr(true)},
		},
		"Command that's not runnable doesn't run": {
			commands:              command.Slice{{"false"}},
			command:               "false",
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				test.BoolPtr(false)},
			wantFails: []*bool{
				test.BoolPtr(false)},
			nextRunnable: time.Now().UTC().Add(time.Minute),
		},
		"Command that's runnable does run": {
			commands:              command.Slice{{"false"}},
			command:               "false",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				test.BoolPtr(false)},
			wantFails: []*bool{
				test.BoolPtr(true)},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
	}

	for name, tc := range tests {
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), false)
			}
			svc.Command = tc.commands
			if len(tc.commands) != 0 {
				svc.CommandController = &command.Controller{}
			}
			svc.Status.Init(
				len(svc.Notify), len(svc.Command), len(svc.WebHook),
				&svc.ID,
				&svc.Dashboard.WebURL)
			for k, v := range tc.fails {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			svc.CommandController.Init(
				&svc.Status,
				&svc.Command,
				nil,
				&svc.Options.Interval)
			svc.DeployedVersionLookup = nil
			for i := range svc.Command {
				svc.CommandController.SetNextRunnable(i, &tc.nextRunnable)
			}

			// WHEN HandleCommand is called on it
			want := svc.Status.LatestVersion()
			svc.HandleCommand(tc.command)
			// wait until all commands have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.Command != nil {
					for j := range svc.Command {
						if test.StringifyPtr(svc.Status.Fails.Command.Get(j)) != test.StringifyPtr(tc.wantFails[j]) {
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
			got := svc.Status.DeployedVersion()
			if !tc.deployedLatest && ((tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want)) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				fails := ""
				for i := range svc.Command {
					fails += fmt.Sprintf("%d=%s, ", i, test.StringifyPtr(svc.Status.Fails.Command.Get(i)))
				}
				t.Logf("commandFails: {%s}", fails[:len(fails)-2])
				for len(*svc.Status.AnnounceChannel) != 0 {
					msg := <-*svc.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the Command fails are as expected
			for i := range tc.wantFails {
				if test.StringifyPtr(svc.Status.Fails.Command.Get(i)) != test.StringifyPtr(tc.wantFails[i]) {
					t.Errorf("got, command[%d]=%s, want %s",
						i, test.StringifyPtr(svc.Status.Fails.Command.Get(i)), test.StringifyPtr(tc.wantFails[i]))
				}
			}
		})
	}
}

func TestService_HandleWebHook(t *testing.T) {
	// GIVEN a Service
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
				"pass": test.BoolPtr(true)},
			wantFails: map[string]*bool{
				"pass": test.BoolPtr(false)},
		},
		"WebHook that passed fails": {
			webhooks: webhook.Slice{
				"fail": testWebHook(true)},
			webhook:               "fail",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"fail": test.BoolPtr(false)},
			wantFails: map[string]*bool{
				"fail": test.BoolPtr(true)},
		},
		"WebHook that's not runnable doesn't run": {
			webhooks: webhook.Slice{
				"pass": testWebHook(true)},
			webhook:               "pass",
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": test.BoolPtr(false)},
			wantFails: map[string]*bool{
				"pass": test.BoolPtr(false)},
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
				"pass": test.BoolPtr(true)},
			wantFails: map[string]*bool{
				"pass": test.BoolPtr(false)},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
	}

	for name, tc := range tests {
		svc := testServiceURL(name)
		svc.Status.Init(
			len(svc.Notify), len(tc.webhooks), 0,
			&svc.ID,
			&svc.Dashboard.WebURL)
		svc.WebHook = tc.webhooks
		svc.WebHook.Init(
			&svc.Status,
			&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{},
			nil,
			&svc.Options.Interval)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), false)
			}
			for k, v := range tc.fails {
				svc.Status.Fails.WebHook.Set(k, v)
			}
			svc.DeployedVersionLookup = nil
			for i := range svc.WebHook {
				svc.WebHook[i].SetNextRunnable(&tc.nextRunnable)
			}

			// WHEN HandleWebHook is called on it
			want := svc.Status.LatestVersion()
			svc.HandleWebHook(tc.webhook)
			// wait until all webhooks have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						if test.StringifyPtr(svc.Status.Fails.WebHook.Get(j)) != test.StringifyPtr(tc.wantFails[j]) {
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
			got := svc.Status.DeployedVersion()
			if !tc.deployedLatest && ((tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want)) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*svc.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*svc.Status.AnnounceChannel))
				fails := ""
				for i := range svc.WebHook {
					fails += fmt.Sprintf("%s=%s, ", i, test.StringifyPtr(svc.Status.Fails.WebHook.Get(i)))
				}
				t.Logf("webhookFails: {%s}", fails[:len(fails)-2])
				for len(*svc.Status.AnnounceChannel) != 0 {
					msg := <-*svc.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the WebHook fails are as expected
			for i := range tc.wantFails {
				if test.StringifyPtr(svc.Status.Fails.WebHook.Get(i)) != test.StringifyPtr(tc.wantFails[i]) {
					t.Errorf("got, webhook[%s]=%s, want %s",
						i, test.StringifyPtr(svc.Status.Fails.WebHook.Get(i)), test.StringifyPtr(tc.wantFails[i]))
				}
			}
		})
	}
}

func TestService_HandleSkip(t *testing.T) {
	// GIVEN a Service
	latestVersion := "1.2.3"
	tests := map[string]struct {
		startVersion         string
		approvedVersion      string
		wantAnnounces        int
		wantDatabaseMessages int
		prepDelete           bool
	}{
		"skip of latest version does nothing": {
			startVersion:         latestVersion,
			approvedVersion:      "",
			wantAnnounces:        0,
			wantDatabaseMessages: 0},
		"skip of version that's not latest skips version and announces to announce and database channels": {
			startVersion:         "1.0.0",
			approvedVersion:      "SKIP_" + latestVersion,
			wantAnnounces:        1,
			wantDatabaseMessages: 1},
		"skip of version that's not latest but PrepDelete has nil'd the database channel": {
			startVersion:         "0.2.3",
			approvedVersion:      "SKIP_" + latestVersion,
			prepDelete:           true,
			wantAnnounces:        0,
			wantDatabaseMessages: 0},
	}

	for name, tc := range tests {
		svc := testServiceURL(name)

		t.Run(name, func(t *testing.T) {
			// t.Parallel() - cannot run in parallel as it uses the same channel

			svc.Status.SetDeployedVersion(tc.startVersion, false)
			svc.Status.SetApprovedVersion("", false)
			svc.Status.SetLatestVersion(latestVersion, false)
			if tc.prepDelete {
				svc.PrepDelete(true)
			}

			// WHEN HandleSkip is called on it
			svc.HandleSkip()

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			if tc.approvedVersion != svc.Status.ApprovedVersion() {
				t.Errorf("ApprovedVersion should have changed to %q not %q",
					tc.approvedVersion, svc.Status.ApprovedVersion())
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
				test.BoolPtr(true),
				test.BoolPtr(true)},
			want: false,
		},
		"commands that have failed/haven't run": {
			command: []*bool{
				test.BoolPtr(true),
				nil},
			want: false,
		},
		"commands that haven't failed": {
			command: []*bool{
				test.BoolPtr(false),
				test.BoolPtr(false)},
			want: true,
		},
		"mix of all command fail states": {
			command: []*bool{
				test.BoolPtr(true),
				test.BoolPtr(false),
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
				"1": test.BoolPtr(true),
				"2": test.BoolPtr(true)},
			want: false,
		},
		"webhooks that have failed/haven't run": {
			webhook: map[string]*bool{
				"1": test.BoolPtr(true),
				"2": nil},
			want: false,
		},
		"webhooks that haven't failed": {
			webhook: map[string]*bool{
				"1": test.BoolPtr(false),
				"2": test.BoolPtr(false)},
			want: true,
		},
		"mix of all webhook fail states": {
			webhook: map[string]*bool{
				"1": test.BoolPtr(true),
				"2": test.BoolPtr(false),
				"3": nil},
			want: false,
		},
		"mix of all webhook and command fail states": {
			command: []*bool{
				test.BoolPtr(true), test.BoolPtr(false), nil},
			webhook: map[string]*bool{
				"1": test.BoolPtr(true),
				"2": test.BoolPtr(false),
				"3": nil},
			want: false,
		},
		"mix of all webhook and command no fails": {
			command: []*bool{
				test.BoolPtr(false), test.BoolPtr(false), test.BoolPtr(false)},
			webhook: map[string]*bool{
				"1": test.BoolPtr(false),
				"2": test.BoolPtr(false),
				"3": test.BoolPtr(false)},
			want: true,
		},
	}

	for name, tc := range tests {
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
			svc.Status.Init(
				0, len(svc.Command), len(svc.WebHook),
				&name,
				nil)
			for k, v := range tc.command {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			for k, v := range tc.webhook {
				svc.Status.Fails.WebHook.Set(k, v)
			}

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
