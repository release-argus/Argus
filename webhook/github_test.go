// Copyright [2024] [Argus]
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
	"crypto/hmac"
	//#nosec G505 -- GitHub's X-Hub-Signature uses SHA-1
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestSetGitHubHeaders(t *testing.T) {
	// GIVEN a secret and a payload
	tests := map[string]struct {
		secret string
	}{
		"empty secret":   {secret: ""},
		"defined secret": {secret: "123"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			payload, _ := json.Marshal(GitHub{
				Ref:    "refs/heads/master",
				Before: "0123456789012345678901234567890123456789",
				After:  "0123456789012345678901234567890123456789",
			})
			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)

			// WHEN SetGitHubHeaders is called
			SetGitHubHeaders(req, payload, tc.secret)

			// THEN the GitHub headers are correctly added
			key := "X-Github-Event"
			want := "^push$"
			if !util.RegexCheck(want, getHeaderKey(req.Header[key])) {
				t.Errorf("%s - Wanted %s, got %s",
					key, want, getHeaderKey(req.Header[key]))
			}
			key = "X-Github-Hook-Id"
			want = "^[0-9]{9}$"
			if !util.RegexCheck(want, getHeaderKey(req.Header[key])) {
				t.Errorf("%s - Wanted %s, got %s",
					key, want, getHeaderKey(req.Header[key]))
			}
			key = "X-Github-Delivery"
			want = "^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}$"
			if !util.RegexCheck(want, getHeaderKey(req.Header[key])) {
				t.Errorf("%s - Wanted %s, got %s",
					key, want, getHeaderKey(req.Header[key]))
			}
			key = "X-Github-Hook-Installation-Target-Id"
			want = "^[0-9]{9}$"
			if !util.RegexCheck(want, getHeaderKey(req.Header[key])) {
				t.Errorf("%s - Wanted %s, got %s",
					key, want, getHeaderKey(req.Header[key]))
			}
			key = "X-Github-Hook-Installation-Target-Type"
			want = "^repository$"
			if !util.RegexCheck(want, getHeaderKey(req.Header[key])) {
				t.Errorf("%s - Wanted %s, got %s",
					key, want, getHeaderKey(req.Header[key]))
			}
			key = "X-Hub-Signature"
			hash := hmac.New(sha1.New, []byte(tc.secret))
			hash.Write(payload)
			wantVal := hex.EncodeToString(hash.Sum(nil))
			want = "sha1=" + wantVal
			if getHeaderKey(req.Header[key]) != want {
				t.Errorf("%s - Wanted %s, got %s",
					key, want, getHeaderKey(req.Header[key]))
			}
			key = "X-Hub-Signature-256"
			hash = hmac.New(sha256.New, []byte(tc.secret))
			hash.Write(payload)
			wantVal = hex.EncodeToString(hash.Sum(nil))
			want = "sha256=" + wantVal
			if getHeaderKey(req.Header[key]) != want {
				t.Errorf("%s - Wanted %s, got %s",
					key, want, getHeaderKey(req.Header[key]))
			}
		})
	}
}

func getHeaderKey(header []string) string {
	if len(header) == 0 {
		return "<nil header>"
	}
	return header[0]
}
