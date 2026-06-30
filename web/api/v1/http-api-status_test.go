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
	"path/filepath"
	"sync"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestHTTP_HTTPWebSocketToken(t *testing.T) {
	// GIVEN: an API without Basic Auth (wsTokens is nil).
	prefix := fmt.Sprintf("%q\nAPI.httpWebSocketToken()", packageName)
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)

	// WHEN: a request is made for a WebSocket token.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws-token", nil)
	w := httptest.NewRecorder()
	api.httpWebSocketToken(w, req)
	res := w.Result()
	t.Cleanup(func() { _ = res.Body.Close() })

	// THEN: 204 is returned (no token needed, Basic Auth not configured).
	if res.StatusCode != http.StatusNoContent {
		t.Errorf(
			"%s\nstatus code mismatch\ngot:  %d\nwant: %d",
			prefix, res.StatusCode, http.StatusNoContent,
		)
	}
}

func TestHTTP_HTTPWebSocketToken__AuthGated(t *testing.T) {
	// GIVEN: an API with Basic Auth, routed through SetupRoutesAPI so that
	// "/api/v1/ws-token" sits behind the basic-auth middleware.
	cfg := config.Config{}
	cfg.Settings.Web.BasicAuth = &config.WebSettingsBasicAuth{
		Username: "test", Password: "123",
	}
	cfg.Settings.Web.BasicAuth.CheckValues()
	api, _ := NewAPI(&cfg)
	api.SetupRoutesAPI()
	ts := httptest.NewServer(api.BaseRouter)
	t.Cleanup(ts.Close)

	tokenURL := ts.URL + "/api/v1/ws-token"

	tests := map[string]struct {
		authHeader       string // raw "Authorization" header value ("" to omit it).
		wantUnauthorized bool
	}{
		"rejects request without basic auth": {
			authHeader:       "",
			wantUnauthorized: true,
		},
		"rejects request with invalid basic auth": {
			authHeader:       "Basic dGVzdDp3cm9uZw==", // "test:wrong".
			wantUnauthorized: true,
		},
		"issues a token with valid basic auth": {
			authHeader:       "Basic dGVzdDoxMjM=", // "test:123".
			wantUnauthorized: false,
		},
	}

	prefix := fmt.Sprintf("%q\nAPI.httpWebSocketToken() (auth-gated)", packageName)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN: "/ws-token" is requested with the given credentials.
			req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
			if err != nil {
				t.Fatalf(
					"%s request creation failed - %v",
					prefix, err,
				)
			}
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf(
					"%s request 'Do' failed - %v",
					prefix, err,
				)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })

			wantStatusCode := http.StatusOK
			if tc.wantUnauthorized {
				wantStatusCode = http.StatusUnauthorized
			}
			// THEN: the basic-auth middleware rejects unauthenticated/invalid credential requests
			// with 401, and a correctly authenticated request passes the middleware.
			if resp.StatusCode != wantStatusCode {
				t.Fatalf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, resp.StatusCode, wantStatusCode,
				)
			}
			// Rejected requests carry no token body to validate.
			if tc.wantUnauthorized {
				return
			}

			// AND: the body is a well-formed token that validates against the store.
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf(
					"%s unexpected error:\n%v",
					prefix, err,
				)
			}
			bodyRegex := fmt.Sprintf(`{"token":"[a-z0-9]{%d}"}`+"\n$", webSocketTokenLength)
			if got := string(data); !util.RegexCheck(bodyRegex, got) {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, got, bodyRegex,
				)
			}
			var tokenResp struct {
				Token string `json:"token"`
			}
			if err := decode.Unmarshal("json", data, &tokenResp); err != nil {
				t.Fatalf("%s failed to unmarshal response: %v", prefix, err)
			}
			if !api.wsTokens.Validate(tokenResp.Token) {
				t.Errorf(
					"%s returned token %q did not validate against the store",
					prefix, tokenResp.Token,
				)
			}
		})
	}
}

func TestHTTP_HTTPRuntimeInfo(t *testing.T) {
	// GIVEN: an API and a request for the runtime info.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}
	bodyRegex := test.TrimJSON(`
		{
			"start_time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[^"]*",
			"cwd":"[^"]+",
			"goroutines":[0-9]+,
			"GOMAXPROCS":[0-9]+
		}`,
	) + "\n$"

	// WHEN: that HTTP request is sent.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status/runtime", nil)
	w := httptest.NewRecorder()
	apiMu.RLock()
	api.httpRuntimeInfo(w, req)
	apiMu.RUnlock()
	res := w.Result()
	t.Cleanup(func() { _ = res.Body.Close() })

	prefix := fmt.Sprintf("%q\nAPI.httpRuntimeInfo()", packageName)

	// THEN: the expected body is returned.
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf(
			"%s unexpected error:\n%v",
			prefix, err,
		)
	}
	if got := string(data); !util.RegexCheck(bodyRegex, got) {
		t.Errorf(
			"%s error mismatch\ngot:  %q\nwant %q",
			prefix, got, bodyRegex,
		)
	}
}
