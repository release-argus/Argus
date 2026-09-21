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

// Endpoint is a valid/invalid URL pair for a test endpoint.
type Endpoint struct {
	URLValid   string
	URLInvalid string
}

// LookupBare is a lookup that gives the next path segment as the version
// (add /1.2.3 to the URL to get version 1.2.3).
var LookupBare = Endpoint{
	URLValid:   ValidCertHTTPS + "/bare",
	URLInvalid: InvalidCertHTTPS + "/bare",
}

// LookupPlain is a lookup that gives a plain text response with versions in the body.
var LookupPlain = Endpoint{
	URLValid:   ValidCertHTTPS + "/plain",
	URLInvalid: InvalidCertHTTPS + "/plain",
}

// plainPOSTEndpoint is an Endpoint that requires a POST request with a body.
type plainPOSTEndpoint struct {
	Endpoint
	DataPass string
	DataFail string
}

// LookupPlainPOST is a lookup that gives a plain text response with versions in the body, and requires a POST request.
var LookupPlainPOST = plainPOSTEndpoint{
	URLValid:   ValidCertHTTPS + "/plain_post",
	URLInvalid: InvalidCertHTTPS + "/plain_post",
	DataPass:   `{"argus":"test"}`,
	DataFail:   `{"argus":"test-"}`,
}

// responseHeaderEndpoint is an Endpoint that returns the version in a response header.
type responseHeaderEndpoint struct {
	Endpoint
	HeaderKeyPass          string
	HeaderKeyPassMixedCase string
	HeaderKeyFail          string
}

// LookupResponseHeader is a lookup for testing Webhooks with versions in their response headers.
var LookupResponseHeader = responseHeaderEndpoint{
	URLValid:               ValidCertHTTPS + "/header",
	URLInvalid:             InvalidCertHTTPS + "/header",
	HeaderKeyPass:          "X-Version-Here",
	HeaderKeyPassMixedCase: "x-VeRSioN-HERe",
	HeaderKeyFail:          "X-Version-Foo",
}

// LookupJSON is a lookup that gives a JSON response with versions in the body.
var LookupJSON = Endpoint{
	URLValid:   ValidCertHTTPS + "/json",
	URLInvalid: InvalidCertHTTPS + "/json",
}

// webhookEndpoint is an Endpoint that verifies a signed Webhook payload.
type webhookEndpoint struct {
	Endpoint
	SecretPass string
	SecretFail string
}

// WebhookGitHub is a lookup for testing Webhooks with versions in their response body.
var WebhookGitHub = webhookEndpoint{
	URLValid:   ValidCertHTTPS + "/hooks/github-style",
	URLInvalid: InvalidCertHTTPS + "/hooks/github-style",
	SecretPass: "argus",
	SecretFail: "argus-",
}

// headerAuthEndpoint is an Endpoint that requires a header for authentication.
type headerAuthEndpoint struct {
	Endpoint
	HeaderKey       string
	HeaderValuePass string
	HeaderValueFail string
}

// LookupWithHeaderAuth is a lookup for testing lookups that require header authentication.
var LookupWithHeaderAuth = headerAuthEndpoint{
	URLValid:        ValidCertHTTPS + "/hooks/single-header",
	URLInvalid:      InvalidCertHTTPS + "/hooks/single-header",
	HeaderKey:       "X-Test",
	HeaderValuePass: "secret",
	HeaderValueFail: "secret-",
}

// basicAuthEndpoint is an Endpoint that requires basic authentication.
type basicAuthEndpoint struct {
	Endpoint
	Username string
	Password string
}

// LookupWithBasicAuth is a lookup for testing lookups that require basic authentication.
var LookupWithBasicAuth = basicAuthEndpoint{
	URLValid:   ValidCertHTTPS + "/basic-auth",
	URLInvalid: InvalidCertHTTPS + "/basic-auth",
	Username:   "test",
	Password:   "123",
}

// gotifyEndpoint is a Gotify server reachable over both the valid and invalid cert domains.
type gotifyEndpoint struct {
	HostValid   string
	HostInvalid string
	Path        string
	// TokenPass is accepted by the server.
	TokenPass string
	// TokenRejected is well-formed, but unknown to the server, so it is rejected with a 401.
	TokenRejected string
	// TokenMalformed is rejected by Shoutrrr before any request is made.
	TokenMalformed string
}

// NotifyGotify is the Gotify endpoint used by the Shoutrrr tests.
var NotifyGotify = gotifyEndpoint{
	HostValid:   ValidCertNoProtocol,
	HostInvalid: InvalidCertNoProtocol,
	Path:        "gotify",
	TokenPass:   ShoutrrrGotifyToken(),
	// trunk-ignore(gitleaks/generic-api-key)
	TokenRejected:  "AGdjFCZugzJGhEG",
	TokenMalformed: "invalid",
}
