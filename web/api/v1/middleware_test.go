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
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestAPI_ClientIPMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		trusted       []string
		remoteAddr    string
		xForwardedFor []string
		wantIP        string
		wantWarn      bool
	}{
		{
			name:          "no trusted proxies/forwarded header is ignored, and warns",
			remoteAddr:    "1.1.1.1:1234",
			xForwardedFor: []string{"2.2.2.2"},
			wantIP:        "1.1.1.1",
			wantWarn:      true,
		},
		{
			name:       "no trusted proxies/no forwarded header does not warn",
			remoteAddr: "1.1.1.1:1234",
			wantIP:     "1.1.1.1",
		},
		{
			name:          "trusted proxies/an unparseable peer is not trusted",
			trusted:       []string{"10.0.0.0/8"},
			remoteAddr:    "not-an-ip",
			xForwardedFor: []string{"2.2.2.2"},
			wantIP:        "",
		},
		{
			name:          "trusted proxies/trusted peer resolves X-Forwarded-For IP",
			trusted:       []string{"10.0.0.0/8"},
			remoteAddr:    "10.1.2.3",
			xForwardedFor: []string{"3.3.3.3"},
			wantIP:        "3.3.3.3",
		},
		{
			name:          "trusted proxies/proxy-appended X-Forwarded-For line beats a spoofed first line",
			trusted:       []string{"10.0.0.0/8"},
			remoteAddr:    "10.1.2.3",
			xForwardedFor: []string{"6.6.6.6", "203.0.113.9"},
			wantIP:        "203.0.113.9",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			proxies := make([]netip.Prefix, 0, len(tc.trusted))
			for _, p := range tc.trusted {
				proxies = append(proxies, netip.MustParsePrefix(p))
			}
			api := &API{trustedProxies: proxies}

			gotIP := "handler not reached"
			handler := api.clientIPMiddleware(
				http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					gotIP = getIP(r)
				}),
			)

			for range 2 {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = tc.remoteAddr
				for _, line := range tc.xForwardedFor {
					req.Header.Add("X-Forwarded-For", line)
				}
				handler.ServeHTTP(httptest.NewRecorder(), req)
			}
			stdout := releaseStdout()

			if gotIP != tc.wantIP {
				t.Errorf("getIP() mismatch\ngot:  %q\nwant: %q", gotIP, tc.wantIP)
			}

			// The warning fires once, not per request.
			wantWarnings := 0
			if tc.wantWarn {
				wantWarnings = 1
			}
			got := strings.Count(stdout, "settings.web.trusted_proxies is empty")
			if got != wantWarnings {
				t.Errorf(
					"warning count mismatch\ngot:  %d\nwant: %d\nstdout: %q",
					got, wantWarnings, stdout,
				)
			}
		})
	}
}

func TestAPI_IsTrustedProxy(t *testing.T) {
	tests := []struct {
		name    string
		trusted []string
		addr    string
		want    bool
	}{
		{
			name: "empty trustedProxies/always false",
			addr: "10.0.0.5",
			want: false,
		},
		{
			name:    "single CIDR/matching address",
			trusted: []string{"10.0.0.0/8"},
			addr:    "10.0.0.5",
			want:    true,
		},
		{
			name:    "single CIDR/non-matching address",
			trusted: []string{"10.0.0.0/8"},
			addr:    "3.3.3.3",
			want:    false,
		},
		{
			name:    "multiple CIDRs/matches second entry",
			trusted: []string{"172.16.0.0/12", "10.0.0.0/8"},
			addr:    "10.0.0.5",
			want:    true,
		},
		{
			name:    "IPv4-mapped IPv6 address/matches IPv4 CIDR",
			trusted: []string{"10.0.0.0/8"},
			addr:    "::ffff:10.0.0.5", // Unmapped to 10.0.0.5 before matching.
			want:    true,
		},
		{
			name:    "IPv6 CIDR/matching IPv6 address",
			trusted: []string{"::1/128"},
			addr:    "::1",
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			proxies := make([]netip.Prefix, 0, len(tc.trusted))
			for _, p := range tc.trusted {
				pfx, err := netip.ParsePrefix(p)
				if err != nil {
					t.Fatalf("netip.ParsePrefix(%q) failed: %v", p, err)
				}
				proxies = append(proxies, pfx)
			}
			api := &API{trustedProxies: proxies}

			addr, err := netip.ParseAddr(tc.addr)
			if err != nil {
				t.Fatalf("netip.ParseAddr() failed: %v", err)
			}

			got := api.isTrustedProxy(addr)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestHTTP_BasicAuthMiddleware(t *testing.T) {
	tests := []struct {
		name                       string
		basicAuth                  *config.WebSettingsBasicAuth
		sendUsername, sendPassword string
		fail                       bool
		noHeader                   bool
	}{
		{
			name:         "pass/no basic auth configured",
			basicAuth:    nil,
			sendUsername: "test", sendPassword: "123",
			fail: false,
		},
		{
			name: "pass/matching credentials",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "123",
			},
			sendUsername: "test", sendPassword: "123",
			fail: false,
		},
		{
			name: "fail/invalid creds",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234",
			},
			sendUsername: "test", sendPassword: "123",
			fail: true,
		},
		{
			name: "fail/no Authorization header",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test", Password: "1234",
			},
			noHeader: true,
			fail:     true,
		},
		{
			name: "pass/no username configured, empty username sent",
			basicAuth: &config.WebSettingsBasicAuth{
				Password: "123",
			},
			sendUsername: "", sendPassword: "123",
			fail: false,
		},
		{
			name: "fail/no username configured, a username sent",
			basicAuth: &config.WebSettingsBasicAuth{
				Password: "123",
			},
			sendUsername: "test", sendPassword: "123",
			fail: true,
		},
		{
			name: "pass/no password configured, empty password sent",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test",
			},
			sendUsername: "test", sendPassword: "",
			fail: false,
		},
		{
			name: "fail/no password configured, a password sent",
			basicAuth: &config.WebSettingsBasicAuth{
				Username: "test",
			},
			sendUsername: "test", sendPassword: "123",
			fail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{}
			cfg.Settings.Web.BasicAuth = tc.basicAuth
			// The middleware compares the hashes that config.WebSettingsBasicAuth.CheckValues computes.
			if cfg.Settings.Web.BasicAuth != nil {
				cfg.Settings.Web.BasicAuth.CheckValues()
			}
			api, _ := NewAPI(&cfg)
			api.Router.HandleFunc("/test", func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})
			ts := httptest.NewServer(api.BaseRouter)
			t.Cleanup(ts.Close)

			client := http.Client{}
			req, err := http.NewRequest(
				http.MethodGet,
				ts.URL+"/test",
				nil,
			)
			if err != nil {
				t.Fatalf("http.NewRequest() failed: %v", err)
			}
			if !tc.noHeader {
				req.SetBasicAuth(tc.sendUsername, tc.sendPassword)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Client.Do() failed: %v", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}

			got, want := resp.StatusCode, http.StatusOK
			if tc.fail {
				want = http.StatusUnauthorized
			}
			if got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}
		})
	}
}

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		middleware bool
	}{
		{name: "with middleware", middleware: true},
		{name: "without middleware", middleware: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)
			if tc.middleware {
				server.Config.Handler = loggerMiddleware(server.Config.Handler)
			}
			t.Cleanup(server.Close)

			client := http.Client{}
			req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
			if err != nil {
				t.Fatalf("http.NewRequest() failed: %v", err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Client.Do() failed: %v", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}

			got := resp.StatusCode
			want := http.StatusOK
			if got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}

			stdout := releaseStdout()
			re := fmt.Sprintf(
				`%s \([0-9a-f.\:]+\) %s .+`,
				req.Method, strings.ReplaceAll(req.URL.Path, "/", `\/`),
			)
			gotLoggerMiddlewareStdout := util.RegexCheck(re, stdout)
			if tc.middleware != gotLoggerMiddlewareStdout {
				if tc.middleware {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, re)
				} else {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: NOT %q", stdout, re)
				}
			}
		})
	}
}
