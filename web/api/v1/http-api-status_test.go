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
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestHTTP_HTTPWebSocketToken(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws-token", nil)
	w := httptest.NewRecorder()
	api.httpWebSocketToken(w, req)
	resp := w.Result()
	t.Cleanup(func() { _ = resp.Body.Close() })

	// No token needed without Basic Auth.
	if got, want := resp.StatusCode, http.StatusNoContent; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}

func TestHTTP_HTTPWebSocketToken__authGated(t *testing.T) {
	cfg := config.Config{}
	cfg.Settings.Web.BasicAuth = &config.WebSettingsBasicAuth{
		Username: "test", Password: "123",
	}
	cfg.Settings.Web.BasicAuth.CheckValues()
	api, _ := NewAPI(&cfg)
	// Routed through SetupRoutesAPI so "/api/v1/ws-token" sits behind the basic-auth middleware.
	api.SetupRoutesAPI()
	ts := httptest.NewServer(api.BaseRouter)
	t.Cleanup(ts.Close)

	tokenURL := ts.URL + "/api/v1/ws-token"

	tests := []struct {
		name             string
		authHeader       string // Raw "Authorization" header value ("" to omit it).
		wantUnauthorized bool
	}{
		{
			name:             "rejects request without basic auth",
			authHeader:       "",
			wantUnauthorized: true,
		},
		{
			name:             "rejects request with invalid basic auth",
			authHeader:       "Basic dGVzdDp3cm9uZw==", // "test:wrong".
			wantUnauthorized: true,
		},
		{
			name:             "issues a token with valid basic auth",
			authHeader:       "Basic dGVzdDoxMjM=", // "test:123".
			wantUnauthorized: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
			if err != nil {
				t.Fatalf("http.NewRequest() failed: %v", err)
			}
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Client.Do() failed: %v", err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}

			wantStatusCode := http.StatusOK
			if tc.wantUnauthorized {
				wantStatusCode = http.StatusUnauthorized
			}

			if got, want := resp.StatusCode, wantStatusCode; got != want {
				t.Fatalf("status code mismatch\ngot:  %d - %q\nwant: %d", got, data, want)
			}
			// Rejected requests carry no token body to validate.
			if tc.wantUnauthorized {
				return
			}

			bodyRegex := fmt.Sprintf(`{"token":"[a-z0-9]{%d}"}`+"\n$", webSocketTokenLength)
			if got := string(data); !util.RegexCheck(bodyRegex, got) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, bodyRegex)
			}
			var tokenResp struct {
				Token string `json:"token"`
			}
			if err := decode.Unmarshal("json", data, &tokenResp); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if !api.wsTokens.Validate(tokenResp.Token) {
				t.Errorf(
					"webSocketTokenStore.Validate(%q) mismatch\ngot:  false\nwant: true",
					tokenResp.Token,
				)
			}
		})
	}
}

func TestHTTP_HTTPRuntimeInfo(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)
	bodyRegex := test.TrimJSON(`
		{
			"start_time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[^"]*",
			"cwd":"[^"]+",
			"goroutines":[0-9]+,
			"GOMAXPROCS":[0-9]+
		}`,
	) + "\n$"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status/runtime", nil)
	w := httptest.NewRecorder()
	api.httpRuntimeInfo(w, req)
	resp := w.Result()
	t.Cleanup(func() { _ = resp.Body.Close() })

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() failed: %v", err)
	}
	if got := string(data); !util.RegexCheck(bodyRegex, got) {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, bodyRegex)
	}
}
