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

//go:build integration

package docker

import (
	"regexp"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestHubRegistry_Check(t *testing.T) {
	tests := []struct {
		name     string
		registry HubRegistry
		version  string
		errRegex string
	}{
		{
			name: "no auth, known image+tag",
			registry: HubRegistry{
				Image: test.ArgusDockerHubRepo,
				Tag:   "{{ version }}",
				Auth:  &HubAuth{},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, known image+tag",
			registry: HubRegistry{
				Image: test.ArgusDockerHubRepo,
				Tag:   "{{ version }}",
				Auth: &HubAuth{
					Username:   test.DockerHubUsername(t),
					Token:      test.DockerHubToken(t),
					queryToken: "123",
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, unknown image",
			registry: HubRegistry{
				Image: test.ArgusDockerHubRepo + "-unknown",
				Tag:   "{{ version }}",
				Auth: &HubAuth{
					Username:   test.DockerHubUsername(t),
					Token:      test.DockerHubToken(t),
					queryToken: "123",
				},
			},
			version:  "latest",
			errRegex: `^` + regexp.QuoteMeta(test.ArgusDockerHubRepo) + `-unknown:latest - tag not found$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.registry.Check(tc.version)
			e := errfmt.FormatError(err)
			if util.RegexCheck(`(?i)rate limit exceeded`, e) {
				t.Skipf("registry rate-limited:\n%s", e)
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestHubRegistry_Check__errors(t *testing.T) {
	tests := []struct {
		name            string
		hubTokenAddress string
		hubQueryURL     string
		registry        HubRegistry
		version         string
		errRegex        string
	}{
		{
			name:            "GetQueryToken error, expired query token (token req fails)",
			hubTokenAddress: "https://	example.com",
			registry: HubRegistry{
				Image: test.ArgusDockerHubRepo,
				Tag:   "{{ version }}",
				Auth: &HubAuth{
					Username:   "test",
					Token:      "test",
					queryToken: "test",
					validUntil: time.Now().UTC().Add(-10 * time.Minute),
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^create docker-hub token request:
					parse "https://.*
						.*invalid control character in URL$`,
			),
		},
		{
			name:        "newRequest error, invalid URL",
			hubQueryURL: "https://	example.com",
			registry: HubRegistry{
				Image: test.ArgusDockerHubRepo + "-unknown",
				Tag:   "{{ version }}",
				Auth: &HubAuth{
					Username:   "test",
					Token:      "test",
					queryToken: "test",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				parse "https://.*
					.*invalid control character in URL$`,
			),
		},
		{
			name:        "http.client.Do error, invalid URL TLD",
			hubQueryURL: "https://example.invalid/%s/%s",
			registry: HubRegistry{
				Image: test.ArgusDockerHubRepo,
				Tag:   "{{ version }}",
				Auth: &HubAuth{
					Username:   "test",
					Token:      "test",
					queryToken: "test",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				` + regexp.QuoteMeta(test.ArgusDockerHubRepo) + `:latest
					Get "https://.*
						dial tcp:
							lookup .*$`,
			),
		},
	}
	_hubTokenAddress := hubTokenAddress
	_hubQueryURL := hubQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps hubTokenAddress and hubQueryURL.

			if tc.hubTokenAddress != "" {
				hubTokenAddress = tc.hubTokenAddress
				t.Cleanup(func() {
					hubTokenAddress = _hubTokenAddress
				})
			}
			if tc.hubQueryURL != "" {
				hubQueryURL = tc.hubQueryURL
				t.Cleanup(func() {
					hubQueryURL = _hubQueryURL
				})
			}

			err := tc.registry.Check(tc.version)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestHubAuth_GetQueryToken__integration(t *testing.T) {
	tests := []struct {
		name     string
		data     *HubAuth
		detail   ContainerDetail
		want     string
		errRegex string
	}{
		{
			name: "valid query token cached",
			data: &HubAuth{
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "query token expired, fail fetching invalid image",
			data: &HubAuth{
				Username:   "test",
				Token:      "foo",
				queryToken: "query-token",
				validUntil: time.Now().Add(-10 * time.Second),
			},
			detail: ContainerDetail{
				Image: test.ArgusDockerHubRepo + "-unknown",
			},
			errRegex: `^docker-hub token request fail`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queryToken, err := tc.data.GetQueryToken(tc.detail)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  len=%d\nwant: %q", len(queryToken), tc.want)
			}

			if tc.data.queryToken != "" && !isUsable(tc.data.queryToken, tc.data.validUntil) {
				t.Errorf(
					"isUsable() mismatch\ngot:  false (validUntil=%s, now=%s)\nwant: true",
					tc.data.validUntil, time.Now().UTC(),
				)
			}
		})
	}
}

func TestHubAuth_RefreshQueryToken__integration(t *testing.T) {
	tests := []struct {
		name            string
		detail          ContainerDetail
		username, token string
		tokenAddress    *string
		emptyQueryToken bool
		errRegex        string
	}{
		{
			name: "no username",
			detail: ContainerDetail{
				Image: test.ArgusDockerHubRepo,
			},
			token:           test.DockerHubToken(t),
			emptyQueryToken: true,
			errRegex:        `^$`,
		},
		{
			name: "no token",
			detail: ContainerDetail{
				Image: test.ArgusDockerHubRepo,
			},
			username:        test.DockerHubUsername(t),
			emptyQueryToken: true,
			errRegex:        `^$`,
		},
		{
			name: "no username, no token",
			detail: ContainerDetail{
				Image: test.ArgusDockerHubRepo,
			},
			emptyQueryToken: true,
			errRegex:        `^$`,
		},
		{
			name: "valid username, valid token",
			detail: ContainerDetail{
				Image: test.ArgusDockerHubRepo,
			},
			username: test.DockerHubUsername(t),
			token:    test.DockerHubToken(t),
			errRegex: `^$`,
		},
		{
			name: "invalid/url",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: new(test.LookupPlain.URLValid + "	123"),
			errRegex:     `^create docker-hub token request`,
		},
		{
			name: "invalid/expired cert",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: new(test.LookupPlain.URLInvalid),
			errRegex:     `^docker-hub token request failed`,
		},
		{
			name: "404 response",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: new(test.WebhookGitHub.URLValid + "/123"),
			errRegex:     `^docker-hub token request failed \(status=404\)`,
		},
		{
			name: "invalid response JSON",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: new(test.LookupPlain.URLValid),
			errRegex:     `^failed to parse docker-hub token response`,
		},
	}
	_hubTokenAddress := hubTokenAddress

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps hubTokenAddress.

			if tc.tokenAddress != nil {
				hubTokenAddress = *tc.tokenAddress
				t.Cleanup(func() {
					hubTokenAddress = _hubTokenAddress
				})
			}

			data := &HubAuth{
				Username: tc.username,
				Token:    tc.token,
			}

			queryToken, err := data.refreshQueryToken(tc.detail)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if gotEmpty := queryToken == ""; err == nil && gotEmpty != tc.emptyQueryToken {
				t.Errorf(
					"queryToken mismatch\ngot:  empty=%t\nwant: empty=%t",
					gotEmpty, tc.emptyQueryToken,
				)
			}
		})
	}
}
