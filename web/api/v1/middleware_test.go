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

func TestHTTP_BasicAuthMiddleware(t *testing.T) {
	// GIVEN: an API with/without Basic Auth credentials.
	tests := []struct {
		name      string
		basicAuth *config.WebSettingsBasicAuth
		fail      bool
		noHeader  bool
	}{
		{
			name:      "No basic auth",
			basicAuth: nil,
			fail:      false,
		},
		{
			name: "basic auth fail invalid creds",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234",
			},
			fail: true,
		},
		{
			name: "basic auth fail no Authorization header",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234",
			},
			noHeader: true,
			fail:     true,
		},
		{
			name: "basic auth pass",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "123",
			},
			fail: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			cfg := config.Config{}
			cfg.Settings.Web.BasicAuth = tc.basicAuth
			// Hash the username/password.
			if cfg.Settings.Web.BasicAuth != nil {
				cfg.Settings.Web.BasicAuth.CheckValues()
			}
			api := NewAPI(&cfg)
			api.Router.HandleFunc("/test", func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})
			ts := httptest.NewServer(api.BaseRouter)
			t.Cleanup(ts.Close)

			// WHEN: a HTTP request is made to this router.
			client := http.Client{}
			req, err := http.NewRequest(
				http.MethodGet,
				ts.URL+"/test",
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}
			if !tc.noHeader {
				req.Header = http.Header{
					// "test:123"
					"Authorization": {"Basic dGVzdDoxMjM="},
				}
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			// THEN: the request passes only when expected.
			got := resp.StatusCode
			want := 200
			if tc.fail {
				want = http.StatusUnauthorized
			}
			if got != want {
				t.Errorf(
					"%s\nstatus code mismatch on page with BasicAuth=%t\ngot:  %d\nwant: %d",
					packageName, tc.basicAuth != nil,
					got, want,
				)
			}
		})
	}
}

func TestLoggerMiddleware(t *testing.T) {
	// GIVEN: a HTTP route with/without a logger middleware.
	tests := []bool{
		true,
		false,
	}

	for _, tc := range tests {
		name := fmt.Sprintf("middleware=%t", tc)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)
			if tc {
				server.Config.Handler = loggerMiddleware(server.Config.Handler)
			}
			t.Cleanup(server.Close)

			// WHEN: a HTTP request is made to this router.
			client := http.Client{}
			req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf(
					"%s\ncould not make HTTP request to test loggerMiddleware(): %v",
					packageName, err,
				)
			}

			// THEN: the request always passes.
			got := resp.StatusCode
			want := http.StatusOK
			if got != want {
				t.Errorf(
					"%s\nloggerMiddleware() status code mismatch\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}

			// AND: the stdout is as expected.
			stdout := releaseStdout()
			re := fmt.Sprintf(
				`%s \([0-9a-f.\:]+\) %s .+`,
				req.Method, strings.ReplaceAll(req.URL.Path, "/", `\/`),
			)
			gotLoggerMiddlewareStdout := util.RegexCheck(re, stdout)
			if tc != gotLoggerMiddlewareStdout {
				if tc {
					t.Errorf(
						"%s\nloggerMiddleware() stdout mismatch\ngot:  %q\nwant: %q",
						packageName, stdout, re,
					)
				} else {
					t.Errorf(
						"%s\nloggerMiddleware() got loggerMiddleware stdout\nwant no loggerMiddleware stdout\nstdout: %q",
						packageName, stdout,
					)
				}
			}
		})
	}
}
