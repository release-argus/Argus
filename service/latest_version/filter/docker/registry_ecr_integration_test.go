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
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestECRRegistry_Check(t *testing.T) {
	tests := []struct {
		name     string
		registry ECRRegistry
		version  string
		errRegex string
	}{
		{
			name: "anonymous, known image+tag",
			registry: ECRRegistry{
				Image: test.ArgusDockerECRRepo,
				Tag:   "{{ version }}",
				Auth:  &ECRAuth{},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "known image, unknown tag",
			registry: ECRRegistry{
				Image: test.ArgusDockerECRRepo,
				Tag:   "{{ version }}-unknown",
				Auth:  &ECRAuth{},
			},
			version:  "latest",
			errRegex: `^` + regexp.QuoteMeta(test.ArgusDockerECRRepo) + `:latest-unknown - tag not found$`,
		},
		{
			name: "unknown image",
			registry: ECRRegistry{
				Image: test.ArgusDockerECRRepo + "-unknown",
				Tag:   "{{ version }}",
				Auth:  &ECRAuth{},
			},
			version: "latest",
			// ECR also answers 404 for an unknown image, which ECRRegistry.parseBody turns into ErrTagNotFound.
			errRegex: `^` + regexp.QuoteMeta(test.ArgusDockerECRRepo) + `-unknown:latest - tag not found$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.registry.Check(tc.version)
			e := errfmt.FormatError(err)
			if util.RegexCheck(`(?i)TOOMANYREQUESTS|rate exceeded`, e) {
				t.Skipf("registry rate-limited:\n%s", e)
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestECRRegistry_Check__errors(t *testing.T) {
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
				Image: test.ArgusDockerECRRepo,
				Tag:   "{{ version }}",
				Auth:  &ECRAuth{},
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
				Image: test.ArgusDockerECRRepo,
				Tag:   "{{ version }}",
				Auth: &ECRAuth{
					queryToken: "test",
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
				Image: test.ArgusDockerECRRepo,
				Tag:   "{{ version }}",
				Auth: &ECRAuth{
					queryToken: "test",
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^` + regexp.QuoteMeta(test.ArgusDockerECRRepo) + `:latest
					Get "https://.*
						dial tcp:
							lookup .*$`,
			),
		},
	}

	_ecrTokenAddress := ecrTokenAddress
	_ecrQueryURL := ecrQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps ecrTokenAddress and ecrQueryURL.

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

			err := tc.registry.Check(tc.version)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestECRAuth_RefreshQueryToken__integration(t *testing.T) {
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
			tokenAddress: new("https://	example.com"),
			errRegex:     `^ecr token refresh fail`,
		},
	}
	_ecrTokenAddress := ecrTokenAddress

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps ecrTokenAddress.

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

			data := &ECRAuth{}

			queryToken, err := data.refreshQueryToken(detail)
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
