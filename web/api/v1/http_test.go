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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

func TestHTTP_Version(t *testing.T) {
	// GIVEN an API and the Version,BuildDate and GoVersion vars defined
	api := API{}
	util.Version = "1.2.3"
	util.BuildDate = "2022-01-01T01:01:01Z"

	// WHEN a HTTP request is made to the httpVersion handler
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()
	api.httpVersion(w, req)
	res := w.Result()
	defer res.Body.Close()

	// THEN the version is returned in JSON format
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v",
			err)
	}
	var got api_type.VersionAPI
	json.Unmarshal(data, &got)
	want := api_type.VersionAPI{
		Version:   util.Version,
		BuildDate: util.BuildDate,
		GoVersion: util.GoVersion,
	}
	if got != want {
		t.Errorf("Version HTTP should have returned %v, not %v",
			want, got)
	}
}

func TestHTTP_BasicAuth(t *testing.T) {
	// GIVEN an API with/without Basic Auth credentials
	tests := map[string]struct {
		basicAuth *config.WebSettingsBasicAuth
		fail      bool
		noHeader  bool
	}{
		"No basic auth": {
			basicAuth: nil,
			fail:      false},
		"basic auth fail invalid creds": {
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234"},
			fail: true},
		"basic auth fail no Authorization header": {
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234"},
			noHeader: true,
			fail:     true},
		"basic auth pass": {
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "123"},
			fail: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			cfg := config.Config{}
			cfg.Settings.Web.BasicAuth = tc.basicAuth
			// Hash the username/password
			if cfg.Settings.Web.BasicAuth != nil {
				cfg.Settings.Web.BasicAuth.CheckValues()
			}
			cfg.Settings.Web.RoutePrefix = test.StringPtr("")
			api := NewAPI(&cfg, util.NewJLog("WARN", false))
			api.Router.HandleFunc("/test", func(rw http.ResponseWriter, req *http.Request) {
				return
			})
			ts := httptest.NewServer(api.BaseRouter)
			defer ts.Close()

			// WHEN a HTTP request is made to this router
			client := http.Client{}
			req, err := http.NewRequest("GET", ts.URL+"/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			if !tc.noHeader {
				req.Header = http.Header{
					// test:123
					"Authorization": {"Basic dGVzdDoxMjM="},
				}
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			// THEN the request passes only when expected
			got := resp.StatusCode
			want := 200
			if tc.fail {
				want = http.StatusUnauthorized
			}
			if got != want {
				t.Errorf("Expected a %d, not a %d",
					want, got)
			}
		})
	}
}

func TestHTTP_SetupRoutesFavicon(t *testing.T) {
	// GIVEN an API with/without favicon overrides
	tests := map[string]struct {
		favicon *config.FaviconSettings
		urlPNG  string
		urlSVG  string
	}{
		"no override": {
			urlPNG: "",
			urlSVG: "",
		},
		"override png": {
			urlPNG: "https://release-argus.io/demo/apple-touch-icon.png",
		},
		"override svg": {
			urlSVG: "https://release-argus.io/demo/favicon.svg",
		},
		"override png and svg": {
			urlPNG: "https://release-argus.io/demo/apple-touch-icon.png",
			urlSVG: "https://release-argus.io/demo/favicon.svg",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			cfg := test.BareConfig(true)
			cfg.Settings.Web.Favicon = testFaviconSettings(tc.urlPNG, tc.urlSVG)
			api := NewAPI(cfg, util.NewJLog("WARN", false))
			api.SetupRoutesFavicon()
			ts := httptest.NewServer(api.Router)
			defer ts.Close()
			client := http.Client{}

			// WHEN a HTTP request is made to this router (apple-touch-icon.png)
			req, err := http.NewRequest("GET", ts.URL+"/apple-touch-icon.png", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			// THEN the status code is as expected
			wantStatus := http.StatusNotFound
			if tc.urlPNG != "" {
				wantStatus = http.StatusOK
			}
			if resp.StatusCode != wantStatus {
				t.Errorf("/apple-touch-icon.png - Expected a %d, not a %d",
					wantStatus, resp.StatusCode)
			}
			if tc.urlPNG != "" && tc.urlPNG != resp.Request.URL.String() {
				t.Errorf("/apple-touch-icon.png - Expected a redirect to %s, not %s",
					tc.urlPNG, resp.Request.URL.String())
			}

			// WHEN a HTTP request is made to this router (favicon.svg)
			req, err = http.NewRequest("GET", ts.URL+"/favicon.svg", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err = client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			// THEN the status code is as expected
			wantStatus = http.StatusNotFound
			if tc.urlSVG != "" {
				wantStatus = http.StatusOK
			}
			if resp.StatusCode != wantStatus {
				t.Errorf("/favicon.svg - Expected a %d, not a %d",
					wantStatus, resp.StatusCode)
			}
			if tc.urlSVG != "" && tc.urlSVG != resp.Request.URL.String() {
				t.Errorf("/favicon.svg - Expected a redirect to %s, not %s",
					tc.urlSVG, resp.Request.URL.String())
			}
		})
	}
}

func TestHTTP_DisableRoutes(t *testing.T) {
	// GIVEN an API and a bunch of routes
	tests := map[string]struct {
		method             string
		path               string
		replaceLastPathDir string
		wantStatus         int
		wantBody           string
	}{
		"-config": {
			method:     http.MethodGet,
			path:       "config",
			wantStatus: http.StatusOK,
			wantBody: `{
				"settings":{.*},
				"defaults":{.*"notify":{.*},
				"webhook":{.*},
				"service":{.*
				}`,
		},
		"-runtime": {
			method:     http.MethodGet,
			path:       "status/runtime",
			wantStatus: http.StatusOK,
			wantBody: `{
				"start_time":"[^"]+",
				"cwd":"[^"]+",
				"goroutines":\d+,
				"GOMAXPROCS":\d+
			}`,
		},
		"-version": {
			method:     http.MethodGet,
			path:       "version",
			wantStatus: http.StatusOK,
			wantBody: `{
				"version":"[^"]*",
				"buildDate":"[^"]*",
				"goVersion":"[^"]*"
			}`,
		},
		"-flags": {
			method:     http.MethodGet,
			path:       "flags",
			wantStatus: http.StatusOK,
			wantBody: `{
				"log.level":"INFO",
				"log.timestamps":false,
				"data.database-file":"[^"]+",
				"web.listen-host":"[\d.]+",
				"web.listen-port":"\d+",
				"web.cert-file":null,
				"web.pkey-file":null
			`,
		},
		"-order": {
			method:     http.MethodGet,
			path:       "service/order",
			wantStatus: http.StatusOK,
			wantBody: `{
				"order": null
			}`,
		},
		"-service_summary": {
			method:             http.MethodGet,
			path:               "service/summary/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusNotFound,
			wantBody: `{
				"message":"service \\"[^"]+\\" not found"
			}`,
		},
		"-service_actions - GET": {
			method:             http.MethodGet,
			path:               "service/actions/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusNotFound,
			wantBody: `{
				"message":"service \\"[^"]+\\" not found"
			}`,
		},
		"service_actions": {
			method:             http.MethodPost,
			path:               "service/actions/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusNotFound,
			wantBody: `{
				"message":"service \\"[^"]+\\" not found"
			}`,
		},
		"-service_update - GET unspecific": {
			method:     http.MethodGet,
			path:       "service/update",
			wantStatus: http.StatusOK,
			wantBody: `{
				"hard_defaults":{.*},
				"defaults":{.*},
				"notify":{.*},
				"webhook":{.*}
			}`,
		},
		"-service_update - GET": {
			method:             http.MethodGet,
			path:               "service/update/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusNotFound,
			wantBody: `{
				"message":"service \\"[^"]+\\" not found"
			}`,
		},
		"lv_refresh_new": {
			method:     http.MethodGet,
			path:       "latest_version/refresh",
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"version":"","error":"values failed [^"]*".*
			}`,
		},
		"dv_refresh_new": {
			method:     http.MethodGet,
			path:       "deployed_version/refresh",
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"version":"","error":"values failed [^"]*".*
			}`,
		},
		"lv_refresh": {
			method:             http.MethodGet,
			path:               "latest_version/refresh/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusNotFound,
			wantBody: `{
				"message":"service \\"[^"]+\\" not found"
			}`,
		},
		"dv_refresh": {
			method:             http.MethodGet,
			path:               "deployed_version/refresh/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusNotFound,
			wantBody: `{
				"message":"service \\"[^"]+\\" not found"
			}`,
		},
		"notify_test": {
			method:     http.MethodPost,
			path:       "notify/test",
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"message":"unexpected end of JSON input"
			}`,
		},
		"service_update": {
			method:             http.MethodPut,
			path:               "service/update/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusBadRequest,
			wantBody: `{
				"message":"edit \\"[^"]+\\" failed[^"]*"
			}`,
		},
		"service_create": {
			method:     http.MethodPost,
			path:       "service/new",
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"message":"create \\"\\" failed[^"]*"
			}`,
		},
		"service_delete": {
			method:             http.MethodDelete,
			path:               "service/delete/{service_name:.+}",
			replaceLastPathDir: "test",
			wantStatus:         http.StatusBadRequest,
			wantBody: `{
				"message":"delete \\"[^"]+\\" failed[^"]*"
			}`,
		},
	}
	log := util.NewJLog("WARN", false)
	log.Testing = true
	disableCombinations := test.Combinations(util.SortedKeys(tests))

	// Split tests into groups
	groupSize := len(disableCombinations) / (runtime.NumCPU())
	numGroups := len(disableCombinations) / groupSize
	for i := 0; i < numGroups; i++ {
		groupStart := i * groupSize
		groupEnd := min((i+1)*groupSize, len(disableCombinations))
		group := disableCombinations[groupStart:groupEnd]

		t.Run(fmt.Sprintf("Group %d", i+1), func(t *testing.T) {
			t.Parallel()
			for j, disabledRoutes := range group {
				// Insane number of tests, so have to skip some
				if j%((len(tests)+1)*10) != 0 {
					continue
				}
				t.Run(strings.Join(disabledRoutes, ";"), func(t *testing.T) {

					cfg := test.BareConfig(false)
					cfg.Settings.Web.DisabledRoutes = disabledRoutes
					// Give every other test a route prefix
					routePrefix := ""
					if j%2 == 0 {
						routePrefix = "/test"
						cfg.Settings.Web.RoutePrefix = &routePrefix
					}
					api := NewAPI(cfg, log)
					api.SetupRoutesAPI()
					ts := httptest.NewServer(api.Router)
					ts.Config.Handler = api.Router
					defer ts.Close()
					client := http.Client{}

					// Test each route for this set of disabled routes
					for name, tc := range tests {
						if !strings.HasPrefix(name, "-") && util.Contains(disabledRoutes, name) {
							tc.wantStatus = http.StatusNotFound
							tc.wantBody = "Route disabled"
						} else {
							tc.wantBody = test.TrimJSON(tc.wantBody)
						}

						path := fmt.Sprintf("%s/api/v1/%s",
							routePrefix, tc.path)
						if tc.replaceLastPathDir != "" {
							parts := strings.Split(path, "/")
							path = strings.Join(parts[:len(parts)-1], "/") + "/" + tc.replaceLastPathDir
						}
						url := ts.URL + path

						// WHEN a HTTP request is made to this router
						req, err := http.NewRequest(tc.method, url, nil)
						if err != nil {
							t.Fatal(err)
						}
						resp, err := client.Do(req)
						if err != nil {
							t.Fatal(err)
						}

						// Read the response body
						body, err := io.ReadAll(resp.Body)
						if err != nil {
							t.Fatal(err)
						}
						resp.Body.Close()

						fail := false
						// THEN the status code is as expected
						if resp.StatusCode != tc.wantStatus {
							t.Errorf("%s - Expected a %d, not a %d",
								path, tc.wantStatus, resp.StatusCode)
							fail = true
						}
						// AND the body is as expected
						if !util.RegexCheck(tc.wantBody, string(body)) {
							t.Errorf("%s - Expected a body of\n%s\nnot\n%s",
								path, tc.wantBody, string(body))
							fail = true
						}

						if fail {
							t.FailNow()
						}
					}
				})
			}
		})
	}
}
