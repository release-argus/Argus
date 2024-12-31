// Copyright [2024] [Argus]
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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
)

func TestNewAPI(t *testing.T) {
	// GIVEN a config
	tests := map[string]struct {
		routePrefix string
	}{
		"prefix=''": {
			routePrefix: "",
		},
		"prefix=/": {
			routePrefix: "/",
		},
		"prefix=/test": {
			routePrefix: "/test",
		},
		"prefix=/my/test": {
			routePrefix: "/my/test",
		},
		"prefix=/my/test/": {
			routePrefix: "/my/test/",
		},
	}
	basicAuthTests := []struct {
		username, password string
	}{
		{"", ""},
		{"user", "pass"},
		{"foo", "bar"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// with/without basic auth
			for _, basicAuthTest := range basicAuthTests {
				t.Log("Testing with basic auth", basicAuthTest)
				cfg := &config.Config{
					Settings: config.Settings{
						SettingsBase: config.SettingsBase{
							Web: config.WebSettings{
								RoutePrefix: tc.routePrefix,
							}},
					},
				}
				if basicAuthTest.username != "" {
					cfg.Settings.Web.BasicAuth = &config.WebSettingsBasicAuth{
						Username: basicAuthTest.username,
						Password: basicAuthTest.password}
					cfg.Settings.Web.BasicAuth.CheckValues()
				}
				// Test as if the routePrefix is always without a trailing slash
				tc.routePrefix = strings.TrimSuffix(tc.routePrefix, "/")

				// WHEN a new API is created
				api := NewAPI(cfg, nil)

				// THEN the healthcheck endpoint is accessible
				req, _ := http.NewRequest("GET", tc.routePrefix+"/api/v1/healthcheck", nil)
				resp := httptest.NewRecorder()
				api.BaseRouter.ServeHTTP(resp, req)
				// 200
				if http.StatusOK != resp.Code {
					t.Errorf("Healthcheck, expected status code %d, got %d",
						http.StatusOK, resp.Code)
				}
				// Alive
				if resp.Body.String() != "Alive" {
					t.Errorf("Healthcheck, expected body %s, got %s",
						"Alive", resp.Body.String())
				}
				// AND the route prefix always has a trailing slash
				req, _ = http.NewRequest("GET", tc.routePrefix, nil)
				resp = httptest.NewRecorder()
				api.BaseRouter.ServeHTTP(resp, req)
				// 308
				expectedStatusCode := http.StatusPermanentRedirect
				if tc.routePrefix == "" {
					expectedStatusCode = http.StatusMovedPermanently
				}
				if expectedStatusCode != resp.Code {
					t.Errorf("trailing slash, expected status code %d, got %d",
						expectedStatusCode, resp.Code)
				}
				// Location
				if resp.Header().Get("Location") != tc.routePrefix+"/" {
					t.Errorf("trailing slash, expected Location %s, got %s",
						tc.routePrefix+"/", resp.Header().Get("Location"))
				}
				// AND basic auth middleware is added when set
				api.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("OK"))
				}).Methods("GET")
				authMiddlewareWanted := basicAuthTest.username != ""
				req, _ = http.NewRequest("GET", tc.routePrefix+"/", nil)
				resp = httptest.NewRecorder()
				api.BaseRouter.ServeHTTP(resp, req)
				middlewareUsed := resp.Body.String() != "OK"
				if middlewareUsed != authMiddlewareWanted {
					t.Errorf("Expected basicAuth middleware to be used: %t, got: %t",
						authMiddlewareWanted, middlewareUsed)
				}
				// Verify the basicAuth middleware is working
				if authMiddlewareWanted {
					req.SetBasicAuth(basicAuthTest.username, basicAuthTest.password)
					resp = httptest.NewRecorder()
					api.BaseRouter.ServeHTTP(resp, req)
					if resp.Body.String() != "OK" {
						t.Errorf("Expected basicAuth middleware to pass, got: %s",
							resp.Body.String())
					}
				}
			}
		})
	}
}