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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/command"
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

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	tests := []struct {
		name         string
		serviceID    *string
		commands     command.Commands
		commandFails []*bool
		webhooks     webhook.Webhooks
		webhookFails map[string]*bool
		wants        wants
	}{
		{
			name:      "service_id=unknown",
			serviceID: new("unknown?"),
			wants: wants{
				stdoutRegex: `^WARNING: .*service "unknown\?" not found`,
				bodyRegex:   `^{"message":"service \\"unknown\?\\" not found"}$`,
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:      "service_id=empty",
			serviceID: new(""),
			wants: wants{
				bodyRegex:  `{"message":"missing required query parameter: service_id"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:     "known service_id/0 command, 0 webhooks",
			commands: command.Commands{},
		},
		{
			name: "known service_id/1 command, 0 webhooks",
			commands: command.Commands{
				testCommand(true),
			},
		},
		{
			name: "known service_id/2 command, 0 webhooks",
			commands: command.Commands{
				testCommand(true),
				testCommand(false),
			},
		},
		{
			name: "known service_id/0 command, 1 webhooks",
			webhooks: webhook.Webhooks{
				"fail0": whtest.Webhook(t, true, false, false),
			},
		},
		{
			name: "known service_id/0 command, 2 webhooks",
			webhooks: webhook.Webhooks{
				"fail0": whtest.Webhook(t, true, false, false),
				"pass0": whtest.Webhook(t, false, false, false),
			},
		},
		{
			name: "known service_id/2 command, 2 webhooks",
			commands: command.Commands{
				testCommand(true),
				testCommand(false),
			},
			webhooks: webhook.Webhooks{
				"fail0": whtest.Webhook(t, true, false, false),
				"pass0": whtest.Webhook(t, false, false, false),
			},
		},
		{
			name: "known service_id/fail states",
			commands: command.Commands{
				testCommand(true),
				testCommand(false),
				{"true"},
			},
			commandFails: []*bool{
				new(true),
				new(false),
				nil,
			},
			webhooks: webhook.Webhooks{
				"fail0":  whtest.Webhook(t, true, false, false),
				"pass0":  whtest.Webhook(t, false, false, false),
				"unsent": whtest.Webhook(t, false, false, false),
			},
			webhookFails: map[string]*bool{
				"fail0": new(true),
				"pass0": new(false),
			},
		},
	}
	cfg := api.Config

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

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
				new("10m"),
			)

			svc.Webhook = tc.webhooks
			svc.Webhook.Init(
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
					WebURL: "https://example.com",
				},
			)
			// Status.Init resets the fail states, so they are set after it.
			for i, failed := range tc.commandFails {
				if failed != nil {
					svc.CommandController.Failed.Set(i, *failed)
				}
			}
			for id, failed := range tc.webhookFails {
				svc.Webhook[id].Failed.Set(id, failed)
			}

			cfg.OrderMu.Lock()
			cfg.Service[tc.name] = svc
			cfg.Order = append(cfg.Order, tc.name)
			cfg.OrderMu.Unlock()
			t.Cleanup(func() { cfg.DeleteService(tc.name) })

			target := "/api/v1/service/actions/"
			params := url.Values{}
			params.Set("service_id", serviceID)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.URL.RawQuery = params.Encode()
			w := httptest.NewRecorder()
			api.httpServiceGetActions(w, req)
			resp := w.Result()
			t.Cleanup(func() { _ = resp.Body.Close() })

			stdout := releaseStdout()
			if tc.wants.stdoutRegex != "" {
				if !util.RegexCheck(tc.wants.stdoutRegex, stdout) {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.wants.stdoutRegex)
				}
			}
			message, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if resp.StatusCode != tc.wants.statusCode {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					resp.StatusCode, message, tc.wants.statusCode,
				)
			}
			if got := string(message); !util.RegexCheck(tc.wants.bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
			}
			var gotStruct apitype.ActionSummary
			if err := decode.Unmarshal("json", message, &gotStruct); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if gotLen, wantLen := len(gotStruct.Command), len(tc.commands); gotLen != wantLen {
				t.Fatalf(
					"ActionSummary.Command length mismatch\ngot:  %d\nwant: %d\nbody: %q",
					gotLen, wantLen, message,
				)
			}
			if gotLen, wantLen := len(gotStruct.Webhook), len(tc.webhooks); gotLen != wantLen {
				t.Fatalf(
					"ActionSummary.Webhook length mismatch\ngot:  %d\nwant: %d\nbody: %q",
					gotLen, wantLen, message,
				)
			}
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
							var failed *bool
							if indexInService < len(tc.commandFails) {
								failed = tc.commandFails[indexInService]
							}
							gotFailed := test.StringifyPtr(got.Failed)
							wantFailed := test.StringifyPtr(failed)
							wantNextRunnable := svc.CommandController.NextRunnable(indexInService)
							if gotFailed != wantFailed || !got.NextRunnable.Equal(wantNextRunnable) {
								t.Fatalf(
									"ActionSummary.Command[%q] mismatch\n"+
										"got:  failed=%s, next_runnable=%s\n"+
										"want: failed=%s, next_runnable=%s\n"+
										"body: %q",
									cmd,
									gotFailed, got.NextRunnable,
									wantFailed, wantNextRunnable,
									message,
								)
							}
							break
						}
					}
					if !found {
						t.Fatalf(
							"ActionSummary.Command[%q] mismatch\ngot:  present\nwant: absent\nbody: %q",
							cmd, message,
						)
					}
				}
			}
			if tc.webhooks != nil {
				for wh, got := range gotStruct.Webhook {
					found := false
					for _, want := range tc.webhooks {
						if wh == want.ID {
							found = true
							gotFailed := test.StringifyPtr(got.Failed)
							wantFailed := test.StringifyPtr(tc.webhookFails[wh])
							wantNextRunnable := want.NextRunnable()
							if gotFailed != wantFailed || !got.NextRunnable.Equal(wantNextRunnable) {
								t.Fatalf(
									"ActionSummary.Webhook[%q] mismatch\n"+
										"got:  failed=%s, next_runnable=%s\n"+
										"want: failed=%s, next_runnable=%s\n"+
										"body: %q",
									wh,
									gotFailed, got.NextRunnable,
									wantFailed, wantNextRunnable,
									message,
								)
							}
							break
						}
					}
					if !found {
						t.Fatalf(
							"ActionSummary.Webhook[%q] mismatch\ngot:  present\nwant: absent\nbody: %q",
							wh, message,
						)
					}
				}
			}
		})
	}
}

func TestHTTP_HTTPServiceGetActions__requiresActionExecute(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	addActionableServices(t, api, "readable", "hidden")
	adminCookie := loginCookie(t, api, "admin", "admin-password")

	actionerCookie := createActioner(t, api, deps)

	createAuthGroup(t, deps, "readers", "", []rbac.Grant{
		{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
	})
	createAuthUser(t, deps, "reader", "reader-password", "readers")
	readerCookie := loginCookie(t, api, "reader", "reader-password")

	tests := []struct {
		name       string
		serviceID  string
		cookie     *http.Cookie
		wantStatus int
	}{
		{
			name:       "valid/execute on a readable service",
			serviceID:  "readable",
			cookie:     actionerCookie,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid/service:read alone cannot list the actions",
			serviceID:  "readable",
			cookie:     readerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid/global execute cannot reach an unreadable service",
			serviceID:  "hidden",
			cookie:     actionerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "valid/admin reads every service",
			serviceID:  "hidden",
			cookie:     adminCookie,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := authedRequest(
				http.MethodGet, "/api/v1/service/actions?service_id="+tc.serviceID, "", tc.cookie,
			)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
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
		upgradesApprovedVersion     bool
		upgradesDeployedVersion     bool
		approveCommandsIndividually bool
		actions                     map[string]int // Runs announced per command/webhook.
	}

	receiver := test.GitHubWebhookServer(t)

	file := t.Name() + ".yml"
	api := testAPI(t, file)
	tests := []struct {
		name          string
		serviceID     *string
		active        *bool
		payload       *string
		target        *string
		commands      command.Commands
		commandFails  []*bool
		webhooks      webhook.Webhooks
		webhookFails  map[string]*bool
		removeDVL     bool
		latestVersion string
		wants         wants
	}{
		{
			name:    "invalid payload",
			payload: new("target: foo"),
			wants: wants{
				stdoutRegex: "^WARNING: .*" + test.TrimYAML(`
					invalid payload:
						jsontext: invalid character`,
				),
				bodyRegex:  `^{"message":"invalid payload:`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "ARGUS_SKIP/known service_id",
			target: new(ActionSkip),
			wants: wants{
				statusCode:      http.StatusOK,
				wantSkipMessage: true,
			},
		},
		{
			name:   "ARGUS_SKIP/inactive service_id",
			active: new(false),
			target: new(ActionSkip),
			wants: wants{
				stdoutRegex:     `^WARNING: .*service is inactive, actions can't be run for it`,
				bodyRegex:       `^{"message":"service is inactive, actions can't be run for it"}$`,
				statusCode:      http.StatusBadRequest,
				wantSkipMessage: false,
			},
		},
		{
			name:      "ARGUS_SKIP/unknown service_id",
			serviceID: new("unknown?"),
			target:    new(ActionSkip),
			wants: wants{
				stdoutRegex: `^WARNING: .*service "unknown\?" not found`,
				bodyRegex:   `^{"message":"service \\"unknown\?\\" not found"}$`,
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:      "ARGUS_SKIP/no service_id provided",
			serviceID: new(""),
			target:    new(ActionSkip),
			wants: wants{
				bodyRegex:  `^{"message":"missing required query parameter: service_id"}$`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "target=nil, known service_id",
			target: nil,
			wants: wants{
				stdoutRegex: `^WARNING: .*invalid payload, target service not provided`,
				bodyRegex:   `^{"message":"invalid payload, target service not provided"}$`,
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with no commands or webhooks",
			target: new(ActionAll),
			wants: wants{
				stdoutRegex: `"[^"]+" does not have any commands\/webhooks to approve`,
				statusCode:  http.StatusOK,
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with command",
			target: new(ActionAll),
			commands: command.Commands{
				{"false", "0"},
			},
			wants: wants{
				statusCode: http.StatusOK,
				actions:    map[string]int{"false 0": 1},
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with webhook",
			target: new(ActionAll),
			webhooks: webhook.Webhooks{
				"known-service-and-webhook": whtest.Webhook(t, true, false, false),
			},
			wants: wants{
				statusCode: http.StatusOK,
				actions:    map[string]int{"known-service-and-webhook": 1},
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with multiple webhooks",
			target: new(ActionAll),
			webhooks: webhook.Webhooks{
				"known-service-and-multiple-webhook-0": whtest.Webhook(t, true, false, false),
				"known-service-and-multiple-webhook-1": whtest.Webhook(t, true, false, false),
			},
			wants: wants{
				statusCode: http.StatusOK,
				actions: map[string]int{
					"known-service-and-multiple-webhook-0": 1,
					"known-service-and-multiple-webhook-1": 1,
				},
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with multiple commands",
			target: new(ActionAll),
			commands: command.Commands{
				{"ls", "-a"},
				{"false", "1"},
			},
			wants: wants{
				statusCode: http.StatusOK,
				actions:    map[string]int{"ls -a": 1, "false 1": 1},
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with dvl and command and webhook that pass upgrades approved_version",
			target: new(ActionAll),
			commands: command.Commands{
				{"ls", "-b"},
			},
			webhooks: webhook.Webhooks{
				"known-service-dvl-webhook-0": whtest.Webhook(t, false, false, false),
			},
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesApprovedVersion: true,
				actions:                 map[string]int{"ls -b": 1, "known-service-dvl-webhook-0": 1},
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with command and webhook that pass upgrades deployed_version",
			target: new(ActionAll),
			commands: command.Commands{
				{"ls", "-c"},
			},
			webhooks: webhook.Webhooks{
				"known-service-upgrade-deployed-version-webhook-0": whtest.Webhook(t, false, false, false),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
				actions: map[string]int{
					"ls -c": 1,
					"known-service-upgrade-deployed-version-webhook-0": 1,
				},
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with passing command and failing webhook doesn't upgrade any versions",
			target: new(ActionAll),
			commands: command.Commands{
				{"ls", "-d"},
			},
			webhooks: webhook.Webhooks{
				"known-service-fail-webhook-0": whtest.Webhook(t, true, false, false),
			},
			wants: wants{
				statusCode: http.StatusOK,
				actions:    map[string]int{"ls -d": 1, "known-service-fail-webhook-0": 1},
			},
		},
		{
			name:   "ARGUS_ALL/known service_id with failing command and passing webhook doesn't upgrade any versions",
			target: new(ActionAll),
			commands: command.Commands{
				{"fail"},
			},
			webhooks: webhook.Webhooks{
				"known-service-pass-webhook-0": whtest.Webhook(t, false, false, false),
			},
			wants: wants{
				statusCode: http.StatusOK,
				actions:    map[string]int{"fail": 1, "known-service-pass-webhook-0": 1},
			},
		},
		{
			name:   "webhook_NAME/known service_id with 1 webhook left to pass does upgrade deployed_version",
			target: new("webhook_will_pass"),
			commands: command.Commands{
				{"ls", "-f"},
			},
			commandFails: []*bool{
				new(false),
			},
			webhooks: webhook.Webhooks{
				"will_pass":  whtest.Webhook(t, false, false, false),
				"would_fail": whtest.Webhook(t, true, false, false),
			},
			webhookFails: map[string]*bool{
				"will_pass":  new(true),
				"would_fail": new(false),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
				actions:                 map[string]int{"will_pass": 1},
			},
		},
		{
			name:   "webhook_NAME/runs the targeted webhook even though it already succeeded",
			target: new("webhook_already_passed"),
			webhooks: webhook.Webhooks{
				"already_passed": whtest.Webhook(t, false, false, false),
				"unsent":         whtest.Webhook(t, false, false, false),
			},
			webhookFails: map[string]*bool{
				"already_passed": new(false),
				"unsent":         new(true),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode: http.StatusOK,
				actions:    map[string]int{"already_passed": 1},
			},
		},
		{
			name:   "command_NAME/known service_id with 1 command left to pass does upgrade deployed_version",
			target: new("command_ls -g"),
			commands: command.Commands{
				{"ls", "/root"},
				{"ls", "-g"},
			},
			commandFails: []*bool{
				new(false),
				new(true),
			},
			webhooks: webhook.Webhooks{
				"would_fail": whtest.Webhook(t, true, false, false),
			},
			webhookFails: map[string]*bool{
				"would_fail": new(false),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode:              http.StatusOK,
				upgradesDeployedVersion: true,
				actions:                 map[string]int{"ls -g": 1},
			},
		},
		{
			name:   "command_NAME/runs the targeted command even though it already succeeded",
			target: new("command_ls -i"),
			commands: command.Commands{
				{"ls", "-i"},
				{"false", "3"},
			},
			commandFails: []*bool{
				new(false),
				new(true),
			},
			removeDVL:     true,
			latestVersion: "0.9.0",
			wants: wants{
				statusCode: http.StatusOK,
				actions:    map[string]int{"ls -i": 1},
			},
		},
		{
			name: "command_NAME/known service_id with multiple commands targeted individually, handle broadcast queue",
			commands: command.Commands{
				{"ls", "-h"},
				{"false", "2"},
				{"true"},
			},
			wants: wants{
				statusCode:                  http.StatusOK,
				approveCommandsIndividually: true,
				actions:                     map[string]int{"ls -h": 1, "false 2": 1, "true": 1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

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
				new("10m"),
			)
			if len(tc.commandFails) != 0 {
				for i := range tc.commandFails {
					if tc.commandFails[i] != nil {
						svc.CommandController.Failed.Set(i, *tc.commandFails[i])
					}
				}
			}

			for _, wh := range tc.webhooks {
				wh.URL = receiver.URL
			}
			svc.Webhook = tc.webhooks
			svc.Webhook.Init(
				&svc.Status,
				whCfg,
				&svc.Notify,
				&svc.Options.Interval,
			)
			if len(tc.webhookFails) != 0 {
				for key := range tc.webhookFails {
					svc.Webhook[key].Failed.Set(key, tc.webhookFails[key])
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

			params := url.Values{}
			params.Set("service_id", serviceID)

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
					t.Logf("send %d", sends)
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
				w := httptest.NewRecorder()
				api.httpServiceRunActions(w, req)
				resp := w.Result()
				data, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("io.ReadAll() failed: %v", err)
				}

				if got, want := resp.StatusCode, tc.wants.statusCode; got != want {
					t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, data, want)
				}
				if got := string(data); !util.RegexCheck(tc.wants.bodyRegex, got) {
					t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wants.bodyRegex)
				}
				time.Sleep(10 * time.Microsecond)
			}
			time.Sleep(time.Duration((len(tc.commands)+len(tc.webhooks))*500) * time.Millisecond)

			wantActions := tc.wants.actions
			if wantActions == nil {
				wantActions = map[string]int{}
			}

			expecting := 0
			for _, count := range wantActions {
				expecting += count
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
			wantMessages := expecting
			messages := make([]apitype.WebSocketMessage, expecting)
			got := 0
			for expecting != 0 {
				var message []byte
				select {
				case msg := <-api.Config.HardDefaults.Service.Status.AnnounceChannel:
					message = msg
				case <-time.After(2 * time.Second):
					t.Fatalf(
						"Status.AnnounceChannel message count mismatch\n"+
							"got:  %d (timeout after 2s)\nwant: %d",
						got, wantMessages,
					)
				}
				if message == nil {
					stdout := releaseStdout()
					t.Errorf(
						"message %d: Status.AnnounceChannel message mismatch\n"+
							"got:  nil\nwant: non-nil\nstdout: %q",
						got, stdout,
					)
					return
				}
				if err := decode.Unmarshal("json", message, &messages[got]); err != nil {
					t.Fatalf("message %d: decode.Unmarshal() failed: %v", got, err)
				}
				t.Logf("message %d: announced: %s", got, message)
				got++
				expecting--
			}
			if extraMessages := len(api.Config.HardDefaults.Service.Status.AnnounceChannel); extraMessages != 0 {
				raw := <-api.Config.HardDefaults.Service.Status.AnnounceChannel
				t.Fatalf(
					"Status.AnnounceChannel extra message count mismatch\n"+
						"got:  %d (first: %s)\nwant: 0",
					extraMessages, raw,
				)
			}
			stdout := releaseStdout()
			if tc.wants.stdoutRegex != "" {
				if !util.RegexCheck(tc.wants.stdoutRegex, stdout) {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.wants.stdoutRegex)
				}
				return
			}
			t.Logf("stdout: %q", stdout)
			if util.DerefOrZero(tc.target) == ActionSkip {
				if tc.wants.wantSkipMessage &&
					messages[0].ServiceData.Status.ApprovedVersion != serviceinfo.SkippedVersion(svc.Status.LatestVersion()) {
					t.Errorf(
						"WebSocketMessage.ServiceData.Status.ApprovedVersion mismatch\ngot:  %q\nwant: %q",
						messages[0].ServiceData.Status.ApprovedVersion,
						serviceinfo.SkippedVersion(svc.Status.LatestVersion()),
					)
				}
			} else {
				for i, message := range messages {
					receivedForAnAction := false
					for _, cmd := range tc.commands {
						if message.CommandData[cmd.String()] != nil {
							receivedForAnAction = true
							break
						}
					}
					if !receivedForAnAction {
						for id := range tc.webhooks {
							if message.WebhookData[id] != nil {
								receivedForAnAction = true
								break
							}
						}
					}
					if !receivedForAnAction {
						if tc.wants.upgradesApprovedVersion && message.Type == "VERSION" && message.SubType == "ACTION" {
							if got, want := message.ServiceData.Status.ApprovedVersion, svc.Status.LatestVersion(); got != want {
								t.Fatalf(
									"message %d: WebSocketMessage.ServiceData.Status.ApprovedVersion mismatch\n"+
										"got:  %q\nwant: %q",
									i, got, want,
								)
							}
							continue
						}
						if tc.wants.upgradesDeployedVersion && message.Type == "VERSION" && message.SubType == "UPDATED" {
							if got, want := message.ServiceData.Status.DeployedVersion, svc.Status.LatestVersion(); got != want {
								t.Fatalf(
									"message %d: WebSocketMessage.ServiceData.Status.DeployedVersion mismatch\n"+
										"got:  %q\nwant: %q",
									i, got, want,
								)
							}
							continue
						}
						raw, err := decode.Marshal("json", message)
						if err != nil {
							t.Fatalf("message %d: decode.Marshal() failed: %v", i, err)
						}
						if want := `{"page":"","type":""}`; string(raw) != want {
							t.Fatalf(
								"message %d: WebSocketMessage mismatch\ngot:  %s\nwant: %s",
								i, raw, want,
							)
						}
					}
				}

				gotActions := make(map[string]int, len(wantActions))
				for _, message := range messages {
					for key := range message.CommandData {
						gotActions[key]++
					}
					for key := range message.WebhookData {
						gotActions[key]++
					}
				}
				if testErr := test.AssertMapEqual(
					t,
					gotActions, wantActions,
					"WebSocketMessage action counts",
				); testErr != nil {
					t.Error(testErr)
				}
			}
		})
	}
}

func TestHTTP_HTTPServiceRunActions__requiresServiceRead(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	addActionableServices(t, api, "readable", "hidden")

	actionerCookie := createActioner(t, api, deps)
	adminCookie := loginCookie(t, api, "admin", "admin-password")

	tests := []struct {
		name       string
		serviceID  string
		cookie     *http.Cookie
		wantStatus int
	}{
		{
			name:      "valid/global execute on a readable service",
			serviceID: "readable",
			cookie:    actionerCookie, wantStatus: http.StatusOK,
		},
		{
			name:      "invalid/global execute cannot reach an unreadable service",
			serviceID: "hidden",
			cookie:    actionerCookie, wantStatus: http.StatusForbidden,
		},
		{
			name:      "valid/admin reads every service",
			serviceID: "hidden",
			cookie:    adminCookie, wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := authedRequest(
				http.MethodPost, "/api/v1/service/actions?service_id="+tc.serviceID,
				`{"target":"`+ActionSkip+`"}`, tc.cookie,
			)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

// addActionableServices adds a service with one command and a pending release
// (2.0.0 deployed, 3.0.0 latest) for each ID.
func addActionableServices(t *testing.T, api *API, serviceIDs ...string) {
	t.Helper()

	for _, serviceID := range serviceIDs {
		svc := testService(t, serviceID, "url", "url", true)
		svc.Defaults = &api.Config.Defaults.Service
		svc.HardDefaults = &api.Config.HardDefaults.Service
		svc.Command = command.Commands{{"true"}}
		svc.CommandController = command.NewController(
			&svc.Status,
			svc.Command,
			svc.Notify,
			new("10m"),
		)
		svc.Status.SetAnnounceChannel(api.Config.HardDefaults.Service.Status.AnnounceChannel)
		svc.Status.SetDeployedVersion("2.0.0", "", false)
		svc.Status.SetLatestVersion("3.0.0", "", false)
		svc.Status.SetApprovedVersion("2.0.0", false)

		api.Config.OrderMu.Lock()
		api.Config.Service[serviceID] = svc
		api.Config.Order = append(api.Config.Order, serviceID)
		api.Config.OrderMu.Unlock()
		t.Cleanup(func() { api.Config.DeleteService(serviceID) })
	}
}

// createActioner creates a user with GLOBAL service_action:execute and
// service:read on "readable" only, and returns its session cookie.
func createActioner(t *testing.T, api *API, deps *AuthDeps) *http.Cookie {
	t.Helper()

	createAuthGroup(t, deps, "actioners", "", []rbac.Grant{
		{Resource: rbac.ResourceServiceAction, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
		{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "readable"}},
	})
	createAuthUser(t, deps, "actioner", "actioner-password", "actioners")
	return loginCookie(t, api, "actioner", "actioner-password")
}
