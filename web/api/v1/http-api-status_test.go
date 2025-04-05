// Copyright [2025] [Argus]
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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestHTTP_httpRuntimeInfo(t *testing.T) {
	// GIVEN an API and a request for the runtime info.
	file := "TestHTTP_httpRuntimeInfo.yml"
	api := testAPI(file)
	apiMutex := sync.RWMutex{}
	t.Cleanup(func() {
		os.RemoveAll(file)
		if api.Config.Settings.Data.DatabaseFile != "" {
			os.RemoveAll(api.Config.Settings.Data.DatabaseFile)
		}
	})
	want := `
		{
			"start_time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[^"]*",
			"cwd":"[^"]+",
			"goroutines":[0-9]+,
			"GOMAXPROCS":[0-9]+
		}\s$`

	// WHEN that HTTP request is sent.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status/runtime", nil)
	w := httptest.NewRecorder()
	apiMutex.RLock()
	api.httpRuntimeInfo(w, req)
	apiMutex.RUnlock()
	res := w.Result()
	t.Cleanup(func() { res.Body.Close() })

	// THEN the expected body is returned as expected.
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("%s\nunexpected error - %v",
			packageName, err)
	}
	got := string(data)
	want = strings.ReplaceAll(want, "\t", "")
	want = strings.ReplaceAll(want, "\n", "")
	if !util.RegexCheck(want, got) {
		t.Errorf("%s\nerror mismatch\nwant %q\ngot:  %q",
			packageName, want, got)
	}
}
