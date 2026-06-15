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

package v1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestHTTP_HTTPServiceGetActions(t *testing.T) {
	whCfg := whtest.PlainConfig(t)
	type wants struct {
		stdoutRegex, bodyRegex string
		statusCode             int
	}

	// GIVEN: an API and a request for the Actions of a Service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	tests := []struct {
		name      string
		serviceID *string
		commands  command.Commands
		webhooks  webhook.WebHooks
		wants     wants
	}{
		{
			name:      "service_id=unknown",
			serviceID: test.Ptr("unknown?"),
			wants: wants{
				stdoutRegex: `service "unknown\?" not found`,
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:      "service_id=nil",
			serviceID: test.Ptr(""),
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:     "known service_id, 0 command, 0 webhooks",
			commands: command.Commands{},
		},
		{
			name: "known service_id, 1 command, 0 webhooks",
			commands: command.Commands{
				testCommand(true),
			},
		},
		{
			name: "known service_id, 2 command, 0 webhooks",
			commands: command.Commands{
				testCommand(true),
				testCommand(false),
			},
		},
		{
			name: "known service_id, 0 command, 1 webhooks",
			webhooks: webhook.WebHooks{
				"fail0": whtest.WebHook(t, true, false, false),
			},
		},
		{
			name: "known service_id, 0 command, 2 webhooks",
			webhooks: webhook.WebHooks{
				"fail0": whtest.WebHook(t, true, false, false),
				"pass0": whtest.WebHook(t, false, false, false),
			},
		},
		{
			name: "known service_id, 2 command, 2 webhooks",
			commands: command.Commands{
				testCommand(true),
				testCommand(false),
			},
			webhooks: webhook.WebHooks{
				"fail0": whtest.WebHook(t, true, false, false),
				"pass0": whtest.WebHook(t, false, false, false),
			},
		},
	}
	cfg := api.Config

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			if tc.wants.statusCode == 0 {
				tc.wants.statusCode = http.StatusOK
			}
			svc := testService(t, tc.name, "url", "url", true)
			serviceID := util.DerefOr(tc.serviceID, svc.ID)
			svc.Defaults = &cfg.Defaults.Service
			svc.HardDefaults = &cfg.HardDefaults.Service

			svc.Command = tc.commands
			svc.CommandController = command.NewController(
				&svc.Status,
				svc.Command,
				svc.Notify,
				test.Ptr("10m"),
			)

			svc.WebHook = tc.webhooks
			svc.WebHook.Init(
				&svc.Status,
				whCfg,
				&svc.Notify,
				&svc.Options.Interval,
			)

			svc.Status.SetAnnounceChannel(cfg.HardDefaults.Service.Status.AnnounceChannel)
			svc.Status.SetDeployedVersion("2.0.0", "", false)
			svc.Status.SetLatestVersion("3.0.0", "", true)
			svc.Status.SetApprovedVersion("2.0.0", false)
			svc.Status.Init(
				len(tc.commands), len(svc.Notify), len(tc.webhooks),
				status.ServiceInfo{
					ID:   serviceID,
					Name: tc.name,
				},
				&dashboard.Options{
					OptionsBase: dashboard.OptionsBase{
						WebURL: "https://example.com",
					},
				},
			)

			cfg.OrderMu.Lock()
			cfg.Service[tc.name] = svc
			cfg.Order = append(cfg.Order, tc.name)
			cfg.OrderMu.Unlock()

			target := "/api/v1/service/actions/"
			params := url.Values{}
			// Set service_id.
			params.Set("service_id", serviceID)

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			wHTTP := httptest.NewRecorder()
			api.httpServiceGetActions(wHTTP, req)
			res := wHTTP.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpServiceGetActions()", packageName)

			// THEN: we get the expected response.
			stdout := releaseStdout()
			// stdout finishes.
			if tc.wants.stdoutRegex != "" {
				tc.wants.stdoutRegex = strings.ReplaceAll(tc.wants.stdoutRegex, "__name__", tc.name)
				if !util.RegexCheck(tc.wants.stdoutRegex, stdout) {
					t.Errorf("%s stdout mismatch\ngot:  %q\nwant: %q", prefix, stdout, tc.wants.stdoutRegex)
				}
			}
			message, _ := io.ReadAll(res.Body)
			if res.StatusCode != tc.wants.statusCode {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d\nbody: %q",
					prefix, res.StatusCode, tc.wants.statusCode, message,
				)
			}
			var gotStruct apitype.ActionSummary
			_ = decode.Unmarshal("json", message, &gotStruct)
			if gotLen, wantLen := len(gotStruct.Command), len(tc.commands); gotLen != wantLen {
				t.Fatalf(
					"%s commands count mismatch\ngot:  %d\nwant: %d\nbody: %q",
					prefix, gotLen, wantLen, message,
				)
			}
			if gotLen, wantLen := len(gotStruct.WebHook), len(tc.webhooks); gotLen != wantLen {
				t.Fatalf(
					"%s webhooks count mismatch\ngot:  %d\nwant: %d\nbody: %q",
					prefix, gotLen, wantLen, message,
				)
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
								t.Fatalf(
									"%s command %q mismatch\ngot:  %+v\nwant: %+v\nbody: %q",
									prefix, cmd,
									got, want,
									message,
								)
							}
							break
						}
					}
					if !found {
						t.Fatalf("%s command %q wasn't found in response\nbody: %q", prefix, cmd, message)
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
								t.Fatalf(
									"%s webhook %q mismatch\ngot:  %+v\nwant: %+v\nbody: %q",
									prefix, wh,
									got, want,
									message,
								)
							}
							break
						}
					}
					if !found {
						t.Fatalf(
							"%s webhook %q wasn't found in response\nbody: %q",
							prefix, wh, message,
						)
					}
				}
			}
		})
	}
}

func TestHTTP_HTTPServiceRunActions(t *testing.T) {
	whCfg := whtest.PlainConfig(t)
	type wants struct {
		statusCode                  int
		stdoutRegex, bodyRegex      string
		wantSkipMessage             bool
		latestVersionIsApproved     bool
		upgradesApprovedVersion     bool
		upgradesDeployedVersion     bool
		approveCommandsIndividually bool
	}

	// GIVEN: an API and a request for the Actions of a Service.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	if api.Config.Settings.Data.DatabaseFile != "" {
		// Give time for save before TempDir clean-up.
		t.Cleanup(func() { time.Sleep(2 * config.DebounceDuration) })
	}
	tests := []struct {
		name          string
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
		{
			name:    "invalid payload",
			payload: test.Ptr("target: foo"),
			wants: wants{
				stdoutRegex: test.TrimYAML(`
					invalid payload:
						jsontext: invalid character`,
				),
			},
		},
		{
			name:   "ARGUS_SKIP, known service_id",
			target: test.Ptr(ActionSkip),
			wants: wants{
				wantSkipMessage: true,
			},
		},
		{
			name:   "ARGUS_SKIP, inactive service_id",
			active: test.Ptr(false),
			target: test.Ptr(ActionSkip),
			wants: wants{
				wantSkipMessage: false,
			},
		},
		{
			name:      "ARGUS_SKIP, unknown service_id",
			serviceID: test.Ptr("unknown?"),
			target:    test.Ptr(ActionSkip),
			wants: wants{
				stdoutRegex: `service "unknown\?" not found`,
			},
		},
		{
			name:      "ARGUS_SKIP, no service_id provided",
			serviceID: test.Ptr(""),
			target:    test.Ptr(ActionSkip),
			wants: wants{
				bodyRegex:  `service "" not found`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "target=nil, known service_id",
			target: nil,
			wants: wants{
				stdoutRegex: `invalid payload, target service not provided`,
				statusCode:  http.StatusOK,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with no commands/webhooks",
			target: test.Ptr(ActionAll),
			wants: wants{
				stdoutRegex: `"[^"]+" does not have any commands\/webhooks to approve`,
				statusCode:  http.StatusOK,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with command",
			target: test.Ptr(ActionAll),
			commands: command.Commands{
				{"false", "0"},
			},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with webhook",
			target: test.Ptr(ActionAll),
			webhooks: webhook.WebHooks{
				"known-service-and-webhook": whtest.WebHook(t, true, false, false),
			},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with multiple webhooks",
			target: test.Ptr(ActionAll),
			webhooks: webhook.WebHooks{
				"known-service-and-multiple-webhook-0": whtest.WebHook(t, true, false, false),
				"known-service-and-multiple-webhook-1": whtest.WebHook(t, true, false, false),
			},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with multiple commands",
			target: test.Ptr(ActionAll),
			commands: command.Commands{
				{"ls", "-a"},
				{"false", "1"},
			},
			wants: wants{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with dvl and command and webhook that pass upgrades approved_version",
			target: test.Ptr(ActionAll),
			commands: command.Commands{
				{"ls", "-b"},
			},
			webhooks: webhook.WebHooks{
				"known-service-dvl-webhook-0": whtest.WebHook(t, false, false, false),
			},
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesApprovedVersion: true,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with command and webhook that pass upgrades deployed_version",
			target: test.Ptr(ActionAll),
			commands: command.Commands{
				{"ls", "-c"},
			},
			webhooks: webhook.WebHooks{
				"known-service-upgrade-deployed-version-webhook-0": whtest.WebHook(t, false, false, false),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with passing command and failing webhook doesn't upgrade any versions",
			target: test.Ptr(ActionAll),
			commands: command.Commands{
				{"ls", "-d"},
			},
			webhooks: webhook.WebHooks{
				"known-service-fail-webhook-0": whtest.WebHook(t, true, false, false),
			},
			wants: wants{
				statusCode:              http.StatusOK,
				latestVersionIsApproved: true,
			},
		},
		{
			name:   "ARGUS_ALL, known service_id with failing command and passing webhook doesn't upgrade any versions",
			target: test.Ptr(ActionAll),
			commands: command.Commands{
				{"fail"},
			},
			webhooks: webhook.WebHooks{
				"known-service-pass-webhook-0": whtest.WebHook(t, false, false, false),
			},
			wants: wants{
				statusCode:              http.StatusOK,
				latestVersionIsApproved: true,
			},
		},
		{
			name:   "webhook_NAME, known service_id with 1 webhook left to pass does upgrade deployed_version",
			target: test.Ptr("webhook_will_pass"),
			commands: command.Commands{
				{"ls", "-f"},
			},
			commandFails: []*bool{
				test.Ptr(false),
			},
			webhooks: webhook.WebHooks{
				"will_pass":  whtest.WebHook(t, false, false, false),
				"would_fail": whtest.WebHook(t, true, false, false),
			},
			webhookFails: map[string]*bool{
				"will_pass":  test.Ptr(true),
				"would_fail": test.Ptr(false),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
			},
		},
		{
			name:   "command_NAME, known service_id with 1 command left to pass does upgrade deployed_version",
			target: test.Ptr("command_ls -g"),
			commands: command.Commands{
				{"ls", "/root"},
				{"ls", "-g"},
			},
			commandFails: []*bool{
				test.Ptr(false),
				test.Ptr(true),
			},
			webhooks: webhook.WebHooks{
				"would_fail": whtest.WebHook(t, true, false, false),
			},
			webhookFails: map[string]*bool{
				"would_fail": test.Ptr(false),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
			},
		},
		{
			name: "command_NAME, known service_id with multiple commands targeted individually (handle broadcast queue)",
			commands: command.Commands{
				{"ls", "-h"},
				{"false", "2"},
				{"true"},
			},
			wants: wants{
				statusCode:                  http.StatusOK,
				approveCommandsIndividually: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			serviceID := util.DerefOr(tc.serviceID, tc.name)
			svc := testService(t, serviceID, "url", "url", true)
			svc.Options.Active = tc.active
			svc.Defaults = &api.Config.Defaults.Service
			svc.HardDefaults = &api.Config.HardDefaults.Service

			if tc.removeDVL {
				svc.DeployedVersionLookup = nil
			}

			svc.Command = tc.commands
			svc.CommandController = command.NewController(
				&svc.Status,
				svc.Command,
				svc.Notify,
				test.Ptr("10m"),
			)
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
				whCfg,
				&svc.Notify,
				&svc.Options.Interval,
			)
			if len(tc.webhookFails) != 0 {
				for key := range tc.webhookFails {
					svc.WebHook[key].Failed.Set(key, tc.webhookFails[key])
				}
			}

			svc.Status.SetAnnounceChannel(api.Config.HardDefaults.Service.Status.AnnounceChannel)
			svc.Status.SetDeployedVersion("2.0.0", "", false)
			svc.Status.SetLatestVersion(util.ValueOr(tc.latestVersion, "3.0.0"), "", false)
			svc.Status.SetApprovedVersion("2.0.0", false)

			api.Config.OrderMu.Lock()
			api.Config.Service[tc.name] = svc
			api.Config.Order = append(api.Config.Order, tc.name)
			api.Config.OrderMu.Unlock()
			t.Cleanup(func() { api.Config.DeleteService(tc.name) })

			// Set service_id.
			params := url.Values{}
			params.Set("service_id", serviceID)

			// WHEN: the HTTP request is sent to run the actions.
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
			var prefix string
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
				prefix = fmt.Sprintf("%s\nAPI.httpServiceRunActions()", packageName)
				t.Logf(
					"%s target=%q\nresult=%q, status_code=%d",
					prefix, util.DerefOr(target, "<nil>"),
					string(data), res.StatusCode,
				)
				time.Sleep(10 * time.Microsecond)
			}
			time.Sleep(time.Duration((len(tc.commands)+len(tc.webhooks))*500) * time.Millisecond)

			// THEN: we get the expected response.
			expecting := 0
			if tc.commands != nil {
				expecting += len(tc.commands)
				if tc.commandFails != nil {
					for i := range tc.commandFails {
						if util.DerefOr(tc.commandFails[i], true) == false {
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
			t.Logf("%s expecting %d messages", prefix, expecting)
			got := 0
			for expecting != 0 {
				var message []byte
				select {
				case msg := <-api.Config.HardDefaults.Service.Status.AnnounceChannel:
					message = msg
				case <-time.After(2 * time.Second):
					t.Fatalf(
						"%s timed out waiting for message\ngot:  %d\nwant: %d more",
						prefix, got, expecting,
					)
				}
				if message == nil {
					stdout := releaseStdout()
					t.Log(time.Now(), stdout)
					t.Errorf(
						"%s expected more messages\ngot:  0\nwant: %d",
						prefix, expecting,
					)
					return
				}
				_ = decode.Unmarshal("json", message, &messages[got])
				raw, _ := decode.Marshal("json", messages[got])
				t.Logf(
					"%s\n%s\n",
					packageName, string(raw),
				)
				got++
				expecting--
			}
			// extra message check.
			if extraMessages := len(api.Config.HardDefaults.Service.Status.AnnounceChannel); extraMessages != 0 {
				raw := <-api.Config.HardDefaults.Service.Status.AnnounceChannel
				t.Fatalf(
					"%s wasn't expecting another message but got %d:\n%s",
					prefix, extraMessages, string(raw),
				)
			}
			stdout := releaseStdout()
			// stdout finishes.
			if tc.wants.stdoutRegex != "" {
				if !util.RegexCheck(tc.wants.stdoutRegex, stdout) {
					t.Errorf(
						"%s stdout mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, tc.wants.stdoutRegex,
					)
				}
				return
			}
			t.Log(stdout)
			// Check version was skipped.
			if util.DerefOrZero(tc.target) == ActionSkip {
				if tc.wants.wantSkipMessage &&
					messages[0].ServiceData.Status.ApprovedVersion != serviceinfo.SkippedVersion(svc.Status.LatestVersion()) {
					t.Errorf(
						"%s LatestVersion %q wasn't skipped. ApprovedVersion=%q\ngot=%q",
						prefix, svc.Status.LatestVersion(),
						svc.Status.ApprovedVersion(),
						messages[0].ServiceData.Status.ApprovedVersion,
					)
				}
			} else {
				// expecting = commands + webhooks that have not failed=false.
				expecting := 0
				if tc.commands != nil {
					expecting += len(tc.commands)
					if tc.commandFails != nil {
						for i := range tc.commandFails {
							if util.DerefOr(tc.commandFails[i], true) == false {
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
					t.Logf(
						"%s - message[%d]=%v",
						prefix, i, message,
					)
					receivedForAnAction := false
					for _, cmd := range tc.commands {
						if message.CommandData[cmd.String()] != nil {
							receivedForAnAction = true
							received = append(received, cmd.String())
							t.Logf(
								"%s FOUND COMMAND %q - failed=%s",
								prefix, cmd.String(),
								test.StringifyPtr(message.CommandData[cmd.String()].Failed),
							)
							break
						}
					}
					if !receivedForAnAction {
						for i := range tc.webhooks {
							if message.WebHookData[i] != nil {
								receivedForAnAction = true
								received = append(received, i)
								t.Logf(
									"%s FOUND WEBHOOK %q - failed=%s",
									prefix, i,
									test.StringifyPtr(message.WebHookData[i].Failed),
								)
								break
							}
						}
					}
					if !receivedForAnAction {
						//  IF we're expecting a message about approvedVersion.
						if tc.wants.upgradesApprovedVersion && message.Type == "VERSION" && message.SubType == "ACTION" {
							if got, want := message.ServiceData.Status.ApprovedVersion, svc.Status.LatestVersion(); got != want {
								t.Fatalf(
									"%s expected approved version to be updated to latest version in the message\n%#v\napproved=%#v, latest=%#v",
									prefix, message,
									got, want,
								)
							}
							continue
						}
						if tc.wants.upgradesDeployedVersion && message.Type == "VERSION" && message.SubType == "UPDATED" {
							if got, want := message.ServiceData.Status.DeployedVersion, svc.Status.LatestVersion(); got != want {
								t.Fatalf(
									"%s expected deployed version to be updated to latest version in the message\n%#v\ndeployed=%#v, latest=%#v",
									prefix, message,
									got, want,
								)
							}
							continue
						}
						raw, _ := decode.Marshal("json", message)
						if string(raw) != `{"page":"","type":""}` {
							t.Fatalf("%s Unexpected message\n%#v\n%s", prefix, message, raw)
						}
					}
				}
			}
		})
	}
}
