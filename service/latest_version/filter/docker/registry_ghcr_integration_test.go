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

func TestGHCRRegistry_Check(t *testing.T) {
	// GIVEN: a GHCRRegistry, and version to check for.
	tests := []struct {
		name     string
		registry GHCRRegistry
		version  string
		errRegex string
	}{
		{
			name: "no auth, known image+tag",
			registry: GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &GHCRAuth{},
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, known image+tag",
			registry: GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      test.GitHubToken(t),
							queryToken: test.GitHubTokenEncoded(t),
						},
					},
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, known image, unknown tag",
			registry: GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo,
						Tag:   "{{ version }}-unknown",
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      test.GitHubToken(t),
							queryToken: test.GitHubTokenEncoded(t),
						},
					},
				},
			},
			version:  "latest",
			errRegex: `^release-argus\/argus:latest-unknown - tag not found$`,
		},
		{
			name: "auth, unknown image",
			registry: GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      test.GitHubToken(t),
							queryToken: test.GitHubTokenEncoded(t),
						},
					},
				},
			},
			version:  "latest",
			errRegex: `^release-argus\/argus-unknown:latest - tag not found$`,
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
					"%s\nGHCRRegistry.Check() error mismatch\ngot:  %q\nwant: %s",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}

func TestGHCRRegistry_Check__errors(t *testing.T) {
	// GIVEN: a GHCRRegistry, and version to check for.
	tests := []struct {
		name             string
		ghcrTokenAddress string
		ghcrQueryURL     string
		registry         GHCRRegistry
		version          string
		errRegex         string
	}{
		{
			name:             "GetQueryToken error, no token (no-op token req fails)",
			ghcrTokenAddress: "https://	example.com",
			registry: GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &GHCRAuth{},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^ghcr token refresh fail:
					parse .*
						.*invalid control character in URL$`,
			),
		},
		{
			name:         "newRequest error, invalid URL",
			ghcrQueryURL: "https://	example.com",
			registry: GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      "test",
							queryToken: "test",
						},
					},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				parse .*
					.*invalid control character in URL$`,
			),
		},
		{
			name:         "http.client.Do error, invalid URL TLD",
			ghcrQueryURL: "https://example.invalid/%s/%s",
			registry: GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      "test",
							queryToken: "test",
						},
					},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^` + test.ArgusDockerGHCRRepo + `:latest
					Get "https://.*
						dial tcp:
							lookup .* no such host$`,
			),
		},
	}
	_ghcrTokenAddress := ghcrTokenAddress
	_ghcrQueryURL := ghcrQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

			if tc.ghcrTokenAddress != "" {
				ghcrTokenAddress = tc.ghcrTokenAddress
				t.Cleanup(func() {
					ghcrTokenAddress = _ghcrTokenAddress
				})
			}
			if tc.ghcrQueryURL != "" {
				ghcrQueryURL = tc.ghcrQueryURL
				t.Cleanup(func() {
					ghcrQueryURL = _ghcrQueryURL
				})
			}

			// WHEN: Check is called with this version.
			err := tc.registry.Check(tc.version)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nGHCRRegistry.Check() error mismatch\ngot:  %q\nwant: %s",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestGHCRAuth_GetQueryToken__integration(t *testing.T) {
	// GIVEN: a GHCRAuth and ContainerDetail to fetch a queryToken for.
	tests := []struct {
		name     string
		data     *GHCRAuth
		detail   ContainerDetail
		want     string
		errRegex string
	}{
		{
			name: "valid query token cached",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					queryToken: "query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "query token expired, fail fetching invalid image",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					queryToken: "query-token",
					validUntil: time.Now().Add(-10 * time.Second),
				},
			},
			detail: ContainerDetail{
				Image: test.ArgusDockerGHCRRepo + "-unknown",
			},
			errRegex: `^ghcr token request failed`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetQueryToken() is called on it.
			queryToken, err := tc.data.GetQueryToken(tc.detail)

			prefix := fmt.Sprintf(
				"%s\nGHCRAuth.GetQueryToken(%+v)",
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

func TestGHCRAuth_RefreshQueryToken__integration(t *testing.T) {
	// GIVEN: a ContainerDetail to fetch a queryToken for.
	tests := []struct {
		name         string
		detail       ContainerDetail
		tokenAddress *string
		errRegex     string
	}{
		{
			name: "valid",
			detail: ContainerDetail{
				Image: test.ArgusDockerGHCRRepo,
			},
			errRegex: `^$`,
		},
		{
			name: "unknown image",
			detail: ContainerDetail{
				Image: test.ArgusDockerGHCRRepo + "-unknown",
			},
			errRegex: `^ghcr token request failed`,
		},
		{
			name: "invalid URL template",
			detail: ContainerDetail{
				Image: "release-argus/	argus",
			},
			errRegex: `^ghcr token refresh fail`,
		},
		{
			name: "unknown response",
			detail: ContainerDetail{
				Image: "",
			},
			tokenAddress: test.Ptr(test.LookupPlain["url_valid"] + "%s"),
			errRegex:     `^failed to parse ghcr token response`,
		},
	}
	_ghcrTokenAddress := ghcrTokenAddress

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

			if tc.tokenAddress != nil {
				ghcrTokenAddress = *tc.tokenAddress
				t.Cleanup(func() {
					ghcrTokenAddress = _ghcrTokenAddress
				})
			}

			// AND: a GHCRAuth to fetch it with.
			data := &GHCRAuth{}

			// WHEN: refreshQueryToken() is called on it.
			queryToken, err := data.refreshQueryToken(tc.detail)

			prefix := fmt.Sprintf(
				"%s\nGHCRAuth.refreshQueryToken(%+v)",
				packageName, tc.detail,
			)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nerror mismatch:\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: if we didn't error, then we received a query token.
			if err == nil && queryToken == "" {
				t.Errorf("%s returned an empty queryToken\nwant: non-empty", prefix)
			}
		})
	}
}
