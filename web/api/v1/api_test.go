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
	tests := []struct {
		name        string
		routePrefix string
	}{
		{
			name:        "empty route prefix",
			routePrefix: "",
		},
		{
			name:        "root route prefix",
			routePrefix: "/",
		},
		{
			name:        "single-segment route prefix",
			routePrefix: "/test",
		},
		{
			name:        "multi-segment route prefix",
			routePrefix: "/my/test",
		},
		{
			name:        "multi-segment route prefix, trailing slash",
			routePrefix: "/my/test/",
		},
	}
	basicAuthTests := []struct {
		name               string
		username, password string
	}{
		{name: "no basic auth"},
		{name: "basic auth/user", username: "user", password: "pass"},
		{name: "basic auth/foo", username: "foo", password: "bar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// NewAPI trims a trailing slash, so requests use the trimmed prefix.
			routePrefix := strings.TrimSuffix(tc.routePrefix, "/")

			for _, basicAuthTest := range basicAuthTests {
				t.Run(basicAuthTest.name, func(t *testing.T) {
					cfg := &config.Config{
						Settings: config.Settings{
							Web: config.WebSettings{
								RoutePrefix: tc.routePrefix,
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

					api, _ := NewAPI(cfg)

					// Served by the base router, outside basic auth.
					req := httptest.NewRequest(http.MethodGet, routePrefix+"/api/v1/healthcheck", nil)
					w := httptest.NewRecorder()
					api.BaseRouter.ServeHTTP(w, req)

					if got, want := w.Code, http.StatusOK; got != want {
						t.Errorf(
							"healthcheck: status code mismatch\ngot:  %d - %q\nwant: %d",
							got, w.Body.String(), want,
						)
					}
					if got, want := w.Body.String(), "Alive"; got != want {
						t.Errorf("healthcheck: body mismatch\ngot:  %q\nwant: %q", got, want)
					}

					req, err := http.NewRequest(http.MethodGet, routePrefix, nil)
					if err != nil {
						t.Fatalf("http.NewRequest() failed: %v", err)
					}
					w = httptest.NewRecorder()
					api.BaseRouter.ServeHTTP(w, req)
					wantStatusCode := http.StatusPermanentRedirect
					if routePrefix == "" {
						wantStatusCode = http.StatusMovedPermanently
					}
					if got, want := w.Code, wantStatusCode; got != want {
						t.Errorf(
							"redirect: status code mismatch\ngot:  %d - %q\nwant: %d",
							got, w.Body.String(), want,
						)
					}
					if got, want := w.Header().Get("Location"), routePrefix+"/"; got != want {
						t.Errorf(
							"redirect: Location header mismatch\ngot:  %q\nwant: %q",
							got, want,
						)
					}

					api.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("OK"))
					}).Methods(http.MethodGet)
					authMiddlewareWanted := basicAuthTest.username != ""
					req = httptest.NewRequest(http.MethodGet, routePrefix+"/", nil)
					w = httptest.NewRecorder()
					api.BaseRouter.ServeHTTP(w, req)

					wantStatusCode = http.StatusOK
					if authMiddlewareWanted {
						wantStatusCode = http.StatusUnauthorized
					}
					if got, want := w.Code, wantStatusCode; got != want {
						t.Errorf(
							"unauthenticated: status code mismatch\ngot:  %d - %q\nwant: %d",
							got, w.Body.String(), want,
						)
					}

					if authMiddlewareWanted {
						req.SetBasicAuth(basicAuthTest.username, basicAuthTest.password)
						w = httptest.NewRecorder()
						api.BaseRouter.ServeHTTP(w, req)
						if got, want := w.Body.String(), "OK"; got != want {
							t.Errorf("authenticated: body mismatch\ngot:  %q\nwant: %q", got, want)
						}
					}
				})
			}
		})
	}
}

func TestAPI__ServiceOp(t *testing.T) {
	api := &API{}
	idA, idB := "A", "B"

	opA1 := api.acquireServiceOp(idA)
	opA2 := api.acquireServiceOp(idA)
	opB := api.acquireServiceOp(idB)

	if opA1 != opA2 {
		t.Errorf("same ID: API.acquireServiceOp() mismatch\ngot:  %p\nwant: %p", opA2, opA1)
	}
	if opA1 == opB {
		t.Errorf(
			"different IDs: API.acquireServiceOp() mismatch\ngot:  %p\nwant: NOT %p",
			opB, opA1,
		)
	}

	if got, want := len(api.serviceOps), 2; got != want {
		t.Errorf("acquired: API.serviceOps length mismatch\ngot:  %d\nwant: %d", got, want)
	}
	if got, want := opA1.refs, 2; got != want {
		t.Errorf("acquired: serviceOp.refs mismatch\ngot:  %d\nwant: %d", got, want)
	}

	api.releaseServiceOp(idA, opA1)

	if _, ok := api.serviceOps[idA]; !ok {
		t.Error("one released: API.serviceOps[idA] mismatch\ngot:  absent\nwant: present")
	}
	if got, want := opA2.refs, 1; got != want {
		t.Errorf("one released: serviceOp.refs mismatch\ngot:  %d\nwant: %d", got, want)
	}

	api.releaseServiceOp(idA, opA2)

	if _, ok := api.serviceOps[idA]; ok {
		t.Error("last released: API.serviceOps[idA] mismatch\ngot:  present\nwant: absent")
	}

	opA3 := api.acquireServiceOp(idA)
	if opA3 == opA1 {
		t.Errorf(
			"fresh acquire: API.acquireServiceOp() mismatch\ngot:  %p\nwant: NOT %p",
			opA3, opA1,
		)
	}

	api.releaseServiceOp(idA, opA3)
	api.releaseServiceOp(idB, opB)

	if got, want := len(api.serviceOps), 0; got != want {
		t.Errorf("all released: API.serviceOps length mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		bodyRegex     string
		input         any
		errRegex      string
		useFailWriter bool
	}{
		{
			name:       "successful JSON encoding",
			statusCode: http.StatusOK,
			bodyRegex:  `^{"key":"value"}` + "\n$",
			input:      map[string]string{"key": "value"},
			errRegex:   `^$`,
		},
		{
			name:       "JSON encoding failure",
			statusCode: http.StatusInternalServerError,
			bodyRegex:  `^{"message":"json: .* marshal from Go chan int"}$`,
			input:      make(chan int), // Invalid type for JSON encoding.
			errRegex:   `^ERROR: json: .* marshal from Go chan int\s$`,
		},
		{
			name:          "response write failure",
			statusCode:    http.StatusOK, // Sent before the write failed.
			bodyRegex:     `^{"key":"value"}` + "\n$",
			input:         map[string]string{"key": "value"},
			errRegex:      `^ERROR: write failed\s$`,
			useFailWriter: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			var rw http.ResponseWriter = httptest.NewRecorder()
			if tc.useFailWriter {
				rw = &failWriter{header: make(http.Header)}
			}
			api := &API{}

			api.writeJSON(rw, tc.input, logx.LogFrom{})

			var code int
			var body string
			if rec, ok := rw.(*httptest.ResponseRecorder); ok {
				code = rec.Code
				body = rec.Body.String()
			} else if fw, ok := rw.(*failWriter); ok {
				code = fw.code
				body = fw.body
			}
			if got, want := code, tc.statusCode; got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}

			if !util.RegexCheck(tc.bodyRegex, body) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", body, tc.bodyRegex)
			}

			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.errRegex)
			}
		})
	}
}
