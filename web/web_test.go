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

//go:build integration

package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/util"
)

var router *mux.Router

func TestMainWithRoutePrefix(t *testing.T) {
	// GIVEN a valid config with a Service
	cfg := testConfig("TestMainWithRoutePrefix.yml", t)
	cfg.Settings.Web.RoutePrefix = "/test"

	// WHEN the Web UI is started with this Config
	go Run(cfg)
	time.Sleep(500 * time.Millisecond)

	// THEN Web UI is accessible
	url := fmt.Sprintf("http://localhost:%s%s/metrics",
		cfg.Settings.Web.ListenPort, cfg.Settings.Web.RoutePrefix)
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
			bodyRegex: `go_gc_duration_`},
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
				if !util.RegexCheck(tc.bodyRegex, body) {
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
			bodyRegex: `go_gc_duration_`},
		"/api/v1/healthcheck": {
			path:      "/api/v1/healthcheck",
			bodyRegex: fmt.Sprintf(`^Alive$`)},
		"/api/v1/version": {
			path: "/api/v1/version",
			bodyRegex: fmt.Sprintf(`"goVersion":"%s"`,
				util.GoVersion)},
	}
	cfg := testConfig("TestAccessibleHTTPS.yml", t)
	cfg.Settings.Web.CertFile = "TestAccessibleHTTPS_cert.pem"
	cfg.Settings.Web.KeyFile = "TestAccessibleHTTPS_key.pem"
	generateCertFiles(cfg.Settings.Web.CertFile, cfg.Settings.Web.KeyFile)
	t.Cleanup(func() {
		os.Remove(cfg.Settings.Web.CertFile)
		os.Remove(cfg.Settings.Web.KeyFile)
	})

	router = newWebUI(cfg)
	go Run(cfg)
	time.Sleep(250 * time.Millisecond)
	address := fmt.Sprintf("https://localhost:%s", cfg.Settings.Web.ListenPort)

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
				if !util.RegexCheck(tc.bodyRegex, body) {
					t.Errorf("expected %q in body\ngot: %q",
						tc.bodyRegex, response.Body.String())
				}
			}
		})
	}
}
