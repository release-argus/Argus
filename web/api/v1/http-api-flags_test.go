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
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestHTTP_HTTPFlags(t *testing.T) {
	file := t.Name() + ".yml"
	api := testAPI(t, file)
	bodyRegex := test.TrimJSON(`
		{
			"config.file": "`+regexp.QuoteMeta(api.Config.File)+`",
			"log.level": "`+regexp.QuoteMeta(api.Config.Settings.LogLevel())+`",
			"log.timestamps": `+strconv.FormatBool(*api.Config.Settings.LogTimestamps())+`,
			"data.database-file": "`+regexp.QuoteMeta(api.Config.Settings.DataDatabaseFile())+`",
			"data.readonly": `+strconv.FormatBool(api.Config.Settings.DataReadonly())+`,
			"web.listen-host": "`+regexp.QuoteMeta(api.Config.Settings.WebListenHost())+`",
			"web.listen-port": "[0-9]{1,5}",
			"web.cert-file": "",
			"web.pkey-file": "",
			"web.route-prefix": "`+regexp.QuoteMeta(api.Config.Settings.WebRoutePrefix())+`"
		}`,
	) + "\n$"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/flags", nil)
	w := httptest.NewRecorder()
	api.httpFlags(w, req)
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
