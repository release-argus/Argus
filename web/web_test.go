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
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/utils"
	api_types "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

var router *mux.Router
var port *string
var cfg *config.Config

func TestMain(m *testing.M) {
	// GIVEN a valid config with a Service
	newCfg := testConfig()
	cfg = &newCfg
	jLog = utils.NewJLog("WARN", false)
	port = cfg.Settings.Web.ListenPort

	// WHEN the Router is fetched for this Config
	router = newWebUI(cfg)
	go http.ListenAndServe("localhost:"+*port, router)

	// THEN Web UI is accessible for the tests
	code := m.Run()
	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func TestWebAccessible(t *testing.T) {
	// GIVEN we have the Web UI Router

	// WHEN we make a request to /approvals
	req, _ := http.NewRequest("GET", "/approvals", nil)
	response := executeRequest(req)

	// THEN we get a Status OK
	if response.Code != http.StatusOK {
		t.Errorf("Expected a 200, got %d",
			response.Code)
	}
}

func TestMetricsHTTP(t *testing.T) {
	// GIVEN we have the Web UI Router

	// WHEN we make a request to /metrics
	req, _ := http.NewRequest("GET", "/metrics", nil)

	// THEN the metrics are returned
	response := executeRequest(req)
	if response.Code != http.StatusOK {
		t.Errorf("Expected a 200, got %d",
			response.Code)
	}
	if !strings.Contains(response.Body.String(), "go_gc_duration_") {
		t.Errorf("Metrics page doesn't appear to have loaded correctly, got\n%s",
			response.Body.String())
	}
}

func TestAPIGETVersion(t *testing.T) {
	// GIVEN we have the Web UI Router

	// WHEN we make a request to /api/v1/version
	req, _ := http.NewRequest("GET", "/api/v1/version", nil)

	// THEN the version is returned
	response := executeRequest(req)
	if response.Code != http.StatusOK {
		t.Errorf("Expected a 200, got %d",
			response.Code)
	}
	want := fmt.Sprintf(`"goVersion":"%s"`, utils.GoVersion)
	got := response.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("/api/v1/version doesn't appear to have loaded correctly, got\n%s",
			got)
	}
}

func TestMainWithRoutePrefix(t *testing.T) {
	// GIVEN a valid config with a Service
	config := testConfig()
	*config.Settings.Web.RoutePrefix = "/test"

	// WHEN the Web UI is started with this Config
	go Run(&config, utils.NewJLog("WARN", false))
	time.Sleep(2 * time.Second)

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
	time.Sleep(time.Second)
	return ws
}

func TestWebSocketInvalidJSON(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message with an unknown Type
	msg := `"version": 1, "key": "value", "key": "value"`
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we log this unknown message Type
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "missing/invalid version key") {
		t.Errorf("Expecting an error about the sent message %q being invalid JSON. Got\n%s",
			msg, output)
	}
}

func TestWebSocketClientMessageWithNoVersionKey(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message with an unknown Type
	msg := `"key": "value"`
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we log this unknown message Type
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "missing/invalid version key") {
		t.Errorf("Expecting an error about the sent message %q not containing a version key. Got\n%s",
			msg, output)
	}
}

func TestWebSocketApprovalsINIT(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to the server (wsService)
	var (
		msgVersion int    = 1
		msgPage    string = "APPROVALS"
		msgType    string = "INIT"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
	}
	data, _ := json.Marshal(msg)

	// THEN we get the expected responses
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	// ORDERING
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if receivedMsg.Order == nil {
		t.Errorf("bad message, didn't contain ordering")
	}
	receivedMsg = api_types.WebSocketMessage{}
	// SERVICE
	for _, key := range *cfg.Order {
		if cfg.Service[key] == nil {
			continue
		}
		_, p, err = ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		json.Unmarshal(p, &receivedMsg)
		if receivedMsg.ServiceData == nil {
			t.Errorf("bad message, didn't contain ServiceData")
		}
	}
}

func TestWebSocketApprovalsWithNilService(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and all Service's are nil
	ws := connectToWebSocket(t)
	defer ws.Close()
	hadOrder := make([]string, len(*cfg.Order)+1)
	for i := range *cfg.Order {
		hadOrder[i] = (*cfg.Order)[i]
	}
	*cfg.Order = append(*cfg.Order, "nilService")

	// WHEN we send a message to the server (wsService)
	var (
		msgVersion int    = 1
		msgPage    string = "APPROVALS"
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
	time.Sleep(time.Second)

	// THEN we get the expected responses
	// ORDERING
	_, p, err := ws.ReadMessage()
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if err != nil {
		t.Errorf("%v",
			err)
	} else if receivedMsg.Order == nil ||
		len(*receivedMsg.Order) == 0 {
		order := "[]"
		if cfg.Order != nil {
			order = fmt.Sprint(*cfg.Order)
		}
		t.Errorf("Didn't receive correct ordering - %s",
			order)
	}
	// Would've crashed here without the nil handling
	cfg.Order = &hadOrder
}

func TestWebSocketApprovalsVersion(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and the DeployedVersion for a Service != LatestVersion
	cfg.Service["test"].Status.DeployedVersion = "0.0.0"
	cfg.Service["test"].Status.LatestVersion = "0.1.0"
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to SKIP this LatestVersion (wsServiceAction)
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = "test"
		msgTarget    string = "ARGUS_SKIP"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}

	// THEN we receive a response acknowledging it
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if !strings.HasPrefix(receivedMsg.ServiceData.Status.ApprovedVersion,
		"SKIP_"+cfg.Service["test"].Status.LatestVersion) {
		t.Errorf("LatestVersion wasn't skipped?\n%v",
			string(p))
	}
}

func TestWebSocketApprovalsVersionWithNoServiceID(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and the DeployedVersion for a Service != LatestVersion
	cfg.Service["test"].Status.DeployedVersion = "0.0.0"
	cfg.Service["test"].Status.LatestVersion = "0.1.0"
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send an invalid message to SKIP this LatestVersion (wsServiceAction)
	// No Service ID
	var (
		msgVersion int    = 1
		msgPage    string = "APPROVALS"
		msgType    string = "VERSION"
		msgTarget  string = "ARGUS_SKIP"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we log a warning acknowledging it
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "service_data.id not provided") {
		t.Errorf("Expecting an error about no service id being provided. Got\n%s",
			output)
	}
}

func TestWebSocketApprovalsVersionWithNoTarget(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and the DeployedVersion for a Service != LatestVersion
	cfg.Service["test"].Status.DeployedVersion = "0.0.0"
	cfg.Service["test"].Status.LatestVersion = "0.1.0"
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send an invalid message to SKIP this LatestVersion (wsServiceAction)
	// with no Target
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = "test"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we log a warning acknowledging it
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "target for command/webhook not provided") {
		t.Errorf("Expecting an error about no target being provided. Got\n%s",
			output)
	}
}

func TestWebSocketApprovalsVersionWithInvalidServiceID(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and the DeployedVersion for a Service != LatestVersion
	cfg.Service["test"].Status.DeployedVersion = "0.0.0"
	cfg.Service["test"].Status.LatestVersion = "0.1.0"
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send an invalid message to SKIP this LatestVersion (wsServiceAction)
	// with an invalid (unknown) Service ID
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = "unknown?"
		msgTarget    string = "ARGUS_SKIP"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we log a warning acknowledging it
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "not a valid service_id") {
		t.Errorf("Expecting an error about the %q service_id being unknown. Got\n%s",
			*msg.ServiceData.ID, output)
	}
}

func TestWebSocketApprovalsVersionWithNoCommandsOrWebHooks(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and the DeployedVersion for a Service != LatestVersion
	testName := "TestWebSocketApprovalsVersionWithNoCommandsOrWebHooks"
	svc := testService(testName)
	svc.Status.DeployedVersion = "0.0.0"
	svc.Status.LatestVersion = "0.1.0"
	svc.Command = nil
	svc.CommandController = nil
	svc.WebHook = nil
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send an invalid message to SKIP this LatestVersion (wsServiceAction)
	// with an invalid (unknown) Service ID
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "ARGUS_SKIP"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we receive a response acknowledging it
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "does not have any commands/webhooks to approve") {
		t.Errorf("Expecting an error about the %q service having no commands/webhooks. Got\n%s",
			*msg.ServiceData.ID, output)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithArgusFailedAndFailedCommandThatWillPass(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for the Command was true
	testName := "TestWebSocketApprovalsVersionWithArgusFailedAndFailedCommandThatWillPass"
	svc := testService(testName)
	want := "0.1.0"
	svc.Status.DeployedVersion = "0.0.0"
	svc.Status.LatestVersion = want
	svc.Command = &command.Slice{testCommandPass()}
	svc.CommandController.Init(jLog, svc.ID, svc.Status, svc.Command, nil, svc.Interval)
	failed := true
	svc.CommandController.Failed[0] = &failed
	svc.WebHook = nil
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send an invalid message to SKIP this LatestVersion (wsServiceAction)
	// with an invalid (unknown) Service ID
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "ARGUS_FAILED"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(2 * time.Second)

	// THEN we receive a response acknowledging it
	if svc.Status.DeployedVersion != want {
		t.Errorf("ARGUS_FAILED should have re-run the Service.Command %q and passed. This should've updated DeployedVersion to %q, not %q",
			(*svc.Command)[0].String(), want, svc.Status.DeployedVersion)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithSpecificCommandThatIsOnlyFailedDidUpdateLatestVersion(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for one Command is true
	testName := "TestWebSocketApprovalsVersionWithSpecificCommandThatIsOnlyFailedDidUpdateLatestVersion"
	svc := testService(testName)
	want := "0.1.0"
	svc.Status.DeployedVersion = "0.0.0"
	svc.Status.LatestVersion = want
	svc.Command = &command.Slice{testCommandPass(), testCommandFail()}
	svc.CommandController.Init(jLog, svc.ID, svc.Status, svc.Command, nil, svc.Interval)
	failed0 := true
	svc.CommandController.Failed[0] = &failed0
	failed1 := false
	svc.CommandController.Failed[1] = &failed1
	svc.WebHook = nil
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to re-run that failed Command
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "command_" + (*svc.Command)[0].String()
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and we receive a response acknowledging the change in DeployedVersion
	if svc.Status.DeployedVersion != want {
		t.Errorf("%q should have re-run the Service.Command %q and passed. This should've updated LatestVersion to %q, not %q",
			msgTarget, (*svc.Command)[0].String(), want, svc.Status.DeployedVersion)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithSpecificCommandThatIsOnlyFailedDidAnnounce(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for one Command is true
	testName := "TestWebSocketApprovalsVersionWithSpecificCommandThatIsOnlyFailedDidAnnounce"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	svc.Status.DeployedVersion = "0.0.0"
	want := "0.1.0"
	svc.Status.LatestVersion = want
	svc.Command = &command.Slice{testCommandPass(), testCommandFail()}
	svc.CommandController.Init(jLog, svc.ID, svc.Status, svc.Command, nil, svc.Interval)
	svc.CommandController.Announce = cfg.Service["test"].Announce
	failed0 := true
	svc.CommandController.Failed[0] = &failed0
	failed1 := false
	svc.CommandController.Failed[1] = &failed1
	svc.WebHook = nil
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to re-run that failed Command
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "command_" + (*svc.Command)[0].String()
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and we receive a response acknowledging the change in DeployedVersion
	// EVENT
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if receivedMsg.CommandData[(*svc.Command)[0].String()].Failed == nil ||
		*receivedMsg.CommandData[(*svc.Command)[0].String()].Failed != false {
		got := "nil"
		if receivedMsg.CommandData[(*svc.Command)[0].String()].Failed != nil {
			got = "true"
		}
		t.Errorf("%q should have re-run the Service.Command %q and passed but got %s in the WebSocket response",
			msgTarget, (*svc.Command)[0].String(), got)
	}
	// VERSION
	_, p, err = ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	receivedMsg = api_types.WebSocketMessage{}
	json.Unmarshal(p, &receivedMsg)
	got := "nil"
	if receivedMsg.ServiceData != nil && receivedMsg.ServiceData.Status != nil {
		got = receivedMsg.ServiceData.Status.DeployedVersion
	}
	if got != want {
		t.Errorf("Was expecting DeployedVersion to become LatestVersion %q, not %q",
			want, got)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithSpecificCommandThatIsNotOnlyFailedDidntAnnounce(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for more than one of the Commands is true
	testName := "TestWebSocketApprovalsVersionWithSpecificCommandThatIsNotOnlyFailedDidAnnounce"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	want := "0.0.0"
	svc.Status.DeployedVersion = want
	svc.Status.LatestVersion = "0.1.0"
	svc.Command = &command.Slice{testCommandPass(), testCommandFail()}
	svc.CommandController.Init(jLog, svc.ID, svc.Status, svc.Command, nil, svc.Interval)
	svc.CommandController.Announce = cfg.Service["test"].Announce
	failed0 := true
	svc.CommandController.Failed[0] = &failed0
	failed1 := true
	svc.CommandController.Failed[1] = &failed1
	svc.WebHook = nil
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to re-run one of the failed Commands
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "command_" + (*svc.Command)[0].String()
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and we receive a response acknowledging it with no change in DeployedVersion
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if receivedMsg.CommandData[(*svc.Command)[0].String()].Failed == nil ||
		*receivedMsg.CommandData[(*svc.Command)[0].String()].Failed != false {
		got := "nil"
		if receivedMsg.CommandData[(*svc.Command)[0].String()].Failed != nil {
			got = "true"
		}
		t.Errorf("%q should have re-run the Service.Command %q and passed but got failed=%s in the WebSocket response",
			msgTarget, (*svc.Command)[0].String(), got)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithSpecificCommandThatIsNotOnlyFailedDidntUpdateLatestVersion(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for more than one of the Commands is true
	testName := "TestWebSocketApprovalsVersionWithSpecificCommandThatIsNotOnlyFailedDidntUpdateLatestVersion"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	want := "0.0.0"
	svc.Status.DeployedVersion = want
	svc.Status.LatestVersion = "0.1.0"
	svc.Command = &command.Slice{testCommandPass(), testCommandFail()}
	svc.CommandController.Init(jLog, svc.ID, svc.Status, svc.Command, nil, svc.Interval)
	svc.CommandController.Announce = cfg.Service["test"].Announce
	failed0 := true
	svc.CommandController.Failed[0] = &failed0
	failed1 := true
	svc.CommandController.Failed[1] = &failed1
	svc.WebHook = nil
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to re-run one of the failed Commands
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "command_" + (*svc.Command)[0].String()
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and the DeployedVersion is unchanged
	if svc.Status.DeployedVersion != want {
		t.Errorf("%q should have re-run the Service.Command %q and passed but not have updated LatestVersion to %q, want %q",
			msgTarget, (*svc.Command)[0].String(), svc.Status.DeployedVersion, want)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithSpecificWebHookThatIsOnlyFailedDidAnnounce(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for one WebHook is true
	testName := "TestWebSocketApprovalsVersionWithSpecificWebHookThatIsOnlyFailedDidAnnounce"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	svc.Status.DeployedVersion = "0.0.0"
	want := "0.1.0"
	svc.Status.LatestVersion = want
	svc.Command = nil
	whPass := testWebHookPass("pass")
	whFail := testWebHookFail("fail")
	whPass.Announce = cfg.Service["test"].Announce
	whFail.Announce = cfg.Service["test"].Announce
	failed0 := true
	whPass.Failed = &failed0
	failed1 := false
	whFail.Failed = &failed1
	svc.WebHook = &webhook.Slice{
		*whPass.ID: &whPass,
		*whFail.ID: &whFail,
	}
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to re-run that failed WebHook
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "webhook_" + *whPass.ID
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and we receive a response acknowledging the change in DeployedVersion
	// EVENT
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if receivedMsg.WebHookData[*whPass.ID].Failed == nil ||
		*receivedMsg.WebHookData[*whPass.ID].Failed != false {
		got := "nil"
		if receivedMsg.WebHookData[*whPass.ID].Failed != nil {
			got = "true"
		}
		t.Errorf("%q should have re-run the Service.WebHook and passed but got %s in the WebSocket response",
			msgTarget, got)
	}
	// VERSION
	_, p, err = ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	receivedMsg = api_types.WebSocketMessage{}
	json.Unmarshal(p, &receivedMsg)
	got := "nil"
	if receivedMsg.ServiceData != nil && receivedMsg.ServiceData.Status != nil {
		got = receivedMsg.ServiceData.Status.DeployedVersion
	}
	if got != want {
		t.Errorf("Was expecting DeployedVersion to become LatestVersion %q, not %q",
			want, got)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithSpecificWebHookThatIsNotOnlyFailedDidAnnounce(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for more than one of the WebHooks is true
	testName := "TestWebSocketApprovalsVersionWithSpecificWebHookThatIsNotOnlyFailedDidAnnounce"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	want := "0.0.0"
	svc.Status.DeployedVersion = want
	svc.Status.LatestVersion = "0.1.0"
	svc.Command = nil
	wh0 := testWebHookPass("wh0")
	wh1 := testWebHookFail("wh1")
	wh0.Announce = cfg.Service["test"].Announce
	wh1.Announce = cfg.Service["test"].Announce
	failed0 := true
	wh0.Failed = &failed0
	failed1 := true
	wh1.Failed = &failed1
	svc.WebHook = &webhook.Slice{
		*wh0.ID: &wh0,
		*wh1.ID: &wh1,
	}
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to re-run one of the failed WebHooks
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "webhook_" + *wh0.ID
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and we receive a response acknowledging it with no change in DeployedVersion
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if receivedMsg.WebHookData[*wh0.ID].Failed == nil ||
		*receivedMsg.WebHookData[*wh0.ID].Failed != false {
		got := "nil"
		if receivedMsg.WebHookData[*wh0.ID].Failed != nil {
			got = "true"
		}
		t.Errorf("%q should have re-run the Service.WebHook and passed but got %s in the WebSocket response",
			msgTarget, got)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsVersionWithSpecificWebHookThatIsNotOnlyFailedDidntUpdateLatestVersion(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// the DeployedVersion for a Service != LatestVersion
	// and the Failed status for more than one of the WebHooks is true
	testName := "TestWebSocketApprovalsVersionWithSpecificWebHookThatIsNotOnlyFailedDidUpdateLatestVersion"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	want := "0.0.0"
	svc.Status.DeployedVersion = want
	svc.Status.LatestVersion = "0.1.0"
	svc.Command = nil
	wh0 := testWebHookPass("wh0")
	wh1 := testWebHookFail("wh1")
	wh0.Announce = cfg.Service["test"].Announce
	wh1.Announce = cfg.Service["test"].Announce
	failed0 := true
	wh0.Failed = &failed0
	failed1 := true
	wh1.Failed = &failed1
	svc.WebHook = &webhook.Slice{
		*wh0.ID: &wh0,
		*wh1.ID: &wh1,
	}
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to re-run one of the failed WebHooks
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "VERSION"
		msgServiceID string = *svc.ID
		msgTarget    string = "webhook" + *wh0.ID
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
		Target: &msgTarget,
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and the DeployedVersion is unchanged
	if svc.Status.DeployedVersion != want {
		t.Errorf("%q should have re-run the Service.WebHook and passed but not have updated LatestVersion to %q, want %q",
			msgTarget, svc.Status.DeployedVersion, want)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsActionsWithNoServiceID(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message to get the Commands+WebHooks
	// with no service_data.id
	var (
		msgVersion int    = 1
		msgPage    string = "APPROVALS"
		msgType    string = "ACTIONS"
	)
	msg := api_types.WebSocketMessage{
		Version:     &msgVersion,
		Page:        &msgPage,
		Type:        &msgType,
		ServiceData: &api_types.ServiceSummary{},
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we log a warning acknowledging it
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if strings.Count(output, "service_data.id not provided") != 2 {
		t.Errorf("Expecting 2 errors about no service id being provided. Got\n%s",
			output)
	}
}

func TestWebSocketApprovalsActionsWithUnknownServiceID(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message to get the Commands+WebHooks
	// with an unknown service_data.id
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "ACTIONS"
		msgServiceID string = "unknown?"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
		},
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN we log a warning acknowledging it
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if strings.Count(output, fmt.Sprintf("%q, service not found", msgServiceID)) != 2 {
		t.Errorf("Expecting an error about no service id being provided. Got\n%s",
			output)
	}
}

func TestWebSocketApprovalsActionsWithNoWebHook(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and there's a Service with WebHook but no CommandController
	testName := "TestWebSocketApprovalsActionsWithNoWebHook"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	svc.Command = &command.Slice{testCommandPass(), testCommandFail()}
	svc.CommandController.Init(jLog, svc.ID, svc.Status, svc.Command, nil, svc.Interval)
	failed0 := true
	svc.CommandController.Failed[0] = &failed0
	failed1 := false
	svc.CommandController.Failed[1] = &failed1
	svc.WebHook = nil
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to get the Commands+WebHooks
	// and the Service has no CommandController
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "ACTIONS"
		msgServiceID string = testName
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
		},
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if len(receivedMsg.CommandData) == 0 {
		t.Errorf("Expected Commands to be received, got %d\n%v",
			len(receivedMsg.CommandData), receivedMsg)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsActionsWithNoCommandController(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	// and there's a Service with WebHook but no CommandController
	testName := "TestWebSocketApprovalsActionsWithNoCommandController"
	svc := testService(testName)
	svc.Announce = cfg.Service["test"].Announce
	svc.Command = nil
	svc.CommandController = nil
	wh0 := testWebHookPass("wh0")
	wh1 := testWebHookFail("wh1")
	failed0 := true
	wh0.Failed = &failed0
	failed1 := true
	wh1.Failed = &failed1
	svc.WebHook = &webhook.Slice{
		*wh0.ID: &wh0,
		*wh1.ID: &wh1,
	}
	cfg.Service[testName] = &svc
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to get the Commands+WebHooks
	// and the Service has no CommandController
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "ACTIONS"
		msgServiceID string = testName
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
		},
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if len(receivedMsg.WebHookData) == 0 {
		t.Errorf("Expected WebHooks to be received, got %d\n%v",
			len(receivedMsg.WebHookData), receivedMsg)
	}
	delete(cfg.Service, testName)
}

func TestWebSocketApprovalsActions(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()

	// WHEN we send a message to get the Commands+WebHooks
	var (
		msgVersion   int    = 1
		msgPage      string = "APPROVALS"
		msgType      string = "ACTIONS"
		msgServiceID string = "test"
	)
	msg := api_types.WebSocketMessage{
		Version: &msgVersion,
		Page:    &msgPage,
		Type:    &msgType,
		ServiceData: &api_types.ServiceSummary{
			ID: &msgServiceID,
			Status: &api_types.Status{
				LatestVersion: cfg.Service["test"].Status.LatestVersion,
			},
		},
	}
	data, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Errorf("%v",
			err)
	}

	// THEN we receive a response acknowledging it
	// WEBHOOK
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if receivedMsg.WebHookData == nil {
		t.Errorf("bad message, expecting WebHookData not %v",
			receivedMsg)
	}
	if (*receivedMsg.WebHookData["test"]).Failed != nil {
		t.Errorf("bad message, expecting WebHook Failed Status to be nil, not %t",
			*(*receivedMsg.WebHookData["test"]).Failed)
	}
	// COMMAND
	_, p, err = ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	receivedMsg = api_types.WebSocketMessage{}
	json.Unmarshal(p, &receivedMsg)
	if receivedMsg.CommandData == nil {
		t.Errorf("bad message, expecting CommandData not %v",
			receivedMsg)
	}
	command := (*cfg.Service["test"].Command)[0].String()
	if (*receivedMsg.CommandData[command]).Failed != nil {
		t.Errorf("bad message, expecting Command Failed Status to be nil, not %t",
			*(*receivedMsg.CommandData[command]).Failed)
	}
}

func TestWebSocketApprovalsUnknownType(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message with an unknown Type
	var (
		msgVersion int    = 1
		msgPage    string = "APPROVALS"
		msgType    string = "UNKNOWN"
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
	time.Sleep(time.Second)

	// THEN we log this unknown message Type
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "Unknown APPROVALS Type") {
		t.Errorf("Expecting an error about the unknown %q Type we used (%s). Got\n%s",
			*msg.Page, *msg.Type, output)
	}
}

func TestWebSocketRuntimeBuildInit(t *testing.T) {
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if *receivedMsg.Page != msgPage {
		t.Fatalf("Received a response for Page %q. Expected %q",
			*receivedMsg.Page, msgPage)
	}
	if receivedMsg.InfoData == nil {
		t.Fatalf("Didn't get any InfoData in the message\n%v",
			receivedMsg)
	}
	if receivedMsg.InfoData.Build.GoVersion != utils.GoVersion {
		t.Errorf("Expected Info.Build.GoVersion to be %q, got %q\n%v",
			utils.GoVersion, receivedMsg.InfoData.Build.GoVersion, receivedMsg)
	}
}

func TestWebSocketRuntimeBuildUnknownType(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message with an unknown Type
	var (
		msgVersion int    = 1
		msgPage    string = "RUNTIME_BUILD"
		msgType    string = "UNKNOWN"
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
	time.Sleep(time.Second)

	// THEN we log this unknown message Type
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "Unknown RUNTIME_BUILD Type") {
		t.Errorf("Expecting an error about the unknown %q Type we used (%s). Got\n%s",
			*msg.Page, *msg.Type, output)
	}
}

func TestWebSocketFlagsInit(t *testing.T) {
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("%v",
			err)
	}
	var receivedMsg api_types.WebSocketMessage
	json.Unmarshal(p, &receivedMsg)
	if *receivedMsg.Page != msgPage {
		t.Fatalf("Received a response for Page %q. Expected %q",
			*receivedMsg.Page, msgPage)
	}
	if receivedMsg.FlagsData == nil {
		t.Fatalf("Didn't get any FlagsData in the message\n%v",
			receivedMsg)
	}
	if receivedMsg.FlagsData.LogLevel != cfg.Settings.GetLogLevel() {
		t.Errorf("Expected FlagsData.LogLevel to be %q, got %q\n%v",
			cfg.Settings.GetLogLevel(), receivedMsg.FlagsData.LogLevel, receivedMsg)
	}
}

func TestWebSocketFlagsUnknownType(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message with an unknown Type
	var (
		msgVersion int    = 1
		msgPage    string = "FLAGS"
		msgType    string = "UNKNOWN"
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
	time.Sleep(time.Second)

	// THEN we log this unknown message Type
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "Unknown FLAGS Type") {
		t.Errorf("Expecting an error about the unknown %q Type we used (%s). Got\n%s",
			*msg.Page, *msg.Type, output)
	}
}

func TestWebSocketConfigInitSettings(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		var receivedMsg api_types.WebSocketMessage
		json.Unmarshal(p, &receivedMsg)
		wantedType := "SETTINGS"
		if *receivedMsg.Page != msgPage ||
			*receivedMsg.Type != wantedType {
			t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
				*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
		} else {
			if receivedMsg.ConfigData == nil {
				t.Errorf("Didn't get any ConfigData in the message\n%v",
					receivedMsg)
			} else {
				if receivedMsg.ConfigData == nil ||
					receivedMsg.ConfigData.Settings == nil ||
					receivedMsg.ConfigData.Settings.Log.Level == nil {
					t.Errorf("Didn't receive ConfigData.Settings.Log.Level from\n%v",
						receivedMsg)
				} else if *receivedMsg.ConfigData.Settings.Log.Level != cfg.Settings.GetLogLevel() {
					t.Errorf("Expected ConfigData.Settings.Log.Level to be %q, got %q\n%v",
						cfg.Settings.GetLogLevel(), *receivedMsg.ConfigData.Settings.Log.Level, receivedMsg)
				}
			}
		}
	}
}

func TestWebSocketConfigUnknownType(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message with an unknown Type
	var (
		msgVersion int    = 1
		msgPage    string = "FLAGS"
		msgType    string = "UNKNOWN"
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
	time.Sleep(time.Second)

	// THEN we log this unknown message Type
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "Unknown FLAGS Type") {
		t.Errorf("Expecting an error about the unknown %q Type we used (%s). Got\n%s",
			*msg.Page, *msg.Type, output)
	}
}

func TestWebSocketConfigInitDefaults(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		wantedType := "DEFAULTS"
		if *receivedMsg.Page != msgPage ||
			*receivedMsg.Type != wantedType {
			t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
				*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
		} else {
			if receivedMsg.ConfigData == nil ||
				receivedMsg.ConfigData.Defaults == nil ||
				receivedMsg.ConfigData.Defaults.Service.Interval == nil {
				t.Errorf("Didn't receive ConfigData.Defaults.Service.Interval from\n%v",
					receivedMsg)
			} else if *receivedMsg.ConfigData.Defaults.Service.Interval != *cfg.Defaults.Service.Interval {
				t.Errorf("Expected ConfigData.Defaults.Service.Interval to be %q, got %q\n%v",
					*cfg.Defaults.Service.Interval, *receivedMsg.ConfigData.Defaults.Service.Interval, receivedMsg)
			}
		}
	}
}

func TestWebSocketConfigInitNotify(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		wantedType := "NOTIFY"
		if *receivedMsg.Page != msgPage ||
			*receivedMsg.Type != wantedType {
			t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
				*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
		} else {
			if receivedMsg.ConfigData == nil ||
				receivedMsg.ConfigData.Notify == nil ||
				(*receivedMsg.ConfigData.Notify)["discord"] == nil {
				t.Errorf("Didn't receive ConfigData.Notify.discord from\n%v",
					receivedMsg)
			} else if (*receivedMsg.ConfigData.Notify)["discord"].Options["message"] != cfg.Notify["discord"].Options["message"] {
				t.Errorf("Expected ConfigData.Notify.discord.Options.message to be %q, got %q\n%v",
					cfg.Notify["discord"].Options["message"], (*receivedMsg.ConfigData.Notify)["discord"].Options["message"], receivedMsg)
			}
		}
	}
}

func TestWebSocketConfigInitWebHook(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		wantedType := "WEBHOOK"
		if *receivedMsg.Page != msgPage ||
			*receivedMsg.Type != wantedType {
			t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
				*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
		} else {
			if receivedMsg.ConfigData == nil ||
				receivedMsg.ConfigData.WebHook == nil ||
				(*receivedMsg.ConfigData.WebHook)["pass"] == nil {
				t.Errorf("Didn't receive ConfigData.WebHook.pass from\n%v",
					receivedMsg)
			} else if *(*receivedMsg.ConfigData.WebHook)["pass"].URL != *cfg.WebHook["pass"].URL {
				t.Errorf("Expected ConfigData.WebHook.pass.URL to be %q, got %q\n%v",
					*cfg.WebHook["pass"].URL, *(*receivedMsg.ConfigData.WebHook)["pass"].URL, receivedMsg)
			}
		}
	}
}

func TestWebSocketConfigInitServiceData(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		ws.ReadMessage()
	}
	{ // SERVICE
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		wantedType := "SERVICE"
		if *receivedMsg.Page != msgPage ||
			*receivedMsg.Type != wantedType {
			t.Errorf("Received a response for Page %s-%s. Expected %s-%s",
				*receivedMsg.Page, *receivedMsg.Type, msgPage, wantedType)
		} else {
			if receivedMsg.ConfigData == nil ||
				receivedMsg.ConfigData.Service == nil ||
				(*receivedMsg.ConfigData.Service)["test"] == nil {
				t.Errorf("Didn't receive ConfigData.Service.test from\n%v",
					receivedMsg)
			}
		}
	}
}

func TestWebSocketConfigInitServiceService(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		ws.ReadMessage()
	}
	{ // SERVICE
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
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
	}
}

func TestWebSocketConfigInitServiceDeployedVersion(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		ws.ReadMessage()
	}
	{ // SERVICE
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		receivedTestService := (*receivedMsg.ConfigData.Service)["test"]
		cfgTestService := cfg.Service["test"]
		// deployed version lookup
		{
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
	}
}

func TestWebSocketConfigInitServiceServiceURLCommands(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		ws.ReadMessage()
	}
	{ // SERVICE
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		receivedTestService := (*receivedMsg.ConfigData.Service)["test"]
		cfgTestService := cfg.Service["test"]
		// url commands
		{
			if receivedTestService.URLCommands == nil {
				t.Errorf("ConfigData.Service.test.URLCommands should've been %v, got %v",
					*cfgTestService.URLCommands, receivedTestService.URLCommands)
			}
			if *(*receivedTestService.URLCommands)[0].Regex != *(*cfgTestService.URLCommands)[0].Regex {
				t.Errorf("ConfigData.Service.test.URLCommands[0].Regex should've been %q, got %q",
					*(*cfgTestService.URLCommands)[0].Regex, *(*receivedTestService.URLCommands)[0].Regex)
			}
		}
	}
}

func TestWebSocketConfigInitServiceServiceNotify(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		ws.ReadMessage()
	}
	{ // SERVICE
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		receivedTestService := (*receivedMsg.ConfigData.Service)["test"]
		cfgTestService := cfg.Service["test"]
		// notify
		{
			if receivedTestService.Notify == nil {
				t.Errorf("ConfigData.Service.test.Notify should've been %v, got %v",
					*cfgTestService.Notify, receivedTestService.Notify)
			}
			if (*receivedTestService.Notify)["test"].Options["message"] != (*cfgTestService.Notify)["test"].Options["message"] {
				t.Errorf("ConfigData.Service.test.Notify.test.Options.message should've been %q, got %q",
					(*cfgTestService.Notify)["test"].Options["message"], (*receivedTestService.Notify)["test"].Options["message"])
			}
		}
	}
}

func TestWebSocketConfigInitServiceServiceWebhook(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		ws.ReadMessage()
	}
	{ // SERVICE
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		receivedTestService := (*receivedMsg.ConfigData.Service)["test"]
		cfgTestService := cfg.Service["test"]
		// webhook
		{
			if receivedTestService.WebHook == nil {
				t.Errorf("ConfigData.Service.test.WebHook should've been %v, got %v",
					*cfgTestService.URLCommands, receivedTestService.URLCommands)
			}
			if *(*receivedTestService.WebHook)["test"].URL != *(*cfgTestService.WebHook)["test"].URL {
				t.Errorf("ConfigData.Service.test.WebHook.pass.URL should've been %q, got %q",
					*(*cfgTestService.WebHook)["test"].URL, *(*receivedTestService.WebHook)["test"].URL)
			}
		}
	}
}

func TestWebSocketConfigInitServiceCommand(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
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
	time.Sleep(time.Second)

	// THEN it passes and we only receive a response with the WebHooks
	{ // SETTINGS
		ws.ReadMessage()
	}
	{ // DEFAULTS
		ws.ReadMessage()
	}
	{ // NOTIFY
		ws.ReadMessage()
	}
	{ // WEBHOOK
		ws.ReadMessage()
	}
	{ // SERVICE
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v",
				err)
		}
		receivedMsg := api_types.WebSocketMessage{}
		json.Unmarshal(p, &receivedMsg)
		receivedTestService := (*receivedMsg.ConfigData.Service)["test"]
		cfgTestService := cfg.Service["test"]
		// command
		{
			if receivedTestService.Command == nil {
				t.Errorf("ConfigData.Service.test.Command should've been %v, got %v",
					*cfgTestService.Command, receivedTestService.Command)
			}
			got := strings.Join((*receivedTestService.Command)[0], " ")
			if got != (*cfgTestService.Command)[0].String() {
				t.Errorf("ConfigData.Service.test.Command[0] should've been %q, got %q",
					(*cfgTestService.Command)[0].String(), got)
			}
		}
	}
}

func TestWebSocketUnknownPage(t *testing.T) {
	// GIVEN WebSocket server is running and we're connected to it
	ws := connectToWebSocket(t)
	defer ws.Close()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we send a message with an unknown Page
	var (
		msgVersion int    = 1
		msgPage    string = "UNKNOWN"
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
	time.Sleep(time.Second)

	// THEN we log this unknown message Type
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "Unknown PAGE ") {
		t.Errorf("Expecting an error about the unknown %q Page we used (%s). Got\n%s",
			*msg.Page, *msg.Type, output)
	}
}
