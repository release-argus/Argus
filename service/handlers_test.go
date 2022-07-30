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
	"github.com/release-argus/Argus/service/deployed_version"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/webhook"
)

func TestUpdateLatestApproved(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		latestVersion        string
		startApprovedVersion string
		wantAnnounces        int
	}{
		"empty ApprovedVersion does announce":   {startApprovedVersion: "", latestVersion: "1.2.3", wantAnnounces: 1},
		"same ApprovedVersion doesn't announce": {startApprovedVersion: "1.2.3", latestVersion: "1.2.3", wantAnnounces: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			service.Status.ApprovedVersion = tc.startApprovedVersion
			service.Status.LatestVersion = tc.latestVersion

			// WHEN UpdateLatestApproved is called on it
			want := service.Status.LatestVersion
			service.UpdateLatestApproved()

			// THEN ApprovedVersion becomes LatestVersion
			got := service.Status.ApprovedVersion
			if got != want {
				t.Errorf("%s:\nLatestVersion should have changed to %q not %q",
					name, want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*service.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("%s:\nExpecting %d announce message but got %d",
					name, tc.wantAnnounces, len(*service.Status.AnnounceChannel))
			}
		})
	}
}

func TestUpdatedVersion(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		commands              command.Slice
		webhooks              webhook.Slice
		fails                 service_status.Fails
		latestIsDeployed      bool
		deployedVersion       *deployed_version.Lookup
		approvedBecomesLatest bool
		deployedBecomesLatest bool
		wantAnnounces         int
	}{
		"doesn't do anything if DeployedVersion == LatestVersion":                             {latestIsDeployed: true, wantAnnounces: 0, deployedBecomesLatest: true},
		"no webhooks/command/deployedVersionLookup does announce and update deployed_version": {wantAnnounces: 1, deployedBecomesLatest: true},
		"commands that have no fails does announce and update deployed_version": {wantAnnounces: 1, deployedBecomesLatest: true,
			commands: command.Slice{{"true"}, {"false"}}, fails: service_status.Fails{Command: []*bool{boolPtr(false), boolPtr(false)}}},
		"commands that haven't run fails doesn't announce or update deployed_version": {wantAnnounces: 0, deployedBecomesLatest: false,
			commands: command.Slice{{"true"}, {"false"}}, fails: service_status.Fails{Command: []*bool{boolPtr(false), nil}}},
		"commands that have failed doesn't announce or update deployed_version": {wantAnnounces: 0, deployedBecomesLatest: false,
			commands: command.Slice{{"true"}, {"false"}}, fails: service_status.Fails{Command: []*bool{boolPtr(false), boolPtr(true)}}},
		"webhooks that have no fails does announce and update deployed_version": {wantAnnounces: 1, deployedBecomesLatest: true,
			webhooks: webhook.Slice{"0": {}, "1": {}}, fails: service_status.Fails{WebHook: map[string]*bool{"0": boolPtr(false), "1": boolPtr(false)}}},
		"webhooks that haven't run fails doesn't announce or update deployed_version": {wantAnnounces: 0, deployedBecomesLatest: false,
			webhooks: webhook.Slice{"0": {}, "1": {}}, fails: service_status.Fails{WebHook: map[string]*bool{"0": boolPtr(false), "1": nil}}},
		"webhooks that have failed doesn't announce or update deployed_version": {wantAnnounces: 0, deployedBecomesLatest: false,
			webhooks: webhook.Slice{"0": {}, "1": {}}, fails: service_status.Fails{WebHook: map[string]*bool{"0": boolPtr(false), "1": boolPtr(true)}}},
		"commands and webhooks that have no fails does announce and update deployed_version": {wantAnnounces: 1, deployedBecomesLatest: true,
			commands: command.Slice{{"true"}, {"false"}}, webhooks: webhook.Slice{"0": {}, "1": {}},
			fails: service_status.Fails{Command: []*bool{boolPtr(false), boolPtr(false)}, WebHook: map[string]*bool{"0": boolPtr(false), "1": boolPtr(false)}}},
		"commands and webhooks that have no fails with deployedVersionLookup does announce and only update approved_version": {wantAnnounces: 1, deployedBecomesLatest: false, approvedBecomesLatest: true,
			commands: command.Slice{{"true"}, {"false"}}, webhooks: webhook.Slice{"0": {}, "1": {}}, deployedVersion: &deployed_version.Lookup{},
			fails: service_status.Fails{Command: []*bool{boolPtr(false), boolPtr(false)}, WebHook: map[string]*bool{"0": boolPtr(false), "1": boolPtr(false)}}},
		"deployedVersionLookup with no commands/webhooks doesn't announce or update deployed_version/approved_version": {wantAnnounces: 0, deployedBecomesLatest: false, deployedVersion: &deployed_version.Lookup{},
			fails: service_status.Fails{Command: []*bool{boolPtr(false), boolPtr(false)}, WebHook: map[string]*bool{"0": boolPtr(false), "1": boolPtr(false)}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			service.Command = tc.commands
			service.WebHook = tc.webhooks
			service.DeployedVersionLookup = tc.deployedVersion
			service.Status.Fails = tc.fails
			if tc.latestIsDeployed {
				service.Status.DeployedVersion = service.Status.LatestVersion
			}

			// WHEN UpdatedVersion is called on it
			want := service.Status.LatestVersion
			service.UpdatedVersion()

			// THEN ApprovedVersion becomes LatestVersion if there's a dvl and commands/webhooks
			got := service.Status.ApprovedVersion
			if (tc.approvedBecomesLatest && got != want) || (!tc.approvedBecomesLatest && got == want) {
				t.Errorf("%s:\nApprovedVersion should have changed to %q not %q",
					name, want, got)
			}
			// THEN DeployedVersion becomes LatestVersion if there's no dvl
			got = service.Status.DeployedVersion
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("%s:\nDeployedVersion should have changed to %q not %q",
					name, want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*service.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("%s:\nExpecting %d announce message but got %d",
					name, tc.wantAnnounces, len(*service.Status.AnnounceChannel))
			}
		})
	}
}

func TestHandleUpdateActions(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		commands              command.Slice
		webhooks              webhook.Slice
		autoApprove           bool
		deployedBecomesLatest bool
		wantAnnounces         int
	}{
		"no auto_approve and no webhooks/command does announce and update deployed_version": {autoApprove: false, wantAnnounces: 1, deployedBecomesLatest: true},
		"no auto_approve but do have webhooks only announces the new version": {autoApprove: false, wantAnnounces: 1, deployedBecomesLatest: false,
			webhooks: webhook.Slice{"fail": testWebHookFailing()}},
		"auto_approve and webhook that fails only announces the fail and doesn't update deployed_version": {autoApprove: true, wantAnnounces: 1, deployedBecomesLatest: false,
			webhooks: webhook.Slice{"fail": testWebHookFailing()}},
		"auto_approve and webhook that passes announces the pass and version change and updates deployed_version": {autoApprove: true, wantAnnounces: 2, deployedBecomesLatest: true,
			webhooks: webhook.Slice{"pass": testWebHookSuccessful()}},
		"auto_approve and command that fails only announces the fail and doesn't update deployed_version": {autoApprove: true, wantAnnounces: 2, deployedBecomesLatest: false,
			commands: command.Slice{{"true"}, {"false"}}},
		"auto_approve and command that passes announces the pass and version change and updates deployed_version": {autoApprove: true, wantAnnounces: 3, deployedBecomesLatest: true,
			commands: command.Slice{{"true"}, {"ls"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			service.Status.Init(len(service.Notify), len(tc.commands), len(tc.webhooks), &service.ID, &service.Dashboard.WebURL)
			service.Command = tc.commands
			if len(tc.commands) != 0 {
				service.CommandController = &command.Controller{}
			}
			service.CommandController.Init(jLog, &service.ID, &service.Status, &service.Command, nil, &service.Options.Interval)
			service.WebHook = tc.webhooks
			service.WebHook.Init(jLog, &service.ID, &service.Status, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{}, nil, &service.Options.Interval)
			service.Dashboard.AutoApprove = &tc.autoApprove
			service.DeployedVersionLookup = nil

			// WHEN HandleUpdateActions is called on it
			want := service.Status.LatestVersion
			service.HandleUpdateActions()
			// wait until all commands/webhooks have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if service.Command != nil {
					for j := range service.Command {
						if service.Status.Fails.Command[j] == nil {
							actionsRan = false
							break
						}
					}
				}
				if service.WebHook != nil {
					for j := range service.WebHook {
						if service.Status.Fails.WebHook[j] == nil {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("%s:\nfinished running after %v",
						name, time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !tc.autoApprove {
				if actionsRan && (len(service.Status.Fails.Command) != 0 || len(service.Status.Fails.WebHook) != 0) {
					t.Fatalf("%s:\nno actions should have run as auto_approve is %t\n%#v",
						name, tc.autoApprove, service.Status.Fails)
				}
			} else if !actionsRan {
				t.Fatalf("%s:\nactions didn't finish running",
					name)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := service.Status.DeployedVersion
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("%s:\nDeployedVersion should have changed to %q not %q",
					name, want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*service.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("%s:\nExpecting %d announce message but got %d",
					name, tc.wantAnnounces, len(*service.Status.AnnounceChannel))
				fails := ""
				if len(service.Status.Fails.Command) != 0 {
					for i := range service.Status.Fails.Command {
						fails += fmt.Sprintf("%d=%t, ", i, *service.Status.Fails.Command[i])
					}
					t.Logf("commandFails: {%s}", fails[:len(fails)-2])
				}
				fails = ""
				if len(service.Status.Fails.WebHook) != 0 {
					for i := range service.Status.Fails.WebHook {
						fails += fmt.Sprintf("%s=%t, ", i, *service.Status.Fails.WebHook[i])
					}
					t.Logf("webhookFails: {%s}", fails[:len(fails)-2])
				}
				for len(*service.Status.AnnounceChannel) != 0 {
					msg := <-*service.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
		})
	}
}

func TestHandleFailedActions(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		commands              command.Slice
		commandNextRunnables  []time.Time
		webhooks              webhook.Slice
		webhookNextRunnables  map[string]time.Time
		fails                 service_status.Fails
		wantFails             service_status.Fails
		deployedBecomesLatest bool
		deployedLatest        bool
		wantAnnounces         int
	}{
		"no webhook fails and no command fails retries all": {wantAnnounces: 2,
			commands: command.Slice{{"false"}}, webhooks: webhook.Slice{"pass": testWebHookFailing()},
			fails:     service_status.Fails{Command: []*bool{boolPtr(false)}, WebHook: map[string]*bool{"pass": boolPtr(false)}},
			wantFails: service_status.Fails{Command: []*bool{boolPtr(true)}, WebHook: map[string]*bool{"pass": boolPtr(true)}}},
		"no webhook fails and have command fails retries only the failed commands": {wantAnnounces: 3, deployedLatest: false,
			commands: command.Slice{{"true"}, {"false"}, {"true"}, {"false"}}, webhooks: webhook.Slice{"pass": testWebHookSuccessful()},
			fails:     service_status.Fails{Command: []*bool{boolPtr(true), boolPtr(false), boolPtr(true), boolPtr(true)}, WebHook: map[string]*bool{"pass": boolPtr(false)}},
			wantFails: service_status.Fails{Command: []*bool{boolPtr(false), boolPtr(false), boolPtr(false), boolPtr(true)}, WebHook: map[string]*bool{"pass": boolPtr(false)}}},
		"command fails before their next_runnable don't run": {wantAnnounces: 1, deployedLatest: false,
			commands: command.Slice{{"true"}, {"false"}, {"true"}, {"false"}}, webhooks: webhook.Slice{"pass": testWebHookSuccessful()},
			fails:                service_status.Fails{Command: []*bool{boolPtr(true), boolPtr(false), boolPtr(true), boolPtr(true)}, WebHook: map[string]*bool{"pass": boolPtr(false)}},
			wantFails:            service_status.Fails{Command: []*bool{boolPtr(false), boolPtr(false), boolPtr(true), boolPtr(true)}, WebHook: map[string]*bool{"pass": boolPtr(false)}},
			commandNextRunnables: []time.Time{time.Now().UTC(), time.Now().UTC(), time.Now().UTC().Add(time.Minute), time.Now().UTC().Add(time.Minute)}},
		"no webhook fails and have command fails retries only the failed commands and updates deployed_version": {wantAnnounces: 2, deployedBecomesLatest: true,
			commands: command.Slice{{"true"}, {"false"}}, webhooks: webhook.Slice{"pass": testWebHookSuccessful()},
			fails:     service_status.Fails{Command: []*bool{boolPtr(true), boolPtr(false)}, WebHook: map[string]*bool{"pass": boolPtr(false)}},
			wantFails: service_status.Fails{Command: []*bool{nil, nil}, WebHook: map[string]*bool{"pass": nil}}},
		"have webhook fails and no command fails retries only the failed commands": {wantAnnounces: 2, deployedLatest: false,
			commands: command.Slice{{"false"}}, webhooks: webhook.Slice{"will_fail": testWebHookFailing(), "will_pass": testWebHookSuccessful(), "would_fail": testWebHookFailing()},
			fails:     service_status.Fails{Command: []*bool{boolPtr(false)}, WebHook: map[string]*bool{"will_fail": boolPtr(true), "will_pass": boolPtr(true), "would_fail": boolPtr(false)}},
			wantFails: service_status.Fails{Command: []*bool{boolPtr(false)}, WebHook: map[string]*bool{"will_fail": boolPtr(true), "will_pass": boolPtr(false), "would_fail": boolPtr(false)}}},
		"webhook fails before their next_runnable don't run": {wantAnnounces: 1, deployedLatest: false,
			commands: command.Slice{{"false"}}, webhooks: webhook.Slice{"is_runnable": testWebHookSuccessful(), "not_runnable": testWebHookFailing(), "would_fail": testWebHookFailing()},
			fails:                service_status.Fails{Command: []*bool{boolPtr(false)}, WebHook: map[string]*bool{"is_runnable": boolPtr(true), "not_runnable": boolPtr(true), "would_fail": boolPtr(false)}},
			wantFails:            service_status.Fails{Command: []*bool{boolPtr(false)}, WebHook: map[string]*bool{"is_runnable": boolPtr(false), "not_runnable": boolPtr(true), "would_fail": boolPtr(false)}},
			webhookNextRunnables: map[string]time.Time{"is_runnable": time.Now().UTC(), "not_runnable": time.Now().UTC().Add(time.Minute)}},
		"have webhook fails and no command fails retries only the failed commands and updates deployed_version": {wantAnnounces: 3, deployedBecomesLatest: true,
			commands: command.Slice{{"false"}}, webhooks: webhook.Slice{"will_pass0": testWebHookSuccessful(), "will_pass1": testWebHookSuccessful(), "would_fail": testWebHookFailing()},
			fails:     service_status.Fails{Command: []*bool{boolPtr(false)}, WebHook: map[string]*bool{"will_pass0": boolPtr(true), "will_pass1": boolPtr(true), "would_fail": boolPtr(false)}},
			wantFails: service_status.Fails{Command: []*bool{nil}, WebHook: map[string]*bool{"will_pass0": nil, "will_pass1": nil, "would_fail": nil}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			service.Status.Init(len(service.Notify), len(tc.commands), len(tc.webhooks), &service.ID, &service.Dashboard.WebURL)
			service.Status.Fails = tc.fails
			if tc.deployedLatest {
				service.Status.DeployedVersion = service.Status.LatestVersion
			}
			service.Command = tc.commands
			if len(tc.commands) != 0 {
				service.CommandController = &command.Controller{}
			}
			service.CommandController.Init(jLog, &service.ID, &service.Status, &service.Command, nil, &service.Options.Interval)
			service.WebHook = tc.webhooks
			service.WebHook.Init(jLog, &service.ID, &service.Status, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{}, nil, &service.Options.Interval)
			service.DeployedVersionLookup = nil
			for i := range tc.commandNextRunnables {
				service.CommandController.NextRunnable[i] = tc.commandNextRunnables[i]
			}
			for i := range tc.webhookNextRunnables {
				service.WebHook[i].NextRunnable = tc.webhookNextRunnables[i]
			}

			// WHEN HandleFailedActions is called on it
			want := service.Status.LatestVersion
			service.HandleFailedActions()
			// wait until all commands/webhooks have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if service.Command != nil {
					for j := range service.Command {
						if stringifyPointer(service.Status.Fails.Command[j]) != stringifyPointer(tc.wantFails.Command[j]) {
							actionsRan = false
							break
						}
					}
				}
				if service.WebHook != nil {
					for j := range service.WebHook {
						if stringifyPointer(service.Status.Fails.WebHook[j]) != stringifyPointer(tc.wantFails.WebHook[j]) {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("%s:\nfinished running after %v",
						name, time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !actionsRan {
				t.Errorf("%s:\nactions didn't finish running or gave unexpected results",
					name)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := service.Status.DeployedVersion
			if (tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*service.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*service.Status.AnnounceChannel))
				fails := ""
				if len(service.Status.Fails.Command) != 0 {
					for i := range service.Status.Fails.Command {
						fails += fmt.Sprintf("%d=%t, ", i, *service.Status.Fails.Command[i])
					}
					t.Logf("commandFails: {%s}", fails[:len(fails)-2])
				}
				fails = ""
				if len(service.Status.Fails.WebHook) != 0 {
					for i := range service.Status.Fails.WebHook {
						fails += fmt.Sprintf("%s=%t, ", i, *service.Status.Fails.WebHook[i])
					}
					t.Logf("webhookFails: {%s}", fails[:len(fails)-2])
				}
				for len(*service.Status.AnnounceChannel) != 0 {
					msg := <-*service.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the Command fails are as expected
			for i := range tc.wantFails.Command {
				if stringifyPointer(service.Status.Fails.Command[i]) != stringifyPointer(tc.wantFails.Command[i]) {
					t.Errorf("got, command[%d]=%s, want %s",
						i, stringifyPointer(service.Status.Fails.Command[i]), stringifyPointer(tc.wantFails.Command[i]))
				}
			}
			// THEN the WebHook fails are as expected
			for i := range tc.wantFails.WebHook {
				if stringifyPointer(service.Status.Fails.WebHook[i]) != stringifyPointer(tc.wantFails.WebHook[i]) {
					t.Errorf("got, webhook[%s]=%s, want %s",
						i, stringifyPointer(service.Status.Fails.WebHook[i]), stringifyPointer(tc.wantFails.WebHook[i]))
				}
			}
		})
	}
}

func TestHandleCommand(t *testing.T) {
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
		"empty Command slice does nothing": {commands: command.Slice{}, wantAnnounces: 0, deployedLatest: true, deployedBecomesLatest: false},
		"Command that failed passes": {commands: command.Slice{{"ls", "-lah"}}, command: "ls -lah", wantAnnounces: 1, deployedLatest: true, deployedBecomesLatest: false,
			fails: []*bool{boolPtr(true)}, wantFails: []*bool{boolPtr(false)}},
		"Command that passed fails": {commands: command.Slice{{"false"}}, command: "false", wantAnnounces: 1, deployedLatest: true, deployedBecomesLatest: false,
			fails: []*bool{boolPtr(false)}, wantFails: []*bool{boolPtr(true)}},
		"Command that's not runnable doesn't run": {commands: command.Slice{{"false"}}, command: "false", wantAnnounces: 0, deployedLatest: true, deployedBecomesLatest: false,
			fails: []*bool{boolPtr(false)}, wantFails: []*bool{boolPtr(false)}, nextRunnable: time.Now().UTC().Add(time.Minute)},
		"Command that's runnable does run": {commands: command.Slice{{"false"}}, command: "false", wantAnnounces: 1, deployedLatest: true, deployedBecomesLatest: false,
			fails: []*bool{boolPtr(false)}, wantFails: []*bool{boolPtr(true)}, nextRunnable: time.Now().UTC().Add(-time.Second)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			service.Status.Init(len(service.Notify), len(tc.commands), 0, &service.ID, &service.Dashboard.WebURL)
			service.Status.Fails.Command = tc.fails
			if tc.deployedLatest {
				service.Status.DeployedVersion = service.Status.LatestVersion
			}
			service.Command = tc.commands
			if len(tc.commands) != 0 {
				service.CommandController = &command.Controller{}
			}
			service.CommandController.Init(jLog, &service.ID, &service.Status, &service.Command, nil, &service.Options.Interval)
			service.DeployedVersionLookup = nil
			for i := range service.Command {
				service.CommandController.NextRunnable[i] = tc.nextRunnable
			}

			// WHEN HandleCommand is called on it
			want := service.Status.LatestVersion
			service.HandleCommand(tc.command)
			// wait until all commands have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if service.Command != nil {
					for j := range service.Command {
						if stringifyPointer(service.Status.Fails.Command[j]) != stringifyPointer(tc.wantFails[j]) {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("%s:\nfinished running after %v",
						name, time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !actionsRan {
				t.Errorf("%s:\nactions didn't finish running or gave unexpected results",
					name)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := service.Status.DeployedVersion
			if !tc.deployedLatest && ((tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want)) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*service.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*service.Status.AnnounceChannel))
				fails := ""
				if len(service.Status.Fails.Command) != 0 {
					for i := range service.Status.Fails.Command {
						fails += fmt.Sprintf("%d=%t, ", i, *service.Status.Fails.Command[i])
					}
					t.Logf("commandFails: {%s}", fails[:len(fails)-2])
				}
				for len(*service.Status.AnnounceChannel) != 0 {
					msg := <-*service.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the Command fails are as expected
			for i := range tc.wantFails {
				if stringifyPointer(service.Status.Fails.Command[i]) != stringifyPointer(tc.wantFails[i]) {
					t.Errorf("got, command[%d]=%s, want %s",
						i, stringifyPointer(service.Status.Fails.Command[i]), stringifyPointer(tc.wantFails[i]))
				}
			}
		})
	}
}

func TestHandleWebHook(t *testing.T) {
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
		"empty WebHook slice does nothing": {webhooks: webhook.Slice{}, wantAnnounces: 0, deployedLatest: true, deployedBecomesLatest: false},
		"WebHook that failed passes": {webhooks: webhook.Slice{"pass": testWebHookSuccessful()}, webhook: "pass", wantAnnounces: 1, deployedLatest: true, deployedBecomesLatest: false,
			fails: map[string]*bool{"pass": boolPtr(true)}, wantFails: map[string]*bool{"pass": boolPtr(false)}},
		"WebHook that passed fails": {webhooks: webhook.Slice{"fail": testWebHookFailing()}, webhook: "fail", wantAnnounces: 1, deployedLatest: true, deployedBecomesLatest: false,
			fails: map[string]*bool{"fail": boolPtr(false)}, wantFails: map[string]*bool{"fail": boolPtr(true)}},
		"WebHook that's not runnable doesn't run": {webhooks: webhook.Slice{"pass": testWebHookFailing()}, webhook: "pass", wantAnnounces: 0, deployedLatest: true, deployedBecomesLatest: false,
			fails: map[string]*bool{"pass": boolPtr(false)}, wantFails: map[string]*bool{"pass": boolPtr(false)}, nextRunnable: time.Now().UTC().Add(time.Minute)},
		"WebHook that's runnable does run": {webhooks: webhook.Slice{"pass": testWebHookSuccessful()}, webhook: "pass", wantAnnounces: 1, deployedLatest: true, deployedBecomesLatest: false,
			fails: map[string]*bool{"pass": boolPtr(true)}, wantFails: map[string]*bool{"pass": boolPtr(false)}, nextRunnable: time.Now().UTC().Add(-time.Second)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			service.Status.Init(len(service.Notify), len(tc.webhooks), 0, &service.ID, &service.Dashboard.WebURL)
			service.Status.Fails.WebHook = tc.fails
			if tc.deployedLatest {
				service.Status.DeployedVersion = service.Status.LatestVersion
			}
			service.WebHook = tc.webhooks
			service.WebHook.Init(jLog, &service.ID, &service.Status, &service.WebHook, &webhook.WebHook{}, &webhook.WebHook{}, nil, &service.Options.Interval)
			service.DeployedVersionLookup = nil
			for i := range service.WebHook {
				service.WebHook[i].NextRunnable = tc.nextRunnable
			}

			// WHEN HandleWebHook is called on it
			want := service.Status.LatestVersion
			service.HandleWebHook(tc.webhook)
			// wait until all webhooks have run
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if service.WebHook != nil {
					for j := range service.WebHook {
						if stringifyPointer(service.Status.Fails.WebHook[j]) != stringifyPointer(tc.wantFails[j]) {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf("%s:\nfinished running after %v",
						name, time.Duration(i*10)*time.Microsecond)
					break
				}
			}
			if !actionsRan {
				t.Errorf("%s:\nactions didn't finish running or gave unexpected results",
					name)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := service.Status.DeployedVersion
			if !tc.deployedLatest && ((tc.deployedBecomesLatest && got != want) || (!tc.deployedBecomesLatest && got == want)) {
				t.Errorf("DeployedVersion should have changed to %q not %q",
					want, got)
			}
			// THEN the correct number of changes are announced to the channel
			if len(*service.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("Expecting %d announce message but got %d",
					tc.wantAnnounces, len(*service.Status.AnnounceChannel))
				fails := ""
				if len(service.Status.Fails.WebHook) != 0 {
					for i := range service.Status.Fails.WebHook {
						fails += fmt.Sprintf("%s=%t, ", i, *service.Status.Fails.WebHook[i])
					}
					t.Logf("webhookFails: {%s}", fails[:len(fails)-2])
				}
				for len(*service.Status.AnnounceChannel) != 0 {
					msg := <-*service.Status.AnnounceChannel
					t.Logf("%#v",
						string(msg))
				}
			}
			// THEN the WebHook fails are as expected
			for i := range tc.wantFails {
				if stringifyPointer(service.Status.Fails.WebHook[i]) != stringifyPointer(tc.wantFails[i]) {
					t.Errorf("got, webhook[%s]=%s, want %s",
						i, stringifyPointer(service.Status.Fails.WebHook[i]), stringifyPointer(tc.wantFails[i]))
				}
			}
		})
	}
}

func TestHandleSkip(t *testing.T) {
	// GIVEN a Service
	testLogging()
	latestVersion := "1.2.3"
	tests := map[string]struct {
		skipVersion          string
		approvedVersion      string
		wantAnnounces        int
		wantDatabaseMessages int
	}{
		"skip of not latest version does nothing": {skipVersion: latestVersion + "-beta"},
		"skip of latest version skips version and announces to announce and database channels": {
			skipVersion: latestVersion, approvedVersion: "SKIP_" + latestVersion, wantAnnounces: 1, wantDatabaseMessages: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			service.Status.ApprovedVersion = ""
			service.Status.LatestVersion = latestVersion

			// WHEN HandleSkip is called on it
			service.HandleSkip(tc.skipVersion)

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			got := service.Status.ApprovedVersion
			if tc.approvedVersion != got {
				t.Errorf("%s:\nApprovedVersion should have changed to %q not %q",
					name, tc.approvedVersion, got)
			}
			// THEN the correct number of changes are announced to the announce channel
			if len(*service.Status.AnnounceChannel) != tc.wantAnnounces {
				t.Errorf("%s:\nExpecting %d announce message but got %d",
					name, tc.wantAnnounces, len(*service.Status.AnnounceChannel))
			}
			// THEN the correct number of messages are announced to the database channel
			if len(*service.Status.DatabaseChannel) != tc.wantDatabaseMessages {
				t.Errorf("%s:\nExpecting %d announce message but got %d",
					name, tc.wantDatabaseMessages, len(*service.Status.DatabaseChannel))
			}
		})
	}
}

func TestShouldRetryAll(t *testing.T) {
	// GIVEN a Service
	testLogging()
	tests := map[string]struct {
		command []*bool
		webhook map[string]*bool
		want    bool
	}{
		"no commands or webhooks":                    {want: true},
		"commands that haven't run":                  {command: []*bool{nil, nil}, want: false},
		"commands that have failed":                  {command: []*bool{boolPtr(true), boolPtr(true)}, want: false},
		"commands that have failed/haven't run":      {command: []*bool{boolPtr(true), nil}, want: false},
		"commands that haven't failed":               {command: []*bool{boolPtr(false), boolPtr(false)}, want: true},
		"mix of all command fail states":             {command: []*bool{boolPtr(true), boolPtr(false), nil}, want: false},
		"webhooks that haven't run":                  {webhook: map[string]*bool{"1": nil, "2": nil}, want: false},
		"webhooks that have failed":                  {webhook: map[string]*bool{"1": boolPtr(true), "2": boolPtr(true)}, want: false},
		"webhooks that have failed/haven't run":      {webhook: map[string]*bool{"1": boolPtr(true), "2": nil}, want: false},
		"webhooks that haven't failed":               {webhook: map[string]*bool{"1": boolPtr(false), "2": boolPtr(false)}, want: true},
		"mix of all webhook fail states":             {webhook: map[string]*bool{"1": boolPtr(true), "2": boolPtr(false), "3": nil}, want: false},
		"mix of all webhook and command fail states": {command: []*bool{boolPtr(true), boolPtr(false), nil}, webhook: map[string]*bool{"1": boolPtr(true), "2": boolPtr(false), "3": nil}, want: false},
		"mix of all webhook and command no fails":    {command: []*bool{boolPtr(false), boolPtr(false), boolPtr(false)}, webhook: map[string]*bool{"1": boolPtr(false), "2": boolPtr(false), "3": boolPtr(false)}, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service := testServiceURL()
			commands := len(tc.command)
			service.Command = command.Slice{}
			for commands != 0 {
				service.Command = append(service.Command, command.Command{})
				commands--
			}
			webhooks := len(tc.webhook)
			service.WebHook = webhook.Slice{}
			for webhooks != 0 {
				service.WebHook[fmt.Sprint(webhooks)] = &webhook.WebHook{}
				webhooks--
			}
			service.Status.Fails.Command = tc.command
			service.Status.Fails.WebHook = tc.webhook

			// WHEN shouldRetryAll is called on it
			got := service.shouldRetryAll()

			// THEN DeployedVersion becomes LatestVersion as there's no dvl
			if tc.want != got {
				t.Errorf("%s:\nwant %t not %t",
					name, tc.want, got)
			}
		})
	}
}
