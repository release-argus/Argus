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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestHTTP_HTTPWebSocketToken(t *testing.T) {
	prefix := fmt.Sprintf("%q\nAPI.httpWebSocketToken()", packageName)
	file := filepath.Join(t.TempDir(), "config.yml")

	t.Run("no basic auth returns 204", func(t *testing.T) {
		// GIVEN: an API without Basic Auth (wsTokens is nil).
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
	})

	t.Run("with basic auth returns token", func(t *testing.T) {
		// GIVEN: an API with Basic Auth (wsTokens set).
		api := testAPI(t, file)
		api.wsTokens = newWebSocketTokenStore()
		bodyRegex := test.TrimJSON(`{"token":"[a-z0-9]{`+fmt.Sprint(webSocketTokenLength)+`}"}`) + "\n$"

		// WHEN: a request is made for a WebSocket token.
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ws-token", nil)
		w := httptest.NewRecorder()
		api.httpWebSocketToken(w, req)
		res := w.Result()
		t.Cleanup(func() { _ = res.Body.Close() })

		// THEN: a token matching the expected format is returned.
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf(
				"%s unexpected error:\n%v",
				prefix, err,
			)
		}
		if got := string(data); !util.RegexCheck(bodyRegex, got) {
			t.Errorf(
				"%s body mismatch\ngot:  %q\nwant: %q",
				prefix, got, bodyRegex,
			)
		}

		// AND: the returned token validates successfully against the store.
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
