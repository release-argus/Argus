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

package webhook

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetGitLabParameter(t *testing.T) {
	// GIVEN: a HTTP Request.
	tests := []struct {
		name        string
		secret      string
		queryParams string
		want        string
	}{
		{
			name:        "query param override ref",
			secret:      "fizz",
			queryParams: "?ref=main",
			want:        "ref=main&token=fizz",
		},
		{
			name:        "query param override secret as well as ref",
			secret:      "fizz",
			queryParams: "?ref=main&token=bang",
			want:        "ref=main&token=bang",
		},
		{
			name:        "query param add other params",
			secret:      "fizz",
			queryParams: "?ref=main&token=bang&foo=bar",
			want:        "foo=bar&ref=main&token=bang",
		},
		{
			name:   "no query params",
			secret: "bang",
			want:   "ref=master&token=bang",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals"+tc.queryParams, nil)

			// WHEN: SetGitLabParameter is called.
			SetGitLabParameter(req, tc.secret)

			// THEN: the function correctly encodes URL.RawQuery.
			got := req.URL.RawQuery
			if got != tc.want {
				t.Errorf(
					"%s\nSetGitLabParameter(req, secret=%q) didn't put secret into RawQuery\ngot:  %q\nwant: %q",
					packageName, tc.secret,
					tc.want, got,
				)
			}
		})
	}
}
