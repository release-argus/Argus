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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

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
