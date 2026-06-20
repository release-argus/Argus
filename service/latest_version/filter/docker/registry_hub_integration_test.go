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
	"fmt"
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
	// GIVEN: a HubRegistry, and version to check for.
	tests := []struct {
		name     string
		registry HubRegistry
		version  string
		errRegex string
	}{
		{
			name: "no auth, known image+tag",
			registry: HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerHubRepo,
						Tag:   "{{ version }}",
					},
					Auth: &HubAuth{},
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, known image+tag",
			registry: HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerHubRepo,
						Tag:   "{{ version }}",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username:   test.DockerHubUsername(t),
							Token:      test.DockerHubToken(t),
							queryToken: "123",
						},
					},
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, unknown image",
			registry: HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerHubRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Token:      test.GitHubToken(t),
							queryToken: "123",
						},
					},
				},
			},
			version:  "latest",
			errRegex: `^` + test.ArgusDockerHubRepo + `-unknown:latest - tag not found$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Check is called with this version.
			err := tc.registry.Check(tc.version)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nHubRegistry.Check() error mismatch\ngot:  %q\nwant: %q",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}

func TestHubRegistry_Check__errors(t *testing.T) {
	// GIVEN: a HubRegistry, and version to check for.
	tests := []struct {
		name            string
		hubTokenAddress string
		hubQueryURL     string
		registry        HubRegistry
		version         string
		errRegex        string
	}{
		{
			name:            "GetQueryToken error, no token (no-op token req fails)",
			hubTokenAddress: "https://	example.com",
			registry: HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerHubRepo,
						Tag:   "{{ version }}",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username:   "test",
							Token:      "test",
							queryToken: "test",
							validUntil: time.Now().UTC().Add(-10 * time.Minute),
						},
					},
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerHubRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username:   "test",
							Token:      "test",
							queryToken: "test",
							validUntil: time.Now().UTC().Add(10 * time.Minute),
						},
					},
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerHubRepo,
						Tag:   "{{ version }}",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username:   "test",
							Token:      "test",
							queryToken: "test",
							validUntil: time.Now().UTC().Add(10 * time.Minute),
						},
					},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				` + test.ArgusDockerHubRepo + `:latest
					Get "https://.*
						dial tcp:
							lookup .* no such host$`,
			),
		},
	}
	_hubTokenAddress := hubTokenAddress
	_hubQueryURL := hubQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

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

			// WHEN: Check is called with this version.
			err := tc.registry.Check(tc.version)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nHubRegistry.Check() error mismatch\ngot:  %q\nwant: %q",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestHubAuth_GetQueryToken__integration(t *testing.T) {
	// GIVEN: a HubAuth and ContainerDetail to fetch a queryToken for.
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
				HubAuthDefaults: HubAuthDefaults{
					queryToken: "query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "query token expired, fail fetching invalid image",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username:   "test",
					Token:      "foo",
					queryToken: "query-token",
					validUntil: time.Now().Add(-10 * time.Second),
				},
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

			// WHEN: GetQueryToken() is called on it.
			queryToken, err := tc.data.GetQueryToken(tc.detail)

			prefix := fmt.Sprintf(
				"%s\nHubAuth.GetQueryToken(%+v)",
				packageName, tc.detail,
			)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch:\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: the query token is returned as expected.
			if queryToken != tc.want {
				t.Errorf(
					"%s mismatch on queryToken\ngot:  %q\nwant: %q",
					prefix, queryToken, tc.want,
				)
			}

			// AND: the query token expiry time is always valid.
			if tc.data.queryToken != "" && !isUsable(tc.data.queryToken, tc.data.validUntil) {
				t.Errorf(
					"%s returned a queryToken that shouldn't be used\nnow:       %s\nvalidUntil: %s",
					prefix, time.Now().UTC(), tc.data.validUntil,
				)
			}
		})
	}
}

func TestHubAuth_RefreshQueryToken__integration(t *testing.T) {
	// GIVEN: a ContainerDetail to fetch a queryToken for.
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
			username:        test.DockerHubUsername(t),
			token:           test.DockerHubToken(t),
			emptyQueryToken: true,
			errRegex:        `^$`,
		},
		{
			name: "invalid/url",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: test.Ptr(test.LookupPlain["url_valid"] + "	123"),
			errRegex:     `^create docker-hub token request`,
		},
		{
			name: "invalid, expired cert",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: test.Ptr(test.LookupPlain["url_invalid"]),
			errRegex:     `^docker-hub token request failed`,
		},
		{
			name: "404 response",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: test.Ptr(test.WebHookGitHub["url_valid"] + "/123"),
			errRegex:     `^docker-hub token request failed \(status=404\)`,
		},
		{
			name: "invalid response JSON",
			detail: ContainerDetail{
				Image: "",
			},
			username:     "u",
			token:        "t",
			tokenAddress: test.Ptr(test.LookupPlain["url_valid"]),
			errRegex:     `^failed to parse docker-hub token response`,
		},
	}
	_hubTokenAddress := hubTokenAddress

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

			if tc.tokenAddress != nil {
				hubTokenAddress = *tc.tokenAddress
				t.Cleanup(func() {
					hubTokenAddress = _hubTokenAddress
				})
			}

			// AND: a HubAuth to fetch it with.
			data := &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: tc.username,
					Token:    tc.token,
				},
			}

			// WHEN: refreshQueryToken() is called on it.
			queryToken, err := data.refreshQueryToken(tc.detail)

			prefix := fmt.Sprintf(
				"%s\nHubAuth.refreshQueryToken(%+v)",
				packageName, tc.detail,
			)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch:\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: if we didn't error, then we received a query token.
			if err == nil && queryToken == "" && tc.emptyQueryToken == false {
				t.Errorf("%s returned an empty queryToken\nwant: non-empty", prefix)
			}
		})
	}
}
