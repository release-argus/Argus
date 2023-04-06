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

//go:build integration

package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

var router *mux.Router
var port *string
var cfg *config.Config

func TestMain(m *testing.M) {
	// GIVEN a valid config with a Service
	testLogging("DEBUG", true)
	cfg = testConfig("TestMain.yml")
	defer os.Remove(cfg.File)
	defer os.Remove(*cfg.Settings.Data.DatabaseFile)
	port = cfg.Settings.Web.ListenPort

	// WHEN the Router is fetched for this Config
	router = newWebUI(cfg)
	go http.ListenAndServe("localhost:"+*port, router)

	// THEN Web UI is accessible for the tests
	code := m.Run()
	os.Exit(code)
}

func TestMainWithRoutePrefix(t *testing.T) {
	// GIVEN a valid config with a Service
	cfg := testConfig("TestMainWithRoutePrefix.yml")
	defer os.Remove(cfg.File)
	defer os.Remove(*cfg.Settings.Data.DatabaseFile)
	*cfg.Settings.Web.RoutePrefix = "/test"

	// WHEN the Web UI is started with this Config
	go Run(cfg, util.NewJLog("WARN", false))
	time.Sleep(100 * time.Millisecond)

	// THEN Web UI is accessible
	url := fmt.Sprintf("http://localhost:%s%s/metrics",
		*cfg.Settings.Web.ListenPort, *cfg.Settings.Web.RoutePrefix)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	if resp.StatusCode != 200 {
		t.Errorf("Should have got a 200 from a GET on %s",
			url)
	}
}

func TestWebAccessible(t *testing.T) {
	// GIVEN we have the Web UI Router from TestMain()
	tests := map[string]struct {
		path      string
		bodyRegex string
	}{
		"/approvals": {
			path: "/approvals"},
		"/metrics": {
			path:      "/metrics",
			bodyRegex: "go_gc_duration_"},
		"/api/v1/healthcheck": {
			path:      "/api/v1/healthcheck",
			bodyRegex: fmt.Sprintf(`^Alive$`)},
		"/api/v1/version": {
			path: "/api/v1/version",
			bodyRegex: fmt.Sprintf(`"goVersion":"%s"`,
				util.GoVersion)},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we make a request to path
			req, _ := http.NewRequest("GET", tc.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			// THEN we get a Status OK
			if response.Code != http.StatusOK {
				t.Errorf("Expected a 200, got %d",
					response.Code)
			}
			if tc.bodyRegex != "" {
				body := response.Body.String()
				re := regexp.MustCompile(tc.bodyRegex)
				match := re.MatchString(body)
				if !match {
					t.Errorf("expected %q in body\ngot: %q",
						tc.bodyRegex, response.Body.String())
				}
			}
		})
	}
}

func TestAccessibleHTTPS(t *testing.T) {
	// GIVEN a bunch of URLs to test and the webserver is running with HTTPS
	tests := map[string]struct {
		path      string
		bodyRegex string
	}{
		"/approvals": {
			path: "/approvals"},
		"/metrics": {
			path:      "/metrics",
			bodyRegex: "go_gc_duration_"},
		"/api/v1/healthcheck": {
			path:      "/api/v1/healthcheck",
			bodyRegex: fmt.Sprintf(`^Alive$`)},
		"/api/v1/version": {
			path: "/api/v1/version",
			bodyRegex: fmt.Sprintf(`"goVersion":"%s"`,
				util.GoVersion)},
	}
	testLogging("DEBUG", true)
	cfg := testConfig("TestAccessibleHTTPS.yml")
	cfg.Settings.Web.CertFile = stringPtr("TestAccessibleHTTPS_cert.pem")
	cfg.Settings.Web.KeyFile = stringPtr("TestAccessibleHTTPS_key.pem")
	generateCertFiles(*cfg.Settings.Web.CertFile, *cfg.Settings.Web.KeyFile)
	defer os.Remove(cfg.File)
	defer os.Remove(*cfg.Settings.Data.DatabaseFile)
	defer os.Remove(*cfg.Settings.Web.CertFile)
	defer os.Remove(*cfg.Settings.Web.KeyFile)

	router = newWebUI(cfg)
	go Run(cfg, util.NewJLog("WARN", false))
	time.Sleep(250 * time.Millisecond)
	address := fmt.Sprintf("https://localhost:%s", *cfg.Settings.Web.ListenPort)

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we make a HTTPS request to path
			req, _ := http.NewRequest("GET", address+tc.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			// THEN we get a Status OK
			if response.Code != http.StatusOK {
				t.Errorf("Expected a 200, got %d",
					response.Code)
			}
			if tc.bodyRegex != "" {
				body := response.Body.String()
				re := regexp.MustCompile(tc.bodyRegex)
				match := re.MatchString(body)
				if !match {
					t.Errorf("expected %q in body\ngot: %q",
						tc.bodyRegex, response.Body.String())
				}
			}
		})
	}
}

///////////////
// WebSocket //
///////////////

func connectToWebSocket(t *testing.T) *websocket.Conn {
	var dialer = websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
	}

	url := fmt.Sprintf("ws://localhost:%s/ws", *port)

	// Connect to the server
	ws, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(10 * time.Millisecond)
	return ws
}

// seeIfMessage will try and get a message from the WebSocket
// if it receives no message within 120*50ms (6s), it will give up and return nil
func seeIfMessage(t *testing.T, ws *websocket.Conn) *api_type.WebSocketMessage {
	var message api_type.WebSocketMessage
	msgChan := make(chan api_type.WebSocketMessage, 1)
	errChan := make(chan error, 1)
	go func() {
		_, p, err := ws.ReadMessage()
		json.Unmarshal(p, &message)
		errChan <- err
		msgChan <- message
	}()
	for i := 0; i < 120; i++ {
		time.Sleep(50 * time.Millisecond)
		if len(errChan) != 0 {
			break
		}
	}
	if len(errChan) == 0 {
		t.Logf("No messages received after %s", time.Duration(50*120)*time.Millisecond)
		return nil
	}
	time.Sleep(time.Millisecond)
	err := <-errChan
	if err != nil {
		t.Logf("err - %s", err.Error())
		return nil
	}
	msg := <-msgChan
	return &msg
}

func TestWebSocket(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	tests := map[string]struct {
		msg         string
		stdoutRegex string
		bodyRegex   string
	}{
		"no version": {
			msg:         `{"key": "value"}`,
			stdoutRegex: `^DEBUG:[^:]+READ \{"key": "value"\}\n$`},
		"no version, unknown type": {
			msg:         `{"page": "APPROVALS", "type": "SHAZAM", "key": "value"}`,
			stdoutRegex: "Unknown TYPE"},
		"invalid JSON": {
			msg:         `{"version": 1, "key": "value"`,
			stdoutRegex: "missing/invalid version key"},
		"unknown page": {
			msg:         `{"version": 1, "page": "foo", "type": "value"}`,
			stdoutRegex: "Unknown PAGE"},
		"APPROVALS - unknown type": {
			msg:         `{"version": 1, "page": "APPROVALS", "type": "value"}`,
			stdoutRegex: "Unknown APPROVALS Type"},
		"RUNTIME_BUILD - unknown type": {
			msg:         `{"version": 1, "page": "RUNTIME_BUILD", "type": "value"}`,
			stdoutRegex: "Unknown RUNTIME_BUILD Type"},
		"FLAGS - unknown type": {
			msg:         `{"version": 1, "page": "FLAGS", "type": "value"}`,
			stdoutRegex: "Unknown FLAGS Type"},
		"CONFIG - unknown type": {
			msg:         `{"version": 1, "page": "CONFIG", "type": "value"}`,
			stdoutRegex: "Unknown CONFIG Type"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			ws := connectToWebSocket(t)
			defer ws.Close()
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN we send a message
			if err := ws.WriteMessage(websocket.TextMessage, []byte(tc.msg)); err != nil {
				t.Errorf("%v",
					err)
			}
			time.Sleep(50 * time.Millisecond)

			// THEN we receive the expected response
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			re := regexp.MustCompile(tc.stdoutRegex)
			match := re.MatchString(output)
			if !match {
				t.Errorf("match on %q not found in\n%q",
					tc.stdoutRegex, output)
			}
		})
	}
}

func TestWebSocketApprovalsINIT(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	tests := map[string]struct {
		order []string
	}{
		"INIT": {
			order: cfg.Order},
		"INIT with nil Service in config.Order": {
			order: append(cfg.Order, "nilService")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			order := cfg.Order
			cfg.Order = tc.order
			ws := connectToWebSocket(t)
			defer ws.Close()

			// WHEN we send a message to the server (wsServiceInit)
			msg := api_type.WebSocketMessage{Version: intPtr(1), Page: "APPROVALS", Type: "INIT"}
			data, _ := json.Marshal(msg)
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Fatalf("error sending message\n%s",
					err.Error())
			}

			// THEN we get the expected responses
			// ORDERING
			message := seeIfMessage(t, ws)
			cfg.Order = order
			if message.Order == nil || len(*message.Order) != len(tc.order) {
				t.Fatalf("want order=%#v\ngot  order=%#v",
					tc.order, *message.Order)
			}
			for i := range tc.order {
				if tc.order[i] != (*message.Order)[i] {
					t.Fatalf("want order=%#v\ngot  order=%#v",
						tc.order, *message.Order)
				}
			}
			// SERVICE
			receivedOrder := *message.Order
			for _, key := range receivedOrder {
				if cfg.Service[key] == nil {
					continue
				}
				message = seeIfMessage(t, ws)
				if message == nil {
					t.Fatal("expecting another message but didn't get one")
				}
				if message.ServiceData == nil {
					t.Errorf("bad message, didn't contain ServiceData for %q",
						key)
				}
			}
			message = seeIfMessage(t, ws)
			if message != nil {
				t.Fatalf("wasn't expecting another message but got one\n%#v",
					message)
			}
		})
	}
}

func TestWebSocketApprovalsVERSION(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	testLogging("VERBOSE", true)
	tests := map[string]struct {
		serviceID                   string
		active                      *bool
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
		"ARGUS_SKIP known service_id": {
			serviceID:       "test",
			target:          stringPtr("ARGUS_SKIP"),
			wantSkipMessage: true,
		},
		"ARGUS_SKIP inactive service_id": {
			serviceID:       "test",
			active:          boolPtr(false),
			target:          stringPtr("ARGUS_SKIP"),
			wantSkipMessage: false,
		},
		"ARGUS_SKIP unknown service_id": {
			serviceID:   "unknown?",
			target:      stringPtr("ARGUS_SKIP"),
			stdoutRegex: "service not found",
		},
		"ARGUS_SKIP empty service_id": {
			serviceID:   "",
			target:      stringPtr("ARGUS_SKIP"),
			stdoutRegex: "service_data.id not provided",
		},
		"target=nil, known service_id": {
			serviceID:   "test",
			target:      nil,
			stdoutRegex: "target for command/webhook not provided",
		},
		"ARGUS_ALL, known service_id with no commands/webhooks": {
			serviceID:   "test",
			target:      stringPtr("ARGUS_ALL"),
			stdoutRegex: "does not have any commands/webhooks to approve",
		},
		"ARGUS_ALL, known service_id with command": {
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"false", "0"}},
		},
		"ARGUS_ALL, known service_id with webhook": {
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
			webhooks: webhook.Slice{
				"known-service-and-webhook": testWebHook(true, "known-service-and-webhook")},
		},
		"ARGUS_ALL, known service_id with multiple webhooks": {
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
			webhooks: webhook.Slice{
				"known-service-and-multiple-webhook-0": testWebHook(true, "known-service-and-multiple-webhook-0"),
				"known-service-and-multiple-webhook-1": testWebHook(true, "known-service-and-multiple-webhook-1")},
		},
		"ARGUS_ALL, known service_id with multiple commands": {
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"ls", "-a"}, {"false", "1"}},
		},
		"ARGUS_ALL, known service_id with dvl and command and webhook that pass upgrades approved_version": {
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"ls", "-b"}},
			webhooks: webhook.Slice{
				"known-service-dvl-webhook-0": testWebHook(false, "known-service-dvl-webhook-0")},
			upgradesApprovedVersion: true,
			latestVersion:           "0.9.0",
		},
		"ARGUS_ALL, known service_id with command and webhook that pass upgrades deployed_version": {
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
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
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"ls", "-d"}},
			webhooks: webhook.Slice{
				"known-service-fail-webhook-0": testWebHook(true, "known-service-fail-webhook-0")},
			latestVersionIsApproved: true,
		},
		"ARGUS_ALL, known service_id with failing command and passing webhook doesn't upgrade any versions": {
			serviceID: "test",
			target:    stringPtr("ARGUS_ALL"),
			commands: command.Slice{
				{"fail"}},
			webhooks: webhook.Slice{
				"known-service-pass-webhook-0": testWebHook(false, "known-service-pass-webhook-0")},
			latestVersionIsApproved: true,
		},
		"webhook_NAME, known service_id with 1 webhook left to pass does upgrade deployed_version": {
			serviceID: "test",
			target:    stringPtr("webhook_will_pass"),
			commands: command.Slice{
				{"ls", "-f"}},
			commandFails: []*bool{
				boolPtr(false)},
			webhooks: webhook.Slice{
				"will_pass":  testWebHook(false, "will_pass"),
				"would_fail": testWebHook(true, "would_fail")},
			webhookFails: map[string]*bool{
				"will_pass":  boolPtr(true),
				"would_fail": boolPtr(false)},
			removeDVL:               true,
			upgradesDeployedVersion: true,
			latestVersion:           "0.9.0",
		},
		"command_NAME, known service_id with 1 command left to pass does upgrade deployed_version": {
			serviceID: "test",
			target:    stringPtr("command_ls -g"),
			commands: command.Slice{
				{"ls", "/root"},
				{"ls", "-g"}},
			commandFails: []*bool{
				boolPtr(false),
				boolPtr(true)},
			webhooks: webhook.Slice{
				"would_fail": testWebHook(true, "would_fail")},
			webhookFails: map[string]*bool{
				"would_fail": boolPtr(false)},
			removeDVL:               true,
			upgradesDeployedVersion: true,
			latestVersion:           "0.9.0",
		},
		"command_NAME, known service_id with multiple commands targeted individually (handle broadcast queue)": {
			serviceID: "test",
			commands: command.Slice{
				{"ls", "-h"},
				{"false", "2"},
				{"true"}},
			approveCommandsIndividually: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// backup Service
			var hadCommandSlice command.Slice
			var hadWebHookSlice webhook.Slice
			var hadDVL deployedver.Lookup
			var hadStatus svcstatus.Status
			svc := cfg.Service[tc.serviceID]
			if svc != nil {
				svc.Options.Active = tc.active
				{ // Copy Status
					hadStatus.SetApprovedVersion(svc.Status.GetApprovedVersion())
					hadStatus.SetDeployedVersion(svc.Status.GetDeployedVersion(), false)
					hadStatus.SetDeployedVersionTimestamp(svc.Status.GetDeployedVersionTimestamp())
					hadStatus.SetLatestVersion(svc.Status.GetLatestVersion(), false)
					hadStatus.SetLatestVersionTimestamp(svc.Status.GetLatestVersionTimestamp())
					hadStatus.SetLastQueried(svc.Status.GetLastQueried())
				}
				svc.Status.Init(
					0, len(tc.commands), len(tc.webhooks),
					&tc.serviceID,
					stringPtr("https://example.com"))
				svc.Status.SetLatestVersion(tc.latestVersion, false)
				hadDVL = *svc.DeployedVersionLookup
				if tc.removeDVL {
					svc.DeployedVersionLookup = nil
				}
				hadCommandSlice = svc.Command
				svc.Command = tc.commands
				svc.CommandController.Init(
					&svc.Status,
					&svc.Command,
					nil,
					stringPtr("10m"))
				if len(tc.commandFails) != 0 {
					for i := range tc.commandFails {
						if tc.commandFails[i] != nil {
							svc.CommandController.Failed.Set(i, *tc.commandFails[i])
						}
					}
				}
				hadWebHookSlice = svc.WebHook
				if tc.webhooks != nil {
					for i := range tc.webhooks {
						tc.webhooks[i].ServiceStatus = &svc.Status
					}
				}
				svc.WebHook = tc.webhooks
				svc.WebHook.Init(
					&svc.Status,
					&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
					nil,
					&svc.Options.Interval)
				if len(tc.webhookFails) != 0 {
					for key := range tc.webhookFails {
						svc.WebHook[key].Failed.Set(key, tc.webhookFails[key])
					}
				}
				// revert Service
				defer func() {
					{ // Status
						svc.Status.SetApprovedVersion(hadStatus.GetApprovedVersion())
						svc.Status.SetDeployedVersion(hadStatus.GetDeployedVersion(), false)
						svc.Status.SetDeployedVersionTimestamp(hadStatus.GetDeployedVersionTimestamp())
						svc.Status.SetLatestVersion(hadStatus.GetLatestVersion(), false)
						svc.Status.SetLatestVersionTimestamp(hadStatus.GetLatestVersionTimestamp())
						svc.Status.SetLastQueried(hadStatus.GetLastQueried())
					}
					svc.DeployedVersionLookup = &hadDVL
					svc.Command = hadCommandSlice
					svc.CommandController.Init(
						&svc.Status,
						&svc.Command,
						nil,
						stringPtr("10m"))
					svc.WebHook = hadWebHookSlice
				}()
			}
			ws := connectToWebSocket(t)
			defer ws.Close()
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN we send a message to the server (wsServiceInit)
			msg := api_type.WebSocketMessage{
				Version:     intPtr(1),
				Page:        "APPROVALS",
				Type:        "VERSION",
				Target:      tc.target,
				ServiceData: &api_type.ServiceSummary{ID: tc.serviceID}}
			if svc != nil {
				msg.ServiceData.Status = &api_type.Status{
					LatestVersion: svc.Status.GetLatestVersion(),
				}
			}
			sends := 1
			if tc.approveCommandsIndividually {
				sends = len(tc.commands)
			}
			for sends != 0 {
				sends--
				if tc.approveCommandsIndividually {
					msg.Target = stringPtr(fmt.Sprintf("command_%s", tc.commands[sends].String()))
				}
				data, _ := json.Marshal(msg)
				t.Log(string(data))
				if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
					t.Fatalf("error sending message\n%s",
						err.Error())
				}
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
			t.Logf("expecting %d messages", expecting)
			got := 0
			for expecting != 0 {
				message := seeIfMessage(t, ws)
				if message == nil {
					w.Close()
					out, _ := io.ReadAll(r)
					os.Stdout = stdout
					output := string(out)
					t.Log(output)
					t.Log(time.Now())
					t.Errorf("expecting %d more messages but got %v",
						expecting, message)
					return
				}
				messages[got] = *message
				raw, _ := json.Marshal(*message)
				t.Logf("\n%s\n", string(raw))
				got++
				expecting--
			}
			// extra message check
			message := seeIfMessage(t, ws)
			if message != nil {
				raw, _ := json.Marshal(*message)
				t.Fatalf("wasn't expecting another message but got one\n%#v\n%s",
					message, string(raw))
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
					messages[0].ServiceData.Status.ApprovedVersion != "SKIP_"+svc.Status.GetLatestVersion() {
					t.Errorf("LatestVersion %q wasn't skipped. approved is %q\ngot=%q",
						svc.Status.GetLatestVersion(),
						svc.Status.GetApprovedVersion(),
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
					t.Logf("message %d",
						i)
					receivedForAnAction := false
					for _, command := range tc.commands {
						if message.CommandData[command.String()] != nil {
							receivedForAnAction = true
							received = append(received, command.String())
							t.Logf("FOUND COMMAND %q - failed=%s", command.String(), stringifyPointer(message.CommandData[command.String()].Failed))
							break
						}
					}
					if !receivedForAnAction {
						for i := range tc.webhooks {
							if message.WebHookData[i] != nil {
								receivedForAnAction = true
								received = append(received, i)
								t.Logf("FOUND WEBHOOK %q - failed=%s", i, stringifyPointer(message.WebHookData[i].Failed))
								break
							}
						}
					}
					if !receivedForAnAction {
						//  IF we're expecting a message about approvedVersion
						if tc.upgradesApprovedVersion && message.Type == "VERSION" && message.SubType == "ACTION" {
							if message.ServiceData.Status.ApprovedVersion != svc.Status.GetLatestVersion() {
								t.Fatalf("expected approved version to be updated to latest version in the message\n%#v\napproved=%#v, latest=%#v",
									message, message.ServiceData.Status.ApprovedVersion, svc.Status.GetLatestVersion())
							}
							continue
						}
						if tc.upgradesDeployedVersion && message.Type == "VERSION" && message.SubType == "UPDATED" {
							if message.ServiceData.Status.DeployedVersion != svc.Status.GetLatestVersion() {
								t.Fatalf("expected deployed version to be updated to latest version in the message\n%#v\ndeployed=%#v, latest=%#v",
									message, message.ServiceData.Status.DeployedVersion, svc.Status.GetLatestVersion())
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

func TestWebSocketApprovalsACTIONS(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	testLogging("WARN", true)
	tests := map[string]struct {
		serviceID   string
		stdoutRegex string
		bodyRegex   string
		commands    command.Slice
		webhooks    webhook.Slice
	}{
		"service_id=unknown": {
			serviceID:   "unknown?",
			stdoutRegex: "service not found",
		},
		"service_id=nil": {
			serviceID:   "",
			stdoutRegex: "service_data.id not provided",
		},
		"known service_id, 0 command, 0 webhooks,": {
			serviceID:   "test",
			stdoutRegex: `DEBUG:[^:]+, READ \{[^}]+\}\}\}\nVERBOSE:[^V]+VERBOSE:[^"]+$`,
			// DEBUG: WebSocket (127.0.0.1), READ {"version":1,"page":"APPROVALS","type":"ACTIONS","service_data":{"id":"test","status":{}}}
			// VERBOSE: wsCommand (127.0.0.1), -\n
			// VERBOSE: wsWebHook (127.0.0.1), -\n
		},
		"known service_id, 1 command, 0 webhooks,": {
			serviceID: "test",
			commands: command.Slice{
				testCommand(true)},
		},
		"known service_id, 2 command, 0 webhooks,": {
			serviceID: "test",
			commands: command.Slice{
				testCommand(true), testCommand(false)},
		},
		"known service_id, 0 command, 1 webhooks,": {
			serviceID: "test",
			webhooks: webhook.Slice{
				"fail0": testWebHook(true, "fail0")},
		},
		"known service_id, 0 command, 2 webhooks,": {
			serviceID: "test",
			webhooks: webhook.Slice{
				"fail0": testWebHook(true, "fail0"),
				"pass0": testWebHook(false, "pass0")},
		},
		"known service_id, 2 command, 2 webhooks,": {
			serviceID: "test",
			commands: command.Slice{
				testCommand(true), testCommand(false)},
			webhooks: webhook.Slice{
				"fail0": testWebHook(true, "fail0"),
				"pass0": testWebHook(false, "pass0")},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// backup Service
			var hadCommandSlice command.Slice
			var hadWebHookSlice webhook.Slice
			var hadStatus svcstatus.Status
			ws := connectToWebSocket(t)
			defer ws.Close()
			svc := cfg.Service[tc.serviceID]
			if svc != nil {
				{ // Copy Status
					hadStatus.SetApprovedVersion(svc.Status.GetApprovedVersion())
					hadStatus.SetDeployedVersion(svc.Status.GetDeployedVersion(), false)
					hadStatus.SetDeployedVersionTimestamp(svc.Status.GetDeployedVersionTimestamp())
					hadStatus.SetLatestVersion(svc.Status.GetLatestVersion(), false)
					hadStatus.SetLatestVersionTimestamp(svc.Status.GetLatestVersionTimestamp())
					hadStatus.SetLastQueried(svc.Status.GetLastQueried())
				}
				hadCommandSlice = svc.Command
				svc.Command = tc.commands
				if tc.commands == nil {
					svc.CommandController = nil
				} else {
					svc.CommandController = &command.Controller{}
					svc.CommandController.Init(
						&svc.Status,
						&svc.Command,
						nil,
						stringPtr("10m"))
				}
				hadWebHookSlice = svc.WebHook
				svc.WebHook = tc.webhooks
				svc.WebHook.Init(
					&svc.Status,
					&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
					nil,
					&svc.Options.Interval)
				defer func() {
					{ // Status
						svc.Status.SetApprovedVersion(hadStatus.GetApprovedVersion())
						svc.Status.SetDeployedVersion(hadStatus.GetDeployedVersion(), false)
						svc.Status.SetDeployedVersionTimestamp(hadStatus.GetDeployedVersionTimestamp())
						svc.Status.SetLatestVersion(hadStatus.GetLatestVersion(), false)
						svc.Status.SetLatestVersionTimestamp(hadStatus.GetLatestVersionTimestamp())
						svc.Status.SetLastQueried(hadStatus.GetLastQueried())
					}
					svc.Command = hadCommandSlice
					svc.CommandController.Init(
						&svc.Status,
						&svc.Command,
						nil,
						stringPtr("10m"))
					svc.WebHook = hadWebHookSlice
				}()
			}
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN we send a message to the server (wsServiceInit)
			msg := api_type.WebSocketMessage{Version: intPtr(1), Page: "APPROVALS", Type: "ACTIONS",
				ServiceData: &api_type.ServiceSummary{ID: tc.serviceID}}
			if svc != nil {
				msg.ServiceData.Status = &api_type.Status{
					LatestVersion: svc.Status.GetLatestVersion(),
				}
			}
			data, _ := json.Marshal(msg)
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Errorf("error sending message\n%s",
					err.Error())
			}

			// THEN we get the expected response
			message := seeIfMessage(t, ws)
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
				// don't want a message
				if message != nil {
					raw, _ := json.Marshal(*message)
					t.Errorf("wasn't expecting another message but got one\n%#v",
						raw)
				}
				return
			}
			expectingC := len(tc.commands) != 0
			expectingWH := len(tc.webhooks) != 0
			for expectingC || expectingWH {
				// didn't get a message
				if message == nil {
					t.Fatalf("expecting message but got %#v\nexpecting commands=%t, expecting webhook=%t",
						message, expectingC, expectingWH)
				}
				if message.CommandData != nil {
					expectingC = false
				} else if message.WebHookData != nil {
					expectingWH = false
				}
				message = seeIfMessage(t, ws)
			}

			// extra message check
			if message != nil {
				raw, _ := json.Marshal(*message)
				t.Fatalf("wasn't expecting another message but got one\n%#v",
					raw)
			}
		})
	}
}

func TestWebSocketRuntimeBuildINIT(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to get the Runtime+Build info
	var (
		msgVersion int    = 1
		msgPage    string = "RUNTIME_BUILD"
		msgType    string = "INIT"
	)
	msg := api_type.WebSocketMessage{
		Version: &msgVersion,
		Page:    msgPage,
		Type:    msgType,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}

	// THEN it passes and we only receive a response with the WebHooks
	message := seeIfMessage(t, ws)
	if message == nil {
		t.Fatal("expecting another message but didn't get one")
	}
	raw, _ := json.Marshal(*message)
	if message.Page != msgPage {
		t.Fatalf("Received a response for Page %q. Expected %q",
			message.Page, msgPage)
	}
	if message.InfoData == nil {
		t.Fatalf("Didn't get any InfoData in the message\n%#v",
			message)
	}
	if message.InfoData.Build.GoVersion != util.GoVersion {
		t.Errorf("Expected Info.Build.GoVersion to be %q, got %q\n%s",
			util.GoVersion, message.InfoData.Build.GoVersion, raw)
	}
}

func TestWebSocketFlagsINIT(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to get the Commands+WebHooks
	var (
		msgVersion int    = 1
		msgPage    string = "FLAGS"
		msgType    string = "INIT"
	)
	msg := api_type.WebSocketMessage{
		Version: &msgVersion,
		Page:    msgPage,
		Type:    msgType,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}

	// THEN it passes and we only receive a response with the WebHooks
	message := seeIfMessage(t, ws)
	if message == nil {
		t.Fatal("expecting another message but didn't get one")
	}
	raw, _ := json.Marshal(*message)
	if message.Page != msgPage {
		t.Fatalf("Received a response for Page %q. Expected %q",
			message.Page, msgPage)
	}
	if message.FlagsData == nil {
		t.Fatalf("Didn't get any FlagsData in the message\n%#v",
			message)
	}
	if message.FlagsData.LogLevel != cfg.Settings.GetLogLevel() {
		t.Errorf("Expected FlagsData.LogLevel to be %q, got %q\n%s",
			cfg.Settings.GetLogLevel(), message.FlagsData.LogLevel, raw)
	}
}

func TestWebSocketConfigINIT(t *testing.T) {
	// GIVEN we have a Config
	tests := map[string]struct {
		nilService       bool
		nilServiceDVL    bool
		nilServiceURLC   bool
		nilServiceNotify bool
		nilServiceC      bool
		nilServiceWH     bool
	}{
		"no Service's": {
			nilService: true},
		"no Service DeployedVersionLookup": {
			nilServiceDVL: true},
		"no Service URLCommands": {
			nilServiceURLC: true},
		"no Service Notify": {
			nilServiceNotify: true},
		"no Service Command": {
			nilServiceC: true},
		"no Service WebHook": {
			nilServiceWH: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// backup changes
			hadService := make(service.Slice, len(cfg.Service))
			for i := range cfg.Service {
				s := cfg.Service[i]
				svc := service.Service{
					ID:            s.ID,
					Comment:       s.Comment,
					Options:       s.Options,
					LatestVersion: s.LatestVersion,
					Dashboard:     s.Dashboard,
					Status:        svcstatus.Status{},
					Defaults:      s.Defaults,
					HardDefaults:  s.HardDefaults,
					Type:          s.Type,
				}
				hadService[i] = &svc
				dvl := *cfg.Service[i].DeployedVersionLookup
				hadService[i].DeployedVersionLookup = &dvl
				if tc.nilServiceDVL {
					cfg.Service[i].DeployedVersionLookup = nil
				}
				urlc := cfg.Service[i].LatestVersion.URLCommands
				hadService[i].LatestVersion.URLCommands = urlc
				if tc.nilServiceURLC {
					cfg.Service[i].LatestVersion.URLCommands = nil
				}
				notify := cfg.Service[i].Notify
				hadService[i].Notify = notify
				if tc.nilServiceNotify {
					cfg.Service[i].Notify = nil
				}
				wh := cfg.Service[i].WebHook
				hadService[i].WebHook = wh
				if tc.nilServiceWH {
					cfg.Service[i].WebHook = nil
				}
				command := cfg.Service[i].Command
				hadService[i].Command = command
				if tc.nilServiceC {
					cfg.Service[i].Command = nil
					cfg.Service[i].CommandController = nil
				} else {
					cfg.Service[i].CommandController.Init(
						&svc.Status,
						&svc.Command,
						nil,
						nil)
				}
				// restore possible changes
				defer func() {
					cfg.Service = hadService
				}()
			}
			if tc.nilService {
				cfg.Service = nil
			}
			ws := connectToWebSocket(t)
			defer ws.Close()

			// WHEN we send a message to get the Config
			var (
				msgVersion int    = 1
				msgPage    string = "CONFIG"
				msgType    string = "INIT"
			)
			msg := api_type.WebSocketMessage{
				Version: &msgVersion,
				Page:    msgPage,
				Type:    msgType,
			}
			data, _ := json.Marshal(msg)
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Errorf("%v",
					err)
			}

			// THEN it passes and we receive the config
			message := seeIfMessage(t, ws)
			if message == nil {
				t.Error("SETTINGS - expecting another message but didn't get one")
			}
			raw, _ := json.Marshal(*message)
			{ // SETTINGS
				wantedType := "SETTINGS"
				if message.Page != msgPage ||
					message.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						message.Page, message.Type, msgPage, wantedType)
				} else {
					if message.ConfigData == nil {
						t.Errorf("Didn't get any ConfigData in the message\n%#v",
							message)
					} else {
						if message.ConfigData == nil ||
							message.ConfigData.Settings == nil ||
							message.ConfigData.Settings.Log.Level == nil {
							t.Errorf("Didn't receive ConfigData.Settings.Log.Level from\n%s",
								raw)
						} else if *message.ConfigData.Settings.Log.Level != cfg.Settings.GetLogLevel() {
							t.Errorf("Expected ConfigData.Settings.Log.Level to be %q, got %q\n%s",
								cfg.Settings.GetLogLevel(), *message.ConfigData.Settings.Log.Level, raw)
						}
					}
				}
			}
			message = seeIfMessage(t, ws)
			if message == nil {
				t.Error("DEFAULTS - expecting another message but didn't get one")
			}
			raw, _ = json.Marshal(*message)
			{ // DEFAULTS
				wantedType := "DEFAULTS"
				if message.Page != msgPage ||
					message.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						message.Page, message.Type, msgPage, wantedType)
				} else {
					if message.ConfigData == nil ||
						message.ConfigData.Defaults == nil ||
						message.ConfigData.Defaults.Service.Options.Interval == "" {
						t.Errorf("Didn't receive ConfigData.Defaults.Service.Interval from\n%#v",
							message)
					} else if message.ConfigData.Defaults.Service.Options.Interval != cfg.Defaults.Service.Options.Interval {
						t.Errorf("Expected ConfigData.Defaults..Service.Options.Interval to be %q, got %q\n%s",
							cfg.Defaults.Service.Options.Interval, message.ConfigData.Defaults.Service.Options.Interval, raw)
					}
				}
			}
			message = seeIfMessage(t, ws)
			if message == nil {
				t.Error("NOTIFY - expecting another message but didn't get one")
			}
			raw, _ = json.Marshal(*message)
			{ // NOTIFY
				wantedType := "NOTIFY"
				if message.Page != msgPage ||
					message.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						message.Page, message.Type, msgPage, wantedType)
				} else {
					if message.ConfigData == nil ||
						message.ConfigData.Notify == nil ||
						(*message.ConfigData.Notify)["discord"] == nil {
						t.Errorf("Didn't receive ConfigData.Notify.discord from\n%#v",
							message)
					} else if (*message.ConfigData.Notify)["discord"].Options["message"] != cfg.Notify["discord"].Options["message"] {
						t.Errorf("Expected ConfigData.Notify.discord.Options.message to be %q, got %q\n%s",
							cfg.Notify["discord"].Options["message"], (*message.ConfigData.Notify)["discord"].Options["message"], raw)
					}
				}
			}
			message = seeIfMessage(t, ws)
			if message == nil {
				t.Error("WEBHOOK - expecting another message but didn't get one")
			}
			raw, _ = json.Marshal(*message)
			{ // WEBHOOK
				wantedType := "WEBHOOK"
				if message.Page != msgPage ||
					message.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						message.Page, message.Type, msgPage, wantedType)
				} else {
					if message.ConfigData == nil ||
						message.ConfigData.WebHook == nil ||
						(*message.ConfigData.WebHook)["test"] == nil {
						t.Errorf("Didn't receive ConfigData.WebHook.test from\n%#v",
							message)
					} else if *(*message.ConfigData.WebHook)["test"].URL != cfg.WebHook["test"].URL {
						t.Errorf("Expected ConfigData.WebHook.test.URL to be %q, got %q\n%s",
							cfg.WebHook["test"].URL, *(*message.ConfigData.WebHook)["test"].URL, raw)
					}
				}
			}
			// SERVICE
			message = seeIfMessage(t, ws)
			if message == nil {
				t.Error("SERVICE - expecting another message but didn't get one")
			}
			raw, _ = json.Marshal(*message)
			if tc.nilService {
				if message.ServiceData != nil {
					t.Errorf("expecting ServiceData to be nil, not \n%#v",
						message.ServiceData)
				}
			} else {
				receivedTestService := (*message.ConfigData.Service)["test"]
				cfgTestService := cfg.Service["test"]
				// service
				if receivedTestService.Comment != cfgTestService.Comment {
					t.Errorf("ConfigData.Service.test.Comment should've been %q, got %q",
						cfgTestService.Comment, receivedTestService.Comment)
				}
				if receivedTestService.LatestVersion.URL != cfgTestService.LatestVersion.URL {
					t.Errorf("ConfigData.Service.test.LatestVersion.URL should've been %q, got %q",
						cfgTestService.LatestVersion.URL, receivedTestService.LatestVersion.URL)
				}
				if receivedTestService.Dashboard.WebURL != cfgTestService.Dashboard.WebURL {
					t.Errorf("ConfigData.Service.test.Dashboard.WebURL should've been %q, got %q",
						cfgTestService.Dashboard.WebURL, receivedTestService.Dashboard.WebURL)
				}
				if receivedTestService.LatestVersion.Require.RegexContent != cfgTestService.LatestVersion.Require.RegexContent {
					t.Errorf("ConfigData.Service.test.LatestVersion.Require.RegexContent should've been %q, got %q",
						cfgTestService.LatestVersion.Require.RegexContent, receivedTestService.LatestVersion.Require.RegexContent)
				}
				if receivedTestService.LatestVersion.Require.RegexVersion != cfgTestService.LatestVersion.Require.RegexVersion {
					t.Errorf("ConfigData.Service.test.LatestVersion.Require.RegexVersion should've been %q, got %q",
						cfgTestService.LatestVersion.Require.RegexVersion, receivedTestService.LatestVersion.Require.RegexVersion)
				}
				if *receivedTestService.Dashboard.AutoApprove != *cfgTestService.Dashboard.AutoApprove {
					t.Errorf("ConfigData.Service.test.Dashboard.AutoApprove should've been %t, got %t",
						*cfgTestService.Dashboard.AutoApprove, *receivedTestService.Dashboard.AutoApprove)
				}
				// deployed version lookup
				if tc.nilServiceDVL {
					if receivedTestService.DeployedVersionLookup != nil {
						t.Errorf("expecting DeployedVersionLookup to be nil, not \n%#v",
							receivedTestService.DeployedVersionLookup)
					}
				} else {
					// url
					if receivedTestService.DeployedVersionLookup.URL != cfgTestService.DeployedVersionLookup.URL {
						t.Errorf("ConfigData.Service.test.DeployedVersionLookup.URL should've been %q, got %q",
							cfgTestService.DeployedVersionLookup.URL, receivedTestService.DeployedVersionLookup.URL)
					}
					// basic auth
					if receivedTestService.DeployedVersionLookup.BasicAuth.Password != "<secret>" {
						t.Errorf("ConfigData.Service.test.DeployedVersionLookup.BasicAuth.Password should've been %q, got %q",
							"<secret>", receivedTestService.DeployedVersionLookup.URL)
					}
					// headers
					if receivedTestService.DeployedVersionLookup.Headers[0].Value != "<secret>" {
						t.Errorf("ConfigData.Service.test.DeployedVersionLookup.Headers[0].Values should've been %q, got %q",
							"<secret>", receivedTestService.DeployedVersionLookup.Headers[0].Value)
					}
					// json
					if receivedTestService.DeployedVersionLookup.JSON != cfgTestService.DeployedVersionLookup.JSON {
						t.Errorf("ConfigData.Service.test.DeployedVersionLookup.JSON should've been %q, got %q",
							cfgTestService.DeployedVersionLookup.JSON, receivedTestService.DeployedVersionLookup.JSON)
					}
					// regex
					if receivedTestService.DeployedVersionLookup.Regex != cfgTestService.DeployedVersionLookup.Regex {
						t.Errorf("ConfigData.Service.test.DeployedVersionLookup.Regex should've been %q, got %q",
							cfgTestService.DeployedVersionLookup.Regex, receivedTestService.DeployedVersionLookup.Regex)
					}
				}
				// url commands
				if tc.nilServiceURLC {
					if receivedTestService.LatestVersion.URLCommands == nil || len(*receivedTestService.LatestVersion.URLCommands) != 0 {
						t.Errorf("expecting URLCommands to be 0, not \n%#v",
							len(*receivedTestService.LatestVersion.URLCommands))
					}
				} else {
					if receivedTestService.LatestVersion.URLCommands == nil {
						t.Errorf("ConfigData.Service.test.URLCommands should've been %#v, got %#v",
							cfgTestService.LatestVersion.URLCommands, receivedTestService.LatestVersion.URLCommands)
					} else if util.DefaultIfNil((*receivedTestService.LatestVersion.URLCommands)[0].Regex) != util.DefaultIfNil(cfgTestService.LatestVersion.URLCommands[0].Regex) {
						t.Errorf("ConfigData.Service.test.URLCommands[0].Regex should've been %q, got %q",
							util.DefaultIfNil(cfgTestService.LatestVersion.URLCommands[0].Regex), util.DefaultIfNil((*receivedTestService.LatestVersion.URLCommands)[0].Regex))
					}
				}
				// notify
				if tc.nilServiceNotify {
					if receivedTestService.Notify != nil && len(*receivedTestService.Notify) != 0 {
						t.Errorf("expecting Notify to be nil, not \n%#v",
							receivedTestService.Notify)
					}
				} else {
					if receivedTestService.Notify == nil {
						t.Errorf("ConfigData.Service.test.Notify should've been %#v, got %#v",
							cfgTestService.Notify, receivedTestService.Notify)
					} else if (*receivedTestService.Notify)["test"].Options["message"] != cfgTestService.Notify["test"].Options["message"] {
						t.Errorf("ConfigData.Service.test.Notify.test.Options.message should've been %q, got %q",
							cfgTestService.Notify["test"].Options["message"], (*receivedTestService.Notify)["test"].Options["message"])
					}
				}
				// webhook
				if tc.nilServiceWH {
					if receivedTestService.WebHook != nil && len(*receivedTestService.WebHook) != 0 {
						t.Errorf("expecting WebHook to be nil, not \n%#v",
							receivedTestService.WebHook)
					}
				} else {
					if receivedTestService.WebHook == nil {
						t.Errorf("ConfigData.Service.test.WebHook should've been %#v, got %#v",
							cfgTestService.WebHook, receivedTestService.WebHook)
					} else if *(*receivedTestService.WebHook)["test"].URL != cfgTestService.WebHook["test"].URL {
						t.Errorf("ConfigData.Service.test.WebHook.test.URL should've been %q, got %q",
							cfgTestService.WebHook["test"].URL, *(*receivedTestService.WebHook)["test"].URL)
					}
				}
				// command
				if tc.nilServiceC {
					if receivedTestService.Command != nil && len(*receivedTestService.Command) != 0 {
						t.Errorf("expecting Command to be nil, not \n%#v",
							receivedTestService.Command)
					}
				} else {
					if receivedTestService.Command == nil {
						t.Errorf("ConfigData.Service.test.Command should've been %#v, got %#v",
							cfgTestService.Command, receivedTestService.Command)
					} else {
						got := strings.Join((*receivedTestService.Command)[0], " ")
						if got != cfgTestService.Command[0].String() {
							t.Errorf("ConfigData.Service.test.Command[0] should've been %q, got %q",
								cfgTestService.Command[0].String(), got)
						}
					}
				}
			}
			message = seeIfMessage(t, ws)
			if message != nil {
				raw, _ := json.Marshal(*message)
				t.Errorf("wasn't expecting another message but got one\n%s",
					raw)
			}
		})
	}
}
