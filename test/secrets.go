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

var ValidCertHTTPS = "https://" + ValidCertNoProtocol
var InvalidCertHTTPS = "https://" + InvalidCertNoProtocol

var LookupPlain = map[string]string{
	"url_valid":   ValidCertHTTPS + "/plain",
	"url_invalid": InvalidCertHTTPS + "/plain"}

var LookupPlainPOST = map[string]string{
	"url_valid":   ValidCertHTTPS + "/plain_post",
	"url_invalid": InvalidCertHTTPS + "/plain_post",
	"data_pass":   `{"argus":"test"}`,
	"data_fail":   `{"argus":"test-"}`}

var LookupHeader = map[string]string{
	"url_valid":                  ValidCertHTTPS + "/header",
	"url_invalid":                InvalidCertHTTPS + "/header",
	"header_key_pass":            "X-Version-Here",
	"header_key_pass_mixed_case": "x-VeRSioN-HERe",
	"header_key_fail":            "X-Version-Foo"}

var LookupJSON = map[string]string{
	"url_valid":   ValidCertHTTPS + "/json",
	"url_invalid": InvalidCertHTTPS + "/json"}

var LookupGitHub = map[string]string{
	"url_valid":   ValidCertHTTPS + "/hooks/github-style",
	"url_invalid": InvalidCertHTTPS + "/hooks/github-style",
	"secret_pass": "argus",
	"secret_fail": "argus-"}

var LookupWithHeaderAuth = map[string]string{
	"url_valid":         ValidCertHTTPS + "/hooks/single-header",
	"url_invalid":       InvalidCertHTTPS + "/hooks/single-header",
	"header_key":        "X-Test",
	"header_value_pass": "secret",
	"header_value_fail": "secret-"}

var LookupBasicAuth = map[string]string{
	"url_valid":   ValidCertHTTPS + "/basic-auth",
	"url_invalid": InvalidCertHTTPS + "/basic-auth",
	"username":    "test",
	"password":    "123"}
