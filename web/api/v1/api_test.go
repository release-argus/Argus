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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestNewAPI(t *testing.T) {
	// GIVEN: a config.
	tests := []struct {
		name        string
		routePrefix string
	}{
		{
			name:        "prefix: ''",
			routePrefix: "",
		},
		{
			name:        "prefix: /",
			routePrefix: "/",
		},
		{
			name:        "prefix: /test",
			routePrefix: "/test",
		},
		{
			name:        "prefix: /my/test",
			routePrefix: "/my/test",
		},
		{
			name:        "prefix: /my/test/",
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// with/without basic auth.
			for _, basicAuthTest := range basicAuthTests {
				t.Logf(
					"%s - Testing with basic auth %+v",
					packageName, basicAuthTest,
				)
				cfg := &config.Config{
					Settings: config.Settings{
						SettingsBase: config.SettingsBase{
							Web: config.WebSettings{
								RoutePrefix: tc.routePrefix,
							},
						},
					},
				}
				if basicAuthTest.username != "" {
					cfg.Settings.Web.BasicAuth = &config.WebSettingsBasicAuth{
						Username: basicAuthTest.username,
						Password: basicAuthTest.password,
					}
					cfg.Settings.Web.BasicAuth.CheckValues()
				}
				// Test as if the routePrefix is always without a trailing slash.
				tc.routePrefix = strings.TrimSuffix(tc.routePrefix, "/")

				// WHEN: a new API is created.
				api := NewAPI(cfg)

				// THEN: the healthcheck endpoint is accessible.
				req, _ := http.NewRequest(
					http.MethodGet,
					tc.routePrefix+"/api/v1/healthcheck",
					nil,
				)
				resp := httptest.NewRecorder()
				api.BaseRouter.ServeHTTP(resp, req)

				prefix := fmt.Sprintf(
					"%s\nNewAPI() %s",
					packageName, req.URL.Path,
				)

				// 200.
				if http.StatusOK != resp.Code {
					t.Errorf(
						"%s status code mismatch\ngot:  %d\nwant: %d",
						prefix, resp.Code, http.StatusOK,
					)
				}
				// Alive.
				if gotBody, wantBody := resp.Body.String(), "Alive"; gotBody != wantBody {
					t.Errorf(
						"%s body mismatch\ngot:  %q\nwant: %q",
						prefix, resp.Body.String(), wantBody,
					)
				}

				// AND: the route prefix always has a trailing slash.
				req, _ = http.NewRequest(http.MethodGet, tc.routePrefix, nil)
				resp = httptest.NewRecorder()
				api.BaseRouter.ServeHTTP(resp, req)
				// 308.
				expectedStatusCode := http.StatusPermanentRedirect
				if tc.routePrefix == "" {
					// 301.
					expectedStatusCode = http.StatusMovedPermanently
				}
				if expectedStatusCode != resp.Code {
					t.Errorf(
						"%s status code mismatch\ngot:  %d\nwant: %d",
						prefix, resp.Code, expectedStatusCode,
					)
				}
				// Location.
				if got, want := resp.Header().Get("Location"), tc.routePrefix+"/"; got != want {
					t.Errorf(
						"%s (trailing slash), Header['Location'] mismatch\ngot:  %s\nwant: %s",
						prefix, got, want,
					)
				}

				prefix = fmt.Sprintf(
					"%s\nNewAPI() %s SetupRoutes() basicAuth middleware",
					packageName, req.URL.Path,
				)

				// AND: basic auth middleware is added when set.
				api.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("OK"))
				}).Methods(http.MethodGet)
				authMiddlewareWanted := basicAuthTest.username != ""
				req, _ = http.NewRequest(
					http.MethodGet,
					tc.routePrefix+"/",
					nil,
				)
				resp = httptest.NewRecorder()
				api.BaseRouter.ServeHTTP(resp, req)

				middlewareUsed := resp.Body.String() != "OK"
				if middlewareUsed != authMiddlewareWanted {
					t.Errorf(
						"%s mismatch\ngot:  used=%t\nwant: used=%t",
						prefix, middlewareUsed, authMiddlewareWanted,
					)
				}

				// Verify the basicAuth middleware is working.
				if authMiddlewareWanted {
					req.SetBasicAuth(basicAuthTest.username, basicAuthTest.password)
					resp = httptest.NewRecorder()
					api.BaseRouter.ServeHTTP(resp, req)
					if got, want := resp.Body.String(), "OK"; got != want {
						t.Errorf(
							"%s failed\ngot: body=%q\nwant: body=%q",
							prefix, got, want,
						)
					}
				}
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	// GIVEN: different input strings to write.
	tests := []struct {
		name          string
		response      *http.Response
		statusCode    int
		expectedBody  string
		input         any
		errRegex      string
		useFailWriter bool
	}{
		{
			name:         "successful JSON encoding",
			response:     &http.Response{},
			statusCode:   http.StatusOK,
			expectedBody: `{"key":"value"}` + "\n",
			input:        map[string]string{"key": "value"},
			errRegex:     `^$`,
		},
		{
			name:         "JSON encoding failure",
			response:     &http.Response{},
			statusCode:   http.StatusInternalServerError,
			expectedBody: `{"error":"json: cannot marshal from Go chan int"}` + "\n",
			input:        make(chan int), // Invalid type for JSON encoding.
			errRegex:     `^ERROR: json: cannot marshal from Go chan int\s$`,
		},
		{
			name:          "response write failure",
			statusCode:    http.StatusOK,
			expectedBody:  "",
			input:         map[string]string{"key": "value"},
			errRegex:      `^ERROR: write failed\s$`,
			useFailWriter: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			var rw http.ResponseWriter = httptest.NewRecorder()
			if tc.useFailWriter {
				rw = &failWriter{header: make(http.Header)}
			}
			api := &API{}

			// WHEN: writeJSON is called with this input.
			api.writeJSON(rw, tc.input, logx.LogFrom{})

			prefix := fmt.Sprintf("%s\nwriteJSON()", packageName)

			// THEN: the status code is as expected.
			var code int
			var body string
			if rec, ok := rw.(*httptest.ResponseRecorder); ok {
				code = rec.Code
				body = rec.Body.String()
			} else if fw, ok := rw.(*failWriter); ok {
				code = fw.code
				body = fw.body
			}
			if code != tc.statusCode {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, code, tc.statusCode,
				)
			}

			if body != tc.expectedBody {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, body, tc.expectedBody,
				)
			}

			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.errRegex,
				)
			}
		})
	}
}
