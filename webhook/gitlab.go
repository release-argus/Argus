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
	"net/url"
)

// SetGitLabParameter of the req based on the secret.
func SetGitLabParameter(req *http.Request, secret string) {
	q := url.Values{}
	if req.URL.Query() != nil && len(req.URL.Query()) != 0 {
		q = req.URL.Query()
	}
	if q["token"] == nil {
		q.Add("token", secret)
	}
	if q["ref"] == nil {
		q.Add("ref", "master")
	}
	req.URL.RawQuery = q.Encode()
}
