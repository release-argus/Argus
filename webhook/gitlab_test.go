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

package webhook

import (
	"net/http"
	"testing"
)

func TestSetGitLabParameter(t *testing.T) {
	{ // GIVEN a URL without query params
		req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
		if err != nil {
			t.Fatalf("http.NewRequest failed - %s", err.Error())
		}
		whSecret := "secret"
		// WHEN SetGitLabParameter is called
		SetGitLabParameter(req, whSecret)
		// THEN the function correctly encodes URL.RawQuery
		want := "ref=master&token=secret"
		got := req.URL.RawQuery
		if got != want {
			t.Fatalf("SetGitLabParameter failed. Want %s, got %s", want, got)
		}
	}

	{ // GIVEN a URL with query params
		req, err := http.NewRequest(http.MethodGet, "https://example.com?test=123", nil)
		if err != nil {
			t.Fatalf("http.NewRequest failed - %s", err.Error())
		}
		whSecret := "secret"
		// WHEN SetGitLabParameter is called
		SetGitLabParameter(req, whSecret)
		// THEN the function correctly encodes URL.RawQuery
		want := "ref=master&test=123&token=secret"
		got := req.URL.RawQuery
		if got != want {
			t.Fatalf("SetGitLabParameter failed. Want %s, got %s", want, got)
		}
	}
}
