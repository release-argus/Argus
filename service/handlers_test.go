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
	"strconv"
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
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			svc.Status.SetDeployedVersion(tc.startVersion, "", false)
			svc.Status.SetLatestVersion(latestVersion, "", false)
			svc.Status.SetApprovedVersion("", false)
			if tc.prepDelete {
				svc.PrepDelete(true)
			}

			svc.HandleSkip()

			if got := svc.Status.ApprovedVersion(); got != tc.approvedVersion {
				t.Errorf(
					"Status.ApprovedVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.approvedVersion,
				)
			}

			if tc.prepDelete {
				if svc.Status.AnnounceChannel != nil || svc.Status.DatabaseChannel != nil {
					t.Errorf(
						"Status channels mismatch\ngot:  announce=%p, database=%p\nwant: both nil",
						svc.Status.AnnounceChannel, svc.Status.DatabaseChannel,
					)
				}
				return
			}

			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantAnnounces,
				)
			}

			if got, want := len(svc.Status.DatabaseChannel), tc.wantDatabaseMessages; got != want {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
		})
	}
}

func TestService_HandleCommand(t *testing.T) {
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
				new(true),
			},
			wantFails: []*bool{
				new(false),
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
				new(false),
			},
			wantFails: []*bool{
				new(true),
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
				new(false),
			},
			wantFails: []*bool{
				new(false),
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
				new(false),
			},
			wantFails: []*bool{
				new(true),
			},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
		{
			name: "Command that passes updates deployed_version",
			commands: command.Commands{
				{"ls", "-lah"},
			},
			command:               "ls -lah",
			wantAnnounces:         2, // 2 = 1 command, 1 deployed version.
			deployedBecomesLatest: true,
			fails: []*bool{
				new(true),
			},
			wantFails: []*bool{
				nil,
			},
		},
		{
			name: "Command that fails doesn't update deployed_version",
			commands: command.Commands{
				{"false"},
			},
			command:               "false",
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			fails: []*bool{
				new(false),
			},
			wantFails: []*bool{
				new(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			// With no DeployedVersionLookup, UpdatedVersion sets DeployedVersion to LatestVersion.
			svc.DeployedVersionLookup = nil
			setCommands(svc, tc.commands)
			initActionStatus(svc, tc.deployedLatest)

			for k, v := range tc.fails {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			for i := range svc.Command {
				svc.CommandController.SetNextRunnable(i, tc.nextRunnable)
			}

			want := svc.Status.LatestVersion()
			svc.HandleCommand(tc.command)
			waitForFails(svc, tc.wantFails, nil)
			time.Sleep(500 * time.Millisecond)

			assertBecameLatest(
				t,
				"Status.DeployedVersion()", svc.Status.DeployedVersion(), want,
				tc.deployedLatest || tc.deployedBecomesLatest,
			)

			if !test.AssertChannelLen(
				t, "Status.AnnounceChannel", svc.Status.AnnounceChannel, tc.wantAnnounces,
			) {
				t.Logf("Fails:\n%s", svc.Status.Fails.String("  "))
			}

			for i := range tc.wantFails {
				gotFail := test.StringifyPtr(svc.Status.Fails.Command.Get(i))
				wantFail := test.StringifyPtr(tc.wantFails[i])
				if gotFail != wantFail {
					t.Errorf(
						"Status.Fails.Command.Get(%d) mismatch\ngot:  %q\nwant: %q",
						i, gotFail, wantFail,
					)
				}
			}
		})
	}
}

func TestService_HandleWebhook(t *testing.T) {
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name                                  string
		webhook                               string
		webhooks                              webhook.Webhooks
		nextRunnable                          time.Time
		fails, wantFails                      map[string]*bool
		deployedBecomesLatest, deployedLatest bool
		wantAnnounces                         int
	}{
		{
			name:                  "empty Webhook map does nothing",
			webhooks:              webhook.Webhooks{},
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
		},
		{
			name: "Webhook that failed passes",
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, false, false, false),
			},
			webhook:               "pass",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": new(true),
			},
			wantFails: map[string]*bool{
				"pass": new(false),
			},
		},
		{
			name: "Webhook that passed fails",
			webhooks: webhook.Webhooks{
				"fail": whtest.Webhook(t, true, false, false),
			},
			webhook:               "fail",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"fail": new(false),
			},
			wantFails: map[string]*bool{
				"fail": new(true),
			},
		},
		{
			name: "Webhook that's not runnable doesn't run",
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, true, false, false),
			},
			webhook:               "pass",
			wantAnnounces:         0,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": new(false),
			},
			wantFails: map[string]*bool{
				"pass": new(false),
			},
			nextRunnable: time.Now().UTC().Add(time.Minute),
		},
		{
			name: "Webhook that's runnable does run",
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, false, false, false),
			},
			webhook:               "pass",
			wantAnnounces:         1,
			deployedLatest:        true,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"pass": new(true),
			},
			wantFails: map[string]*bool{
				"pass": new(false),
			},
			nextRunnable: time.Now().UTC().Add(-time.Second),
		},
		{
			name: "Webhook that passes updates deployed_version",
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, false, false, false),
			},
			webhook:               "pass",
			wantAnnounces:         2, // 2 = 1 webhook, 1 deployed version.
			deployedBecomesLatest: true,
			fails: map[string]*bool{
				"pass": new(true),
			},
			wantFails: map[string]*bool{
				"pass": nil,
			},
		},
		{
			name: "Webhook that fails doesn't update deployed_version",
			webhooks: webhook.Webhooks{
				"fail": whtest.Webhook(t, true, false, false),
			},
			webhook:               "fail",
			wantAnnounces:         1,
			deployedBecomesLatest: false,
			fails: map[string]*bool{
				"fail": new(false),
			},
			wantFails: map[string]*bool{
				"fail": new(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			// With no DeployedVersionLookup, UpdatedVersion sets DeployedVersion to LatestVersion.
			svc.DeployedVersionLookup = nil
			setWebhooks(t, svc, tc.webhooks, whCfg)
			initActionStatus(svc, tc.deployedLatest)

			for k, v := range tc.fails {
				svc.Status.Fails.Webhook.Set(k, v)
			}
			for i := range svc.Webhook {
				svc.Webhook[i].SetNextRunnable(tc.nextRunnable)
			}

			want := svc.Status.LatestVersion()
			svc.HandleWebhook(tc.webhook)
			waitForFails(svc, nil, tc.wantFails)
			time.Sleep(500 * time.Millisecond)

			assertBecameLatest(
				t,
				"Status.DeployedVersion()", svc.Status.DeployedVersion(), want,
				tc.deployedLatest || tc.deployedBecomesLatest,
			)

			if !test.AssertChannelLen(
				t, "Status.AnnounceChannel", svc.Status.AnnounceChannel, tc.wantAnnounces,
			) {
				t.Logf("Fails:\n%s", svc.Status.Fails.String("  "))
			}

			for i := range tc.wantFails {
				gotFail := test.StringifyPtr(svc.Status.Fails.Webhook.Get(i))
				wantFail := test.StringifyPtr(tc.wantFails[i])
				if gotFail != wantFail {
					t.Errorf(
						"Status.Fails.Webhook.Get(%q) mismatch\ngot:  %q\nwant: %q",
						i, gotFail, wantFail,
					)
				}
			}
		})
	}
}

func TestService_HandleUpdateActions(t *testing.T) {
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name                               string
		commands                           command.Commands
		webhooks                           webhook.Webhooks
		autoApprove, deployedBecomesLatest bool
		wantAnnounces                      int
	}{
		{
			name:                  "no auto_approve, commands or webhooks - announces and updates deployed_version",
			autoApprove:           false,
			wantAnnounces:         1, // 1 = 1 deployed version.
			deployedBecomesLatest: true,
		},
		{
			name:                  "no auto_approve but do have webhooks only announces the new version",
			autoApprove:           false,
			wantAnnounces:         1, // 1 = 1 approval request.
			deployedBecomesLatest: false,
			webhooks: webhook.Webhooks{
				"fail": whtest.Webhook(t, true, false, false),
			},
		},
		{
			name:                  "auto_approve and webhook that fails only announces the fail and doesn't update deployed_version",
			autoApprove:           true,
			wantAnnounces:         1, // 1 = 1 webhook run.
			deployedBecomesLatest: false,
			webhooks: webhook.Webhooks{
				"fail": whtest.Webhook(t, true, false, false),
			},
		},
		{
			name:                  "auto_approve and webhook that passes announces the pass and version change and updates deployed_version",
			autoApprove:           true,
			wantAnnounces:         2, // 2 = 1 webhook run, 1 deployed version.
			deployedBecomesLatest: true,
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, false, false, false),
			},
		},
		{
			name:                  "auto_approve and command that fails only announces the fail and doesn't update deployed_version",
			autoApprove:           true,
			wantAnnounces:         2, // 2 = 2 command runs.
			deployedBecomesLatest: false,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
		},
		{
			name:                  "auto_approve and command that passes announces the pass and version change and updates deployed_version",
			autoApprove:           true,
			wantAnnounces:         3, // 3 = 2 command runs, 1 deployed version.
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

			// With no DeployedVersionLookup, UpdatedVersion sets DeployedVersion to LatestVersion.
			svc.DeployedVersionLookup = nil
			setCommands(svc, tc.commands)
			setWebhooks(t, svc, tc.webhooks, whCfg)
			initActionStatus(svc, false)

			svc.Dashboard.AutoApprove = &tc.autoApprove

			want := svc.Status.LatestVersion()
			svc.HandleUpdateActions(true)
			// With auto_approve, wait until all commands/webhooks have run.
			if tc.autoApprove {
				var actionsRan bool
				for range 499 {
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
					if svc.Webhook != nil {
						for j := range svc.Webhook {
							webhookFailed := svc.Status.Fails.Webhook.Get(j)
							if (tc.deployedBecomesLatest && webhookFailed != nil) ||
								(!tc.deployedBecomesLatest && webhookFailed == nil) {
								actionsRan = false
								break
							}
						}
					}
					deployedSettled := !tc.deployedBecomesLatest || svc.Status.DeployedVersion() == want
					if actionsRan && deployedSettled {
						break
					}
				}
				if !actionsRan {
					wantFails := "non-nil"
					if tc.deployedBecomesLatest {
						wantFails = "nil"
					}
					t.Fatalf(
						"Status.Fails mismatch\ngot:\n%s\nwant: all %s",
						svc.Status.Fails.String("  "), wantFails,
					)
				}
			}
			time.Sleep(500 * time.Millisecond)

			if !tc.autoApprove {
				ranAction := false
				for i := range svc.Command {
					if svc.Status.Fails.Command.Get(i) != nil {
						ranAction = true
						break
					}
				}
				for i := range svc.Webhook {
					if svc.Status.Fails.Webhook.Get(i) != nil {
						ranAction = true
						break
					}
				}
				if ranAction {
					t.Fatalf(
						"Status.Fails mismatch\ngot:\n%s\nwant: all nil",
						svc.Status.Fails.String("  "),
					)
				}
			}

			assertBecameLatest(
				t,
				"Status.DeployedVersion()", svc.Status.DeployedVersion(), want,
				tc.deployedBecomesLatest,
			)

			if !test.AssertChannelLen(
				t, "Status.AnnounceChannel", svc.Status.AnnounceChannel, tc.wantAnnounces,
			) {
				t.Logf("Fails:\n%s", svc.Status.Fails.String("  "))
			}
		})
	}
}

func TestService_HandleFailedActions(t *testing.T) {
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name                                string
		commands                            command.Commands
		commandNextRunnable                 []time.Time
		webhooks                            webhook.Webhooks
		webhookNextRunnable                 map[string]time.Time
		startFailsCommand, wantFailsCommand []*bool
		startFailsWebhook, wantFailsWebhook map[string]*bool
		deployedBecomesLatest               bool
		wantAnnounces                       int
	}{
		{
			name:                  "no commands or webhooks updates deployed_version",
			wantAnnounces:         1, // 1 = 1 deployed version.
			deployedBecomesLatest: true,
		},
		{
			name:          "no command or webhooks fails retries all",
			wantAnnounces: 3, // 3 = 2 command fails, 1 webhook fail.
			commands: command.Commands{
				{"false"},
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"will_fail": whtest.Webhook(t, true, false, false),
			},
			startFailsCommand: []*bool{
				nil, nil,
			},
			wantFailsCommand: []*bool{
				new(true), new(true),
			},
			startFailsWebhook: map[string]*bool{
				"will_fail": nil,
			},
			wantFailsWebhook: map[string]*bool{
				"will_fail": new(true),
			},
		},
		{
			name:          "have command fails and no webhook fails retries only the failed commands",
			wantAnnounces: 3, // 3 = 2 command passes, 1 command fail.
			commands: command.Commands{
				{"true"},
				{"false"},
				{"true"},
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, false, false, false),
			},
			startFailsCommand: []*bool{
				new(true),
				new(false),
				new(true),
				new(true),
			},
			wantFailsCommand: []*bool{
				new(false),
				new(false),
				new(false),
				new(true),
			},
			startFailsWebhook: map[string]*bool{
				"pass": new(false),
			},
			wantFailsWebhook: map[string]*bool{
				"pass": new(false),
			},
		},
		{
			name:          "command fails before their next_runnable don't run",
			wantAnnounces: 1, // 1 = 1 command pass.
			commands: command.Commands{
				{"true"},
				{"false"},
				{"true"},
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, false, false, false),
			},
			startFailsCommand: []*bool{
				new(true),
				new(false),
				new(true),
				new(true),
			},
			wantFailsCommand: []*bool{
				new(false),
				new(false),
				new(true),
				new(true),
			},
			startFailsWebhook: map[string]*bool{
				"pass": new(false),
			},
			wantFailsWebhook: map[string]*bool{
				"pass": new(false),
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
			webhooks: webhook.Webhooks{
				"pass": whtest.Webhook(t, false, false, false),
			},
			startFailsCommand: []*bool{new(true), new(false)},
			wantFailsCommand: []*bool{
				nil, nil,
			},
			startFailsWebhook: map[string]*bool{
				"pass": new(false),
			},
			wantFailsWebhook: map[string]*bool{
				"pass": nil,
			},
		},
		{
			name:          "have webhook fails and no command fails retries only the failed webhooks",
			wantAnnounces: 2, // 2 = 2 webhook runs.
			commands: command.Commands{
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"will_fail":  whtest.Webhook(t, true, false, false),
				"will_pass":  whtest.Webhook(t, false, false, false),
				"would_fail": whtest.Webhook(t, true, false, false),
			},
			startFailsCommand: []*bool{
				new(false),
			},
			wantFailsCommand: []*bool{
				new(false),
			},
			startFailsWebhook: map[string]*bool{
				"will_fail":  new(true),
				"will_pass":  new(true),
				"would_fail": new(false),
			},
			wantFailsWebhook: map[string]*bool{
				"will_fail":  new(true),
				"will_pass":  new(false),
				"would_fail": new(false),
			},
		},
		{
			name:          "webhook fails before their next_runnable don't run",
			wantAnnounces: 1, // 1 = 1 webhook pass.
			commands: command.Commands{
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"is_runnable":  whtest.Webhook(t, false, false, false),
				"not_runnable": whtest.Webhook(t, true, false, false),
				"would_fail":   whtest.Webhook(t, true, false, false),
			},
			startFailsCommand: []*bool{
				new(false),
			},
			wantFailsCommand: []*bool{
				new(false),
			},
			startFailsWebhook: map[string]*bool{
				"is_runnable":  new(true),
				"not_runnable": new(true),
				"would_fail":   new(false),
			},
			wantFailsWebhook: map[string]*bool{
				"is_runnable":  new(false),
				"not_runnable": new(true),
				"would_fail":   new(false),
			},
			webhookNextRunnable: map[string]time.Time{
				"is_runnable":  time.Now().UTC(),
				"not_runnable": time.Now().UTC().Add(time.Minute),
			},
		},
		{
			name:                  "have webhook fails and no command fails retries only the failed webhooks and updates deployed_version",
			wantAnnounces:         3, // 3 = 2 webhook runs, 1 deployed.
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"will_pass0": whtest.Webhook(t, false, false, false),
				"will_pass1": whtest.Webhook(t, false, false, false),
				"would_fail": whtest.Webhook(t, true, false, false),
			},
			startFailsCommand: []*bool{
				new(false),
			},
			wantFailsCommand: []*bool{
				nil,
			},
			startFailsWebhook: map[string]*bool{
				"will_pass0": new(true),
				"will_pass1": new(true),
				"would_fail": new(false),
			},
			wantFailsWebhook: map[string]*bool{
				"will_pass0": nil,
				"will_pass1": nil,
				"would_fail": nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			// With no DeployedVersionLookup, UpdatedVersion sets DeployedVersion to LatestVersion.
			svc.DeployedVersionLookup = nil
			setCommands(svc, tc.commands)
			setWebhooks(t, svc, tc.webhooks, whCfg)
			initActionStatus(svc, false)

			for k, v := range tc.startFailsCommand {
				if v != nil {
					svc.Status.Fails.Command.Set(k, *v)
				}
			}
			for k, v := range tc.startFailsWebhook {
				svc.Status.Fails.Webhook.Set(k, v)
			}

			for i := range tc.commandNextRunnable {
				nextRunnable := tc.commandNextRunnable[i]
				svc.CommandController.SetNextRunnable(i, nextRunnable)
			}
			for i := range tc.webhookNextRunnable {
				nextRunnable := tc.webhookNextRunnable[i]
				svc.Webhook[i].SetNextRunnable(nextRunnable)
			}

			want := svc.Status.LatestVersion()
			done := make(chan struct{})
			go func() {
				svc.HandleFailedActions()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("Service.HandleFailedActions() did not return")
			}

			waitForFails(svc, tc.wantFailsCommand, tc.wantFailsWebhook)
			time.Sleep(500 * time.Millisecond)

			assertBecameLatest(
				t,
				"Status.DeployedVersion()", svc.Status.DeployedVersion(), want,
				tc.deployedBecomesLatest,
			)

			if !test.AssertChannelLen(
				t, "Status.AnnounceChannel", svc.Status.AnnounceChannel, tc.wantAnnounces,
			) {
				t.Logf("Fails:\n%s", svc.Status.Fails.String("  "))
			}

			for i := range tc.wantFailsCommand {
				wantFail := test.StringifyPtr(tc.wantFailsCommand[i])
				gotFail := test.StringifyPtr(svc.Status.Fails.Command.Get(i))
				if gotFail != wantFail {
					t.Errorf(
						"Status.Fails.Command.Get(%d) mismatch\ngot:  %q\nwant: %q",
						i, gotFail, wantFail,
					)
				}
			}

			for i := range tc.wantFailsWebhook {
				wantFail := test.StringifyPtr(tc.wantFailsWebhook[i])
				gotFail := test.StringifyPtr(svc.Status.Fails.Webhook.Get(i))
				if gotFail != wantFail {
					t.Errorf(
						"Status.Fails.Webhook.Get(%q) mismatch\ngot:  %q\nwant: %q",
						i, gotFail, wantFail,
					)
				}
			}
		})
	}
}

func TestService_ShouldRetryAll(t *testing.T) {
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
				new(true),
				new(true),
			},
			want: false,
		},
		{
			name: "commands that have failed, haven't run",
			command: []*bool{
				new(true),
				nil,
			},
			want: false,
		},
		{
			name: "commands that haven't failed",
			command: []*bool{
				new(false),
				new(false),
			},
			want: true,
		},
		{
			name: "mix of all command fail states",
			command: []*bool{
				new(true),
				new(false),
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
				"1": new(true),
				"2": new(true),
			},
			want: false,
		},
		{
			name: "webhook that has failed and webhook that hasn't run",
			webhook: map[string]*bool{
				"1": new(true),
				"2": nil,
			},
			want: false,
		},
		{
			name: "webhooks that haven't failed",
			webhook: map[string]*bool{
				"1": new(false),
				"2": new(false),
			},
			want: true,
		},
		{
			name: "mix of all webhook fail states",
			webhook: map[string]*bool{
				"1": new(true),
				"2": new(false),
				"3": nil,
			},
			want: false,
		},
		{
			name: "mix of all webhook and command fail states",
			command: []*bool{
				new(true),
				new(false),
				nil,
			},
			webhook: map[string]*bool{
				"1": new(true),
				"2": new(false),
				"3": nil,
			},
			want: false,
		},
		{
			name: "mix of all webhook and command no fails",
			command: []*bool{
				new(false),
				new(false),
				new(false),
			},
			webhook: map[string]*bool{
				"1": new(false),
				"2": new(false),
				"3": new(false),
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
			svc.Webhook = webhook.Webhooks{}
			for webhooks != 0 {
				svc.Webhook[strconv.Itoa(webhooks)] = &webhook.Webhook{}
				webhooks--
			}

			svc.Status.Init(
				len(svc.Command), len(svc.Notify), len(svc.Webhook),
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
				svc.Status.Fails.Webhook.Set(k, v)
			}

			got := svc.shouldRetryAll()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestService_UpdatedVersion(t *testing.T) {
	tests := []struct {
		name                  string
		commands              command.Commands
		commandFails          []*bool
		webhooks              webhook.Webhooks
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
			name:                  "no commands, webhooks or deployedVersionLookup - announces and updates deployed_version",
			wantAnnounces:         1, // 1 = 1 deployed version.
			deployedBecomesLatest: true,
		},
		{
			name:                  "commands that have no fails does announce and update deployed_version",
			wantAnnounces:         1, // 1 = 1 deployed version.
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			commandFails: []*bool{
				new(false),
				new(false),
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
				new(false),
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
				new(false),
				new(true),
			},
		},
		{
			name:                  "webhooks that have no fails does announce and update deployed_version",
			wantAnnounces:         1, // 1 = 1 deployed version.
			deployedBecomesLatest: true,
			webhooks: webhook.Webhooks{
				"0": {},
				"1": {},
			},
			webhookFails: map[string]*bool{
				"0": new(false),
				"1": new(false),
			},
		},
		{
			name:                  "webhooks that haven't run fails doesn't announce or update deployed_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.Webhooks{
				"0": {},
				"1": {},
			},
			webhookFails: map[string]*bool{
				"0": new(false),
				"1": nil,
			},
		},
		{
			name:                  "webhooks that have failed doesn't announce or update deployed_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			webhooks: webhook.Webhooks{
				"0": {},
				"1": {},
			},
			webhookFails: map[string]*bool{
				"0": new(false),
				"1": new(true),
			},
		},
		{
			name:                  "commands and webhooks that have no fails does announce and update deployed_version",
			wantAnnounces:         1, // 1 = 1 deployed version.
			deployedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"0": {},
				"1": {},
			},
			commandFails: []*bool{
				new(false),
				new(false),
			},
			webhookFails: map[string]*bool{
				"0": new(false),
				"1": new(false),
			},
		},
		{
			name:                  "commands and webhooks that have no fails with deployedVersionLookup does announce and only update approved_version",
			wantAnnounces:         1, // 1 = 1 approved version.
			deployedBecomesLatest: false,
			approvedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
				{"false"},
			},
			webhooks: webhook.Webhooks{
				"0": {},
				"1": {},
			},
			deployedVersion: &dvmanual.Lookup{},
			commandFails: []*bool{
				new(false),
				new(false),
			},
			webhookFails: map[string]*bool{
				"0": new(false),
				"1": new(false),
			},
		},
		{
			name:                  "commands only that have no fails with deployedVersionLookup does announce and only update approved_version",
			wantAnnounces:         1, // 1 = 1 approved version.
			deployedBecomesLatest: false,
			approvedBecomesLatest: true,
			commands: command.Commands{
				{"true"},
			},
			deployedVersion: &dvmanual.Lookup{},
			commandFails: []*bool{
				new(false),
			},
		},
		{
			name:                  "deployedVersionLookup with no commands or webhooks doesn't announce or update deployed_version or approved_version",
			wantAnnounces:         0,
			deployedBecomesLatest: false,
			deployedVersion:       &dvmanual.Lookup{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := testService(t, tc.name, "url", "url")

			svc.DeployedVersionLookup = tc.deployedVersion
			svc.Command = tc.commands
			svc.Webhook = tc.webhooks

			initActionStatus(svc, tc.latestIsDeployed)

			for i := range tc.commandFails {
				if tc.commandFails[i] != nil {
					svc.Status.Fails.Command.Set(i, *tc.commandFails[i])
				}
			}
			for i := range tc.webhookFails {
				svc.Status.Fails.Webhook.Set(i, tc.webhookFails[i])
			}

			startLV := svc.Status.LatestVersion()

			svc.UpdatedVersion(true)

			assertBecameLatest(
				t,
				"Status.ApprovedVersion()", svc.Status.ApprovedVersion(), startLV,
				tc.approvedBecomesLatest,
			)
			assertBecameLatest(
				t,
				"Status.DeployedVersion()", svc.Status.DeployedVersion(), startLV,
				tc.deployedBecomesLatest,
			)

			if got := len(svc.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantAnnounces,
				)
			}
		})
	}
}

func TestService_UpdateLatestApproved(t *testing.T) {
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

			want := svc.Status.LatestVersion()
			svc.UpdateLatestApproved()

			got := svc.Status.ApprovedVersion()
			if got != want {
				t.Errorf("Status.ApprovedVersion() mismatch\ngot:  %q\nwant: %q", got, want)
			}

			if gotAnnounces := len(svc.Status.AnnounceChannel); gotAnnounces != tc.wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotAnnounces, tc.wantAnnounces,
				)
			}
		})
	}
}

// setCommands gives svc the commands, and a CommandController for them.
func setCommands(svc *Service, commands command.Commands) {
	svc.Command = commands
	svc.CommandController = command.NewController(
		&svc.Status,
		svc.Command,
		nil,
		&svc.Options.Interval,
	)
}

// setWebhooks gives svc the webhooks, each sending to a test.GitHubWebhookServer.
func setWebhooks(t *testing.T, svc *Service, webhooks webhook.Webhooks, whCfg webhook.Config) {
	svc.Webhook = webhooks
	webhookServer := test.GitHubWebhookServer(t)
	for _, wh := range svc.Webhook {
		wh.URL = webhookServer.URL
	}
	svc.Webhook.Init(
		&svc.Status,
		whCfg,
		nil,
		&svc.Options.Interval,
	)
}

// initActionStatus sets the versions the action tests start from, then inits the Status.
func initActionStatus(svc *Service, deployedLatest bool) {
	svc.Status.SetDeployedVersion("1.2.2", "", false)
	svc.Status.SetLatestVersion("1.2.3", "", false)
	svc.Status.SetApprovedVersion("1.2.1", false)
	if deployedLatest {
		svc.Status.SetDeployedVersion(svc.Status.LatestVersion(), "", false)
	}
	svc.Status.Init(
		len(svc.Command), len(svc.Notify), len(svc.Webhook),
		status.ServiceInfo{
			ID: svc.ID,
		},
		&svc.Dashboard,
	)
}

// waitForFails polls for up to ~5s until svc's command and webhook fails match the wants.
func waitForFails(svc *Service, wantCommand []*bool, wantWebhook map[string]*bool) {
	for range 499 {
		time.Sleep(10 * time.Millisecond)
		if failsMatch(svc, wantCommand, wantWebhook) {
			return
		}
	}
}

func failsMatch(svc *Service, wantCommand []*bool, wantWebhook map[string]*bool) bool {
	for j := range svc.Command {
		var wantFail *bool
		if j < len(wantCommand) {
			wantFail = wantCommand[j]
		}
		if test.StringifyPtr(svc.Status.Fails.Command.Get(j)) != test.StringifyPtr(wantFail) {
			return false
		}
	}
	for j := range svc.Webhook {
		gotFail := test.StringifyPtr(svc.Status.Fails.Webhook.Get(j))
		if gotFail != test.StringifyPtr(wantWebhook[j]) {
			return false
		}
	}
	return true
}

// assertBecameLatest checks that got equals latest when becomes, and differs from it otherwise.
func assertBecameLatest(t *testing.T, subject, got, latest string, becomes bool) {
	t.Helper()

	if becomes && got != latest {
		t.Errorf("%s mismatch\ngot:  %q\nwant: %q", subject, got, latest)
	} else if !becomes && got == latest {
		t.Errorf("%s mismatch\ngot:  %q\nwant: NOT %q", subject, got, latest)
	}
}
