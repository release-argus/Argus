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

//go:build unit || integration

package test

import (
	"fmt"
	"os"
)

// ShoutrrrGotifyToken returns the token for the Gotify test.
func ShoutrrrGotifyToken() (token string) {
	token = os.Getenv("ARGUS_TEST_GOTIFY_TOKEN")
	if token == "" {
		// trunk-ignore(gitleaks/generic-api-key)
		token = "AGE-LlHU89Q56uQ"
	}
	return
}

var ValidCertNoProtocol = "valid.release-argus.io"
var InvalidCertNoProtocol = "invalid.release-argus.io"

var ValidCertHTTPS = fmt.Sprintf("https://%s", ValidCertNoProtocol)
var InvalidCertHTTPS = fmt.Sprintf("https://%s", InvalidCertNoProtocol)

var LookupPlain = map[string]string{
	"url_valid":   fmt.Sprintf("%s/plain", ValidCertHTTPS),
	"url_invalid": fmt.Sprintf("%s/plain", InvalidCertHTTPS)}

var LookupPlainPOST = map[string]string{
	"url_valid":   fmt.Sprintf("%s/plain_post", ValidCertHTTPS),
	"url_invalid": fmt.Sprintf("%s/plain_post", InvalidCertHTTPS),
	"data_pass":   `{"argus":"test"}`,
	"data_fail":   `{"argus":"test-"}`}

var LookupHeader = map[string]string{
	"url_valid":                  fmt.Sprintf("%s/header", ValidCertHTTPS),
	"url_invalid":                fmt.Sprintf("%s/header", InvalidCertHTTPS),
	"header_key_pass":            "X-Version-Here",
	"header_key_pass_mixed_case": "x-VeRSioN-HERe",
	"header_key_fail":            "X-Version-Foo"}

var LookupJSON = map[string]string{
	"url_valid":   fmt.Sprintf("%s/json", ValidCertHTTPS),
	"url_invalid": fmt.Sprintf("%s/json", InvalidCertHTTPS)}

var LookupGitHub = map[string]string{
	"url_valid":   fmt.Sprintf("%s/hooks/github-style", ValidCertHTTPS),
	"url_invalid": fmt.Sprintf("%s/hooks/github-style", InvalidCertHTTPS),
	"secret_pass": "argus",
	"secret_fail": "argus-"}

var LookupWithHeaderAuth = map[string]string{
	"url_valid":         fmt.Sprintf("%s/hooks/single-header", ValidCertHTTPS),
	"url_invalid":       fmt.Sprintf("%s/hooks/single-header", InvalidCertHTTPS),
	"header_key":        "X-Test",
	"header_value_pass": "secret",
	"header_value_fail": "secret-"}

var LookupBasicAuth = map[string]string{
	"url_valid":   fmt.Sprintf("%s/basic-auth", ValidCertHTTPS),
	"url_invalid": fmt.Sprintf("%s/basic-auth", InvalidCertHTTPS),
	"username":    "test",
	"password":    "123"}
