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

//go:build unit || integration

package test

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
)

func get(t *testing.T, key string) string {
	if t != nil {
		t.Helper()
	}
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return v
}

// ShoutrrrGotifyToken returns the token for the Gotify test.
func ShoutrrrGotifyToken() (token string) {
	token = os.Getenv("ARGUS_TEST_GOTIFY_TOKEN")
	if token == "" {
		// trunk-ignore(gitleaks/generic-api-key)
		token = "AGE-LlHU89Q56uQ"
	}
	return
}

var ArgusGitHubRepo = "release-argus/Argus"

// Amazon ECR Public Gallery Repo used for tests (AWS-owned, long-lived).
var ArgusDockerECRRepo = "docker/library/busybox"

// GHCR Repo for Argus.
var ArgusDockerGHCRRepo = "release-argus/argus"

// Docker Hub Repo for Argus.
var ArgusDockerHubRepo = "releaseargus/argus"

// Quay Repo for Argus.
var ArgusDockerQuayRepo = "argus-io/argus"

func DockerHubUsername(t *testing.T) string {
	t.Helper()
	return get(t, "DOCKER_HUB_USERNAME")
}
func DockerHubToken(t *testing.T) string {
	t.Helper()
	return get(t, "DOCKER_HUB_TOKEN")
}
func DockerQuayToken(t *testing.T) string {
	t.Helper()
	return get(t, "DOCKER_QUAY_TOKEN")
}
func GitHubToken(t *testing.T) string {
	if t != nil {
		t.Helper()
	}
	return get(t, "GITHUB_TOKEN")
}

// GitHubTokenEncoded is the base64-encoded GitHub token for GHCR queries.
func GitHubTokenEncoded(t *testing.T) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString([]byte(GitHubToken(t)))
}

// ValidCertNoProtocol is a domain with a valid TLS certificate, without the protocol prefix.
var ValidCertNoProtocol = "valid.release-argus.io"

// InvalidCertNoProtocol is a domain with an invalid TLS certificate, without the protocol prefix.
var InvalidCertNoProtocol = "invalid.release-argus.io"

// ValidCertHTTPS is a URL with a valid TLS certificate, using the HTTPS protocol.
var ValidCertHTTPS = "https://" + ValidCertNoProtocol

// InvalidCertHTTPS is a URL with an invalid TLS certificate, using the HTTPS protocol.
var InvalidCertHTTPS = "https://" + InvalidCertNoProtocol

// LookupBare is a lookup that gives the next path segment as the version
// (add /1.2.3 to the URL to get version 1.2.3).
var LookupBare = map[string]string{
	"url_valid":   ValidCertHTTPS + "/bare",
	"url_invalid": InvalidCertHTTPS + "/bare",
}

// LookupPlain is a lookup that gives a plain text response with versions in the body.
var LookupPlain = map[string]string{
	"url_valid":   ValidCertHTTPS + "/plain",
	"url_invalid": InvalidCertHTTPS + "/plain",
}

// LookupPlainPOST is a lookup that gives a plain text response with versions in the body, and requires a POST request.
var LookupPlainPOST = map[string]string{
	"url_valid":   ValidCertHTTPS + "/plain_post",
	"url_invalid": InvalidCertHTTPS + "/plain_post",
	"data_pass":   `{"argus":"test"}`,
	"data_fail":   `{"argus":"test-"}`,
}

// LookupResponseHeader is a lookup for testing WebHooks with versions in their response headers.
var LookupResponseHeader = map[string]string{
	"url_valid":                  ValidCertHTTPS + "/header",
	"url_invalid":                InvalidCertHTTPS + "/header",
	"header_key_pass":            "X-Version-Here",
	"header_key_pass_mixed_case": "x-VeRSioN-HERe",
	"header_key_fail":            "X-Version-Foo",
}

// LookupJSON is a lookup that gives a JSON response with versions in the body.
var LookupJSON = map[string]string{
	"url_valid":   ValidCertHTTPS + "/json",
	"url_invalid": InvalidCertHTTPS + "/json",
}

// WebHookGitHub is a lookup for testing WebHooks with versions in their response body.
var WebHookGitHub = map[string]string{
	"url_valid":   ValidCertHTTPS + "/hooks/github-style",
	"url_invalid": InvalidCertHTTPS + "/hooks/github-style",
	"secret_pass": "argus",
	"secret_fail": "argus-",
}

// LookupWithHeaderAuth is a lookup for testing lookups that require header authentication.
var LookupWithHeaderAuth = map[string]string{
	"url_valid":         ValidCertHTTPS + "/hooks/single-header",
	"url_invalid":       InvalidCertHTTPS + "/hooks/single-header",
	"header_key":        "X-Test",
	"header_value_pass": "secret",
	"header_value_fail": "secret-",
}

// LookupWithBasicAuth is a lookup for testing lookups that require basic authentication.
var LookupWithBasicAuth = map[string]string{
	"url_valid":   ValidCertHTTPS + "/basic-auth",
	"url_invalid": InvalidCertHTTPS + "/basic-auth",
	"username":    "test",
	"password":    "123",
}
