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
	"crypto/hmac"
	"fmt"

	//#nosec G505 -- GitHub's X-Hub-Signature uses SHA-1
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

func TestSetGitHubHeaders(t *testing.T) {
	// GIVEN: a secret and a payload to send.
	tests := []struct {
		name   string
		secret string
	}{
		{
			name:   "empty secret",
			secret: "",
		},
		{
			name:   "defined secret",
			secret: "123",
		},
	}

	// AND: headers to verify.
	type headerTest struct {
		key   string
		regex string
		exact func(payload []byte, secret string) string
	}
	headerTests := []headerTest{
		{
			key:   "X-Github-Event",
			regex: "^push$",
		},
		{
			key:   "X-Github-Hook-Id",
			regex: "^[0-9]{9}$",
		},
		{
			key:   "X-Github-Delivery",
			regex: "^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}$",
		},
		{
			key:   "X-Github-Hook-Installation-Target-Id",
			regex: "^[0-9]{9}$",
		},
		{
			key:   "X-Github-Hook-Installation-Target-Type",
			regex: "^repository$",
		},
		{
			key: "X-Hub-Signature",
			exact: func(payload []byte, secret string) string {
				hash := hmac.New(sha1.New, []byte(secret))
				hash.Write(payload)
				return "sha1=" + hex.EncodeToString(hash.Sum(nil))
			},
		},
		{
			key: "X-Hub-Signature-256",
			exact: func(payload []byte, secret string) string {
				hash := hmac.New(sha256.New, []byte(secret))
				hash.Write(payload)
				return "sha256=" + hex.EncodeToString(hash.Sum(nil))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payload, _ := decode.Marshal(
				"json", GitHub{
					Ref:    "refs/heads/master",
					Before: "0123456789012345678901234567890123456789",
					After:  "0123456789012345678901234567890123456789",
				},
			)
			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)

			// WHEN: SetGitHubHeaders is called.
			SetGitHubHeaders(req, payload, tc.secret)

			prefix := fmt.Sprintf(
				"%s\nSetGitHubHeaders(secret=%q)",
				packageName, tc.secret,
			)

			// THEN: the GitHub headers are correctly added.
			for _, hTC := range headerTests {
				got := getHeaderKey(req.Header[hTC.key])

				switch {
				case hTC.regex != "":
					if !util.RegexCheck(hTC.regex, got) {
						t.Errorf(
							"%s %q regex mismatch\ngot:  %q\nwant regex: %q",
							prefix, hTC.key,
							got, hTC.regex,
						)
					}
				case hTC.exact != nil:
					want := hTC.exact(payload, tc.secret)
					if got != want {
						t.Errorf(
							"%s key=%q value mismatch\ngot:  %q\nwant: %q",
							prefix, hTC.key,
							got, want,
						)
					}
				}
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
