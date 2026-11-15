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
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	config_test "github.com/release-argus/Argus/config/test"
)

func TestHTTP_SetupRoutesNodeJS(t *testing.T) {
	tests := []struct {
		name        string
		route       string
		wantStatus  int
		wantContent string
	}{
		{
			name:        "account profile route",
			route:       "/account/profile",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "account tokens route",
			route:       "/account/tokens",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "account defaults route",
			route:       "/account/defaults",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "admin users route",
			route:       "/admin/users",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "nested route/its bare parent is served too",
			route:       "/admin",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "approvals route",
			route:       "/approvals",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "config route",
			route:       "/config",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "flags route",
			route:       "/flags",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "status route",
			route:       "/status",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "catch-all route/file not found",
			route:       "/some/random/path",
			wantStatus:  http.StatusNotFound,
			wantContent: "text/plain",
		},
		{
			name:        "catch-all route/file exists",
			route:       "/robots.txt",
			wantStatus:  http.StatusOK,
			wantContent: "text/plain",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: resets the config flag vars.

			cfg := config_test.BareConfig(true)
			api, _ := NewAPI(cfg)
			api.SetupRoutesNodeJS()
			ts := httptest.NewServer(api.Router)
			t.Cleanup(ts.Close)
			client := http.Client{}

			req, err := http.NewRequest(http.MethodGet, ts.URL+tc.route, nil)
			if err != nil {
				t.Fatalf("http.NewRequest() failed: %v", err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Client.Do() failed: %v", err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}

			if got, want := resp.StatusCode, tc.wantStatus; got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}

			if tc.wantContent != "" {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tc.wantContent) {
					t.Errorf(
						"Content-Type header mismatch\ngot:  %q\nwant: %q",
						contentType, tc.wantContent,
					)
				}
			}
		})
	}
}

func TestHTTP_SetupRoutesNodeJS__routePrefix(t *testing.T) {
	tests := []struct {
		name  string
		route string
	}{
		{name: "prefix root", route: "/test"},
		{name: "prefix root/with a trailing slash", route: "/test/"},
		{name: "a page", route: "/test/approvals"},
		{name: "a nested page", route: "/test/admin/users"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: resets the config flag vars.

			cfg := config_test.BareConfig(true)
			cfg.Settings.Web.RoutePrefix = "/test"
			api, _ := NewAPI(cfg)
			api.SetupRoutesNodeJS()
			ts := httptest.NewServer(api.BaseRouter)
			t.Cleanup(ts.Close)

			resp, err := http.Get(ts.URL + tc.route)
			if err != nil {
				t.Fatalf("http.Get() failed: %v", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}

			if got, want := resp.StatusCode, http.StatusOK; got != want {
				t.Fatalf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}

			// Relative asset URLs resolve under <base href>, not the domain root.
			gotBase := regexp.MustCompile(`<base[^>]*>`).Find(body)
			if want := `<base href="/test/" />`; string(gotBase) != want {
				t.Errorf("base href mismatch\ngot:  %q\nwant: %q", gotBase, want)
			}
		})
	}
}

func TestHTTP_SetupRoutesNodeJS__routePrefixAssets(t *testing.T) {
	// The shell's asset URLs resolve against its <base href>, so arrive prefixed.
	cfg := config_test.BareConfig(true)
	cfg.Settings.Web.RoutePrefix = "/test"
	api, _ := NewAPI(cfg)
	api.SetupRoutesNodeJS()

	tests := []struct {
		name            string
		route           string
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "known asset/under the prefix",
			route:           "/test/robots.txt",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain",
		},
		{
			name:       "known asset/no prefix",
			route:      "/robots.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unknown asset/under the prefix",
			route:      "/test/something.js",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unknown asset/no prefix",
			route:      "/something.js",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(api.BaseRouter)
			t.Cleanup(ts.Close)

			resp, err := http.Get(ts.URL + tc.route)
			if err != nil {
				t.Fatalf("http.Get() failed: %v", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}

			if got, want := resp.StatusCode, tc.wantStatus; got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}

			// Content-Type tells the asset from the HTML shell.
			if tc.wantContentType != "" {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tc.wantContentType) {
					t.Errorf(
						"Content-Type header mismatch\ngot:  %q\nwant: %q",
						contentType, tc.wantContentType,
					)
				}
			}
		})
	}
}

func TestHTTP_SetupRoutesFavicon(t *testing.T) {
	tests := []struct {
		name           string
		urlPNG, urlSVG string
	}{
		{
			name:   "no override",
			urlPNG: "",
			urlSVG: "",
		},
		{
			name:   "override png",
			urlPNG: "https://release-argus.io/demo/apple-touch-icon.png",
		},
		{
			name:   "override svg",
			urlSVG: "https://release-argus.io/demo/favicon.svg",
		},
		{
			name:   "override png and svg",
			urlPNG: "https://release-argus.io/demo/apple-touch-icon.png",
			urlSVG: "https://release-argus.io/demo/favicon.svg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: resets the config flag vars.

			cfg := config_test.BareConfig(true)
			cfg.Settings.Web.Favicon = testFaviconSettings(tc.urlPNG, tc.urlSVG)
			api, _ := NewAPI(cfg)
			api.SetupRoutesFavicon()
			ts := httptest.NewServer(api.Router)
			t.Cleanup(ts.Close)
			client := http.Client{
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			for _, asset := range []struct{ path, wantLocation string }{
				{path: "/apple-touch-icon.png", wantLocation: tc.urlPNG},
				{path: "/favicon.svg", wantLocation: tc.urlSVG},
			} {
				t.Run(strings.TrimPrefix(asset.path, "/"), func(t *testing.T) {
					req, err := http.NewRequest(http.MethodGet, ts.URL+asset.path, nil)
					if err != nil {
						t.Fatalf("http.NewRequest() failed: %v", err)
					}
					resp, err := client.Do(req)
					if err != nil {
						t.Fatalf("Client.Do() failed: %v", err)
					}
					t.Cleanup(func() { _ = resp.Body.Close() })
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						t.Fatalf("io.ReadAll() failed: %v", err)
					}

					// Only an overridden asset is routed; any other 404s.
					wantStatus := http.StatusNotFound
					if asset.wantLocation != "" {
						wantStatus = http.StatusPermanentRedirect
					}
					if got, want := resp.StatusCode, wantStatus; got != want {
						t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
					}

					if got, want := resp.Header.Get("Location"), asset.wantLocation; got != want {
						t.Errorf("Location header mismatch\ngot:  %q\nwant: %q", got, want)
					}
				})
			}
		})
	}
}
