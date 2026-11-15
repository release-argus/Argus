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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/release-argus/Argus/util"
	v1 "github.com/release-argus/Argus/web/api/v1"
)

func TestWebSocketHandler(t *testing.T) {
	url := fmt.Sprintf("ws://%s:%s/ws", host, port)

	t.Run("ConnectWebSocket", func(t *testing.T) {
		ws, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			status := 0
			if resp != nil {
				status = resp.StatusCode
			}
			t.Fatalf("Dialer.Dial(%q) failed: %v (HTTP status %d)", url, err, status)
		}
		defer ws.Close()

		if ws.NetConn() == nil {
			t.Error("Conn.NetConn() mismatch\ngot:  nil\nwant: non-nil")
		}

		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Errorf("Conn.WriteMessage() failed: %v", err)
		}
	})
}

func TestRun_Error(t *testing.T) {
	cfg := testConfig(t, filepath.Join(t.TempDir(), "config.yml"))
	cfg.Settings.Web.ListenHost = host
	// TestMain's live server already holds port, so Run cannot listen.
	cfg.Settings.Web.ListenPort = port

	errChan := make(chan error, 1)
	go func() {
		errChan <- Run(t.Context(), cfg, nil)
	}()

	select {
	case err := <-errChan:
		if err == nil {
			t.Fatal("Run() error mismatch\ngot:  nil\nwant: non-nil")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Run() error mismatch\ngot:  timeout after 1s\nwant: non-nil")
	}
}

func TestMainWithRoutePrefix(t *testing.T) {
	cfg := testConfig(t, filepath.Join(t.TempDir(), "config.yml"))
	cfg.Settings.Web.RoutePrefix = "/test"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfg, nil)
	}()

	url := fmt.Sprintf(
		"http://localhost:%s%s/metrics",
		cfg.Settings.Web.ListenPort, cfg.Settings.Web.RoutePrefix,
	)
	if err := waitForServer(url, time.Second); err != nil {
		t.Fatalf("waitForServer() failed: %v", err)
	}
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http.Get(%q) failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() failed: %v", err)
	}
	wantStatusCode := http.StatusOK
	if resp.StatusCode != wantStatusCode {
		t.Errorf(
			"status code mismatch\ngot:  %d - %q\nwant: %d",
			resp.StatusCode, body, wantStatusCode,
		)
	}

	assertServerShutdown(t, cancel, errCh, url)
}

func TestWebAccessible(t *testing.T) {
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
			path:      "/api/v1/version",
			bodyRegex: fmt.Sprintf(`"goVersion":"%s"`, util.GoVersion),
		},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, tc.path, nil)
			if err != nil {
				t.Fatalf("http.NewRequest() failed: %v", err)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			wantStatusCode := http.StatusOK
			if response.Code != wantStatusCode {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					response.Code, response.Body.String(), wantStatusCode,
				)
			}

			if tc.bodyRegex != "" {
				body := response.Body.String()
				if !util.RegexCheck(tc.bodyRegex, body) {
					t.Errorf("body mismatch\ngot:  %q\nwant: %q", body, tc.bodyRegex)
				}
			}
		})
	}
}

func TestAccessibleHTTPS(t *testing.T) {
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
			path:      "/api/v1/version",
			bodyRegex: fmt.Sprintf(`"goVersion":"%s"`, util.GoVersion),
		},
	}
	cfg := testConfig(t, filepath.Join(t.TempDir(), "config.yml"))
	cfg.Settings.Web.CertFile = t.Name() + "_Cert.pem"
	cfg.Settings.Web.KeyFile = t.Name() + "_Key.pem"
	t.Cleanup(func() {
		_ = os.Remove(cfg.Settings.Web.CertFile)
		_ = os.Remove(cfg.Settings.Web.KeyFile)
	})
	generateCertFiles(t, cfg.Settings.Web.CertFile, cfg.Settings.Web.KeyFile)

	address := "https://localhost:" + cfg.Settings.Web.ListenPort

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfg, nil)
	}()

	if err := waitForServer(address, time.Second); err != nil {
		t.Fatalf("waitForServer() failed: %v", err)
	}

	client := insecureClient()
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			// Serial: shares the server, which the parent stops on return.

			resp, err := client.Get(address + tc.path)
			if err != nil {
				t.Fatalf("Client.Get() failed: %v", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}

			wantStatusCode := http.StatusOK
			if resp.StatusCode != wantStatusCode {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					resp.StatusCode, body, wantStatusCode,
				)
			}

			if tc.bodyRegex != "" {
				if !util.RegexCheck(tc.bodyRegex, string(body)) {
					t.Errorf("body mismatch\ngot:  %q\nwant: %q", body, tc.bodyRegex)
				}
			}
		})
	}

	assertServerShutdown(t, cancel, errCh, address)
}

func TestNewWebUI__auth(t *testing.T) {
	tests := []struct {
		name     string
		authDeps *v1.AuthDeps
		wantAuth bool
	}{
		{
			name:     "no deps/auth routes are not served",
			authDeps: nil,
			wantAuth: false,
		},
		{
			name:     "deps/auth routes are served",
			authDeps: &v1.AuthDeps{},
			wantAuth: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := testConfig(t, filepath.Join(t.TempDir(), "config.yml"))

			webUI := newWebUI(t.Context(), cfg, tc.authDeps)

			got := hasRoute(t, webUI, "/api/v1/auth/login")
			if got != tc.wantAuth {
				t.Errorf("hasRoute() mismatch\ngot:  %t\nwant: %t", got, tc.wantAuth)
			}
		})
	}
}
