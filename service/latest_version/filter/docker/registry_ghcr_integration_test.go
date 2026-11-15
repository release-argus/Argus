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

func TestGHCRRegistry_Check(t *testing.T) {
	tests := []struct {
		name     string
		registry GHCRRegistry
		version  string
		errRegex string
	}{
		{
			name: "no auth, known image+tag",
			registry: GHCRRegistry{
				Image: test.ArgusDockerGHCRRepo,
				Tag:   "{{ version }}",
				Auth:  &GHCRAuth{},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, known image+tag",
			registry: GHCRRegistry{
				Image: test.ArgusDockerGHCRRepo,
				Tag:   "{{ version }}",
				Auth: &GHCRAuth{
					Token:      test.GitHubToken(t),
					queryToken: test.GitHubTokenEncoded(t),
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth, known image, unknown tag",
			registry: GHCRRegistry{
				Image: test.ArgusDockerGHCRRepo,
				Tag:   "{{ version }}-unknown",
				Auth: &GHCRAuth{
					Token:      test.GitHubToken(t),
					queryToken: test.GitHubTokenEncoded(t),
				},
			},
			version:  "latest",
			errRegex: `^` + regexp.QuoteMeta(test.ArgusDockerGHCRRepo) + `:latest-unknown - tag not found$`,
		},
		{
			name: "auth, unknown image",
			registry: GHCRRegistry{
				Image: test.ArgusDockerGHCRRepo + "-unknown",
				Tag:   "{{ version }}",
				Auth: &GHCRAuth{
					Token:      test.GitHubToken(t),
					queryToken: test.GitHubTokenEncoded(t),
				},
			},
			version:  "latest",
			errRegex: `^` + regexp.QuoteMeta(test.ArgusDockerGHCRRepo) + `-unknown:latest - tag not found$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.registry.Check(tc.version)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestGHCRRegistry_Check__errors(t *testing.T) {
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
				Image: test.ArgusDockerGHCRRepo,
				Tag:   "{{ version }}",
				Auth:  &GHCRAuth{},
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
				Image: test.ArgusDockerGHCRRepo + "-unknown",
				Tag:   "{{ version }}",
				Auth: &GHCRAuth{
					Token:      "test",
					queryToken: "test",
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
				Image: test.ArgusDockerGHCRRepo,
				Tag:   "{{ version }}",
				Auth: &GHCRAuth{
					Token:      "test",
					queryToken: "test",
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^` + regexp.QuoteMeta(test.ArgusDockerGHCRRepo) + `:latest
					Get "https://.*
						dial tcp:
							lookup .*$`,
			),
		},
	}
	_ghcrTokenAddress := ghcrTokenAddress
	_ghcrQueryURL := ghcrQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps ghcrTokenAddress and ghcrQueryURL.

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

func TestGHCRAuth_GetQueryToken__integration(t *testing.T) {
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
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "query token expired, fail fetching invalid image",
			data: &GHCRAuth{
				queryToken: "query-token",
				validUntil: time.Now().Add(-10 * time.Second),
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

func TestGHCRAuth_RefreshQueryToken__integration(t *testing.T) {
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
			tokenAddress: new(test.LookupPlain.URLValid + "%s"),
			errRegex:     `^failed to parse ghcr token response`,
		},
	}
	_ghcrTokenAddress := ghcrTokenAddress

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps ghcrTokenAddress.

			if tc.tokenAddress != nil {
				ghcrTokenAddress = *tc.tokenAddress
				t.Cleanup(func() {
					ghcrTokenAddress = _ghcrTokenAddress
				})
			}

			data := &GHCRAuth{}

			queryToken, err := data.refreshQueryToken(tc.detail)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if err == nil && queryToken == "" {
				t.Error("queryToken mismatch\ngot:  \"\"\nwant: non-empty")
			}
		})
	}
}
