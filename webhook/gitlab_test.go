// Copyright [2022] [Argus]
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetGitLabParameter(t *testing.T) {
	// GIVEN a HTTP Request
	tests := map[string]struct {
		secret      string
		queryParams string
		want        string
	}{
		"query param override ref":                   {secret: "fizz", queryParams: "?ref=main", want: "ref=main&token=fizz"},
		"query param override secret as well as ref": {secret: "fizz", queryParams: "?ref=main&token=bang", want: "ref=main&token=bang"},
		"query param add other params":               {secret: "fizz", queryParams: "?ref=main&token=bang&foo=bar", want: "foo=bar&ref=main&token=bang"},
		"no query params":                            {secret: "bang", want: "ref=master&token=bang"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/approvals%s", tc.queryParams), nil)

			// WHEN SetGitLabParameter is called
			SetGitLabParameter(req, tc.secret)

			// THEN the function correctly encodes URL.RawQuery
			got := req.URL.RawQuery
			if got != tc.want {
				t.Errorf("SetGitLabParameter failed. Want %s, got %s",
					tc.want, got)
			}
		})
	}
}
