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
	"net/url"
	"runtime"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	config_test "github.com/release-argus/Argus/config/test"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/option"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestHTTP_SetupRoutesAPI__DisableRoutes(t *testing.T) {
	// GIVEN: an API and a bunch of routes.
	tests := map[string]struct {
		method, path, body string
		queryParams        url.Values
		wantStatus         int
		wantBody           string
	}{
		"-config": {
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
		"-runtime": {
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
		"-version": {
			method:     http.MethodGet,
			path:       "version",
			wantStatus: http.StatusOK,
			wantBody: test.TrimJSON(`{
				"version":"[^"]*",
				"buildDate":"[^"]*",
				"goVersion":"[^"]*"
			}`),
		},
		"-flags": {
			method:     http.MethodGet,
			path:       "flags",
			wantStatus: http.StatusOK,
			wantBody: test.TrimJSON(`{
				"log.level":"DEBUG",
				"log.timestamps":false,
				"data.database-file":"[^"]+",
				"web.listen-host":"[\d.]+",
				"web.listen-port":"\d+",
				"web.cert-file":"",
				"web.pkey-file":""
			`),
		},
		"-order": {
			method:     http.MethodGet,
			path:       "service/order",
			wantStatus: http.StatusOK,
			wantBody:   `{"order":\["test"\]}`,
		},
		"order_edit": {
			method:     http.MethodPut,
			path:       "service/order",
			body:       `{"order":["test"]}`,
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"order updated[^"]*"}`,
		},
		"-service_summary": {
			method:      http.MethodGet,
			path:        "service/summary",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		"-service_actions - GET": {
			method:      http.MethodGet,
			path:        "service/actions",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		"service_actions": {
			method:      http.MethodPost,
			path:        "service/actions",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		"-service_update - GET unspecific": {
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
		"-service_update - GET": {
			method:      http.MethodGet,
			path:        "service/config",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		"lv_refresh_new": {
			method:     http.MethodGet,
			path:       "latest_version/refresh_uncreated",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"overrides: .*required.*"}`,
		},
		"dv_refresh_new": {
			method:     http.MethodGet,
			path:       "deployed_version/refresh_uncreated",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"overrides: .*required.*"}`,
		},
		"lv_refresh": {
			method:      http.MethodGet,
			path:        "latest_version/refresh",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		"dv_refresh": {
			method:      http.MethodGet,
			path:        "deployed_version/refresh",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"service \\"[^"]+\\" not found"}`,
		},
		"notify_test": {
			method:     http.MethodPost,
			path:       "notify/test",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"name and/or name_previous are required"}`,
		},
		"service_update": {
			method:      http.MethodPut,
			path:        "service/config",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"edit \\"[^"]+\\" failed[^"]*"}`,
		},
		"service_create": {
			method:     http.MethodPut,
			path:       "service/new",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"create \\"\\" failed[^"]*"}`,
		},
		"service_delete": {
			method:      http.MethodDelete,
			path:        "service/delete",
			queryParams: url.Values{"service_id": {"unknown_service_id"}},
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"message":"delete \\"[^"]+\\" failed[^"]*"}`,
		},
	}
	disableCombinations := test.Combinations(util.SortedKeys(tests))

	// Split tests into groups.
	groupSize := max(1, len(disableCombinations)/runtime.NumCPU())
	numGroups := len(disableCombinations) / groupSize
	for i := range numGroups {
		groupStart := i * groupSize
		groupEnd := min((i+1)*groupSize, len(disableCombinations))
		group := disableCombinations[groupStart:groupEnd]

		name := fmt.Sprintf("Group %d", i+1)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for j, disabledRoutes := range group {
				// Insane number of tests, so have to skip some.
				if j%((len(tests)+1)*10) != 0 {
					continue
				}
				name := strings.Join(disabledRoutes, ";")
				t.Run(name, func(t *testing.T) {
					cfg := config_test.BareConfig(false)
					// Give values.
					cfg.Defaults.Default()
					cfg.WebHook = make(webhook.WebHooksDefaults)
					cfg.WebHook["test"], _ = webhook.DecodeDefaults("yaml", nil)
					cfg.WebHook["test"].Default()
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
					cfg.Service["test"].LatestVersion, _ = latestver.Decode(
						"yaml", []byte(`url: `+test.ArgusGitHubRepo),
						&option.Options{},
						svcStatus,
						lvbase.DefaultsConfig{
							Soft: &cfg.Defaults.Service.LatestVersion,
							Hard: &cfg.HardDefaults.Service.LatestVersion,
						},
					)
					cfg.Order = []string{"test"}

					announceChannel := cfg.HardDefaults.Service.Status.AnnounceChannel
					saveChannel := cfg.HardDefaults.Service.Status.SaveChannel
					cfg.Settings.Web.DisabledRoutes = disabledRoutes
					// Give every other test a route prefix.
					routePrefix := ""
					if j%2 == 0 {
						routePrefix = "/test"
						cfg.Settings.Web.RoutePrefix = routePrefix
					}
					api := NewAPI(cfg)
					api.SetupRoutesAPI()
					ts := httptest.NewServer(api.Router)
					ts.Config.Handler = api.Router
					t.Cleanup(ts.Close)
					client := http.Client{}

					// Test each route for this set of disabled routes.
					for name, tc := range tests {
						if len(announceChannel) != 0 {
							<-(announceChannel)
						}
						if len(saveChannel) != 0 {
							<-(saveChannel)
						}

						if !strings.HasPrefix(name, "-") && util.Contains(disabledRoutes, name) {
							tc.wantStatus = http.StatusNotFound
							tc.wantBody = "route disabled"
						}

						path := fmt.Sprintf(
							"%s/api/v1/%s",
							routePrefix, tc.path,
						)
						target := ts.URL + path

						reqBody := io.NopCloser(strings.NewReader(tc.body))
						if tc.body == "" {
							reqBody = nil
						}
						// WHEN: a HTTP request is made to this router.
						req, err := http.NewRequest(tc.method, target, reqBody)
						req.URL.RawQuery = tc.queryParams.Encode()
						if err != nil {
							t.Fatalf("%s\n%v", packageName, err)
						}
						resp, err := client.Do(req)
						if err != nil {
							t.Fatalf("%s\n%v", packageName, err)
						}

						// Read the response bodyRegex.
						body, err := io.ReadAll(resp.Body)
						if err != nil {
							t.Fatal(err)
						}
						_ = resp.Body.Close()

						fail := false
						prefix := fmt.Sprintf(
							"%s\nAPI.SetupRoutesAPI() method=%s, path=%s",
							packageName, tc.method, path,
						)

						// THEN: the status code is as expected.
						if got := resp.StatusCode; got != tc.wantStatus {
							t.Errorf("%s - status code mismatch\ngot:  %d\nwant: %d", prefix, got, tc.wantStatus)
							fail = true
						}

						// AND: the body is as expected.
						if got := string(body); !util.RegexCheck(tc.wantBody, got) {
							t.Errorf("%s - body mismatch\ngot:  %q\nwant: %q", prefix, got, tc.wantBody)
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

func TestHTTP_SetupRoutesNodeJS(t *testing.T) {
	// GIVEN: an API with NodeJS routes.
	tests := []struct {
		name        string
		route       string
		wantStatus  int
		wantContent string
	}{
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
			name:        "catch-all route - file not found",
			route:       "/some/random/path",
			wantStatus:  http.StatusNotFound,
			wantContent: "text/plain",
		},
		{
			name:        "catch-all route - file exists",
			route:       "/robots.txt",
			wantStatus:  http.StatusOK,
			wantContent: "text/plain",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config_test.BareConfig(true)
			api := NewAPI(cfg)
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
				t.Errorf("%s status code mismatch\ngot:  %d\nwant: %d", prefix, got, tc.wantStatus)
			}

			// AND: the content type is as expected.
			if tc.wantContent != "" {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tc.wantContent) {
					t.Errorf("%s Content-Type mismatch\ngot:  %q\nwant: %q", prefix, contentType, tc.wantContent)
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

			cfg := config_test.BareConfig(true)
			cfg.Settings.Web.Favicon = testFaviconSettings(tc.urlPNG, tc.urlSVG)
			api := NewAPI(cfg)
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
				t.Fatalf("%s\n%v", packageName, err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("%s\n%v", packageName, err)
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
				t.Fatalf("%s\n%v", packageName, err)
			}
			resp, err = client.Do(req)
			if err != nil {
				t.Fatalf("%s\n%v", packageName, err)
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

func TestHTTP_Version(t *testing.T) {
	// GIVEN: an API and the Version,BuildDate and GoVersion vars defined.
	api := API{}
	util.Version = "1.2.3"
	util.BuildDate = "2022-01-01T01:01:01Z"

	// WHEN: a HTTP request is made to the httpVersion handler.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()
	api.httpVersion(w, req)
	res := w.Result()
	t.Cleanup(func() { _ = res.Body.Close() })

	prefix := fmt.Sprintf("%s\nAPI.httpVersion()", packageName)

	// THEN: the version is returned in JSON format.
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: nil",
			prefix, err,
		)
	}
	var got apitype.VersionAPI
	_ = decode.Unmarshal("json", data, &got)
	want := apitype.VersionAPI{
		Version:   util.Version,
		BuildDate: util.BuildDate,
		GoVersion: util.GoVersion,
	}
	if got != want {
		t.Errorf(
			"%s body mismatch\ngot:  %+v\nwant: %+v",
			prefix, got, want,
		)
	}
}

func TestFailRequest(t *testing.T) {
	// GIVEN: an error.
	testErr := fmt.Errorf("request failed")
	wantBody, err := decode.Marshal(
		"json", map[string]string{
			"message": errfmt.FormatError(testErr),
		},
	)
	if err != nil {
		t.Fatalf("decode.Marshal() = %v, want nil", err)
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
			err:        testErr,
			statusCode: http.StatusBadRequest,
			wantBody:   string(wantBody),
			errRegex:   `^$`,
		},
		{
			name:          "write error",
			err:           testErr,
			statusCode:    http.StatusInternalServerError,
			wantBody:      "",
			errRegex:      `ERROR: failRequest, write failed`,
			useFailWriter: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			// AND: a response writer to write this error.
			var rw http.ResponseWriter = httptest.NewRecorder()
			if tc.useFailWriter {
				rw = &failWriter{header: make(http.Header)}
			}

			// WHEN: failRequest is called.
			failRequest(&rw, tc.err, tc.statusCode)

			prefix := fmt.Sprintf("%s\nfailRequest(%s)", packageName, tc.name)

			var gotCode int
			var gotBody string
			switch w := rw.(type) {
			case *httptest.ResponseRecorder:
				gotCode = w.Code
				gotBody = w.Body.String()
			case *failWriter:
				gotCode = w.code
				gotBody = w.body
			default:
				t.Fatalf("%s unexpected ResponseWriter type %T", prefix, rw)
			}

			// THEN: the status code is as expected.
			if gotCode != tc.statusCode {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, gotCode, tc.statusCode,
				)
			}

			// AND: the body is as expected.
			if gotBody != tc.wantBody {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, gotBody, tc.wantBody,
				)
			}

			// AND: stdout is as expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf(
					"%s log mismatch\ngot:  %q\nwant pattern: %q",
					prefix, stdout, tc.errRegex,
				)
			}
		})
	}
}

func TestFailRequest__MarshalError(t *testing.T) {
	// GIVEN: a failing marshal function.
	original := marshalFailRequestBody
	marshalFailRequestBody = func(v map[string]string) ([]byte, error) {
		return nil, fmt.Errorf("marshal failed")
	}
	t.Cleanup(func() { marshalFailRequestBody = original })

	releaseStdout := test.CaptureLog(t, logx.Default())

	w := httptest.NewRecorder()
	var rw http.ResponseWriter = w

	// WHEN: failRequest is called.
	failRequest(
		&rw,
		fmt.Errorf("original error"),
		http.StatusBadRequest,
	)

	prefix := fmt.Sprintf("%s\nfailRequest(marshal error)", packageName)

	// THEN: the status code is unchanged.
	if got := w.Code; got != http.StatusBadRequest {
		t.Errorf(
			"%s status code mismatch\ngot:  %d\nwant: %d",
			prefix, got, http.StatusBadRequest,
		)
	}

	// AND: a fallback JSON body is written.
	got := w.Body.String()
	want := `{"message":"failed to encode error response"}`
	if got != want {
		t.Errorf(
			"%s body mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// AND: the marshal error is logged.
	stdout := releaseStdout()
	if !util.RegexCheck(`ERROR: failRequest, marshal failed`, stdout) {
		t.Errorf(
			"%s log mismatch\ngot:  %q\nwant pattern: %q",
			prefix, stdout, `ERROR: failRequest, marshal failed`,
		)
	}
}
