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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRequireQueryParam(t *testing.T) {
	tests := []struct {
		name           string
		queryParamName string
		queryValue     string
		omitParam      bool
		statusCode     int
	}{
		{
			name:           "Simple query parameter",
			queryParamName: "key",
			queryValue:     "value",
			statusCode:     http.StatusOK,
		},
		{
			name:           "Complex query parameter",
			queryParamName: "key",
			queryValue:     `test/app&123%2F%x?name=John+Doe#section\né漢字🚀`,
			statusCode:     http.StatusOK,
		},
		{
			name:           "Query parameter missing",
			queryParamName: "key",
			omitParam:      true,
			statusCode:     http.StatusBadRequest,
		},
		{
			name:           "Query parameter present but empty",
			queryParamName: "key",
			queryValue:     "",
			statusCode:     http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			params := url.Values{"app": {"argus"}}
			if !tc.omitParam {
				params.Set(tc.queryParamName, tc.queryValue)
			}

			req := &http.Request{URL: &url.URL{RawQuery: params.Encode()}}
			w := httptest.NewRecorder()

			value, ok := requireQueryParam(w, req, tc.queryParamName)

			resp := w.Result()
			defer resp.Body.Close()

			if value != tc.queryValue {
				t.Errorf("value mismatch\ngot:  %q\nwant: %q", value, tc.queryValue)
			}

			if got, want := resp.StatusCode, tc.statusCode; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			wantOk := tc.statusCode == http.StatusOK
			if ok != wantOk {
				t.Errorf("ok mismatch\ngot:  %t\nwant: %t", ok, wantOk)
			}
		})
	}
}
