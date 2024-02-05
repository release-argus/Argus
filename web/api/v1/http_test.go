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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

func TestHTTP_Version(t *testing.T) {
	// GIVEN an API and the Version,BuildDate and GoVersion vars defined
	api := API{}
	api.Log = util.NewJLog("WARN", false)
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
			cfg.Settings.Web.RoutePrefix = stringPtr("")
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

			cfg := testBareConfig()
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
