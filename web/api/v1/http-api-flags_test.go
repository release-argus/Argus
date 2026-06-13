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
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestHTTP_HTTPFlags(t *testing.T) {
	// GIVEN: an API and a request for the flag var values.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)
	apiMu := sync.RWMutex{}
	bodyRegex := test.TrimJSON(`
		{
			"config.file": "`+file+`",
			"log.level": "`+api.Config.Settings.LogLevel()+`",
			"log.timestamps": `+strconv.FormatBool(*api.Config.Settings.LogTimestamps())+`,
			"data.database-file": "`+api.Config.Settings.DataDatabaseFile()+`",
			"web.listen-host": "`+api.Config.Settings.WebListenHost()+`",
			"web.listen-port": "[0-9]{1,5}",
			"web.cert-file": "",
			"web.pkey-file": "",
			"web.route-prefix": "`+strings.ReplaceAll(api.Config.Settings.WebRoutePrefix(), "/", `\/`)+`"
		}`,
	) + "\n$"

	// WHEN: that HTTP request is sent.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/flags", nil)
	w := httptest.NewRecorder()
	apiMu.RLock()
	api.httpFlags(w, req)
	apiMu.RUnlock()
	res := w.Result()
	t.Cleanup(func() { _ = res.Body.Close() })

	prefix := fmt.Sprintf("%s\nAPI.httpFlags()", packageName)

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
			"%s response body mismatch\ngot:  %q\nwant: %q",
			prefix, got, bodyRegex,
		)
	}
}
