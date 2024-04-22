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

//go:build integration

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

var router *mux.Router

func TestMainWithRoutePrefix(t *testing.T) {
	// GIVEN a valid config with a Service
	cfg := testConfig("TestMainWithRoutePrefix.yml", nil, t)
	*cfg.Settings.Web.RoutePrefix = "/test"

	// WHEN the Web UI is started with this Config
	go Run(cfg, nil)
	time.Sleep(500 * time.Millisecond)

	// THEN Web UI is accessible
	url := fmt.Sprintf("http://localhost:%s%s/metrics",
		*cfg.Settings.Web.ListenPort, *cfg.Settings.Web.RoutePrefix)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error making request: %s", err)
	}
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
	cfg := testConfig("TestAccessibleHTTPS.yml", nil, t)
	cfg.Settings.Web.CertFile = test.StringPtr("TestAccessibleHTTPS_cert.pem")
	cfg.Settings.Web.KeyFile = test.StringPtr("TestAccessibleHTTPS_key.pem")
	generateCertFiles(*cfg.Settings.Web.CertFile, *cfg.Settings.Web.KeyFile)
	defer os.Remove(*cfg.Settings.Web.CertFile)
	defer os.Remove(*cfg.Settings.Web.KeyFile)

	router = newWebUI(cfg)
	go Run(cfg, nil)
	time.Sleep(250 * time.Millisecond)
	address := fmt.Sprintf("https://localhost:%s", *cfg.Settings.Web.ListenPort)

	for name, tc := range tests {
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
	start := time.Now()
	for i := 0; i < 120; i++ {
		time.Sleep(50 * time.Millisecond)
		if len(errChan) != 0 {
			break
		}
	}
	if len(errChan) == 0 {
		t.Logf("No messages received after %s", time.Since(start))
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
			stdoutRegex: `^DEBUG:[^:]+READ \{"key": "value"\}\nVERBOSE: WebSocket`},
		"no version, unknown type": {
			msg:         `{"page": "APPROVALS", "type": "SHAZAM", "key": "value"}`,
			stdoutRegex: "Unknown TYPE"},
		"invalid JSON": {
			msg:         `{"version": 1, "key": "value"`,
			stdoutRegex: "missing/invalid version key"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			ws := connectToWebSocket(t)

			// WHEN we send a message
			if err := ws.WriteMessage(websocket.TextMessage, []byte(tc.msg)); err != nil {
				t.Errorf("%v",
					err)
			}
			time.Sleep(50 * time.Millisecond)

			// THEN we receive the expected response
			ws.Close()
			time.Sleep(250 * time.Millisecond)
			stdout := releaseStdout()
			re := regexp.MustCompile(tc.stdoutRegex)
			match := re.MatchString(stdout)
			if !match {
				t.Errorf("match on %q not found in\n\n%s",
					tc.stdoutRegex, stdout)
			}
		})
	}
}
