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
	"io/ioutil"
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
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	api_types "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

var router *mux.Router
var port *string
var cfg config.Config

func TestMain(m *testing.M) {
	// GIVEN a valid config with a Service
	cfg = testConfig()
	jLog = utils.NewJLog("WARN", false)
	port = cfg.Settings.Web.ListenPort

	// WHEN the Router is fetched for this Config
	router = newWebUI(&cfg)
	go http.ListenAndServe("localhost:"+*port, router)

	// THEN Web UI is accessible for the tests
	code := m.Run()
	os.Exit(code)
}

func TestMainWithRoutePrefix(t *testing.T) {
	// GIVEN a valid config with a Service
	config := testConfig()
	*config.Settings.Web.RoutePrefix = "/test"

	// WHEN the Web UI is started with this Config
	go Run(&config, utils.NewJLog("WARN", false))
	time.Sleep(50 * time.Millisecond)

	// THEN Web UI is accessible
	url := fmt.Sprintf("http://localhost:%s%s/metrics", *config.Settings.Web.ListenPort, *config.Settings.Web.RoutePrefix)
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
		"/approvals":      {path: "/approvals"},
		"/metrics":        {path: "/metrics", bodyRegex: "go_gc_duration_"},
		"/api/v1/version": {path: "/api/v1/version", bodyRegex: fmt.Sprintf(`"goVersion":"%s"`, utils.GoVersion)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN we make a request to path
			req, _ := http.NewRequest("GET", tc.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			// THEN we get a Status OK
			if response.Code != http.StatusOK {
				t.Errorf("%s:\nExpected a 200, got %d",
					name, response.Code)
			}
			if tc.bodyRegex != "" {
				body := response.Body.String()
				re := regexp.MustCompile(tc.bodyRegex)
				match := re.MatchString(body)
				if !match {
					t.Errorf("%s:\nexpected %q in body\ngot: %q",
						name, tc.bodyRegex, response.Body.String())
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
	return ws
}

// seeIfMessage will try and get a message from the WebSocket
// if it receives no message withing 100ms, it will give up and return nil
func seeIfMessage(t *testing.T, ws *websocket.Conn) *[]byte {
	receiver := make(chan []byte, 4)
	go func() {
		_, p, _ := ws.ReadMessage()
		receiver <- p
	}()
	for i := 0; i < 50; i++ {
		time.Sleep(10 * time.Millisecond)
		if len(receiver) != 0 {
			break
		}
	}
	if len(receiver) == 0 {
		return nil
	}
	got := <-receiver
	return &got
}

func TestWebSocket(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	tests := map[string]struct {
		msg         string
		stdoutRegex string
		bodyRegex   string
	}{
		"no version":                   {msg: `{"key": "value"}`, stdoutRegex: "^$"},
		"invalid JSON":                 {msg: `{"version": 1, "key": "value"`, stdoutRegex: "missing/invalid version key"},
		"unknown page":                 {msg: `{"version": 1, "page": "", "type": "value"}`, stdoutRegex: "Unknown PAGE"},
		"APPROVALS - unknown type":     {msg: `{"version": 1, "page": "APPROVALS", "type": "value"}`, stdoutRegex: "Unknown APPROVALS Type"},
		"RUNTIME_BUILD - unknown type": {msg: `{"version": 1, "page": "RUNTIME_BUILD", "type": "value"}`, stdoutRegex: "Unknown RUNTIME_BUILD Type"},
		"FLAGS - unknown type":         {msg: `{"version": 1, "page": "FLAGS", "type": "value"}`, stdoutRegex: "Unknown FLAGS Type"},
		"CONFIG - unknown type":        {msg: `{"version": 1, "page": "CONFIG", "type": "value"}`, stdoutRegex: "Unknown CONFIG Type"},
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
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			re := regexp.MustCompile(tc.stdoutRegex)
			match := re.MatchString(output)
			if !match {
				t.Errorf("%s:\nmatch on %q not found in\n%q",
					name, tc.stdoutRegex, output)
			}
		})
	}
}

func TestWebSocketApprovalsINIT(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	tests := map[string]struct {
		order []string
	}{
		"INIT":                                  {order: *cfg.Order},
		"INIT with nil Service in config.Order": {order: append(*cfg.Order, "nilService")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			order := cfg.Order
			cfg.Order = &tc.order
			ws := connectToWebSocket(t)
			defer ws.Close()

			// WHEN we send a message to the server (wsService)
			msg := api_types.WebSocketMessage{Version: intPtr(1), Page: stringPtr("APPROVALS"), Type: stringPtr("INIT")}
			data, _ := json.Marshal(msg)
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Fatalf("%s:\nerror sending message\n%s",
					name, err.Error())
			}

			// THEN we get the expected responses
			// ORDERING
			p := seeIfMessage(t, ws)
			cfg.Order = order
			var receivedMsg api_types.WebSocketMessage
			json.Unmarshal(*p, &receivedMsg)
			if receivedMsg.Order == nil || len(*receivedMsg.Order) != len(tc.order) {
				t.Fatalf("%s:\nwant order=%#v\ngot  order=%#v",
					name, tc.order, *receivedMsg.Order)
			}
			for i := range tc.order {
				if tc.order[i] != (*receivedMsg.Order)[i] {
					t.Fatalf("%s:\nwant order=%#v\ngot  order=%#v",
						name, tc.order, *receivedMsg.Order)
				}
			}
			// SERVICE
			receivedOrder := *receivedMsg.Order
			for _, key := range receivedOrder {
				if cfg.Service[key] == nil {
					continue
				}
				p := seeIfMessage(t, ws)
				if p == nil {
					t.Fatalf("%s:\nexpecting another message but didn't get one",
						name)
				}
				receivedMsg = api_types.WebSocketMessage{}
				json.Unmarshal(*p, &receivedMsg)
				if receivedMsg.ServiceData == nil {
					t.Errorf("%s:\nbad message, didn't contain ServiceData for %q",
						name, key)
				}
			}
			p = seeIfMessage(t, ws)
			if p != nil {
				receivedMsg = api_types.WebSocketMessage{}
				json.Unmarshal(*p, &receivedMsg)
				t.Fatalf("%s:\nwasn't expecting another message but got one\n%#v",
					name, receivedMsg)
			}
		})
	}
}

func TestWebSocketApprovalsVERSION(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	var logInitCommands *command.Controller
	logInitCommands.Init(jLog, nil, nil, nil, nil, nil)
	var logInitWebHooks *webhook.Slice
	logInitWebHooks.Init(jLog, nil, nil, nil, nil, nil, nil, nil)
	tests := map[string]struct {
		serviceID                   *string
		target                      *string
		stdoutRegex                 string
		bodyRegex                   string
		commands                    *command.Slice
		commandFails                []*bool
		webhooks                    *webhook.Slice
		webhookFails                map[string]*bool
		removeDVL                   bool
		upgradesApprovedVersion     bool
		upgradesDeployedVersion     bool
		approveCommandsIndividually bool
	}{
		"target=ARGUS_SKIP service_id=known": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_SKIP")},
		"target=ARGUS_SKIP service_id=unknown": {serviceID: stringPtr("unknown?"), target: stringPtr("ARGUS_SKIP"),
			stdoutRegex: "service not found"},
		"target=ARGUS_SKIP service_id=nil": {serviceID: nil, target: stringPtr("ARGUS_SKIP"),
			stdoutRegex: "service_data.id not provided"},
		"target=nil, service_id=known": {serviceID: stringPtr("test"), target: nil,
			stdoutRegex: "target for command/webhook not provided"},
		"target=ARGUS_ALL, service_id=known - service has no commands/webhooks": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"), stdoutRegex: "does not have any commands/webhooks to approve"},
		"target=ARGUS_ALL, service_id=known - service has command": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			commands: &command.Slice{{"false"}}},
		"target=ARGUS_ALL, service_id=known - service has webhook": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			webhooks: &webhook.Slice{"0": testWebHookFail("0")}},
		"target=ARGUS_ALL, service_id=known - service has multiple webhooks": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			webhooks: &webhook.Slice{"0": testWebHookFail("0"), "1": testWebHookFail("1")}},
		"target=ARGUS_ALL, service_id=known - service has multiple commands": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			commands: &command.Slice{{"ls"}, {"false"}}},
		"target=ARGUS_ALL, service_id=known - service with dvl and command and webhook that pass upgrades approved_version": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			commands: &command.Slice{{"ls"}}, webhooks: &webhook.Slice{"0": testWebHookPass("0")}, upgradesApprovedVersion: true},
		"target=ARGUS_ALL, service_id=known - service with command and webhook that pass upgrades deployed_version": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			commands: &command.Slice{{"ls"}}, webhooks: &webhook.Slice{"0": testWebHookPass("0")}, removeDVL: true, upgradesDeployedVersion: true},
		"target=ARGUS_ALL, service_id=known - service with passing command and failing webhook doesn't upgrade any versions": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			commands: &command.Slice{{"ls"}}, webhooks: &webhook.Slice{"0": testWebHookFail("0")}},
		"target=ARGUS_ALL, service_id=known - service with failing command and passing webhook doesn't upgrade any versions": {serviceID: stringPtr("test"), target: stringPtr("ARGUS_ALL"),
			commands: &command.Slice{{"fail"}}, webhooks: &webhook.Slice{"0": testWebHookPass("0")}},
		"target=webhook_only_failed, service_id=known - service with 1 webhook left to pass does upgrade deployed_version": {serviceID: stringPtr("test"), target: stringPtr("webhook_only_failed"),
			commands: &command.Slice{{"ls"}}, commandFails: []*bool{boolPtr(false)},
			webhooks: &webhook.Slice{"only_failed": testWebHookPass("only_failed"), "would_fail": testWebHookFail("would_fail")}, webhookFails: map[string]*bool{"only_failed": boolPtr(true), "would_fail": boolPtr(false)},
			removeDVL: true, upgradesDeployedVersion: true},
		"target=command_ls, service_id=known - service with 1 command left to pass does upgrade deployed_version": {serviceID: stringPtr("test"), target: stringPtr("command_ls"),
			commands: &command.Slice{{"ls", "/root"}, {"ls"}}, commandFails: []*bool{boolPtr(false), boolPtr(true)},
			webhooks: &webhook.Slice{"would_fail": testWebHookFail("would_fail")}, webhookFails: map[string]*bool{"would_fail": boolPtr(false)},
			removeDVL: true, upgradesDeployedVersion: true},
		"target=command_x, service_id=known - service has multiple commands targetted individually (handle broadcast queue)": {serviceID: stringPtr("test"),
			commands: &command.Slice{{"ls"}, {"false"}, {"true"}}, approveCommandsIndividually: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svcLog := service.Service{ID: stringPtr("service_for_log")}
			svcLog.Init(jLog, &service.Service{}, &service.Service{})
			// backup Service
			var hadCommandSlice command.Slice
			var hadWebHookSlice webhook.Slice
			var hadDVL service.DeployedVersionLookup
			var hadStatus service_status.Status
			if cfg.Service[utils.DefaultIfNil(tc.serviceID)] != nil {
				hadStatus = *cfg.Service[*tc.serviceID].Status
				hadDVL = *cfg.Service[*tc.serviceID].DeployedVersionLookup
				if tc.removeDVL {
					cfg.Service[*tc.serviceID].DeployedVersionLookup = nil
				}
				hadCommandSlice = *cfg.Service[*tc.serviceID].Command
				cfg.Service[*tc.serviceID].Command = tc.commands
				cfg.Service[*tc.serviceID].CommandController.Init(jLog, cfg.Service[*tc.serviceID].ID, nil, cfg.Service[*tc.serviceID].Command, nil, stringPtr("10m"))
				if len(tc.commandFails) != 0 {
					for i := range tc.commandFails {
						cfg.Service[*tc.serviceID].CommandController.Failed[i] = tc.commandFails[i]
					}
				}
				hadWebHookSlice = *cfg.Service[*tc.serviceID].WebHook
				if tc.webhooks != nil {
					for i := range *tc.webhooks {
						(*tc.webhooks)[i].Announce = cfg.Service[*tc.serviceID].Announce
					}
				}
				cfg.Service[*tc.serviceID].WebHook = tc.webhooks
				cfg.Service[*tc.serviceID].WebHook.Init(jLog, cfg.Service[*tc.serviceID].ID, cfg.Service[*tc.serviceID].Status, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{}, nil, cfg.Service[*tc.serviceID].Interval)
				if len(tc.webhookFails) != 0 {
					for key := range tc.webhookFails {
						(*cfg.Service[*tc.serviceID].WebHook)[key].Failed = tc.webhookFails[key]
					}
				}
				// revert Service
				defer func() {
					cfg.Service[*tc.serviceID].Status = &hadStatus
					cfg.Service[*tc.serviceID].DeployedVersionLookup = &hadDVL
					cfg.Service[*tc.serviceID].Command = &hadCommandSlice
					cfg.Service[*tc.serviceID].CommandController.Init(jLog, cfg.Service[*tc.serviceID].ID, nil, cfg.Service[*tc.serviceID].Command, nil, stringPtr("10m"))
					cfg.Service[*tc.serviceID].WebHook = &hadWebHookSlice
				}()
			}
			var receivedMsg api_types.WebSocketMessage
			ws := connectToWebSocket(t)
			defer ws.Close()
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN we send a message to the server (wsService)
			msg := api_types.WebSocketMessage{Version: intPtr(1), Page: stringPtr("APPROVALS"), Type: stringPtr("VERSION"),
				Target: tc.target, ServiceData: &api_types.ServiceSummary{ID: tc.serviceID}}
			if cfg.Service[utils.DefaultIfNil(tc.serviceID)] != nil {
				msg.ServiceData.Status = &api_types.Status{
					LatestVersion: cfg.Service[*tc.serviceID].Status.LatestVersion,
				}
			}
			sends := 1
			if tc.approveCommandsIndividually {
				sends = len(*tc.commands)
			}
			for sends != 0 {
				sends--
				if tc.approveCommandsIndividually {
					msg.Target = stringPtr(fmt.Sprintf("command_%s", (*tc.commands)[sends].String()))
				}
				data, _ := json.Marshal(msg)
				if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
					t.Errorf("%s:\nerror sending message\n%s",
						name, err.Error())
				}
				time.Sleep(10 * time.Microsecond)
			}

			// THEN we get the expected response
			p := seeIfMessage(t, ws)
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			// stdout finishes
			if tc.stdoutRegex != "" {
				re := regexp.MustCompile(tc.stdoutRegex)
				match := re.MatchString(output)
				if !match {
					t.Errorf("%s:\nmatch on %q not found in\n%q",
						name, tc.stdoutRegex, output)
				}
				// don't want a message
				if p != nil {
					receivedMsg = api_types.WebSocketMessage{}
					json.Unmarshal(*p, &receivedMsg)
					t.Errorf("%s:\nwasn't expecting another message but got one\n%#v",
						name, receivedMsg)
				}
				return
			}
			// didn't get a message
			if p == nil {
				t.Fatalf("%s:\nexpecting message but got %#v",
					name, p)
			} else {
				receivedMsg = api_types.WebSocketMessage{}
				json.Unmarshal(*p, &receivedMsg)
			}

			// Check version was skipped
			if utils.DefaultIfNil(tc.target) == "ARGUS_SKIP" {
				if receivedMsg.ServiceData.Status.ApprovedVersion != "SKIP_"+cfg.Service[*tc.serviceID].Status.LatestVersion {
					t.Errorf("%s:\nLatestVersion %q wasn't skipped. approved is %q\ngot=%q",
						name, cfg.Service[*tc.serviceID].Status.LatestVersion, cfg.Service[*tc.serviceID].Status.ApprovedVersion, receivedMsg.ServiceData.Status.ApprovedVersion)
				}
			} else {
				// expecting = commands + webhooks that have not failed=false
				expecting := 0
				if tc.commands != nil {
					expecting += len(*tc.commands)
					if tc.commandFails != nil {
						for i := range tc.commandFails {
							if utils.EvalNilPtr(tc.commandFails[i], true) == false {
								expecting--
							}
						}
					}
				}
				if tc.webhooks != nil {
					expecting += len(*tc.webhooks)
					if tc.webhookFails != nil {
						for i := range tc.webhookFails {
							if tc.webhookFails[i] != nil && *tc.webhookFails[i] == false {
								expecting--
							}
						}
					}
				}
				// until we've receieved all the messages we're expecting
				for expecting != 0 {
					receivedForAnAction := false
					if tc.commands != nil {
						for i := range *tc.commands {
							if receivedMsg.CommandData[(*tc.commands)[i].String()] != nil {
								receivedForAnAction = true
								break
							}
						}
					}
					if !receivedForAnAction && tc.webhooks != nil {
						for i := range *tc.webhooks {
							if receivedMsg.WebHookData[i] != nil {
								receivedForAnAction = true
								break
							}
						}
					}
					if !receivedForAnAction {
						t.Fatalf("%s:\n%d remaining and message for an unknown action received\n%#v",
							name, expecting, receivedMsg)
					}
					expecting--
					if expecting != 0 {
						p := seeIfMessage(t, ws)
						if p == nil {
							t.Fatalf("%s:\nexpecting %d more messages but got %#v",
								name, expecting, p)
						}
						receivedMsg = api_types.WebSocketMessage{}
						json.Unmarshal(*p, &receivedMsg)
					}
				}
			}

			if tc.upgradesApprovedVersion {
				p = seeIfMessage(t, ws)
				receivedMsg = api_types.WebSocketMessage{}
				if p == nil {
					t.Fatalf("%s:\nexpecting message announcing approved version change but got %#v",
						name, p)
				}
				json.Unmarshal(*p, &receivedMsg)
				if receivedMsg.ServiceData.Status.ApprovedVersion != cfg.Service[*tc.serviceID].Status.LatestVersion {
					t.Fatalf("%s:\nexpected approved version to be updated to latest version in the message\n%#v\napproved=%#v, latest=%#v",
						name, receivedMsg, receivedMsg.ServiceData.Status.ApprovedVersion, cfg.Service[*tc.serviceID].Status.LatestVersion)
				}
			}
			if tc.upgradesDeployedVersion {
				p = seeIfMessage(t, ws)
				receivedMsg = api_types.WebSocketMessage{}
				if p == nil {
					t.Fatalf("%s:\nexpecting message announcing deployed version change but got %#v",
						name, p)
				}
				json.Unmarshal(*p, &receivedMsg)
				if receivedMsg.ServiceData.Status.DeployedVersion != cfg.Service[*tc.serviceID].Status.LatestVersion {
					t.Fatalf("%s:\nexpected deployed version to be updated to latest version in the message\n%#v\ndeployed=%#v, latest=%#v",
						name, receivedMsg, receivedMsg.ServiceData.Status.DeployedVersion, cfg.Service[*tc.serviceID].Status.LatestVersion)
				}
			}

			// extra message check
			p = seeIfMessage(t, ws)
			if p != nil {
				receivedMsg = api_types.WebSocketMessage{}
				json.Unmarshal(*p, &receivedMsg)
				t.Fatalf("%s:\nwasn't expecting another message but got one\n%#v",
					name, receivedMsg)
			}
		})
	}
}

func TestWebSocketApprovalsACTIONS(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	var logInitCommands *command.Controller
	logInitCommands.Init(jLog, nil, nil, nil, nil, nil)
	var logInitWebHooks *webhook.Slice
	logInitWebHooks.Init(jLog, nil, nil, nil, nil, nil, nil, nil)
	tests := map[string]struct {
		serviceID   *string
		stdoutRegex string
		bodyRegex   string
		commands    *command.Slice
		webhooks    *webhook.Slice
	}{
		"service_id=unknown": {serviceID: stringPtr("unknown?"), stdoutRegex: "service not found"},
		"service_id=nil":     {serviceID: nil, stdoutRegex: "service_data.id not provided"},
		"service_id=known, commands=[], webhooks=[],":  {serviceID: stringPtr("test"), stdoutRegex: "^$"},
		"service_id=known, commands=[1], webhooks=[],": {serviceID: stringPtr("test"), commands: &command.Slice{testCommandFail()}},
		"service_id=known, commands=[2], webhooks=[],": {serviceID: stringPtr("test"), commands: &command.Slice{testCommandFail(), testCommandPass()}},
		"service_id=known, commands=[], webhooks=[1],": {serviceID: stringPtr("test"), webhooks: &webhook.Slice{"fail0": testWebHookFail("fail0")}},
		"service_id=known, commands=[], webhooks=[2],": {serviceID: stringPtr("test"), webhooks: &webhook.Slice{"fail0": testWebHookFail("fail0"), "pass0": testWebHookPass("pass0")}},
		"service_id=known, commands=[2], webhooks=[2],": {serviceID: stringPtr("test"), commands: &command.Slice{testCommandFail(), testCommandPass()},
			webhooks: &webhook.Slice{"fail0": testWebHookFail("fail0"), "pass0": testWebHookPass("pass0")}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// backup Service
			var hadCommandSlice *command.Slice
			var hadWebHookSlice *webhook.Slice
			var hadStatus *service_status.Status
			if cfg.Service[utils.DefaultIfNil(tc.serviceID)] != nil {
				hadStatus = cfg.Service[*tc.serviceID].Status
				hadCommandSlice = cfg.Service[*tc.serviceID].Command
				cfg.Service[*tc.serviceID].Command = tc.commands
				if tc.commands == nil {
					cfg.Service[*tc.serviceID].CommandController = nil
				} else {
					cfg.Service[*tc.serviceID].CommandController = &command.Controller{}
					cfg.Service[*tc.serviceID].CommandController.Init(jLog, cfg.Service[*tc.serviceID].ID, nil, cfg.Service[*tc.serviceID].Command, nil, stringPtr("10m"))
				}
				hadWebHookSlice = cfg.Service[*tc.serviceID].WebHook
				if tc.webhooks != nil {
					for i := range *tc.webhooks {
						(*tc.webhooks)[i].Announce = cfg.Service[*tc.serviceID].Announce
					}
				}
				cfg.Service[*tc.serviceID].WebHook = tc.webhooks
				cfg.Service[*tc.serviceID].WebHook.Init(jLog, cfg.Service[*tc.serviceID].ID, cfg.Service[*tc.serviceID].Status, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{}, nil, cfg.Service[*tc.serviceID].Interval)
				defer func() {
					cfg.Service[*tc.serviceID].Status = hadStatus
					cfg.Service[*tc.serviceID].Command = hadCommandSlice
					cfg.Service[*tc.serviceID].CommandController.Init(jLog, cfg.Service[*tc.serviceID].ID, nil, cfg.Service[*tc.serviceID].Command, nil, stringPtr("10m"))
					cfg.Service[*tc.serviceID].WebHook = hadWebHookSlice
				}()
			}
			var receivedMsg api_types.WebSocketMessage
			ws := connectToWebSocket(t)
			defer ws.Close()
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN we send a message to the server (wsService)
			msg := api_types.WebSocketMessage{Version: intPtr(1), Page: stringPtr("APPROVALS"), Type: stringPtr("ACTIONS"),
				ServiceData: &api_types.ServiceSummary{ID: tc.serviceID}}
			if cfg.Service[utils.DefaultIfNil(tc.serviceID)] != nil {
				msg.ServiceData.Status = &api_types.Status{
					LatestVersion: cfg.Service[*tc.serviceID].Status.LatestVersion,
				}
			}
			data, _ := json.Marshal(msg)
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Errorf("%s:\nerror sending message\n%s",
					name, err.Error())
			}

			// THEN we get the expected response
			p := seeIfMessage(t, ws)
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			// stdout finishes
			if tc.stdoutRegex != "" {
				re := regexp.MustCompile(tc.stdoutRegex)
				match := re.MatchString(output)
				if !match {
					t.Errorf("%s:\nmatch on %q not found in\n%q",
						name, tc.stdoutRegex, output)
				}
				// don't want a message
				if p != nil {
					receivedMsg = api_types.WebSocketMessage{}
					json.Unmarshal(*p, &receivedMsg)
					t.Errorf("%s:\nwasn't expecting another message but got one\n%#v",
						name, receivedMsg)
				}
				return
			}
			expectingC := tc.commands != nil
			expectingWH := tc.webhooks != nil
			for expectingC || expectingWH {
				// didn't get a message
				if p == nil {
					t.Fatalf("%s:\nexpecting message but got %#v\nexpecting commands=%t, expecting webhook=%t",
						name, p, expectingC, expectingWH)
				}
				receivedMsg = api_types.WebSocketMessage{}
				json.Unmarshal(*p, &receivedMsg)
				if receivedMsg.CommandData != nil {
					expectingC = false
				} else if receivedMsg.WebHookData != nil {
					expectingWH = false
				}
				p = seeIfMessage(t, ws)
			}

			// extra message check
			if p != nil {
				receivedMsg = api_types.WebSocketMessage{}
				json.Unmarshal(*p, &receivedMsg)
				t.Fatalf("%s:\nwasn't expecting another message but got one\n%#v",
					name, receivedMsg)
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
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}

	// THEN it passes and we only receive a response with the WebHooks
	p := seeIfMessage(t, ws)
	if p == nil {
		t.Fatal("expecting another message but didn't get one")
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(*p, &receivedMsg)
	if *receivedMsg.Page != msgPage {
		t.Fatalf("Received a response for Page %q. Expected %q",
			*receivedMsg.Page, msgPage)
	}
	if receivedMsg.InfoData == nil {
		t.Fatalf("Didn't get any InfoData in the message\n%#v",
			receivedMsg)
	}
	if receivedMsg.InfoData.Build.GoVersion != utils.GoVersion {
		t.Errorf("Expected Info.Build.GoVersion to be %q, got %q\n%#v",
			utils.GoVersion, receivedMsg.InfoData.Build.GoVersion, receivedMsg)
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
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}

	// THEN it passes and we only receive a response with the WebHooks
	p := seeIfMessage(t, ws)
	if p == nil {
		t.Fatal("expecting another message but didn't get one")
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(*p, &receivedMsg)
	if *receivedMsg.Page != msgPage {
		t.Fatalf("Received a response for Page %q. Expected %q",
			*receivedMsg.Page, msgPage)
	}
	if receivedMsg.FlagsData == nil {
		t.Fatalf("Didn't get any FlagsData in the message\n%#v",
			receivedMsg)
	}
	if receivedMsg.FlagsData.LogLevel != cfg.Settings.GetLogLevel() {
		t.Errorf("Expected FlagsData.LogLevel to be %q, got %q\n%#v",
			cfg.Settings.GetLogLevel(), receivedMsg.FlagsData.LogLevel, receivedMsg)
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
		"no Service's":                     {nilService: true},
		"no Service DeployedVersionLookup": {nilServiceDVL: true},
		"no Service URLCommands":           {nilServiceURLC: true},
		"no Service Notify":                {nilServiceNotify: true},
		"no Service Command":               {nilServiceC: true},
		"no Service WebHook":               {nilServiceWH: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// backup changes
			hadService := make(service.Slice, len(cfg.Service))
			for i := range cfg.Service {
				svc := *cfg.Service[i]
				hadService[i] = &svc
				if tc.nilServiceDVL {
					dvl := *cfg.Service[i].DeployedVersionLookup
					cfg.Service[i].DeployedVersionLookup = nil
					hadService[i].DeployedVersionLookup = &dvl
				}
				if tc.nilServiceURLC {
					urlc := *cfg.Service[i].URLCommands
					cfg.Service[i].URLCommands = nil
					hadService[i].URLCommands = &urlc
				}
				if tc.nilServiceNotify {
					notify := *cfg.Service[i].Notify
					cfg.Service[i].Notify = nil
					hadService[i].Notify = &notify
				}
				wh := *cfg.Service[i].WebHook
				hadService[i].WebHook = &wh
				if tc.nilServiceWH {
					cfg.Service[i].WebHook = nil
				}
				if tc.nilServiceC {
					command := *cfg.Service[i].Command
					cfg.Service[i].Command = nil
					hadService[i].Command = &command
					cfg.Service[i].CommandController = nil
				} else {
					cfg.Service[i].CommandController.Init(jLog, &i, svc.Status, svc.Command, nil, nil)
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

			// WHEN we send a message to get the Commands+WebHooks
			var (
				msgVersion int    = 1
				msgPage    string = "CONFIG"
				msgType    string = "INIT"
			)
			msg := api_types.WebSocketMessage{
				Version: &msgVersion,
				Page:    &msgPage,
				Type:    &msgType,
			}
			data, _ := json.Marshal(msg)
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Errorf("%v",
					err)
			}

			// THEN it passes and we only receive a response with the WebHooks
			p := seeIfMessage(t, ws)
			if p == nil {
				t.Errorf("%s:\nSETTINGS - expecting another message but didn't get one",
					name)
			}
			var receivedMsg api_types.WebSocketMessage
			json.Unmarshal(*p, &receivedMsg)
			{ // SETTINGS
				wantedType := "SETTINGS"
				if *receivedMsg.Page != msgPage ||
					*receivedMsg.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
				} else {
					if receivedMsg.ConfigData == nil {
						t.Errorf("Didn't get any ConfigData in the message\n%#v",
							receivedMsg)
					} else {
						if receivedMsg.ConfigData == nil ||
							receivedMsg.ConfigData.Settings == nil ||
							receivedMsg.ConfigData.Settings.Log.Level == nil {
							t.Errorf("Didn't receive ConfigData.Settings.Log.Level from\n%#v",
								receivedMsg)
						} else if *receivedMsg.ConfigData.Settings.Log.Level != cfg.Settings.GetLogLevel() {
							t.Errorf("Expected ConfigData.Settings.Log.Level to be %q, got %q\n%#v",
								cfg.Settings.GetLogLevel(), *receivedMsg.ConfigData.Settings.Log.Level, receivedMsg)
						}
					}
				}
			}
			p = seeIfMessage(t, ws)
			if p == nil {
				t.Errorf("%s:\nDEFAULTS - expecting another message but didn't get one",
					name)
			}
			receivedMsg = api_types.WebSocketMessage{}
			json.Unmarshal(*p, &receivedMsg)
			{ // DEFAULTS
				wantedType := "DEFAULTS"
				if *receivedMsg.Page != msgPage ||
					*receivedMsg.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
				} else {
					if receivedMsg.ConfigData == nil ||
						receivedMsg.ConfigData.Defaults == nil ||
						receivedMsg.ConfigData.Defaults.Service.Interval == nil {
						t.Errorf("Didn't receive ConfigData.Defaults.Service.Interval from\n%#v",
							receivedMsg)
					} else if *receivedMsg.ConfigData.Defaults.Service.Interval != *cfg.Defaults.Service.Interval {
						t.Errorf("Expected ConfigData.Defaults.Service.Interval to be %q, got %q\n%#v",
							*cfg.Defaults.Service.Interval, *receivedMsg.ConfigData.Defaults.Service.Interval, receivedMsg)
					}
				}
			}
			p = seeIfMessage(t, ws)
			if p == nil {
				t.Errorf("%s:\nNOTIFY - expecting another message but didn't get one",
					name)
			}
			receivedMsg = api_types.WebSocketMessage{}
			json.Unmarshal(*p, &receivedMsg)
			{ // NOTIFY
				wantedType := "NOTIFY"
				if *receivedMsg.Page != msgPage ||
					*receivedMsg.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
				} else {
					if receivedMsg.ConfigData == nil ||
						receivedMsg.ConfigData.Notify == nil ||
						(*receivedMsg.ConfigData.Notify)["discord"] == nil {
						t.Errorf("Didn't receive ConfigData.Notify.discord from\n%#v",
							receivedMsg)
					} else if (*receivedMsg.ConfigData.Notify)["discord"].Options["message"] != cfg.Notify["discord"].Options["message"] {
						t.Errorf("Expected ConfigData.Notify.discord.Options.message to be %q, got %q\n%#v",
							cfg.Notify["discord"].Options["message"], (*receivedMsg.ConfigData.Notify)["discord"].Options["message"], receivedMsg)
					}
				}
			}
			p = seeIfMessage(t, ws)
			if p == nil {
				t.Errorf("%s:\nWEBHOOK - expecting another message but didn't get one",
					name)
			}
			receivedMsg = api_types.WebSocketMessage{}
			json.Unmarshal(*p, &receivedMsg)
			{ // WEBHOOK
				wantedType := "WEBHOOK"
				if *receivedMsg.Page != msgPage ||
					*receivedMsg.Type != wantedType {
					t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
						*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
				} else {
					if receivedMsg.ConfigData == nil ||
						receivedMsg.ConfigData.WebHook == nil ||
						(*receivedMsg.ConfigData.WebHook)["pass"] == nil {
						t.Errorf("Didn't receive ConfigData.WebHook.pass from\n%#v",
							receivedMsg)
					} else if *(*receivedMsg.ConfigData.WebHook)["pass"].URL != *cfg.WebHook["pass"].URL {
						t.Errorf("Expected ConfigData.WebHook.pass.URL to be %q, got %q\n%#v",
							*cfg.WebHook["pass"].URL, *(*receivedMsg.ConfigData.WebHook)["pass"].URL, receivedMsg)
					}
				}
			}
			// SERVICE
			p = seeIfMessage(t, ws)
			if p == nil {
				t.Errorf("%s:\nSERVICE - expecting another message but didn't get one",
					name)
			}
			receivedMsg = api_types.WebSocketMessage{}
			json.Unmarshal(*p, &receivedMsg)
			if tc.nilService {
				if receivedMsg.ServiceData != nil {
					t.Errorf("%s\n:expecting ServiceData to be nil, not \n%#v",
						name, receivedMsg.ServiceData)
				}
			} else {
				receivedTestService := (*receivedMsg.ConfigData.Service)["test"]
				cfgTestService := cfg.Service["test"]
				// service
				if *receivedTestService.Comment != *cfgTestService.Comment {
					t.Errorf("ConfigData.Service.test.Comment should've been %q, got %q",
						*cfgTestService.Comment, *receivedTestService.Comment)
				}
				if *receivedTestService.URL != *cfgTestService.URL {
					t.Errorf("ConfigData.Service.test.URL should've been %q, got %q",
						*cfgTestService.URL, *receivedTestService.URL)
				}
				if *receivedTestService.WebURL != *cfgTestService.WebURL {
					t.Errorf("ConfigData.Service.test.WebURL should've been %q, got %q",
						*cfgTestService.WebURL, *receivedTestService.WebURL)
				}
				if *receivedTestService.RegexContent != *cfgTestService.RegexContent {
					t.Errorf("ConfigData.Service.test.RegexContent should've been %q, got %q",
						*cfgTestService.RegexContent, *receivedTestService.RegexContent)
				}
				if *receivedTestService.RegexVersion != *cfgTestService.RegexVersion {
					t.Errorf("ConfigData.Service.test.RegexVersion should've been %q, got %q",
						*cfgTestService.RegexVersion, *receivedTestService.RegexVersion)
				}
				if *receivedTestService.AutoApprove != *cfgTestService.AutoApprove {
					t.Errorf("ConfigData.Service.test.AutoApprove should've been %t, got %t",
						*cfgTestService.AutoApprove, *receivedTestService.AutoApprove)
				}
				// deployed version lookup
				if tc.nilServiceDVL {
					if receivedTestService.DeployedVersionLookup != nil {
						if receivedTestService.DeployedVersionLookup != nil {
							t.Errorf("%s\n:expecting DeployedVersionLookup to be nil, not \n%#v",
								name, receivedTestService.DeployedVersionLookup)
						}
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
					if receivedTestService.URLCommands != nil {
						if receivedTestService.URLCommands != nil {
							t.Errorf("%s\n:expecting URLCommands to be nil, not \n%#v",
								name, receivedTestService.URLCommands)
						}
					}
				} else {
					if receivedTestService.URLCommands == nil {
						t.Errorf("ConfigData.Service.test.URLCommands should've been %#v, got %#v",
							*cfgTestService.URLCommands, receivedTestService.URLCommands)
					} else if *(*receivedTestService.URLCommands)[0].Regex != *(*cfgTestService.URLCommands)[0].Regex {
						t.Errorf("ConfigData.Service.test.URLCommands[0].Regex should've been %q, got %q",
							*(*cfgTestService.URLCommands)[0].Regex, *(*receivedTestService.URLCommands)[0].Regex)
					}
				}
				// notify
				if tc.nilServiceNotify {
					if receivedTestService.Notify != nil {
						if receivedTestService.Notify != nil {
							t.Errorf("%s\n:expecting Notify to be nil, not \n%#v",
								name, receivedTestService.Notify)
						}
					}
				} else {
					if receivedTestService.Notify == nil {
						t.Errorf("ConfigData.Service.test.Notify should've been %#v, got %#v",
							*cfgTestService.Notify, receivedTestService.Notify)
					} else if (*receivedTestService.Notify)["test"].Options["message"] != (*cfgTestService.Notify)["test"].Options["message"] {
						t.Errorf("ConfigData.Service.test.Notify.test.Options.message should've been %q, got %q",
							(*cfgTestService.Notify)["test"].Options["message"], (*receivedTestService.Notify)["test"].Options["message"])
					}
				}
				// webhook
				if tc.nilServiceWH {
					if receivedTestService.WebHook != nil {
						if receivedTestService.WebHook != nil {
							t.Errorf("%s\n:expecting WebHook to be nil, not \n%#v",
								name, receivedTestService.WebHook)
						}
					}
				} else {
					if receivedTestService.WebHook == nil {
						t.Errorf("ConfigData.Service.test.WebHook should've been %#v, got %#v",
							*cfgTestService.WebHook, receivedTestService.WebHook)
					} else if *(*receivedTestService.WebHook)["test"].URL != *(*cfgTestService.WebHook)["test"].URL {
						t.Errorf("ConfigData.Service.test.WebHook.test.URL should've been %q, got %q",
							*(*cfgTestService.WebHook)["test"].URL, *(*receivedTestService.WebHook)["test"].URL)
					}
				}
				// command
				if tc.nilServiceC {
					if receivedTestService.Command != nil {
						if receivedTestService.Command != nil {
							t.Errorf("%s\n:expecting Command to be nil, not \n%#v",
								name, receivedTestService.Command)
						}
					}
				} else {
					if receivedTestService.Command == nil {
						t.Errorf("ConfigData.Service.test.Command should've been %#v, got %#v",
							*cfgTestService.Command, receivedTestService.Command)
					} else {
						got := strings.Join((*receivedTestService.Command)[0], " ")
						if got != (*cfgTestService.Command)[0].String() {
							t.Errorf("ConfigData.Service.test.Command[0] should've been %q, got %q",
								(*cfgTestService.Command)[0].String(), got)
						}
					}
				}
			}
			p = seeIfMessage(t, ws)
			if p != nil {
				receivedMsg = api_types.WebSocketMessage{}
				json.Unmarshal(*p, &receivedMsg)
				t.Errorf("%s:\nwasn't expecting another message but got one\n%#v",
					name, receivedMsg)
			}
		})
	}
}
