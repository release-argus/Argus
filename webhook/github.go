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
	"crypto/hmac"
	//#nosec G505 -- GitHub's X-Hub-Signature uses SHA-1
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/release-argus/Argus/util"
)

// SetGitHubHeaders of the req based on the payload and secret.
func SetGitHubHeaders(req *http.Request, payload []byte, secret string) {
	req.Header.Set("X-Github-Event", "push")
	req.Header.Set("X-Github-Hook-Id", util.RandNumeric(9))
	req.Header.Set("X-Github-Delivery", fmt.Sprintf("%s-%s-%s-%s-%s",
		util.RandAlphaNumericLower(8),
		util.RandAlphaNumericLower(4),
		util.RandAlphaNumericLower(4),
		util.RandAlphaNumericLower(4),
		util.RandAlphaNumericLower(12)))
	req.Header.Set("X-Github-Hook-Installation-Target-Id", util.RandNumeric(9))
	req.Header.Set("X-Github-Hook-Installation-Target-Type", "repository")

	// X-Hub-Signature-256.
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(payload)
	req.Header.Set("X-Hub-Signature-256", fmt.Sprintf("sha256=%s", hex.EncodeToString(hash.Sum(nil))))

	// X-Hub-Signature.
	hash = hmac.New(sha1.New, []byte(secret))
	hash.Write(payload)
	req.Header.Set("X-Hub-Signature", fmt.Sprintf("sha1=%s", hex.EncodeToString(hash.Sum(nil))))
}
