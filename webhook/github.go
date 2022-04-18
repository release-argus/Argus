// Copyright [2022] [Hymenaios]
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
	"crypto/hmac"
	//#nosec G505 -- GitHub's X-Hub-Signature uses SHA-1
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/hymenaios-io/Hymenaios/utils"
)

// GitHub is the WebHook payload to emulate GitHub.
type GitHub struct {
	Ref    string `json:"ref"`    // "refs/heads/master"
	Before string `json:"before"` // "RandAlphaNumericLower(40)"
	After  string `json:"after"`  // "RandAlphaNumericLower(40)"
}

// SetGitHubHeaders of the req based on the payload and secret.
func SetGitHubHeaders(req *http.Request, payload []byte, secret string) *http.Request {
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-GitHub-Hook-ID", utils.RandNumeric(9))
	req.Header.Set("X-GitHub-Delivery", fmt.Sprintf("%s-%s-%s-%s-%s", utils.RandAlphaNumericLower(8), utils.RandAlphaNumericLower(4), utils.RandAlphaNumericLower(4), utils.RandAlphaNumericLower(4), utils.RandAlphaNumericLower(12)))
	req.Header.Set("X-GitHub-Hook-Installation-Target-ID", utils.RandNumeric(9))
	req.Header.Set("X-GitHub-Hook-Installation-Target-Type", "repository")

	// X-Hub-Signature-256.
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(payload)
	req.Header.Set("X-Hub-Signature-256", fmt.Sprintf("sha256=%s", hex.EncodeToString(hash.Sum(nil))))

	// X-Hub-Signature.
	hash = hmac.New(sha1.New, []byte(secret))
	hash.Write(payload)
	req.Header.Set("X-Hub-Signature", fmt.Sprintf("sha1=%s", hex.EncodeToString(hash.Sum(nil))))

	return req
}
