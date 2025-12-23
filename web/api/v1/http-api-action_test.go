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

package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
	webhook_test "github.com/release-argus/Argus/webhook/test"
)

func TestHTTP_httpServiceGetActions(t *testing.T) {
	type wants struct {
		stdoutRegex, bodyRegex string
		statusCode             int
	}

	// GIVEN an API and a request for the Actions of a Service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	t.Cleanup(func() {
		if api.Config.Settings.Data.DatabaseFile != "" {
			_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})
	tests := map[string]struct {
		serviceID *string
		commands  command.Commands
		webhooks  webhook.WebHooks
		wants     wants
	}{
		"service_id=unknown": {
			serviceID: test.StringPtr("unknown?"),
			wants: wants{
				stdoutRegex: `service "unknown\?" not found`,
				statusCode:  http.StatusNotFound,
			},
		},
		"service_id=nil": {
			serviceID: test.StringPtr(""),
			wants: wants{
				bodyRegex:  `\{"message":"missing required query parameter: service_id"\}`,
				statusCode: http.StatusBadRequest,
			},
		},
		"known service_id, 0 command, 0 webhooks": {
			commands: command.Commands{},
		},
		"known service_id, 1 command, 0 webhooks": {
			commands: command.Commands{
				testCommand(true)},
		},
		"known service_id, 2 command, 0 webhooks": {
			commands: command.Commands{
				testCommand(true), testCommand(false)},
		},
		"known service_id, 0 command, 1 webhooks": {
			webhooks: webhook.WebHooks{
				"fail0": webhook_test.WebHook(true, false, false)},
		},
		"known service_id, 0 command, 2 webhooks": {
			webhooks: webhook.WebHooks{
				"fail0": webhook_test.WebHook(true, false, false),
				"pass0": webhook_test.WebHook(false, false, false)},
		},
		"known service_id, 2 command, 2 webhooks": {
			commands: command.Commands{
				testCommand(true), testCommand(false)},
			webhooks: webhook.WebHooks{
				"fail0": webhook_test.WebHook(true, false, false),
				"pass0": webhook_test.WebHook(false, false, false)},
		},
	}
	cfg := api.Config

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			if tc.wants.statusCode == 0 {
				tc.wants.statusCode = http.StatusOK
			}
			svc := testService(name, true)
			serviceID := util.DereferenceOrValue(tc.serviceID, svc.ID)
			svc.Defaults = &cfg.Defaults.Service
			svc.HardDefaults = &cfg.HardDefaults.Service
			svc.Status.Init(
				len(svc.Notify), len(tc.commands), len(tc.webhooks),
				serviceID, "", "",
				&dashboard.Options{
					WebURL: "https://example.com"})
			svc.Status.SetAnnounceChannel(cfg.HardDefaults.Service.Status.AnnounceChannel)
			svc.Status.SetApprovedVersion("2.0.0", false)
			svc.Status.SetDeployedVersion("2.0.0", "", false)
			svc.Status.SetLatestVersion("3.0.0", "", true)
			svc.Command = tc.commands
			svc.CommandController = &command.Controller{
				Command: &tc.commands,
				Notifiers: command.Notifiers{
					Shoutrrr: &svc.Notify},
				ServiceStatus:  &svc.Status,
				ParentInterval: test.StringPtr("10m")}
			svc.CommandController.Init(
				&svc.Status,
				&svc.Command,
				&svc.Notify,
				test.StringPtr("10m"))
			if tc.commands == nil {
				svc.CommandController = nil
			}
			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
				&svc.Notify,
				&svc.Options.Interval)
			cfg.OrderMutex.Lock()
			cfg.Service[name] = svc
			cfg.Order = append(cfg.Order, name)
			cfg.OrderMutex.Unlock()
			target := "/api/v1/service/actions/"
			params := url.Values{}
			// Set service_id.
			params.Set("service_id", serviceID)

			// WHEN that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			wHTTP := httptest.NewRecorder()
			api.httpServiceGetActions(wHTTP, req)
			res := wHTTP.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			// THEN we get the expected response.
			stdout := releaseStdout()
			// stdout finishes.
			if tc.wants.stdoutRegex != "" {
				tc.wants.stdoutRegex = strings.ReplaceAll(tc.wants.stdoutRegex, "__name__", name)
				if !util.RegexCheck(tc.wants.stdoutRegex, stdout) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.wants.stdoutRegex, stdout)
				}
			}
			message, _ := io.ReadAll(res.Body)
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d\nbody: %q",
					packageName, tc.wants.statusCode, res.StatusCode, message)
			} else if tc.wants.statusCode != http.StatusOK {
				return
			}
			var gotStruct apitype.ActionSummary
			_ = json.Unmarshal(message, &gotStruct)
			if len(gotStruct.Command) != len(tc.commands) {
				t.Fatalf("%s\ncommands mismatch\nwant: %d\ngot:  %d\nbody: %q",
					packageName, len(tc.commands), len(gotStruct.Command), message)
			}
			if len(gotStruct.WebHook) != len(tc.webhooks) {
				t.Fatalf("%s\nwebhooks mismatch\nwant: %d\ngot:  %d\nbody: %q",
					packageName, len(tc.webhooks), len(gotStruct.WebHook), message)
			}
			// Check commands.
			if tc.commands != nil {
				for cmd, got := range gotStruct.Command {
					found := false
					for _, want := range tc.commands {
						if cmd == want.String() {
							found = true
							indexInService := 0
							for i, c := range tc.commands {
								if c.String() == cmd {
									indexInService = i
									break
								}
							}
							if got.Failed != svc.CommandController.Failed.Get(indexInService) ||
								got.NextRunnable != svc.CommandController.NextRunnable(indexInService) {
								t.Fatalf("%s\ncommand %q mismatch\nwant: %+v\ngot:  %+v\nbody: %q",
									packageName, cmd,
									want, got,
									message)
							}
							break
						}
					}
					if !found {
						t.Fatalf("%s\ncommand %q wasn't sent\nbody: %q",
							packageName, cmd, message)
					}
				}
			}
			// Check webhooks.
			if tc.webhooks != nil {
				for wh, got := range gotStruct.WebHook {
					found := false
					for _, want := range tc.webhooks {
						if wh == want.ID {
							found = true
							if got.Failed != want.ServiceStatus.Fails.WebHook.Get(wh) ||
								got.NextRunnable != want.NextRunnable() {
								t.Fatalf("%s\nwebhook %q mismatch\nwant: %+v\ngot:  %+v\nbody: %q",
									packageName, wh,
									want, got,
									message)
							}
							break
						}
					}
					if !found {
						t.Fatalf("%s\nwebhook %q wasn't sent\nbody: %q",
							packageName, wh, message)
					}
				}
			}
		})
	}
}

func TestHTTP_httpServiceRunActions(t *testing.T) {
	type wants struct {
		statusCode                  int
		stdoutRegex, bodyRegex      string
		wantSkipMessage             bool
		latestVersionIsApproved     bool
		upgradesApprovedVersion     bool
		upgradesDeployedVersion     bool
		approveCommandsIndividually bool
	}

	// GIVEN an API and a request for the Actions of a Service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	if api.Config.Settings.Data.DatabaseFile != "" {
		_ = os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		// Give time for save before TempDir clean-up.
		t.Cleanup(func() { time.Sleep(2 * config.DebounceDuration) })
	}
	tests := map[string]struct {
		serviceID     *string
		active        *bool
		payload       *string
		target        *string
		commands      command.Commands
		commandFails  []*bool
		webhooks      webhook.WebHooks
		webhookFails  map[string]*bool
		removeDVL     bool
		latestVersion string
		wants         wants
	}{
		"invalid payload": {
			payload: test.StringPtr("target: foo"),
			wants: wants{
				stdoutRegex: `Invalid payload - invalid character`,
			},
		},
		"ARGUS_SKIP, known service_id": {
			target: test.StringPtr("ARGUS_SKIP"),
			wants: wants{
				wantSkipMessage: true,
			},
		},
		"ARGUS_SKIP, inactive service_id": {
			active: test.BoolPtr(false),
			target: test.StringPtr("ARGUS_SKIP"),
			wants: wants{
				wantSkipMessage: false,
			},
		},
		"ARGUS_SKIP, unknown service_id": {
			serviceID: test.StringPtr("unknown?"),
			target:    test.StringPtr("ARGUS_SKIP"),
			wants: wants{
				stdoutRegex: `service "unknown\?" not found`,
			},
		},
		"ARGUS_SKIP, no service_id provided": {
			serviceID: test.StringPtr(""),
			target:    test.StringPtr("ARGUS_SKIP"),
			wants: wants{
				bodyRegex:  `service "" not found`,
				statusCode: http.StatusBadRequest,
			},
		},
		"target=nil, known service_id": {
			target: nil,
			wants: wants{
				stdoutRegex: `invalid payload, target service not provided`,
				statusCode:  http.StatusOK,
			},
		},
		"ARGUS_ALL, known service_id with no commands/webhooks": {
			target: test.StringPtr("ARGUS_ALL"),
			wants: wants{
				stdoutRegex: `"[^"]+" does not have any commands\/webhooks to approve`,
				statusCode:  http.StatusOK,
			},
		},
		"ARGUS_ALL, known service_id with command": {
			target: test.StringPtr("ARGUS_ALL"),
			commands: command.Commands{
				{"false", "0"}},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		"ARGUS_ALL, known service_id with webhook": {
			target: test.StringPtr("ARGUS_ALL"),
			webhooks: webhook.WebHooks{
				"known-service-and-webhook": webhook_test.WebHook(true, false, false)},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		"ARGUS_ALL, known service_id with multiple webhooks": {
			target: test.StringPtr("ARGUS_ALL"),
			webhooks: webhook.WebHooks{
				"known-service-and-multiple-webhook-0": webhook_test.WebHook(true, false, false),
				"known-service-and-multiple-webhook-1": webhook_test.WebHook(true, false, false)},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		"ARGUS_ALL, known service_id with multiple commands": {
			target: test.StringPtr("ARGUS_ALL"),
			commands: command.Commands{
				{"ls", "-a"}, {"false", "1"}},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		"ARGUS_ALL, known service_id with dvl and command and webhook that pass upgrades approved_version": {
			target: test.StringPtr("ARGUS_ALL"),
			commands: command.Commands{
				{"ls", "-b"}},
			webhooks: webhook.WebHooks{
				"known-service-dvl-webhook-0": webhook_test.WebHook(false, false, false)},
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesApprovedVersion: true,
			},
		},
		"ARGUS_ALL, known service_id with command and webhook that pass upgrades deployed_version": {
			target: test.StringPtr("ARGUS_ALL"),
			commands: command.Commands{
				{"ls", "-c"}},
			webhooks: webhook.WebHooks{
				"known-service-upgrade-deployed-version-webhook-0": webhook_test.WebHook(false, false, false)},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
			},
		},
		"ARGUS_ALL, known service_id with passing command and failing webhook doesn't upgrade any versions": {
			target: test.StringPtr("ARGUS_ALL"),
			commands: command.Commands{
				{"ls", "-d"}},
			webhooks: webhook.WebHooks{
				"known-service-fail-webhook-0": webhook_test.WebHook(true, false, false)},
			wants: wants{
				statusCode:              http.StatusOK,
				latestVersionIsApproved: true,
			},
		},
		"ARGUS_ALL, known service_id with failing command and passing webhook doesn't upgrade any versions": {
			target: test.StringPtr("ARGUS_ALL"),
			commands: command.Commands{
				{"fail"}},
			webhooks: webhook.WebHooks{
				"known-service-pass-webhook-0": webhook_test.WebHook(false, false, false)},
			wants: wants{
				statusCode:              http.StatusOK,
				latestVersionIsApproved: true,
			},
		},
		"webhook_NAME, known service_id with 1 webhook left to pass does upgrade deployed_version": {
			target: test.StringPtr("webhook_will_pass"),
			commands: command.Commands{
				{"ls", "-f"}},
			commandFails: []*bool{
				test.BoolPtr(false)},
			webhooks: webhook.WebHooks{
				"will_pass":  webhook_test.WebHook(false, false, false),
				"would_fail": webhook_test.WebHook(true, false, false)},
			webhookFails: map[string]*bool{
				"will_pass":  test.BoolPtr(true),
				"would_fail": test.BoolPtr(false)},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
			},
		},
		"command_NAME, known service_id with 1 command left to pass does upgrade deployed_version": {
			target: test.StringPtr("command_ls -g"),
			commands: command.Commands{
				{"ls", "/root"},
				{"ls", "-g"}},
			commandFails: []*bool{
				test.BoolPtr(false),
				test.BoolPtr(true)},
			webhooks: webhook.WebHooks{
				"would_fail": webhook_test.WebHook(true, false, false)},
			webhookFails: map[string]*bool{
				"would_fail": test.BoolPtr(false)},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
			},
		},
		"command_NAME, known service_id with multiple commands targeted individually (handle broadcast queue)": {
			commands: command.Commands{
				{"ls", "-h"},
				{"false", "2"},
				{"true"}},
			wants: wants{
				statusCode:                  http.StatusOK,
				approveCommandsIndividually: true,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			serviceID := util.DereferenceOrValue(tc.serviceID, name)
			svc := testService(serviceID, true)
			svc.Options.Active = tc.active
			svc.Defaults = &api.Config.Defaults.Service
			svc.HardDefaults = &api.Config.HardDefaults.Service
			svc.Status.Init(
				len(svc.Notify), len(tc.commands), len(tc.webhooks),
				serviceID, name, "",
				&dashboard.Options{
					WebURL: "https://example.com"})
			svc.Status.SetAnnounceChannel(api.Config.HardDefaults.Service.Status.AnnounceChannel)
			svc.Status.SetApprovedVersion("2.0.0", false)
			svc.Status.SetDeployedVersion("2.0.0", "", false)
			svc.Status.SetLatestVersion(util.ValueOrValue(tc.latestVersion, "3.0.0"), "", false)
			if tc.removeDVL {
				svc.DeployedVersionLookup = nil
			}
			svc.Command = tc.commands
			svc.CommandController = &command.Controller{}
			svc.CommandController.Init(
				&svc.Status,
				&svc.Command,
				&svc.Notify,
				test.StringPtr("10m"))
			if tc.commands == nil {
				svc.CommandController = nil
			}
			if len(tc.commandFails) != 0 {
				for i := range tc.commandFails {
					if tc.commandFails[i] != nil {
						svc.CommandController.Failed.Set(i, *tc.commandFails[i])
					}
				}
			}
			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
				&svc.Notify,
				&svc.Options.Interval)
			if len(tc.webhookFails) != 0 {
				for key := range tc.webhookFails {
					svc.WebHook[key].Failed.Set(key, tc.webhookFails[key])
				}
			}
			api.Config.OrderMutex.Lock()
			api.Config.Service[name] = svc
			api.Config.Order = append(api.Config.Order, name)
			api.Config.OrderMutex.Unlock()
			t.Cleanup(func() { api.Config.DeleteService(name) })
			// Set service_id.
			params := url.Values{}
			params.Set("service_id", serviceID)

			// WHEN the HTTP request is sent to run the actions.
			target := tc.target
			if tc.wants.approveCommandsIndividually {
				commandTarget := "command_" + tc.commands[0].String()
				target = &commandTarget
			}
			targetURL := "/api/v1/service/actions"
			sends := 1
			if tc.wants.approveCommandsIndividually {
				sends = len(tc.commands)
			}
			for sends != 0 {
				sends--
				if tc.wants.approveCommandsIndividually {
					*target = "command_" + tc.commands[sends].String()
				}
				body := []byte(`{}`)
				if target != nil {
					body = []byte(`{"target":"` + *target + `"}`)
				}
				if tc.payload != nil {
					body = []byte(*tc.payload)
				}
				req := httptest.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(body))
				req.URL.RawQuery = params.Encode()
				wHTTP := httptest.NewRecorder()
				api.httpServiceRunActions(wHTTP, req)
				res := wHTTP.Result()
				data, _ := io.ReadAll(res.Body)
				t.Logf("%s\ntarget=%q\nresult=%q, status_code=%d",
					packageName, util.DereferenceOrValue(target, "<nil>"), string(data), res.StatusCode)
				time.Sleep(10 * time.Microsecond)
			}
			time.Sleep(time.Duration((len(tc.commands)+len(tc.webhooks))*500) * time.Millisecond)

			// THEN we get the expected response.
			expecting := 0
			if tc.commands != nil {
				expecting += len(tc.commands)
				if tc.commandFails != nil {
					for i := range tc.commandFails {
						if util.DereferenceOrValue(tc.commandFails[i], true) == false {
							expecting--
						}
					}
				}
			}
			if tc.webhooks != nil {
				expecting += len(tc.webhooks)
				if tc.webhookFails != nil {
					for i := range tc.webhookFails {
						if tc.webhookFails[i] != nil && *tc.webhookFails[i] == false {
							expecting--
						}
					}
				}
			}
			if tc.wants.upgradesApprovedVersion {
				expecting++
			}
			if tc.wants.upgradesDeployedVersion {
				expecting++
			}
			if tc.wants.wantSkipMessage {
				expecting++
			}
			messages := make([]apitype.WebSocketMessage, expecting)
			t.Logf("%s\nexpecting %d messages",
				packageName, expecting)
			got := 0
			for expecting != 0 {
				message := <-api.Config.HardDefaults.Service.Status.AnnounceChannel
				if message == nil {
					stdout := releaseStdout()
					t.Log(time.Now(), stdout)
					t.Errorf("%s\nwant: %d more messages\ngot:  %v",
						packageName, expecting, message)
					return
				}
				_ = json.Unmarshal(message, &messages[got])
				raw, _ := json.Marshal(messages[got])
				t.Logf("%s\n%s\n",
					packageName, string(raw))
				got++
				expecting--
			}
			// extra message check.
			extraMessages := len(api.Config.HardDefaults.Service.Status.AnnounceChannel)
			if extraMessages != 0 {
				raw := <-api.Config.HardDefaults.Service.Status.AnnounceChannel
				t.Fatalf("%s\nwasn't expecting another message but got one\n%#v\n%s",
					packageName, extraMessages, string(raw))
			}
			stdout := releaseStdout()
			// stdout finishes.
			if tc.wants.stdoutRegex != "" {
				if !util.RegexCheck(tc.wants.stdoutRegex, stdout) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.wants.stdoutRegex, stdout)
				}
				return
			}
			t.Log(stdout)
			// Check version was skipped.
			if util.DereferenceOrDefault(tc.target) == "ARGUS_SKIP" {
				if tc.wants.wantSkipMessage &&
					messages[0].ServiceData.Status.ApprovedVersion != "SKIP_"+svc.Status.LatestVersion() {
					t.Errorf("%s\nLatestVersion %q wasn't skipped. ApprovedVersion=%q\ngot=%q",
						packageName, svc.Status.LatestVersion(),
						svc.Status.ApprovedVersion(),
						messages[0].ServiceData.Status.ApprovedVersion)
				}
			} else {
				// expecting = commands + webhooks that have not failed=false.
				expecting := 0
				if tc.commands != nil {
					expecting += len(tc.commands)
					if tc.commandFails != nil {
						for i := range tc.commandFails {
							if util.DereferenceOrValue(tc.commandFails[i], true) == false {
								expecting--
							}
						}
					}

				}
				if tc.webhooks != nil {
					expecting += len(tc.webhooks)
					if tc.webhookFails != nil {
						for i := range tc.webhookFails {
							if tc.webhookFails[i] != nil && *tc.webhookFails[i] == false {
								expecting--
							}
						}
					}
				}
				if tc.wants.upgradesApprovedVersion {
					expecting++
				}
				var received []string
				for i, message := range messages {
					t.Logf("%s - message[%d]=%v",
						packageName, i, message)
					receivedForAnAction := false
					for _, cmd := range tc.commands {
						if message.CommandData[cmd.String()] != nil {
							receivedForAnAction = true
							received = append(received, cmd.String())
							t.Logf("%s\nFOUND COMMAND %q - failed=%s",
								packageName, cmd.String(),
								test.StringifyPtr(message.CommandData[cmd.String()].Failed))
							break
						}
					}
					if !receivedForAnAction {
						for i := range tc.webhooks {
							if message.WebHookData[i] != nil {
								receivedForAnAction = true
								received = append(received, i)
								t.Logf("%s\nFOUND WEBHOOK %q - failed=%s",
									packageName, i,
									test.StringifyPtr(message.WebHookData[i].Failed))
								break
							}
						}
					}
					if !receivedForAnAction {
						//  IF we're expecting a message about approvedVersion.
						if tc.wants.upgradesApprovedVersion && message.Type == "VERSION" && message.SubType == "ACTION" {
							if message.ServiceData.Status.ApprovedVersion != svc.Status.LatestVersion() {
								t.Fatalf("%s\nexpected approved version to be updated to latest version in the message\n%#v\napproved=%#v, latest=%#v",
									packageName, message, message.ServiceData.Status.ApprovedVersion, svc.Status.LatestVersion())
							}
							continue
						}
						if tc.wants.upgradesDeployedVersion && message.Type == "VERSION" && message.SubType == "UPDATED" {
							if message.ServiceData.Status.DeployedVersion != svc.Status.LatestVersion() {
								t.Fatalf("%s\nexpected deployed version to be updated to latest version in the message\n%#v\ndeployed=%#v, latest=%#v",
									packageName, message, message.ServiceData.Status.DeployedVersion, svc.Status.LatestVersion())
							}
							continue
						}
						raw, _ := json.Marshal(message)
						if string(raw) != `{"page":"","type":""}` {
							t.Fatalf("%s\nUnexpected message\n%#v\n%s",
								packageName, message, raw)
						}
					}
				}
			}
		})
	}
}
