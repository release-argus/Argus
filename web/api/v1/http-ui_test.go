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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
	config_test "github.com/release-argus/Argus/config/test"
)

func TestHTTP_SetupRoutesNodeJS(t *testing.T) {
	// GIVEN: an API with NodeJS routes.
	tests := []struct {
		name        string
		route       string
		wantStatus  int
		wantContent string
	}{
		{
			name:        "account tokens route",
			route:       "/account/tokens",
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
			cfg := config_test.BareConfig(t, true)
			api, _ := NewAPI(cfg)
			api.SetupRoutesNodeJS()
			ts := httptest.NewServer(api.Router)
			t.Cleanup(ts.Close)
			client := http.Client{}

			// WHEN: a HTTP request is made to this router.
			req, err := http.NewRequest(http.MethodGet, ts.URL+tc.route, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			prefix := fmt.Sprintf(
				"%s\nAPI.SetupRoutesNodeJS() %q",
				packageName, req.URL.Path,
			)

			// THEN: the status code is as expected.
			if got := resp.StatusCode; got != tc.wantStatus {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantStatus,
				)
			}

			// AND: the content type is as expected.
			if tc.wantContent != "" {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tc.wantContent) {
					t.Errorf(
						"%s Content-Type mismatch\ngot:  %q\nwant: %q",
						prefix, contentType, tc.wantContent,
					)
				}
			}
		})
	}
}

func TestHTTP_SetupRoutesNodeJS__routePrefix(t *testing.T) {
	// GIVEN: an API served under a route prefix.
	tests := []struct {
		name  string
		route string
	}{
		// The prefix's own root is where a visitor lands first, and used to
		// fall through to the catch-all, which serves the shell unrewritten.
		{name: "prefix root", route: "/test"},
		{name: "prefix root/with a trailing slash", route: "/test/"},
		{name: "a page", route: "/test/approvals"},
		{name: "a nested page", route: "/test/admin/users"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config_test.BareConfig(t, true)
			cfg.Settings.Web.RoutePrefix = "/test"
			api, _ := NewAPI(cfg)
			api.SetupRoutesNodeJS()
			ts := httptest.NewServer(api.BaseRouter)
			t.Cleanup(ts.Close)

			// WHEN: the route is requested.
			resp, err := http.Get(ts.URL + tc.route)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			prefix := fmt.Sprintf(
				"%s\nAPI.SetupRoutesNodeJS() %q",
				packageName, tc.route,
			)

			// THEN: the shell is served.
			if resp.StatusCode != http.StatusOK {
				t.Fatalf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, resp.StatusCode, http.StatusOK,
				)
			}

			// AND: its <base href> points at the prefix, so the relative asset
			// URLs inside resolve under it rather than at the domain root.
			want := `<base href="/test/" />`
			if !bytes.Contains(body, []byte(want)) {
				t.Errorf(
					"%s base href mismatch\nwant to contain: %q\ngot:             %q",
					prefix, want, body,
				)
			}
		})
	}
}

func TestHTTP_SetupRoutesNodeJS__routePrefixAssets(t *testing.T) {
	// GIVEN: an API served under a route prefix. The shell's asset URLs resolve
	// against its <base href>, so they arrive prefixed and must serve from there.
	cfg := config_test.BareConfig(t, true)
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

			// WHEN: the asset is requested.
			resp, err := http.Get(ts.URL + tc.route)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			prefix := fmt.Sprintf(
				"%s\nAPI.SetupRoutesNodeJS() %q",
				packageName, tc.route,
			)

			// THEN: the status code is as expected.
			if resp.StatusCode != tc.wantStatus {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, resp.StatusCode, tc.wantStatus,
				)
			}

			// AND: it is the asset, not the shell.
			if tc.wantContentType != "" {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tc.wantContentType) {
					t.Errorf(
						"%s Content-Type mismatch\ngot:  %q\nwant: %q",
						prefix, contentType, tc.wantContentType,
					)
				}
			}
		})
	}
}

func TestHTTP_SetupRoutesFavicon(t *testing.T) {
	// GIVEN: an API with/without favicon overrides.
	tests := []struct {
		name           string
		favicon        *config.FaviconSettings
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

			cfg := config_test.BareConfig(t, true)
			cfg.Settings.Web.Favicon = testFaviconSettings(tc.urlPNG, tc.urlSVG)
			api, _ := NewAPI(cfg)
			api.SetupRoutesFavicon()
			ts := httptest.NewServer(api.Router)
			t.Cleanup(ts.Close)
			client := http.Client{}

			// WHEN: a HTTP request is made to this router (apple-touch-icon.png).
			req, err := http.NewRequest(
				http.MethodGet,
				ts.URL+"/apple-touch-icon.png",
				nil,
			)
			if err != nil {
				t.Fatalf(
					"%s\n%v",
					packageName, err,
				)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf(
					"%s\n%v",
					packageName, err,
				)
			}

			prefix := fmt.Sprintf("%s\n/apple-touch-icon.png", packageName)

			// THEN: the status code is as expected.
			wantStatus := http.StatusNotFound
			if tc.urlPNG != "" {
				wantStatus = http.StatusOK
			}
			if resp.StatusCode != wantStatus {
				t.Errorf(
					"%s - status code mismatch\ngot:  %d\nwant: %d",
					prefix, resp.StatusCode, wantStatus,
				)
			}
			if got := resp.Request.URL.String(); tc.urlPNG != "" && got != tc.urlPNG {
				t.Errorf(
					"%s - redirect mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.urlPNG,
				)
			}

			// WHEN: a HTTP request is made to this router (favicon.svg).
			req, err = http.NewRequest(http.MethodGet, ts.URL+"/favicon.svg", nil)
			if err != nil {
				t.Fatalf(
					"%s\n%v",
					packageName, err,
				)
			}
			resp, err = client.Do(req)
			if err != nil {
				t.Fatalf(
					"%s\n%v",
					packageName, err,
				)
			}

			prefix = fmt.Sprintf("%s\n/favicon.svg", packageName)

			// THEN: the status code is as expected.
			wantStatus = http.StatusNotFound
			if tc.urlSVG != "" {
				wantStatus = http.StatusOK
			}
			if resp.StatusCode != wantStatus {
				t.Errorf(
					"%s - status code mismatch\ngot:  %d\nwant: %d",
					prefix, resp.StatusCode, wantStatus,
				)
			}
			if got := resp.Request.URL.String(); tc.urlSVG != "" && got != tc.urlSVG {
				t.Errorf(
					"%s - redirect mismatch\ngot:  %s\nwant: %s",
					prefix, got, tc.urlSVG,
				)
			}
		})
	}
}
