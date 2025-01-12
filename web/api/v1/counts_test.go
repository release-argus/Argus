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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/web/metric"
)

var router *mux.Router

func TestHTTP_Counts(t *testing.T) {
	// GIVEN values of metrics.
	tests := map[string]struct {
		ServiceCountCurrent     int
		UpdatesCurrentAvailable int
		UpdatesCurrentSkipped   int
	}{
		"empty": {},
		"ServiceCount": {
			ServiceCountCurrent: 9,
		},
		"UpdatesCurrent('AVAILABLE')": {
			UpdatesCurrentAvailable: 5,
		},
		"UpdatesCurrent('SKIPPED')": {
			UpdatesCurrentSkipped: 8,
		},
		"all": {
			ServiceCountCurrent:     5,
			UpdatesCurrentAvailable: 2,
			UpdatesCurrentSkipped:   7,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			api := API{}

			metric.ServiceCountCurrent.Set(float64(tc.ServiceCountCurrent))
			metric.UpdatesCurrent.WithLabelValues("AVAILABLE").Set(float64(tc.UpdatesCurrentAvailable))
			metric.UpdatesCurrent.WithLabelValues("SKIPPED").Set(float64(tc.UpdatesCurrentSkipped))
			wantJSON := test.TrimJSON(fmt.Sprintf(`{
				"service_count": %d,
				"updates_available": %d,
				"updates_skipped": %d
			}`,
				tc.ServiceCountCurrent,
				tc.UpdatesCurrentAvailable,
				tc.UpdatesCurrentSkipped,
			)) + "\n"

			// WHEN a HTTP request is sent to the /counts endpoint.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/counts", nil)
			w := httptest.NewRecorder()
			api.httpCounts(w, req)
			res := w.Result()
			t.Cleanup(func() { res.Body.Close() })

			// THEN the set values are returned in the JSON response.
			data, _ := io.ReadAll(res.Body)
			if dataStr := string(data); dataStr != wantJSON {
				t.Errorf("/api/v1/counts body mismatch\n%q\nwant:\n%q",
					dataStr, wantJSON)
			}
		})
	}
}
