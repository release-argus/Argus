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

package v1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRequireQueryParam(t *testing.T) {
	// GIVEN a query parameter.
	tests := []struct {
		name           string
		queryParamName string
		queryValue     string
		ok             bool
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
			queryValue:     `release-argus/argus&123%2F%x?name=John+Doe#section\né漢字🚀`,
			statusCode:     http.StatusOK,
		},
		{
			name:           "Query parameter missing",
			queryParamName: "key",
			queryValue:     "",
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

			params := url.Values{
				"app":             {"argus"},
				tc.queryParamName: {tc.queryValue}}
			// AND a request made with this query parameter.
			req := &http.Request{URL: &url.URL{RawQuery: params.Encode()}}
			w := httptest.NewRecorder()

			// WHEN requireQueryParam is called on it.
			value, ok := requireQueryParam(w, req, tc.queryParamName)

			resp := w.Result()
			defer resp.Body.Close()

			// THEN the expected value is returned.
			if value != tc.queryValue {
				t.Errorf("%s\nquery parameter value mismatch\nwant: %q\ngot:  %q",
					tc.name, tc.queryValue, value)
			}
			// AND the expected status code is returned.
			if resp.StatusCode != tc.statusCode {
				t.Errorf("%s\nstatus code mismatch\nwant: %d\ngot:  %d",
					tc.name, tc.statusCode, resp.StatusCode)
			}
			// AND the expected ok status is returned.
			wantOk := tc.statusCode == http.StatusOK
			if ok != wantOk {
				t.Errorf("%s\nok mismatch\nwant: %v\ngot:  %v",
					tc.name, wantOk, ok)
			}
		})
	}
}
