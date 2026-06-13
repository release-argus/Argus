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

//go:build unit

package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvmanual "github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestService_HandleSkip(t *testing.T) {
	// GIVEN: a Service.
	latestVersion := "1.2.3"
	tests := []struct {
		name                                string
		startVersion                        string
		approvedVersion                     string
		wantAnnounces, wantDatabaseMessages int
		prepDelete                          bool
	}{
		{
			name:                 "skip of latest version does nothing",
			startVersion:         latestVersion,
			approvedVersion:      "",
			wantAnnounces:        0,
			wantDatabaseMessages: 0,
		},
		{
			name:                 "skip of version that's not latest skips version and announces to announce and database channels",
			startVersion:         "1.0.0",
			approvedVersion:      serviceinfo.SkippedVersion(latestVersion),
			wantAnnounces:        1,
			wantDatabaseMessages: 1,
		},
		{
			name:                 "skip of version that's not latest, but Service deletion has started",
			startVersion:         "0.2.3",
			approvedVersion:      "",
			prepDelete:           true,
			wantAnnounces:        0,
			wantDatabaseMessages: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using a shared channel.
			svc := testService(t, tc.name, "url", "url")

			svc.Status.SetDeployedVersion(tc.startVersion, "", false)
			svc.Status.SetLatestVersion(latestVersion, "", false)
			svc.Status.SetApprovedVersion("", false)
			if tc.prepDelete {
				svc.PrepDelete(true)
			}

			// WHEN: HandleSkip is called on it.
			svc.HandleSkip()

			prefix := fmt.Sprintf("%s\nHandleSkip()", packageName)

			// THEN: DeployedVersion becomes LatestVersion as there is no dvl.
			if got := svc.Status.ApprovedVersion(); got != tc.approvedVersion {
				t.Errorf(
					"%s ApprovedVersion() should have changed\ngot:  %q\nwant: %q",
					prefix, got, tc.approvedVersion,
				)
			}

			// AND: the correct amount of changes are queued in the announce channel.
			if tc.prepDelete {
				if svc.Status.AnnounceChannel != nil || svc.Status.DatabaseChannel != nil {
					t.Errorf(
						"%s AnnounceChannel and DatabaseChannel nil-state mismatch\ngot:  non-nil (%+v)\nwant: nil",
						prefix, &svc.Status,
					)
				}
				return
			}

			// AND: the correct amount of changes are queued in the announce channel.
			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantAnnounces,
				)
			}

			// AND: the correct amount of messages are queued in the database channel.
			if got := len(svc.Status.DatabaseChannel); got != tc.wantDatabaseMessages {
				t.Errorf(
					"%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					packageName, got, tc.wantDatabaseMessages,
				)
			}
		})
	}
}

func TestService_HandleCommand(t *testing.T) {
	// GIVEN: a Service.
	tests := []struct {
		name                                  string
		command                               string
		commands                              command.Commands
		nextRunnable                          time.Time
		fails, wantFails                      []*bool
		deployedBecomesLatest, deployedLatest bool
		wantAnnounces                         int
	}{
		{
			name:                  "empty Command slice does nothing",
			commands:              command.Commands{},
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
		},
		{
			name: "Command that failed passes",
			commands: command.Commands{
				{"ls", "-lah"},
			},
			command:               "ls -lah",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				test.Ptr(true),
			},
			wantFails: []*bool{
				test.Ptr(false),
			},
		},
		{
			name: "Command that passed fails",
			commands: command.Commands{
				{"false"},
			},
			command:               "false",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				test.Ptr(false),
			},
			wantFails: []*bool{
				test.Ptr(true),
			},
		},
		{
			name: "Command that's not runnable doesn't run",
			commands: command.Commands{
				{"false"},
			},
			command:               "false",
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				test.Ptr(false),
			},
			wantFails: []*bool{
				test.Ptr(false),
			},
			nextRunnable: time.Now().UTC().Add(time.Minute),
		},
		{
			name: "Command that's runnable does run",
			commands: command.Commands{
				{"false"},
			},
			command:               "false",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: []*bool{
				test.Ptr(false),
			},
			wantFails: []*bool{
				test.Ptr(true),
			},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Service.
			svc := testService(t, tc.name, "url", "url")

			svc.DeployedVersionLookup = nil

			svc.Command = tc.commands
			svc.CommandController = command.NewController(
				&svc.Status,
				svc.Command,
				nil,
				&svc.Options.Interval,
			)

			svc.Status.SetDeployedVersion("1.2.2", "", false)
			svc.Status.SetLatestVersion("1.2.3", "", false)
			svc.Status.SetApprovedVersion("1.2.1", false)
			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), "", false)
			}
			svc.Status.Init(
				len(svc.Command), len(svc.Notify), len(svc.WebHook),
				status.ServiceInfo{
					ID: svc.ID,
				},
				&svc.Dashboard,
			)

			for k, v := range tc.fails {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			for i := range svc.Command {
				svc.CommandController.SetNextRunnable(i, tc.nextRunnable)
			}

			prefix := fmt.Sprintf(
				"%s\nHandleCommand(%q)",
				packageName, tc.commands,
			)

			// WHEN: HandleCommand is called on it.
			want := svc.Status.LatestVersion()
			svc.HandleCommand(tc.command)
			// wait until all commands have run.
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.Command != nil {
					for j := range svc.Command {
						gotFail := test.StringifyPtr(svc.Status.Fails.Command.Get(j))
						wantFail := test.StringifyPtr(tc.wantFails[j])
						if gotFail != wantFail {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf(
						"%s\nfinished running after %v",
						prefix, time.Duration(i*10)*time.Microsecond,
					)
					break
				}
			}
			if !actionsRan {
				t.Errorf("%s\nactions didn't finish running or gave unexpected results", prefix)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN: DeployedVersion becomes LatestVersion as there is no dvl.
			got := svc.Status.DeployedVersion()
			if !tc.deployedLatest &&
				((tc.deployedBecomesLatest && got != want) ||
					(!tc.deployedBecomesLatest && got == want)) {
				t.Errorf(
					"%s DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the correct amount of changes are queued in the channel.
			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					packageName, got, tc.wantAnnounces,
				)
				fails := ""
				for i := range svc.Command {
					fails += fmt.Sprintf(
						"%d=%s, ",
						i, test.StringifyPtr(svc.Status.Fails.Command.Get(i)),
					)
				}
				t.Logf("%s commandFails: {%s}", prefix, fails[:len(fails)-2])
				for len(svc.Status.AnnounceChannel) != 0 {
					msg := <-svc.Status.AnnounceChannel
					t.Logf(
						"%s - service.Service.HandleCommand() %#v",
						prefix, string(msg),
					)
				}
			}

			// AND: the Command fails are as want.
			for i := range tc.wantFails {
				gotFail := test.StringifyPtr(svc.Status.Fails.Command.Get(i))
				wantFail := test.StringifyPtr(tc.wantFails[i])
				if gotFail != wantFail {
					t.Errorf(
						"%s Command[%d] fail state mismatch\ngot:  %s\nwant: %s",
						prefix, i,
						gotFail, wantFail,
					)
				}
			}
		})
	}
}

func TestService_HandleWebHook(t *testing.T) {
	whCfg := whtest.PlainConfig()
	// GIVEN: a Service.
	tests := []struct {
		name                                  string
		webhook                               string
		webhooks                              webhook.WebHooks
		nextRunnable                          time.Time
		fails, wantFails                      map[string]*bool
		deployedBecomesLatest, deployedLatest bool
		wantAnnounces                         int
	}{
		{
			name:                  "empty WebHook map does nothing",
			webhooks:              webhook.WebHooks{},
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
		},
		{
			name: "WebHook that failed passes",
			webhooks: webhook.WebHooks{
				"pass": whtest.WebHook(false, false, false),
			},
			webhook:               "pass",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": test.Ptr(true),
			},
			wantFails: map[string]*bool{
				"pass": test.Ptr(false),
			},
		},
		{
			name: "WebHook that passed fails",
			webhooks: webhook.WebHooks{
				"fail": whtest.WebHook(true, false, false),
			},
			webhook:               "fail",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"fail": test.Ptr(false),
			},
			wantFails: map[string]*bool{
				"fail": test.Ptr(true),
			},
		},
		{
			name: "WebHook that's not runnable doesn't run",
			webhooks: webhook.WebHooks{
				"pass": whtest.WebHook(true, false, false),
			},
			webhook:               "pass",
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": test.Ptr(false),
			},
			wantFails: map[string]*bool{
				"pass": test.Ptr(false),
			},
			nextRunnable: time.Now().UTC().Add(time.Minute),
		},
		{
			name: "WebHook that's runnable does run",
			webhooks: webhook.WebHooks{
				"pass": whtest.WebHook(false, false, false),
			},
			webhook:               "pass",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": test.Ptr(true),
			},
			wantFails: map[string]*bool{
				"pass": test.Ptr(false),
			},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Service.
			svc := testService(t, tc.name, "url", "url")

			svc.DeployedVersionLookup = nil

			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				whCfg,
				nil,
				&svc.Options.Interval,
			)

			svc.Status.SetDeployedVersion("1.2.2", "", false)
			svc.Status.SetLatestVersion("1.2.3", "", false)
			svc.Status.SetApprovedVersion("1.2.1", false)
			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), "", false)
			}
			svc.Status.Init(
				len(svc.Command), len(svc.Notify), len(svc.WebHook),
				status.ServiceInfo{
					ID: svc.ID,
				},
				&svc.Dashboard,
			)

			for k, v := range tc.fails {
				svc.Status.Fails.WebHook.Set(k, v)
			}
			for i := range svc.WebHook {
				svc.WebHook[i].SetNextRunnable(tc.nextRunnable)
			}

			prefix := fmt.Sprintf(
				"%s\nHandleWebHook(%q)",
				packageName, tc.webhook,
			)

			// WHEN: HandleWebHook is called on it.
			want := svc.Status.LatestVersion()
			svc.HandleWebHook(tc.webhook)
			// wait until all webhooks have run.
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						gotFail := test.StringifyPtr(svc.Status.Fails.WebHook.Get(j))
						wantFail := test.StringifyPtr(tc.wantFails[j])
						if gotFail != wantFail {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf(
						"%s finished running after %v",
						prefix, time.Duration(i*10)*time.Microsecond,
					)
					break
				}
			}
			if !actionsRan {
				t.Errorf("%s actions didn't finish running or gave unexpected results", prefix)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN: DeployedVersion becomes LatestVersion as there is no dvl.
			if !tc.deployedLatest {
				got := svc.Status.DeployedVersion()
				if tc.deployedBecomesLatest && got != want {
					t.Errorf(
						"%s DeployedVersion() did not update to latest\ngot:  %q\nwant: %q",
						prefix, got, want,
					)
				}
				if !tc.deployedBecomesLatest && got == want {
					t.Errorf(
						"%s DeployedVersion() updated unexpectedly\ngot:  %q\nwant: %q",
						prefix, got, want,
					)
				}
			}
			// THEN: the correct amount of changes are queued in the channel.
			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, tc.wantAnnounces, got,
				)
				fails := ""
				for i := range svc.WebHook {
					fails += fmt.Sprintf(
						"%s=%s, ",
						i, test.StringifyPtr(svc.Status.Fails.WebHook.Get(i)),
					)
				}
				t.Logf(
					"%s webhookFails: {%s}",
					prefix, fails[:len(fails)-2],
				)
				for len(svc.Status.AnnounceChannel) != 0 {
					msg := <-svc.Status.AnnounceChannel
					t.Logf(
						"%s - %#v",
						prefix, string(msg),
					)
				}
			}
			// THEN: the WebHook fails are as want.
			for i := range tc.wantFails {
				gotFail := test.StringifyPtr(svc.Status.Fails.WebHook.Get(i))
				wantFail := test.StringifyPtr(tc.wantFails[i])
				if gotFail != wantFail {
					t.Errorf(
						"%s WebHook[%q] value mismatch\ngot:  %s\nwant: %s",
						prefix, i,
						gotFail, wantFail,
					)
				}
			}
		})
	}
}

func TestService_HandleUpdateActions(t *testing.T) {
	whCfg := whtest.PlainConfig()
	// GIVEN: a Service.
	tests := []struct {
		name                               string
		commands                           command.Commands
		webhooks                           webhook.WebHooks
		autoApprove, deployedBecomesLatest bool
		wantAnnounces                      int
	}{
		{
			name:                  "no auto_approve and no webhooks/command does announce and update deployed_version",
			autoApprove:           false,
			wantAnnounces:         1,
			deployedBecomesLatest: true,
		},
		{
			name:                  "no auto_approve but do have webhooks only announces the new version",
			autoApprove:           false,
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			webhooks: webhook.WebHooks{
				"fail": whtest.WebHook(true, false, false),
			},
		},
		{
			name:                  "auto_approve and webhook that fails only announces the fail and doesn't update deployed_version",
			autoApprove:           true,
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			webhooks: webhook.WebHooks{
				"fail": whtest.WebHook(true, false, false),
			},
		},
		{
			name:                  "auto_approve and webhook that passes announces the pass and version change and updates deployed_version",
			autoApprove:           true,
			wantAnnounces:         2,
			deployedBecomesLatest: true,
			webhooks: webhook.WebHooks{
				"pass": whtest.WebHook(false, false, false),
			},
		},
		{
			name:                  "auto_approve and command that fails only announces the fail and doesn't update deployed_version",
			autoApprove:           true,
			wantAnnounces:         2,
			deployedBecomesLatest: false,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
		},
		{
			name:                  "auto_approve and command that passes announces the pass and version change and updates deployed_version",
			autoApprove:           true,
			wantAnnounces:         3,
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"ls"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			svc.DeployedVersionLookup = nil

			svc.Command = tc.commands
			svc.CommandController = command.NewController(
				&svc.Status,
				svc.Command,
				nil,
				&svc.Options.Interval,
			)

			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				whCfg,
				nil,
				&svc.Options.Interval,
			)

			svc.Status.SetDeployedVersion("1.2.2", "", false)
			svc.Status.SetLatestVersion("1.2.3", "", false)
			svc.Status.SetApprovedVersion("1.2.1", false)
			svc.Status.Init(
				len(svc.Command), len(svc.Notify), len(svc.WebHook),
				status.ServiceInfo{
					ID: svc.ID,
				},
				&svc.Dashboard,
			)

			svc.Dashboard.AutoApprove = &tc.autoApprove

			prefix := fmt.Sprintf("%s\nHandleUpdateActions(true)", packageName)

			// WHEN: HandleUpdateActions is called on it.
			want := svc.Status.LatestVersion()
			svc.HandleUpdateActions(true)
			// wait until all commands/webhooks have run.
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
					t.Logf(
						"%s finished running after %v",
						prefix, time.Duration(i*10)*time.Microsecond,
					)
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
						t.Fatalf(
							"%s should have run no actions as auto_approve is %t\nfails:\n%s",
							prefix, tc.autoApprove, svc.Status.Fails.String("  "),
						)
					}
				}
			} else if !actionsRan {
				t.Fatalf("%s actions didn't finish running", prefix)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN: DeployedVersion becomes LatestVersion as there is no dvl.
			got := svc.Status.DeployedVersion()
			if (tc.deployedBecomesLatest && got != want) ||
				(!tc.deployedBecomesLatest && got == want) {
				t.Errorf(
					"%s DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}
			// THEN: the correct amount of changes are queued in the channel.
			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantAnnounces,
				)
				t.Logf(
					"%s - Fails:\n%s",
					prefix, svc.Status.Fails.String("  "),
				)
				for len(svc.Status.AnnounceChannel) != 0 {
					msg := <-svc.Status.AnnounceChannel
					t.Logf(
						"%s - AnnounceChannel message: %#v",
						prefix, string(msg),
					)
				}
			}
		})
	}
}

func TestService_HandleFailedActions(t *testing.T) {
	whCfg := whtest.PlainConfig()
	// GIVEN: a Service.
	tests := []struct {
		name                                  string
		commands                              command.Commands
		commandNextRunnable                   []time.Time
		webhooks                              webhook.WebHooks
		webhookNextRunnable                   map[string]time.Time
		startFailsCommand, wantFailsCommand   []*bool
		startFailsWebHook, wantFailsWebHook   map[string]*bool
		deployedBecomesLatest, deployedLatest bool
		wantAnnounces                         int
	}{
		{
			name:          "no command or webhooks fails retries all",
			wantAnnounces: 3, // 3 = 2 command fails, 1 webhook fail.
			commands: command.Commands{
				{"false"},
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"will_fail": whtest.WebHook(true, false, false),
			},
			startFailsCommand: []*bool{
				nil, nil,
			},
			wantFailsCommand: []*bool{
				test.Ptr(true), test.Ptr(true),
			},
			startFailsWebHook: map[string]*bool{
				"will_fail": nil,
			},
			wantFailsWebHook: map[string]*bool{
				"will_fail": test.Ptr(true),
			},
		},
		{
			name:           "have command fails and no webhook fails retries only the failed commands",
			wantAnnounces:  3, // 3 = 2 command passes, 1 command fail.
			deployedLatest: false,
			commands: command.Commands{
				{"true"},
				{"false"},
				{"true"},
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"pass": whtest.WebHook(false, false, false),
			},
			startFailsCommand: []*bool{
				test.Ptr(true),
				test.Ptr(false),
				test.Ptr(true),
				test.Ptr(true),
			},
			wantFailsCommand: []*bool{
				test.Ptr(false),
				test.Ptr(false),
				test.Ptr(false),
				test.Ptr(true),
			},
			startFailsWebHook: map[string]*bool{
				"pass": test.Ptr(false),
			},
			wantFailsWebHook: map[string]*bool{
				"pass": test.Ptr(false),
			},
		},
		{
			name:           "command fails before their next_runnable don't run",
			wantAnnounces:  1, // 0 = no runs.
			deployedLatest: false,
			commands: command.Commands{
				{"true"},
				{"false"},
				{"true"},
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"pass": whtest.WebHook(false, false, false),
			},
			startFailsCommand: []*bool{
				test.Ptr(true),
				test.Ptr(false),
				test.Ptr(true),
				test.Ptr(true),
			},
			wantFailsCommand: []*bool{
				test.Ptr(false),
				test.Ptr(false),
				test.Ptr(true),
				test.Ptr(true),
			},
			startFailsWebHook: map[string]*bool{
				"pass": test.Ptr(false),
			},
			wantFailsWebHook: map[string]*bool{
				"pass": test.Ptr(false),
			},
			commandNextRunnable: []time.Time{
				time.Now().UTC(),
				time.Now().UTC(),
				time.Now().UTC().Add(time.Minute),
				time.Now().UTC().Add(time.Minute),
			},
		},
		{
			name:                  "have command fails no webhook fails and retries only the failed commands and updates deployed_version",
			wantAnnounces:         2, // 2 = 1 command, 1 deployed.
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"pass": whtest.WebHook(false, false, false),
			},
			startFailsCommand: []*bool{test.Ptr(true), test.Ptr(false)},
			wantFailsCommand: []*bool{
				nil, nil,
			},
			startFailsWebHook: map[string]*bool{
				"pass": test.Ptr(false),
			},
			wantFailsWebHook: map[string]*bool{
				"pass": nil,
			},
		},
		{
			name:           "have webhook fails and no command fails retries only the failed commands",
			wantAnnounces:  2, // 2 = 2 webhook runs.
			deployedLatest: false,
			commands: command.Commands{
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"will_fail":  whtest.WebHook(true, false, false),
				"will_pass":  whtest.WebHook(false, false, false),
				"would_fail": whtest.WebHook(true, false, false),
			},
			startFailsCommand: []*bool{
				test.Ptr(false),
			},
			wantFailsCommand: []*bool{
				test.Ptr(false),
			},
			startFailsWebHook: map[string]*bool{
				"will_fail":  test.Ptr(true),
				"will_pass":  test.Ptr(true),
				"would_fail": test.Ptr(false),
			},
			wantFailsWebHook: map[string]*bool{
				"will_fail":  test.Ptr(true),
				"will_pass":  test.Ptr(false),
				"would_fail": test.Ptr(false),
			},
		},
		{
			name:           "webhook fails before their next_runnable don't run",
			wantAnnounces:  1, // 0 runs.
			deployedLatest: false,
			commands: command.Commands{
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"is_runnable":  whtest.WebHook(false, false, false),
				"not_runnable": whtest.WebHook(true, false, false),
				"would_fail":   whtest.WebHook(true, false, false),
			},
			startFailsCommand: []*bool{
				test.Ptr(false),
			},
			wantFailsCommand: []*bool{
				test.Ptr(false),
			},
			startFailsWebHook: map[string]*bool{
				"is_runnable":  test.Ptr(true),
				"not_runnable": test.Ptr(true),
				"would_fail":   test.Ptr(false),
			},
			wantFailsWebHook: map[string]*bool{
				"is_runnable":  test.Ptr(false),
				"not_runnable": test.Ptr(true),
				"would_fail":   test.Ptr(false),
			},
			webhookNextRunnable: map[string]time.Time{
				"is_runnable":  time.Now().UTC(),
				"not_runnable": time.Now().UTC().Add(time.Minute),
			},
		},
		{
			name:                  "have webhook fails and no command fails retries only the failed commands and updates deployed_version",
			wantAnnounces:         3, // 2 webhook runs.
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"will_pass0": whtest.WebHook(false, false, false),
				"will_pass1": whtest.WebHook(false, false, false),
				"would_fail": whtest.WebHook(true, false, false),
			},
			startFailsCommand: []*bool{
				test.Ptr(false),
			},
			wantFailsCommand: []*bool{
				nil,
			},
			startFailsWebHook: map[string]*bool{
				"will_pass0": test.Ptr(true),
				"will_pass1": test.Ptr(true),
				"would_fail": test.Ptr(false),
			},
			wantFailsWebHook: map[string]*bool{
				"will_pass0": nil,
				"will_pass1": nil,
				"would_fail": nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Service.
			svc := testService(t, tc.name, "url", "url")

			svc.DeployedVersionLookup = nil

			svc.Command = tc.commands
			svc.CommandController = command.NewController(
				&svc.Status,
				svc.Command,
				nil,
				&svc.Options.Interval,
			)

			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				whCfg,
				nil,
				&svc.Options.Interval,
			)

			svc.Status.SetDeployedVersion("1.2.2", "", false)
			svc.Status.SetLatestVersion("1.2.3", "", false)
			svc.Status.SetApprovedVersion("1.2.1", false)
			if tc.deployedLatest {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), "", false)
			}
			svc.Status.Init(
				len(svc.Command), len(svc.Notify), len(svc.WebHook),
				status.ServiceInfo{
					ID: svc.ID,
				},
				&svc.Dashboard,
			)

			for k, v := range tc.startFailsCommand {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			for k, v := range tc.startFailsWebHook {
				svc.Status.Fails.WebHook.Set(k, v)
			}

			for i := range tc.commandNextRunnable {
				nextRunnable := tc.commandNextRunnable[i]
				svc.CommandController.SetNextRunnable(i, nextRunnable)
			}
			for i := range tc.webhookNextRunnable {
				nextRunnable := tc.webhookNextRunnable[i]
				svc.WebHook[i].SetNextRunnable(nextRunnable)
			}

			// WHEN: HandleFailedActions is called on it.
			want := svc.Status.LatestVersion()
			svc.HandleFailedActions()

			prefix := fmt.Sprintf("%s\nHandleFailedActions()", packageName)

			// wait until all commands/webhooks have run.
			var actionsRan bool
			for i := 1; i < 500; i++ {
				actionsRan = true
				time.Sleep(10 * time.Millisecond)
				if svc.Command != nil {
					for j := range svc.Command {
						gotFail := test.StringifyPtr(svc.Status.Fails.Command.Get(j))
						wantFail := test.StringifyPtr(tc.wantFailsCommand[j])
						if gotFail != wantFail {
							actionsRan = false
							break
						}
					}
				}
				if svc.WebHook != nil {
					for j := range svc.WebHook {
						gotFail := test.StringifyPtr(svc.Status.Fails.WebHook.Get(j))
						wantFail := test.StringifyPtr(tc.wantFailsWebHook[j])
						if gotFail != wantFail {
							actionsRan = false
							break
						}
					}
				}
				if actionsRan {
					t.Logf(
						"%s finished running after %v",
						prefix, time.Duration(i*10)*time.Microsecond,
					)
					break
				}
			}
			if !actionsRan {
				t.Errorf("%s actions didn't finish running or gave unexpected results", prefix)
			}
			time.Sleep(500 * time.Millisecond)

			// THEN: DeployedVersion becomes LatestVersion as there is no dvl.
			gotDV := svc.Status.DeployedVersion()
			if tc.deployedBecomesLatest && gotDV != want {
				t.Errorf(
					"%s DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					packageName, gotDV, want,
				)
			} else if !tc.deployedBecomesLatest && gotDV == want {
				t.Errorf(
					"%s DeployedVersion() mismatch\ngot:  %q\nwant: NOT %q",
					packageName, gotDV, want,
				)
			}

			// AND: the correct amount of changes are queued in the channel.
			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					packageName, got, tc.wantAnnounces,
				)
				t.Logf(
					"%s - Fails:\n%s",
					prefix, svc.Status.Fails.String("  "),
				)
				for len(svc.Status.AnnounceChannel) != 0 {
					msg := <-svc.Status.AnnounceChannel
					t.Logf(
						"%s - %#v",
						prefix, string(msg),
					)
				}
			}

			// AND: the Command fails are as want.
			for i := range tc.wantFailsCommand {
				wantFail := test.StringifyPtr(tc.wantFailsCommand[i])
				gotFail := test.StringifyPtr(svc.Status.Fails.Command.Get(i))
				if gotFail != wantFail {
					t.Errorf(
						"%s Command[%d] mismatch\ngot:  %s\nwant: %s",
						prefix, i,
						gotFail, wantFail,
					)
				}
			}

			// AND: the WebHook fails are as want.
			for i := range tc.wantFailsWebHook {
				wantFail := test.StringifyPtr(tc.wantFailsWebHook[i])
				gotFail := test.StringifyPtr(svc.Status.Fails.WebHook.Get(i))
				if gotFail != wantFail {
					t.Errorf(
						"%s WebHook[%q] mismatch\ngot:  %s\nwant: %s",
						prefix, i,
						gotFail, wantFail,
					)
				}
			}
		})
	}
}

func TestService_ShouldRetryAll(t *testing.T) {
	// GIVEN: a Service.
	tests := []struct {
		name    string
		command []*bool
		webhook map[string]*bool
		want    bool
	}{
		{
			name: "no commands or webhooks",
			want: true,
		},
		{
			name: "commands that haven't run",
			command: []*bool{
				nil,
				nil,
			},
			want: false,
		},
		{
			name: "commands that have failed",
			command: []*bool{
				test.Ptr(true),
				test.Ptr(true),
			},
			want: false,
		},
		{
			name: "commands that have failed/haven't run",
			command: []*bool{
				test.Ptr(true),
				nil,
			},
			want: false,
		},
		{
			name: "commands that haven't failed",
			command: []*bool{
				test.Ptr(false),
				test.Ptr(false),
			},
			want: true,
		},
		{
			name: "mix of all command fail states",
			command: []*bool{
				test.Ptr(true),
				test.Ptr(false),
				nil,
			},
			want: false,
		},
		{
			name: "webhooks that haven't run",
			webhook: map[string]*bool{
				"1": nil,
				"2": nil,
			},
			want: false,
		},
		{
			name: "webhooks that have failed",
			webhook: map[string]*bool{
				"1": test.Ptr(true),
				"2": test.Ptr(true),
			},
			want: false,
		},
		{
			name: "webhooks that have failed/haven't run",
			webhook: map[string]*bool{
				"1": test.Ptr(true),
				"2": nil,
			},
			want: false,
		},
		{
			name: "webhooks that haven't failed",
			webhook: map[string]*bool{
				"1": test.Ptr(false),
				"2": test.Ptr(false),
			},
			want: true,
		},
		{
			name: "mix of all webhook fail states",
			webhook: map[string]*bool{
				"1": test.Ptr(true),
				"2": test.Ptr(false),
				"3": nil,
			},
			want: false,
		},
		{
			name: "mix of all webhook and command fail states",
			command: []*bool{
				test.Ptr(true),
				test.Ptr(false),
				nil,
			},
			webhook: map[string]*bool{
				"1": test.Ptr(true),
				"2": test.Ptr(false),
				"3": nil,
			},
			want: false,
		},
		{
			name: "mix of all webhook and command no fails",
			command: []*bool{
				test.Ptr(false),
				test.Ptr(false),
				test.Ptr(false),
			},
			webhook: map[string]*bool{
				"1": test.Ptr(false),
				"2": test.Ptr(false),
				"3": test.Ptr(false),
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			commands := len(tc.command)
			svc.Command = command.Commands{}
			for commands != 0 {
				svc.Command = append(svc.Command, command.Command{})
				commands--
			}

			webhooks := len(tc.webhook)
			svc.WebHook = webhook.WebHooks{}
			for webhooks != 0 {
				svc.WebHook[fmt.Sprint(webhooks)] = &webhook.WebHook{}
				webhooks--
			}

			svc.Status.Init(
				len(svc.Command), len(svc.Notify), len(svc.WebHook),
				status.ServiceInfo{
					ID: svc.ID,
				},
				&dashboard.Options{},
			)

			for k, v := range tc.command {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			for k, v := range tc.webhook {
				svc.Status.Fails.WebHook.Set(k, v)
			}

			// WHEN: shouldRetryAll is called on it.
			got := svc.shouldRetryAll()

			prefix := fmt.Sprintf("%s\nService.shouldRetryAll()", packageName)

			// THEN: DeployedVersion becomes LatestVersion as there is no dvl.
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestService_UpdateLatestApproved(t *testing.T) {
	// GIVEN: a Service.
	tests := []struct {
		name                 string
		latestVersion        string
		startApprovedVersion string
		wantAnnounces        int
	}{
		{
			name:                 "empty ApprovedVersion does announce",
			startApprovedVersion: "",
			latestVersion:        "1.2.3",
			wantAnnounces:        1,
		},
		{
			name:                 "same ApprovedVersion doesn't announce",
			startApprovedVersion: "1.2.3",
			latestVersion:        "1.2.3",
			wantAnnounces:        0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			svc.Status.SetLatestVersion(tc.latestVersion, "", false)
			svc.Status.SetApprovedVersion(tc.startApprovedVersion, false)

			// WHEN: UpdateLatestApproved is called on it.
			want := svc.Status.LatestVersion()
			svc.UpdateLatestApproved()

			prefix := fmt.Sprintf("%s\nService.UpdateLatestApproved()", packageName)

			// THEN: ApprovedVersion becomes LatestVersion.
			got := svc.Status.ApprovedVersion()
			if got != want {
				t.Errorf(
					"%s\nApprovedVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the correct amount of changes are queued in the channel.
			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantAnnounces,
				)
			}
		})
	}
}

func TestService_UpdatedVersion(t *testing.T) {
	// GIVEN: a Service.
	tests := []struct {
		name                  string
		commands              command.Commands
		commandFails          []*bool
		webhooks              webhook.WebHooks
		webhookFails          map[string]*bool
		latestIsDeployed      bool
		deployedVersion       deployedver.Lookup
		approvedBecomesLatest bool
		deployedBecomesLatest bool
		wantAnnounces         int
	}{
		{
			name:                  "doesn't do anything if DeployedVersion == LatestVersion",
			latestIsDeployed:      true,
			wantAnnounces:         0,
			deployedBecomesLatest: true,
		},
		{
			name:                  "no webhooks/command/deployedVersionLookup does announce and update deployed_version",
			wantAnnounces:         1,
			deployedBecomesLatest: true,
		},
		{
			name:                  "commands that have no fails does announce and update deployed_version",
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			commandFails: []*bool{
				test.Ptr(false),
				test.Ptr(false),
			},
		},
		{
			name:                  "commands that haven't run fails doesn't announce or update deployed_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			commandFails: []*bool{
				test.Ptr(false),
				nil,
			},
		},
		{
			name:                  "commands that have failed doesn't announce or update deployed_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			commandFails: []*bool{
				test.Ptr(false),
				test.Ptr(true),
			},
		},
		{
			name:                  "webhooks that have no fails does announce and update deployed_version",
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			webhooks: webhook.WebHooks{
				"0": {},
				"1": {},
			},
			webhookFails: map[string]*bool{
				"0": test.Ptr(false),
				"1": test.Ptr(false),
			},
		},
		{
			name:                  "webhooks that haven't run fails doesn't announce or update deployed_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.WebHooks{
				"0": {},
				"1": {},
			},
			webhookFails: map[string]*bool{
				"0": test.Ptr(false),
				"1": nil,
			},
		},
		{
			name:                  "webhooks that have failed doesn't announce or update deployed_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.WebHooks{
				"0": {},
				"1": {},
			},
			webhookFails: map[string]*bool{
				"0": test.Ptr(false),
				"1": test.Ptr(true),
			},
		},
		{
			name:                  "commands and webhooks that have no fails does announce and update deployed_version",
			wantAnnounces:         1,
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"0": {},
				"1": {},
			},
			commandFails: []*bool{
				test.Ptr(false),
				test.Ptr(false),
			},
			webhookFails: map[string]*bool{
				"0": test.Ptr(false),
				"1": test.Ptr(false),
			},
		},
		{
			name:                  "commands and webhooks that have no fails with deployedVersionLookup does announce and only update approved_version",
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			approvedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			webhooks: webhook.WebHooks{
				"0": {},
				"1": {},
			},
			deployedVersion: &dvmanual.Lookup{},
			commandFails: []*bool{
				test.Ptr(false),
				test.Ptr(false),
			},
			webhookFails: map[string]*bool{
				"0": test.Ptr(false),
				"1": test.Ptr(false),
			},
		},
		{
			name:                  "deployedVersionLookup with no commands/webhooks doesn't announce or update deployed_version/approved_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			deployedVersion:       &dvmanual.Lookup{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Service with the supplied parameters.
			svc := testService(t, tc.name, "url", "url")

			svc.DeployedVersionLookup = tc.deployedVersion
			svc.Command = tc.commands
			svc.WebHook = tc.webhooks

			svc.Status.SetDeployedVersion("1.2.2", "", false)
			svc.Status.SetLatestVersion("1.2.3", "", false)
			svc.Status.SetApprovedVersion("1.2.1", false)
			if tc.latestIsDeployed {
				svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), "", false)
			}
			svc.Status.Init(
				len(svc.Command), len(svc.Notify), len(svc.WebHook),
				status.ServiceInfo{
					ID: svc.ID,
				},
				&svc.Dashboard,
			)

			for i := range tc.commandFails {
				if tc.commandFails[i] != nil {
					svc.Status.Fails.Command.Set(i, *tc.commandFails[i])
				}
			}
			for i := range tc.webhookFails {
				svc.Status.Fails.WebHook.Set(i, tc.webhookFails[i])
			}

			startLV := svc.Status.LatestVersion()

			// WHEN: UpdatedVersion is called on it.
			svc.UpdatedVersion(true)

			prefix := fmt.Sprintf("%s\nService.UpdatedVersion(true)", packageName)

			// THEN: ApprovedVersion becomes LatestVersion if there is a dvl and commands/webhooks.
			gotAV := svc.Status.ApprovedVersion()
			if (tc.approvedBecomesLatest && gotAV != startLV) ||
				(!tc.approvedBecomesLatest && gotAV == startLV) {
				t.Errorf(
					"%s ApprovedVersion mismatch\ngot:  %q\nwant: %q",
					prefix, gotAV, startLV,
				)
			}

			// AND: DeployedVersion becomes LatestVersion if there is no dvl.
			gotDV := svc.Status.DeployedVersion()
			if (tc.deployedBecomesLatest && gotDV != startLV) ||
				(!tc.deployedBecomesLatest && gotDV == startLV) {
				t.Errorf(
					"%s DeployedVersion()\ngot:  %q\nwant: %q",
					prefix, gotDV, startLV,
				)
			}

			// AND: the correct amount of changes are queued in the channel.
			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"%s\nAnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantAnnounces,
				)
			}
		})
	}
}
