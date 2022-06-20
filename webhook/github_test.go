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
	"regexp"

	//#nosec G505 -- GitHub's X-Hub-Signature uses SHA-1
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"

	service_status "github.com/release-argus/Argus/service/status"
)

func TestSetCustomHeaders(t *testing.T) {
	{ // GIVEN CustomHeaders are nil
		req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
		if err != nil {
			t.Fatalf("http.NewRequest failed - %s", err.Error())
		}
		webhook := WebHook{CustomHeaders: nil}
		// WHEN SetCustomHeaders is called
		webhook.SetCustomHeaders(req)
		// THEN the function exits without setting any headers
		want := 0
		got := len(req.Header)
		if got != want {
			t.Fatalf("SetCustomHeaders of nil altered the Header count. Want nil, got %v", got)
		}
	}

	{ // GIVEN CustomHeaders contain some Jinja templates
		req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
		if err != nil {
			t.Fatalf("http.NewRequest failed - %s", err.Error())
		}
		webhook := WebHook{
			CustomHeaders: &map[string]string{
				"test":      "foo",
				"jinja_var": "bar {{ version }}-",
				"jinja_exp": "bang {% if version == '1.2.3' %}success{% endif %} bang",
			},
			ServiceStatus: &service_status.Status{LatestVersion: "1.2.3"},
		}
		// WHEN SetCustomHeaders is called
		webhook.SetCustomHeaders(req)
		// THEN the headers are all set correctly
		want := map[string]string{
			"test":      "foo",
			"jinja_var": "bar 1.2.3-",
			"jinja_exp": "bang success bang",
		}
		for key := range *webhook.CustomHeaders {
			if req.Header[key][0] != want[key] {
				t.Fatalf("Pongo2 template not evaluated correctly. %s - Wanted %s, got %s", key, want[key], req.Header[key][0])
			}
		}
	}
}

func TestSetGitHubHeaders(t *testing.T) {
	{ // GIVEN a secret and GitHub payload
		secret := "secret"
		req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
		if err != nil {
			t.Fatalf("http.NewRequest failed - %s", err.Error())
		}
		payload, err := json.Marshal(GitHub{
			Ref:    "refs/heads/master",
			Before: "0123456789012345678901234567890123456789",
			After:  "0123456789012345678901234567890123456789",
		})
		if err != nil {
			t.Fatalf("json.Marshal failed - %s", err.Error())
		}
		// WHEN SetGitHubHeaders is called
		SetGitHubHeaders(req, payload, secret)
		// THEN the GitHub headers are correctly added
		// X-Github-Event
		// X-Github-Hook-Installation-Target-Type
		want := map[string]string{
			"X-Github-Event":                         "push",
			"X-Github-Hook-Installation-Target-Type": "repository",
		}
		for key := range want {
			if req.Header[key][0] != want[key] {
				t.Fatalf("Headers don't match. %s - Wanted %s, got %s", key, want[key], req.Header[key][0])
			}
		}
		// X-Github-Hook-Id
		regex := "^[0-9]{9}$"
		key := "X-Github-Hook-Id"
		if match, _ := regexp.MatchString(regex, req.Header[key][0]); !match {
			t.Fatalf("%s - Wanted %s, got %s", key, regex, req.Header[key][0])
		}
		// X-Github-Delivery
		regex = "^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$"
		key = "X-Github-Delivery"
		if match, _ := regexp.MatchString(regex, req.Header[key][0]); !match {
			t.Fatalf("%s - Wanted %s, got %s", key, regex, req.Header[key][0])
		}
		// X-Github-Hook-Installation-Target-Id
		regex = "^[0-9]{9}$"
		key = "X-Github-Hook-Installation-Target-Id"
		if match, _ := regexp.MatchString(regex, req.Header[key][0]); !match {
			t.Fatalf("%s - Wanted %s, got %s", key, regex, req.Header[key][0])
		}
		// X-Hub-Signature-256.
		key = "X-Hub-Signature-256"
		hash := hmac.New(sha256.New, []byte(secret))
		hash.Write(payload)
		wantVal := hex.EncodeToString(hash.Sum(nil))
		if req.Header[key][0] != "sha256="+wantVal {
			t.Fatalf("%s - Wanted %s, got %s", key, regex, req.Header[key][0])
		}
		// X-Hub-Signature.
		key = "X-Hub-Signature"
		hash = hmac.New(sha1.New, []byte(secret))
		hash.Write(payload)
		wantVal = hex.EncodeToString(hash.Sum(nil))
		if req.Header[key][0] != "sha1="+wantVal {
			t.Fatalf("%s - Wanted %s, got %s", key, regex, req.Header[key][0])
		}
	}
}
