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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	config_test "github.com/release-argus/Argus/config/test"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/option"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestHTTP_SetupRoutesAPI__disableRoutes(t *testing.T) {
	// Each route is named by its disabled_routes key, or '-'-prefixed if it has none.
	type route struct {
		name               string
		method, path, body string
		queryParams        url.Values
		wantStatus         int
		wantBody           string
	}
	tests := []route{
		{
			name:       "-config",
			method:     http.MethodGet,
			path:       "config",
			wantStatus: http.StatusOK,
			wantBody: test.TrimJSON(`{
				"settings":{.*},
				"defaults":{.*"service":{.*},"notify":{.*},"webhook":{.*}},
				"notify":{.*},
				"webhook":{.*},
				"service":{.*
			}`),
		},
		{
			name:       "-runtime",
			method:     http.MethodGet,
			path:       "status/runtime",
			wantStatus: http.StatusOK,
			wantBody: test.TrimJSON(`{
				"start_time":"[^"]+",
				"cwd":"[^"]+",
				"goroutines":\d+,
				"GOMAXPROCS":\d+
			}`),
		},
		{
			name:       "-version",
			method:     http.MethodGet,
			path:       "version",
			wantStatus: http.StatusOK,
			wantBody: test.TrimJSON(`{
				"version":"[^"]*",
				"buildDate":"[^"]*",
				"goVersion":"[^"]*"
			}`),
		},
		{
			name:       "-flags",
			method:     http.MethodGet,
			path:       "flags",
			wantStatus: http.StatusOK,
			wantBody: test.TrimJSON(`{
				"log.level":"DEBUG",
				"log.timestamps":false,
				"data.database-file":"[^"]+",
				"data.readonly":false,
				"web.listen-host":"[\d.]+",
				"web.listen-port":"\d+",
				"web.cert-file":"",
				"web.pkey-file":"",
				"web.route-prefix":"/"
			}`),
		},
		{
			name:       "-order",
			method:     http.MethodGet,
			path:       "service/order",
			wantStatus: http.StatusOK,
			wantBody:   `{"order":\["test"\]}`,
		},
		{
			name:       "order_edit",
			method:     http.MethodPut,
			path:       "service/order",
			body:       `{"order":["test"]}`,
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"order updated[^"]*"}`,
		},
		{
			name:        "-service_summary",
			method:      http.MethodGet,
			path:        "service/summary",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		{
			name:        "-service_actions/GET",
			method:      http.MethodGet,
			path:        "service/actions",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		{
			name:        "service_actions",
			method:      http.MethodPost,
			path:        "service/actions",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		{
			name:       "-service_update/GET unspecific",
			method:     http.MethodGet,
			path:       "service/defaults",
			wantStatus: http.StatusOK,
			wantBody: test.TrimJSON(`{
				"hard_defaults":{.*},
				"defaults":{.*},
				"notify":{.*},
				"webhook":{.*}
			}`),
		},
		{
			name:        "-service_update/GET",
			method:      http.MethodGet,
			path:        "service/config",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		{
			name:       "lv_refresh_new",
			method:     http.MethodGet,
			path:       "latest_version/refresh_uncreated",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"overrides: .*required.*"}`,
		},
		{
			name:       "dv_refresh_new",
			method:     http.MethodGet,
			path:       "deployed_version/refresh_uncreated",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"overrides: .*required.*"}`,
		},
		{
			name:        "lv_refresh",
			method:      http.MethodGet,
			path:        "latest_version/refresh",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		{
			name:        "dv_refresh",
			method:      http.MethodGet,
			path:        "deployed_version/refresh",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		{
			name:       "notify_test",
			method:     http.MethodPost,
			path:       "notify/test",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"body required"}`,
		},
		{
			name:        "service_update",
			method:      http.MethodPut,
			path:        "service/config",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"edit \\"[^"]+\\" failed[^"]*"}`,
		},
		{
			name:       "service_create",
			method:     http.MethodPut,
			path:       "service/new",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"create \\"\\" failed[^"]*"}`,
		},
		{
			name:        "service_delete",
			method:      http.MethodDelete,
			path:        "service/delete",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"delete \\"[^"]+\\" failed[^"]*"}`,
		},
	}
	names := test.AttributesSorted(tests, func(r route) string { return r.name })
	disableCombinations := test.Combinations(names)

	groupSize := max(1, len(disableCombinations)/runtime.NumCPU())
	numGroups := (len(disableCombinations) + groupSize - 1) / groupSize
	for i := range numGroups {
		groupStart := i * groupSize
		groupEnd := min((i+1)*groupSize, len(disableCombinations))
		group := disableCombinations[groupStart:groupEnd]

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			sampleEvery := (len(tests) + 1) * 10
			for j, disabledRoutes := range group {
				// Too many combinations to run each, so sample them.
				if j%sampleEvery != 0 {
					continue
				}
				name := strings.Join(disabledRoutes, ";")
				t.Run(name, func(t *testing.T) {
					// Serial: bounds concurrency to one server per group.

					cfg := config_test.BareConfig(false)
					cfg.Defaults.Default()
					cfg.Webhook = make(webhook.WebhooksDefaults)
					whDefaults, _ := webhook.DecodeDefaults("yaml", nil)
					cfg.Webhook["test"] = whDefaults
					cfg.Webhook["test"].Default()
					cfg.Notify = make(map[string]*shoutrrr.Defaults)
					cfg.Notify["test"] = shoutrrr.NewDefaults(
						"discord",
						nil,
						nil,
						shoutrrr.MapStringStringOmitNull{
							"test": "123",
						},
					)
					cfg.Service = make(service.Services)
					cfg.Service["test"] = &service.Service{}
					svcStatus, _ := statustest.New("yaml", nil)
					var err error
					cfg.Service["test"].LatestVersion, err = latestver.Decode(
						"yaml", []byte(`url: `+test.ArgusGitHubRepo),
						&option.Options{},
						svcStatus,
						latestver.DefaultsConfig{
							Soft: &cfg.Defaults.Service.LatestVersion,
							Hard: &cfg.HardDefaults.Service.LatestVersion,
						},
					)
					if err != nil {
						t.Fatalf("latestver.Decode() failed: %v", err)
					}
					cfg.Order = []string{"test"}

					announceChannel := cfg.HardDefaults.Service.Status.AnnounceChannel
					saveChannel := cfg.HardDefaults.Service.Status.SaveChannel
					cfg.Settings.Web.DisabledRoutes = disabledRoutes
					// Give every other sampled run a route prefix.
					routePrefix := ""
					if (j/sampleEvery)%2 == 0 {
						routePrefix = "/test"
						cfg.Settings.Web.RoutePrefix = routePrefix
					}
					api, _ := NewAPI(cfg)
					api.SetupRoutesAPI()
					ts := httptest.NewServer(api.Router)
					t.Cleanup(ts.Close)
					client := http.Client{}

					for _, tc := range tests {
						t.Run(tc.name, func(t *testing.T) {
							// Serial: shares api.Config.

							for len(announceChannel) != 0 {
								<-announceChannel
							}
							for len(saveChannel) != 0 {
								<-saveChannel
							}

							if routePrefix != "" {
								tc.wantBody = strings.Replace(
									tc.wantBody,
									`"web.route-prefix":"/"`,
									`"web.route-prefix":"`+routePrefix+`"`,
									1,
								)
							}
							if !strings.HasPrefix(tc.name, "-") && slices.Contains(disabledRoutes, tc.name) {
								tc.wantStatus = http.StatusNotFound
								tc.wantBody = "route disabled"
							}

							target := ts.URL + routePrefix + "/api/v1/" + tc.path

							reqBody := io.NopCloser(strings.NewReader(tc.body))
							if tc.body == "" {
								reqBody = nil
							}

							req, err := http.NewRequest(tc.method, target, reqBody)
							if err != nil {
								t.Fatalf("http.NewRequest() failed: %v", err)
							}
							req.URL.RawQuery = tc.queryParams.Encode()
							resp, err := client.Do(req)
							if err != nil {
								t.Fatalf("Client.Do() failed: %v", err)
							}

							body, err := io.ReadAll(resp.Body)
							if err != nil {
								t.Fatalf("io.ReadAll() failed: %v", err)
							}
							_ = resp.Body.Close()

							if got, want := resp.StatusCode, tc.wantStatus; got != want {
								t.Errorf(
									"status code mismatch\ngot:  %d - %q\nwant: %d",
									got, body, want,
								)
							}

							if got := string(body); !util.RegexCheck(tc.wantBody, got) {
								t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wantBody)
							}
						})
					}
				})
			}
		})
	}
}

func TestAPI_SetupWebSocket(t *testing.T) {
	tests := []struct {
		name             string
		basicAuth        bool
		staticToken      string // Literal '?token=X' value.
		mintToken        bool   // Mint a fresh token from wsTokens and use it.
		wantUnauthorized bool
	}{
		{
			name:             "no basic auth, no token",
			basicAuth:        false,
			wantUnauthorized: false,
		},
		{
			name:             "basic auth/no token",
			basicAuth:        true,
			wantUnauthorized: true,
		},
		{
			name:             "basic auth/invalid token",
			basicAuth:        true,
			staticToken:      "not-a-real-token",
			wantUnauthorized: true,
		},
		{
			name:             "basic auth/valid token",
			basicAuth:        true,
			mintToken:        true,
			wantUnauthorized: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Config{}
			if tc.basicAuth {
				cfg.Settings.Web.BasicAuth = &config.WebSettingsBasicAuth{
					Username: "test", Password: "123",
				}
				cfg.Settings.Web.BasicAuth.CheckValues()
			}
			api, wsRoute := NewAPI(&cfg)
			api.SetupWebSocket(NewHub(), wsRoute)
			ts := httptest.NewServer(api.BaseRouter)
			t.Cleanup(ts.Close)

			token := tc.staticToken
			if tc.mintToken {
				token = api.wsTokens.New()
			}
			wsURL := ts.URL + "/ws"
			if token != "" {
				wsURL += "?token=" + token
			}

			resp, err := http.Get(wsURL)
			if err != nil {
				t.Fatalf("first request: http.Get() failed: %v", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("first request: io.ReadAll() failed: %v", err)
			}

			if tc.wantUnauthorized {
				if got, want := resp.StatusCode, http.StatusUnauthorized; got != want {
					t.Errorf(
						"first request: status code mismatch\ngot:  %d - %q\nwant: %d",
						got, body, want,
					)
				}
			} else if resp.StatusCode == http.StatusUnauthorized {
				t.Errorf("first request: status code mismatch\ngot:  401 - %q\nwant: NOT 401", body)
			}

			// /ws sits outside the basic-auth subrouter, so nothing re-prompts.
			if got, want := resp.Header.Get("WWW-Authenticate"), ""; got != want {
				t.Errorf(
					"first request: WWW-Authenticate header mismatch\ngot:  %q\nwant: %q",
					got, want,
				)
			}

			// A token is single-use: Validate() consumes it.
			if tc.mintToken {
				resp2, err := http.Get(wsURL)
				if err != nil {
					t.Fatalf("token re-use: http.Get() failed: %v", err)
				}
				defer resp2.Body.Close()
				body2, err := io.ReadAll(resp2.Body)
				if err != nil {
					t.Fatalf("token re-use: io.ReadAll() failed: %v", err)
				}

				if got, want := resp2.StatusCode, http.StatusUnauthorized; got != want {
					t.Errorf(
						"token re-use: status code mismatch\ngot:  %d - %q\nwant: %d",
						got, body2, want,
					)
				}
			}
		})
	}
}

func TestAPI_SetupWebSocket__originCheck(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")
	server := httptest.NewServer(api.BaseRouter)
	t.Cleanup(server.Close)
	wsURL := strings.Replace(server.URL, "http:", "ws:", 1) + "/ws"

	tests := []struct {
		name        string
		origin      string
		wantRefused bool
	}{
		{
			name:        "valid/absent Origin - non-browser client",
			origin:      "",
			wantRefused: false,
		},
		{
			name:        "valid/same-origin handshake",
			origin:      server.URL,
			wantRefused: false,
		},
		{
			name:        "invalid/cross-site Origin",
			origin:      "http://evil.example.com",
			wantRefused: true,
		},
		{
			name:        "invalid/opaque null Origin",
			origin:      "null",
			wantRefused: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header := http.Header{"Cookie": {cookie.String()}}
			if tc.origin != "" {
				header.Set("Origin", tc.origin)
			}

			conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
			if conn != nil {
				t.Cleanup(func() { _ = conn.Close() })
			}
			if resp != nil {
				t.Cleanup(func() { _ = resp.Body.Close() })
			}

			if !tc.wantRefused {
				if err != nil {
					t.Errorf("Dialer.Dial(%q) failed: %v", wsURL, err)
				}
				return
			}
			if err == nil {
				t.Fatal("Dialer.Dial() error mismatch\ngot:  nil\nwant: non-nil")
			}
			if resp == nil {
				t.Fatalf("Dialer.Dial() response mismatch\ngot:  nil (err=%v)\nwant: non-nil", err)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got, want := resp.StatusCode, http.StatusForbidden; got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}
		})
	}
}

func TestAPI_SetupWebSocket__originUncheckedWithoutAuth(t *testing.T) {
	cfg := config.Config{}
	api, wsRoute := NewAPI(&cfg)
	api.SetupWebSocket(NewHub(), wsRoute)
	server := httptest.NewServer(api.BaseRouter)
	t.Cleanup(server.Close)
	wsURL := strings.Replace(server.URL, "http:", "ws:", 1) + "/ws"

	// An Origin not matching Host, as behind a reverse proxy that rewrites Host.
	header := http.Header{"Origin": {"http://argus.example.com"}}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if conn != nil {
		t.Cleanup(func() { _ = conn.Close() })
	}
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}

	if err != nil {
		t.Errorf("Dialer.Dial(%q) failed: %v", wsURL, err)
	}
}

func TestAPI_SetupWebSocket__sessionAuth(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")
	adminID := adminContext(t, api, deps).User.ID
	server := httptest.NewServer(api.BaseRouter)
	t.Cleanup(server.Close)

	wsURL := strings.Replace(server.URL, "http:", "ws:", 1) + "/ws"
	header := http.Header{"Cookie": {cookie.String()}}
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("Dialer.Dial(%q) failed: %v", wsURL, err)
	}
	t.Cleanup(func() { _ = clientConn.Close() })

	var registered *Client
	for i := 0; i < 100 && registered == nil; i++ {
		time.Sleep(10 * time.Millisecond)
		registered = hubClientForTest(t, api.hub)
	}
	if registered == nil {
		t.Fatal("hubClientForTest() mismatch\ngot:  nil\nwant: non-nil")
	}
	gotAlive := registered.sessionAlive != nil && registered.sessionAlive()
	fieldTests := []test.FieldAssertion{
		{Name: "userID", Got: registered.userID, Want: adminID, Mode: test.CompareEqual},
		{Name: "sessionHash", Got: registered.sessionHash, Want: auth.HashToken(cookie.Value), Mode: test.CompareEqual},
		{Name: "sessionAlive()", Got: gotAlive, Want: true, Mode: test.CompareEqual},
	}
	if testErr := test.AssertFields(t, fieldTests, "Client"); testErr != nil {
		t.Error(testErr)
	}
}

func TestHTTP_Version(t *testing.T) {
	api := API{}
	version, buildDate := util.Version, util.BuildDate
	t.Cleanup(func() { util.Version, util.BuildDate = version, buildDate })
	util.Version = "1.2.3"
	util.BuildDate = "2022-01-01T01:01:01Z"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()
	api.httpVersion(w, req)
	resp := w.Result()
	t.Cleanup(func() { _ = resp.Body.Close() })

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() failed: %v", err)
	}
	var got apitype.VersionAPI
	if err := decode.Unmarshal("json", data, &got); err != nil {
		t.Fatalf("decode.Unmarshal() failed: %v", err)
	}
	want := apitype.VersionAPI{
		Version:   "1.2.3",
		BuildDate: "2022-01-01T01:01:01Z",
		GoVersion: util.GoVersion,
	}
	if got != want {
		t.Errorf("body mismatch\ngot:  %s\nwant: %s", data, decode.ToJSONString(want))
	}
}

func TestFailRequest(t *testing.T) {
	requestErr := errors.New("request failed")
	wantBody, err := decode.Marshal(
		"json", map[string]string{
			"message": errfmt.FormatError(requestErr),
		},
	)
	if err != nil {
		t.Fatalf("decode.Marshal() failed: %v", err)
	}

	tests := []struct {
		name          string
		err           error
		statusCode    int
		wantBody      string
		errRegex      string
		useFailWriter bool
	}{
		{
			name:       "success",
			err:        requestErr,
			statusCode: http.StatusBadRequest,
			wantBody:   string(wantBody),
			errRegex:   `^$`,
		},
		{
			name:          "write error",
			err:           requestErr,
			statusCode:    http.StatusInternalServerError,
			wantBody:      string(wantBody),
			errRegex:      `ERROR: failRequest, write failed`,
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

			failRequest(&rw, tc.err, tc.statusCode)

			var gotCode int
			var gotBody string
			switch w := rw.(type) {
			case *httptest.ResponseRecorder:
				gotCode = w.Code
				gotBody = w.Body.String()
			case *failWriter:
				gotCode = w.code
				gotBody = w.body
			}

			if got, want := gotCode, tc.statusCode; got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, gotBody, want)
			}

			if gotBody != tc.wantBody {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", gotBody, tc.wantBody)
			}

			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.errRegex)
			}
		})
	}
}

func TestFailRequest__marshalError(t *testing.T) {
	original := marshalFailRequestBody
	marshalFailRequestBody = func(v map[string]string) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}
	t.Cleanup(func() { marshalFailRequestBody = original })

	releaseStdout := test.CaptureLog(t, logx.Default())

	w := httptest.NewRecorder()
	var rw http.ResponseWriter = w

	failRequest(
		&rw,
		errors.New("original error"),
		http.StatusBadRequest,
	)

	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	got := w.Body.String()
	want := `{"message":"failed to encode error response"}`
	if got != want {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, want)
	}

	stdout := releaseStdout()
	wantStdout := `ERROR: failRequest, marshal failed`
	if !util.RegexCheck(wantStdout, stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, wantStdout)
	}
}
