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

package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestHTTP_httpServiceGetActions(t *testing.T) {
	// GIVEN an API and a request for the Actions of a Service
	file := "TestHTTP_httpServiceGetActions.yml"
	api := testAPI(file)
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()
	tests := map[string]struct {
		serviceID   string
		stdoutRegex string
		bodyRegex   string
		commands    command.Slice
		webhooks    webhook.Slice
		statusCode  int
	}{
		"service_id=unknown": {
			serviceID:   "unknown?",
			stdoutRegex: `service "unknown\?" not found`,
			statusCode:  http.StatusNotFound,
		},
		"service_id=nil": {
			serviceID:   "",
			stdoutRegex: `service "" not found`,
			statusCode:  http.StatusNotFound,
		},
		"known service_id, 0 command, 0 webhooks": {
			serviceID:   "__name__",
			stdoutRegex: `VERBOSE: httpServiceActions \([^)]+\), __name__\s$`,
			// DEBUG: WebSocket (127.0.0.1), READ {"version":1,"page":"APPROVALS","type":"ACTIONS","service_data":{"id":"__name__","status":{}}}
			// VERBOSE: wsCommand (127.0.0.1), -\n
			// VERBOSE: wsWebHook (127.0.0.1), -\n
		},
		"known service_id, 1 command, 0 webhooks": {
			serviceID: "__name__",
			commands: command.Slice{
				testCommand(true)},
		},
		"known service_id, 2 command, 0 webhooks": {
			serviceID: "__name__",
			commands: command.Slice{
				testCommand(true), testCommand(false)},
		},
		"known service_id, 0 command, 1 webhooks": {
			serviceID: "__name__",
			webhooks: webhook.Slice{
				"fail0": testWebHook(true, "fail0")},
		},
		"known service_id, 0 command, 2 webhooks": {
			serviceID: "__name__",
			webhooks: webhook.Slice{
				"fail0": testWebHook(true, "fail0"),
				"pass0": testWebHook(false, "pass0")},
		},
		"known service_id, 2 command, 2 webhooks": {
			serviceID: "__name__",
			commands: command.Slice{
				testCommand(true), testCommand(false)},
			webhooks: webhook.Slice{
				"fail0": testWebHook(true, "fail0"),
				"pass0": testWebHook(false, "pass0")},
		},
	}
	cfg := api.Config

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			test.StdoutMutex.Lock()
			defer test.StdoutMutex.Unlock()

			if tc.statusCode == 0 {
				tc.statusCode = http.StatusOK
			}
			svc := testService(name)
			tc.serviceID = strings.ReplaceAll(tc.serviceID, "__name__", name)
			svc.Defaults = &cfg.Defaults.Service
			svc.HardDefaults = &cfg.HardDefaults.Service
			svc.Status.Init(
				len(svc.Notify), len(tc.commands), len(tc.webhooks),
				&tc.serviceID,
				test.StringPtr("https://example.com"))
			svc.Status.SetAnnounceChannel(cfg.HardDefaults.Service.Status.AnnounceChannel)
			svc.Status.SetApprovedVersion("2.0.0", false)
			svc.Status.SetDeployedVersion("2.0.0", false)
			svc.Status.SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
			svc.Status.SetLatestVersion("3.0.0", true)
			svc.Status.SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
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
				&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{},
				&svc.Notify,
				&svc.Options.Interval)
			cfg.OrderMutex.Lock()
			cfg.Service[name] = svc
			cfg.Order = append(cfg.Order, name)
			cfg.OrderMutex.Unlock()
			defer cfg.DeleteService(name)
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			target := "/api/v1/service/actions/"
			target += url.QueryEscape(tc.serviceID)

			// WHEN that HTTP request is sent
			req := httptest.NewRequest(http.MethodGet, target, nil)
			vars := map[string]string{
				"service_name": tc.serviceID}
			req = mux.SetURLVars(req, vars)
			wHTTP := httptest.NewRecorder()
			api.httpServiceGetActions(wHTTP, req)
			res := wHTTP.Result()
			defer res.Body.Close()

			// THEN we get the expected response
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			// stdout finishes
			if tc.stdoutRegex != "" {
				tc.stdoutRegex = strings.ReplaceAll(tc.stdoutRegex, "__name__", name)
				re := regexp.MustCompile(tc.stdoutRegex)
				match := re.MatchString(output)
				if !match {
					t.Errorf("match on %q not found in\n%q",
						tc.stdoutRegex, output)
				}
			}
			message, _ := io.ReadAll(res.Body)
			if res.StatusCode != tc.statusCode {
				t.Errorf("expected status code %d but got %d\n%s",
					tc.statusCode, res.StatusCode, message)
			} else if tc.statusCode != http.StatusOK {
				return
			}
			var gotStruct api_type.ActionSummary
			_ = json.Unmarshal(message, &gotStruct)
			if len(gotStruct.Command) != len(tc.commands) {
				t.Fatalf("expected %d commands but got %d\n%s",
					len(tc.commands), len(gotStruct.Command), message)
			}
			if len(gotStruct.WebHook) != len(tc.webhooks) {
				t.Fatalf("expected %d webhooks but got %d\n%s",
					len(tc.webhooks), len(gotStruct.WebHook), message)
			}
			// Check commands
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
								t.Fatalf("command %q: expected %+v but got %+v\n%s",
									cmd, want, got, message)
							}
							break
						}
					}
					if !found {
						t.Fatalf("command %q wasn't sent\n%s",
							cmd, message)
					}
				}
			}
			// Check webhooks
			if tc.webhooks != nil {
				for wh, got := range gotStruct.WebHook {
					found := false
					for _, want := range tc.webhooks {
						if wh == want.ID {
							found = true
							if got.Failed != want.ServiceStatus.Fails.WebHook.Get(wh) ||
								got.NextRunnable != want.NextRunnable() {
								t.Fatalf("webhook %q: expected %+v but got %+v\n%s",
									wh, want, got, message)
							}
							break
						}
					}
					if !found {
						t.Fatalf("webhook %q wasn't sent\n%s",
							wh, message)
					}
				}
			}
		})
	}
}

func TestHTTP_httpServiceRunActions(t *testing.T) {
	// GIVEN an API and a request for the Actions of a Service
	file := "TestHTTP_httpServiceRunActions.yml"
	api := testAPI(file)
	defer func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != nil {
			os.RemoveAll(*api.Config.Settings.Data.DatabaseFile)
		}
	}()
	tests := map[string]struct {
		serviceID                   string
		active                      *bool
		payload                     *string
		target                      *string
		wantSkipMessage             bool
		stdoutRegex                 string
		bodyRegex                   string
		commands                    command.Slice
		commandFails                []*bool
		webhooks                    webhook.Slice
		webhookFails                map[string]*bool
		removeDVL                   bool
		latestVersion               string
		latestVersionIsApproved     bool
		upgradesApprovedVersion     bool
		upgradesDeployedVersion     bool
		approveCommandsIndividually bool
	}{
		"invalid payload": {
			serviceID:   "__name__",
			payload:     test.StringPtr("target: foo"), // not JSON
			stdoutRegex: `Invalid payload - invalid character`,
		},
		"ARGUS_SKIP known service_id": {
			serviceID:       "__name__",
			target:          test.StringPtr("ARGUS_SKIP"),
			wantSkipMessage: true,
		},
		"ARGUS_SKIP inactive service_id": {
			serviceID:       "__name__",
			active:          test.BoolPtr(false),
			target:          test.StringPtr("ARGUS_SKIP"),
			wantSkipMessage: false,
		},
		"ARGUS_SKIP unknown service_id": {
			serviceID:   "unknown?",
			target:      test.StringPtr("ARGUS_SKIP"),
			stdoutRegex: `service "unknown\?" not found`,
		},
		"ARGUS_SKIP empty service_id": {
			serviceID:   "",
			target:      test.StringPtr("ARGUS_SKIP"),
			stdoutRegex: `service "" not found`,
		},
		"target=nil, known service_id": {
			serviceID:   "__name__",
			target:      nil,
			stdoutRegex: `invalid payload, target service not provided`,
		},
		"ARGUS_ALL, known service_id with no commands/webhooks": {
			serviceID:   "__name__",
			target:      test.StringPtr("ARGUS_ALL"),
			stdoutRegex: `"[^"]+" does not have any commands\/webhooks to approve`,
		},
		"ARGUS_ALL, known service_id with command": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"false", "0"}},
		},
		"ARGUS_ALL, known service_id with webhook": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			webhooks: webhook.Slice{
				"known-service-and-webhook": testWebHook(true, "known-service-and-webhook")},
		},
		"ARGUS_ALL, known service_id with multiple webhooks": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			webhooks: webhook.Slice{
				"known-service-and-multiple-webhook-0": testWebHook(true, "known-service-and-multiple-webhook-0"),
				"known-service-and-multiple-webhook-1": testWebHook(true, "known-service-and-multiple-webhook-1")},
		},
		"ARGUS_ALL, known service_id with multiple commands": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"ls", "-a"}, {"false", "1"}},
		},
		"ARGUS_ALL, known service_id with dvl and command and webhook that pass upgrades approved_version": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"ls", "-b"}},
			webhooks: webhook.Slice{
				"known-service-dvl-webhook-0": testWebHook(false, "known-service-dvl-webhook-0")},
			upgradesApprovedVersion: true,
			latestVersion:           "0.9.0",
		},
		"ARGUS_ALL, known service_id with command and webhook that pass upgrades deployed_version": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"ls", "-c"}},
			webhooks: webhook.Slice{
				"known-service-upgrade-deployed-version-webhook-0": testWebHook(false,
					"known-service-upgrade-deployed-version-webhook-0")},
			removeDVL:               true,
			upgradesDeployedVersion: true,
			latestVersion:           "0.9.0",
		},
		"ARGUS_ALL, known service_id with passing command and failing webhook doesn't upgrade any versions": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"ls", "-d"}},
			webhooks: webhook.Slice{
				"known-service-fail-webhook-0": testWebHook(true, "known-service-fail-webhook-0")},
			latestVersionIsApproved: true,
		},
		"ARGUS_ALL, known service_id with failing command and passing webhook doesn't upgrade any versions": {
			serviceID: "__name__",
			target:    test.StringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"fail"}},
			webhooks: webhook.Slice{
				"known-service-pass-webhook-0": testWebHook(false, "known-service-pass-webhook-0")},
			latestVersionIsApproved: true,
		},
		"webhook_NAME, known service_id with 1 webhook left to pass does upgrade deployed_version": {
			serviceID: "__name__",
			target:    test.StringPtr("webhook_will_pass"),
			commands: command.Slice{
				{"ls", "-f"}},
			commandFails: []*bool{
				test.BoolPtr(false)},
			webhooks: webhook.Slice{
				"will_pass":  testWebHook(false, "will_pass"),
				"would_fail": testWebHook(true, "would_fail")},
			webhookFails: map[string]*bool{
				"will_pass":  test.BoolPtr(true),
				"would_fail": test.BoolPtr(false)},
			removeDVL:               true,
			upgradesDeployedVersion: true,
			latestVersion:           "0.9.0",
		},
		"command_NAME, known service_id with 1 command left to pass does upgrade deployed_version": {
			serviceID: "__name__",
			target:    test.StringPtr("command_ls -g"),
			commands: command.Slice{
				{"ls", "/root"},
				{"ls", "-g"}},
			commandFails: []*bool{
				test.BoolPtr(false),
				test.BoolPtr(true)},
			webhooks: webhook.Slice{
				"would_fail": testWebHook(true, "would_fail")},
			webhookFails: map[string]*bool{
				"would_fail": test.BoolPtr(false)},
			removeDVL:               true,
			upgradesDeployedVersion: true,
			latestVersion:           "0.9.0",
		},
		"command_NAME, known service_id with multiple commands targeted individually (handle broadcast queue)": {
			serviceID: "__name__",
			commands: command.Slice{
				{"ls", "-h"},
				{"false", "2"},
				{"true"}},
			approveCommandsIndividually: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			test.StdoutMutex.Lock()
			defer test.StdoutMutex.Unlock()

			tc.serviceID = strings.ReplaceAll(tc.serviceID, "__name__", name)
			svc := testService(tc.serviceID)
			svc.Options.Active = tc.active
			svc.Defaults = &api.Config.Defaults.Service
			svc.HardDefaults = &api.Config.HardDefaults.Service
			svc.Status.Init(
				len(svc.Notify), len(tc.commands), len(tc.webhooks),
				&tc.serviceID,
				test.StringPtr("https://example.com"))
			svc.Status.SetAnnounceChannel(api.Config.HardDefaults.Service.Status.AnnounceChannel)
			svc.Status.SetApprovedVersion("2.0.0", false)
			svc.Status.SetDeployedVersion("2.0.0", false)
			svc.Status.SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
			svc.Status.SetLatestVersion("3.0.0", true)
			svc.Status.SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
			if tc.latestVersion != "" {
				svc.Status.SetLatestVersion(tc.latestVersion, false)
			}
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
				&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{},
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
			defer api.Config.DeleteService(name)
			// Stdout setup
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN the HTTP request is sent to run the action(s)
			target := tc.target
			if tc.approveCommandsIndividually {
				commandTarget := "command_" + tc.commands[0].String()
				target = &commandTarget
			}
			targetURL := "/api/v1/service/actions/" + url.QueryEscape(tc.serviceID)
			sends := 1
			if tc.approveCommandsIndividually {
				sends = len(tc.commands)
			}
			for sends != 0 {
				sends--
				if tc.approveCommandsIndividually {
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
				vars := map[string]string{
					"service_name": tc.serviceID,
				}
				req = mux.SetURLVars(req, vars)
				wHTTP := httptest.NewRecorder()
				api.httpServiceRunActions(wHTTP, req)
				res := wHTTP.Result()
				data, _ := io.ReadAll(res.Body)
				t.Log(fmt.Sprintf("target=%q\nresult=%q, status_code=%d",
					util.PtrValueOrValue(target, "<nil>"), string(data), res.StatusCode))
				time.Sleep(10 * time.Microsecond)
			}
			time.Sleep(time.Duration((len(tc.commands)+len(tc.webhooks))*500) * time.Millisecond)

			// THEN we get the expected response
			expecting := 0
			if tc.commands != nil {
				expecting += len(tc.commands)
				if tc.commandFails != nil {
					for i := range tc.commandFails {
						if util.EvalNilPtr(tc.commandFails[i], true) == false {
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
			if tc.upgradesApprovedVersion {
				expecting++
			}
			if tc.upgradesDeployedVersion {
				expecting++
			}
			if tc.wantSkipMessage {
				expecting++
			}
			messages := make([]api_type.WebSocketMessage, expecting)
			t.Logf("expecting %d messages",
				expecting)
			got := 0
			for expecting != 0 {
				message := <-*api.Config.HardDefaults.Service.Status.AnnounceChannel
				if message == nil {
					w.Close()
					out, _ := io.ReadAll(r)
					os.Stdout = stdout
					output := string(out)
					t.Log(time.Now(), output)
					t.Errorf("expecting %d more messages but got %v",
						expecting, message)
					return
				}
				json.Unmarshal(message, &messages[got])
				raw, _ := json.Marshal(messages[got])
				t.Logf("\n%s\n", string(raw))
				got++
				expecting--
			}
			// extra message check
			extraMessages := len(*api.Config.HardDefaults.Service.Status.AnnounceChannel)
			if extraMessages != 0 {
				raw := <-*api.Config.HardDefaults.Service.Status.AnnounceChannel
				t.Fatalf("wasn't expecting another message but got one\n%#v\n%s",
					extraMessages, string(raw))
			}
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			// stdout finishes
			if tc.stdoutRegex != "" {
				re := regexp.MustCompile(tc.stdoutRegex)
				match := re.MatchString(output)
				if !match {
					t.Errorf("match on %q not found in\n%q",
						tc.stdoutRegex, output)
				}
				return
			}
			t.Log(output)
			// Check version was skipped
			if util.DefaultIfNil(tc.target) == "ARGUS_SKIP" {
				if tc.wantSkipMessage &&
					messages[0].ServiceData.Status.ApprovedVersion != "SKIP_"+svc.Status.LatestVersion() {
					t.Errorf("LatestVersion %q wasn't skipped. approved is %q\ngot=%q",
						svc.Status.LatestVersion(),
						svc.Status.ApprovedVersion(),
						messages[0].ServiceData.Status.ApprovedVersion)
				}
			} else {
				// expecting = commands + webhooks that have not failed=false
				expecting := 0
				if tc.commands != nil {
					expecting += len(tc.commands)
					if tc.commandFails != nil {
						for i := range tc.commandFails {
							if util.EvalNilPtr(tc.commandFails[i], true) == false {
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
				if tc.upgradesApprovedVersion {
					expecting++
				}
				var received []string
				for i, message := range messages {
					t.Logf("message %d - %v",
						i, message)
					receivedForAnAction := false
					for _, command := range tc.commands {
						if message.CommandData[command.String()] != nil {
							receivedForAnAction = true
							received = append(received, command.String())
							t.Logf("FOUND COMMAND %q - failed=%s",
								command.String(), test.StringifyPtr(message.CommandData[command.String()].Failed))
							break
						}
					}
					if !receivedForAnAction {
						for i := range tc.webhooks {
							if message.WebHookData[i] != nil {
								receivedForAnAction = true
								received = append(received, i)
								t.Logf("FOUND WEBHOOK %q - failed=%s",
									i, test.StringifyPtr(message.WebHookData[i].Failed))
								break
							}
						}
					}
					if !receivedForAnAction {
						//  IF we're expecting a message about approvedVersion
						if tc.upgradesApprovedVersion && message.Type == "VERSION" && message.SubType == "ACTION" {
							if message.ServiceData.Status.ApprovedVersion != svc.Status.LatestVersion() {
								t.Fatalf("expected approved version to be updated to latest version in the message\n%#v\napproved=%#v, latest=%#v",
									message, message.ServiceData.Status.ApprovedVersion, svc.Status.LatestVersion())
							}
							continue
						}
						if tc.upgradesDeployedVersion && message.Type == "VERSION" && message.SubType == "UPDATED" {
							if message.ServiceData.Status.DeployedVersion != svc.Status.LatestVersion() {
								t.Fatalf("expected deployed version to be updated to latest version in the message\n%#v\ndeployed=%#v, latest=%#v",
									message, message.ServiceData.Status.DeployedVersion, svc.Status.LatestVersion())
							}
							continue
						}
						raw, _ := json.Marshal(message)
						if string(raw) != `{"page":"","type":""}` {
							t.Fatalf("Unexpected message\n%#v\n%s",
								message, raw)
						}
					}
				}
			}
		})
	}
}
