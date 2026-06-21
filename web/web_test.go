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

//go:build integration

package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/release-argus/Argus/util"
)

var router *mux.Router

func TestWebSocketHandler(t *testing.T) {
	// GIVEN: a WebSocket is running (TestMain) and we have the URL.
	url := fmt.Sprintf("ws://%s:%s/ws", host, port)

	t.Run("ConnectWebSocket", func(t *testing.T) {
		// WHEN: we attempt to connect.
		ws, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("failed to connect to WebSocket: %v (HTTP status %d)", err, resp.StatusCode)
		}
		defer ws.Close()

		// THEN: connection should be open.
		if ws.NetConn() == nil {
			t.Errorf("expected WebSocket underlying connection to be non-nil")
		}

		// Optional: test sending/receiving a message
		testMsg := "hello"
		if err := ws.WriteMessage(websocket.TextMessage, []byte(testMsg)); err != nil {
			t.Errorf("failed to send message: %v", err)
		}
	})
}

func TestRun_Error(t *testing.T) {
	// GIVEN: a config that wants to use a port that is already in use (TestMain).
	cfg := testConfig(t, filepath.Join(t.TempDir(), "config.yml"))
	cfg.Settings.Web.ListenHost = host
	cfg.Settings.Web.ListenPort = port

	// WHEN: Run is called.
	errChan := make(chan error, 1)
	go func() {
		errChan <- Run(t.Context(), cfg)
	}()

	prefix := fmt.Sprintf("%s\nRun()", packageName)

	// THEN: it should return an error since port already in use.
	select {
	case err := <-errChan:
		if err == nil {
			t.Fatalf(
				"%s should have failed as port was in use\ngot:  nil\nwant: error",
				prefix,
			)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf(
			"%s should have failed as port was in use\ngot:  timeout waiting for Run to return\nwant: error",
			prefix,
		)
	}
}

func TestMainWithRoutePrefix(t *testing.T) {
	// GIVEN: a valid config with a Service.
	cfg := testConfig(t, filepath.Join(t.TempDir(), "config.yml"))
	cfg.Settings.Web.RoutePrefix = "/test"

	// AND: a cancellable context for shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WHEN: the Web UI is started with this Config.
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfg)
	}()

	// THEN: Web UI is accessible.
	url := fmt.Sprintf(
		"http://localhost:%s%s/metrics",
		cfg.Settings.Web.ListenPort, cfg.Settings.Web.RoutePrefix,
	)
	if err := waitForServer(url, time.Second); err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf(
			"%s\nError making request: %s",
			packageName, err,
		)
	}
	wantStatusCode := http.StatusOK
	if resp.StatusCode != wantStatusCode {
		t.Errorf(
			"%s\nRun(%q) status code mismatch\ngot:  %d\nwant: %d",
			packageName, url,
			resp.StatusCode, wantStatusCode,
		)
	}

	// AND: the server shuts down cleanly.
	assertServerShutdown(t, cancel, errCh, url)
}

func TestWebAccessible(t *testing.T) {
	// GIVEN: we have the Web UI Router from TestMain().
	tests := []struct {
		path      string
		bodyRegex string
	}{
		{
			path: "/approvals",
		},
		{
			path:      "/metrics",
			bodyRegex: `go_gc_duration_`},
		{
			path:      "/api/v1/healthcheck",
			bodyRegex: `^Alive$`},
		{
			path: "/api/v1/version",
			bodyRegex: fmt.Sprintf(`"goVersion":"%s"`,
				util.GoVersion,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			// WHEN: we make a request to path.
			req, _ := http.NewRequest(http.MethodGet, tc.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			// THEN: we get a Status OK.
			wantStatusCode := http.StatusOK
			if response.Code != wantStatusCode {
				t.Errorf(
					"%s\nGET %q status code mismatch\ngot:  %d\nwant: %d",
					packageName, tc.path,
					response.Code, wantStatusCode,
				)
			}

			// AND: the body matches the expected string RegEx.
			if tc.bodyRegex != "" {
				body := response.Body.String()
				if !util.RegexCheck(tc.bodyRegex, body) {
					t.Errorf(
						"%s\nGET %q body mismatch\ngot:  %q\nwant: %q",
						packageName, tc.path,
						body, tc.bodyRegex,
					)
				}
			}
		})
	}
}

func TestAccessibleHTTPS(t *testing.T) {
	// GIVEN: a bunch of URLs to test and the webserver is running with HTTPS.
	tests := []struct {
		name      string
		path      string
		bodyRegex string
	}{
		{
			path: "/approvals",
		},
		{
			path:      "/metrics",
			bodyRegex: `go_gc_duration_`},
		{
			path:      "/api/v1/healthcheck",
			bodyRegex: `^Alive$`},
		{
			path: "/api/v1/version",
			bodyRegex: fmt.Sprintf(`"goVersion":"%s"`,
				util.GoVersion,
			),
		},
	}
	cfg := testConfig(t, filepath.Join(t.TempDir(), "config.yml"))
	cfg.Settings.Web.CertFile = "TestAccessibleHTTPS_Cert.pem"
	cfg.Settings.Web.KeyFile = "TestAccessibleHTTPS_Key.pem"
	_ = generateCertFiles(cfg.Settings.Web.CertFile, cfg.Settings.Web.KeyFile)
	t.Cleanup(func() {
		_ = os.Remove(cfg.Settings.Web.CertFile)
		_ = os.Remove(cfg.Settings.Web.KeyFile)
	})

	address := fmt.Sprintf("https://localhost:%s", cfg.Settings.Web.ListenPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router = newWebUI(cfg)
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfg)
	}()

	if err := waitForServer(address, time.Second); err != nil {
		t.Fatal(err)
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			// WHEN: we make a HTTPS request to path.
			req, _ := http.NewRequest(http.MethodGet, address+tc.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			// THEN: we get a Status OK.
			wantStatusCode := http.StatusOK
			if response.Code != wantStatusCode {
				t.Errorf(
					"%s\nGET %q status code mismatch\ngot:  %d\nwant: %d",
					packageName, address+tc.path,
					response.Code, wantStatusCode,
				)
			}

			// AND: the body matches the expected string RegEx.
			if tc.bodyRegex != "" {
				body := response.Body.String()
				if !util.RegexCheck(tc.bodyRegex, body) {
					t.Errorf(
						"%s\nGET %q body mismatch\ngot:  %q\nwant: %q",
						packageName, address+tc.path,
						body, tc.bodyRegex,
					)
				}
			}
		})
	}

	// AND: the server shuts down cleanly.
	assertServerShutdown(t, cancel, errCh, address)
}
