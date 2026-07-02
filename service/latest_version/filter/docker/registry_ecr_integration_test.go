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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestECRRegistry_Check(t *testing.T) {
	// GIVEN: an ECRRegistry, and version to check for.
	tests := []struct {
		name     string
		registry ECRRegistry
		version  string
		errRegex string
	}{
		{
			name: "anonymous, known image+tag",
			registry: ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerECRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &ECRAuth{},
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "known image, unknown tag",
			registry: ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerECRRepo,
						Tag:   "{{ version }}-unknown",
					},
					Auth: &ECRAuth{},
				},
			},
			version:  "latest",
			errRegex: `^` + test.ArgusDockerECRRepo + `:latest-unknown - tag not found$`,
		},
		{
			name: "unknown image",
			registry: ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerECRRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &ECRAuth{},
				},
			},
			version:  "latest",
			errRegex: `^` + test.ArgusDockerECRRepo + `-unknown:latest - tag not found$`,
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
					"%s\nECRRegistry.Check() error mismatch\ngot:  %q\nwant: %s",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}

func TestECRRegistry_Check__errors(t *testing.T) {
	// GIVEN: an ECRRegistry, and version to check for.
	tests := []struct {
		name            string
		ecrTokenAddress string
		ecrQueryURL     string
		registry        ECRRegistry
		version         string
		errRegex        string
	}{
		{
			name:            "GetQueryToken error, invalid token URL",
			ecrTokenAddress: "https://	example.com",
			registry: ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerECRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &ECRAuth{},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^ecr token refresh fail:
					parse .*
						.*invalid control character in URL$`,
			),
		},
		{
			name:        "newRequest error, invalid query URL",
			ecrQueryURL: "https://	example.com",
			registry: ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerECRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &ECRAuth{
						ECRAuthDefaults: ECRAuthDefaults{
							queryToken: "test",
						},
					},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^parse .*
					.*invalid control character in URL$`,
			),
		},
		{
			name:        "http.client.Do error, invalid URL TLD",
			ecrQueryURL: "https://example.invalid/%s/%s",
			registry: ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerECRRepo,
						Tag:   "{{ version }}",
					},
					Auth: &ECRAuth{
						ECRAuthDefaults: ECRAuthDefaults{
							queryToken: "test",
						},
					},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^` + test.ArgusDockerECRRepo + `:latest
					Get "https://.*
						dial tcp:
							lookup .* no such host$`,
			),
		},
	}

	_ecrTokenAddress := ecrTokenAddress
	_ecrQueryURL := ecrQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

			if tc.ecrTokenAddress != "" {
				ecrTokenAddress = tc.ecrTokenAddress
				t.Cleanup(func() {
					ecrTokenAddress = _ecrTokenAddress
				})
			}
			if tc.ecrQueryURL != "" {
				ecrQueryURL = tc.ecrQueryURL
				t.Cleanup(func() {
					ecrQueryURL = _ecrQueryURL
				})
			}

			// WHEN: Check is called with this version.
			err := tc.registry.Check(tc.version)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nECRRegistry.Check() error mismatch\ngot:  %q\nwant: %s",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}

func TestECRAuth_RefreshQueryToken__integration(t *testing.T) {
	// GIVEN: an ECRAuth to fetch a queryToken with.
	tests := []struct {
		name         string
		handler      http.HandlerFunc // Served via httptest; overrides ecrTokenAddress.
		tokenAddress *string
		errRegex     string
	}{
		{
			name:     "valid",
			errRegex: `^$`,
		},
		{
			name: "non-200 response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			errRegex: `^ecr token request failed`,
		},
		{
			name: "unparseable response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("not json"))
			},
			errRegex: `^failed to parse ecr token response`,
		},
		{
			name:         "invalid token URL",
			tokenAddress: test.Ptr("https://	example.com"),
			errRegex:     `^ecr token refresh fail`,
		},
	}
	_ecrTokenAddress := ecrTokenAddress

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

			switch {
			case tc.handler != nil:
				srv := httptest.NewServer(tc.handler)
				t.Cleanup(srv.Close)
				ecrTokenAddress = srv.URL
				t.Cleanup(func() {
					ecrTokenAddress = _ecrTokenAddress
				})
			case tc.tokenAddress != nil:
				ecrTokenAddress = *tc.tokenAddress
				t.Cleanup(func() {
					ecrTokenAddress = _ecrTokenAddress
				})
			}
			detail := ContainerDetail{Image: test.ArgusDockerECRRepo}

			// AND: an ECRAuth to fetch it with.
			data := &ECRAuth{}

			// WHEN: refreshQueryToken() is called on it.
			queryToken, err := data.refreshQueryToken(detail)

			prefix := fmt.Sprintf(
				"%s\nECRAuth.refreshQueryToken(%+v)",
				packageName, detail,
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
			if err == nil && queryToken == "" {
				t.Errorf("%s returned an empty queryToken\nwant: non-empty", prefix)
			}
		})
	}
}
