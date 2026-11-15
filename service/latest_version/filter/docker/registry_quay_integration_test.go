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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestQuayRegistry_Check(t *testing.T) {
	tests := []struct {
		name     string
		registry QuayRegistry
		version  string
		errRegex string
	}{
		{
			name: "no auth, known image+tag",
			registry: QuayRegistry{
				Image: test.ArgusDockerQuayRepo,
				Tag:   "{{ version }}",
				Auth:  &QuayAuth{},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth/known image+tag",
			registry: QuayRegistry{
				Image: test.ArgusDockerQuayRepo,
				Tag:   "{{ version }}",
				Auth: &QuayAuth{
					Token: test.DockerQuayToken(t),
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth/known image, unknown tag",
			registry: QuayRegistry{
				Image: test.ArgusDockerQuayRepo,
				Tag:   "{{ version }}-unknown",
				Auth: &QuayAuth{
					Token: test.DockerQuayToken(t),
				},
			},
			version: "latest",
			// Quay answers 200 with an empty tag list, which QuayRegistry.parseBody turns into ErrTagNotFound.
			errRegex: `^` + regexp.QuoteMeta(test.ArgusDockerQuayRepo) + `:latest-unknown - tag not found$`,
		},
		{
			name: "auth/unknown image",
			registry: QuayRegistry{
				Image: test.ArgusDockerQuayRepo + "-unknown",
				Tag:   "{{ version }}",
				Auth: &QuayAuth{
					Token: test.DockerQuayToken(t),
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				` + regexp.QuoteMeta(test.ArgusDockerQuayRepo) + `-unknown:latest
					{"detail": "[^"]+".*}$`,
			),
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

func TestQuayRegistry_Check__errors(t *testing.T) {
	tests := []struct {
		name         string
		quayQueryURL string
		registry     QuayRegistry
		version      string
		errRegex     string
	}{
		{
			name:         "newRequest error, invalid URL",
			quayQueryURL: "https://	example.com",
			registry: QuayRegistry{
				Image: test.ArgusDockerQuayRepo + "-unknown",
				Tag:   "{{ version }}",
				Auth: &QuayAuth{
					Token: "test",
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^parse "https://.*
					.*invalid control character in URL$`,
			),
		},
		{
			name:         "http.client.Do error, invalid URL TLD",
			quayQueryURL: "https://example.invalid/%s/%s",
			registry: QuayRegistry{
				Image: test.ArgusDockerQuayRepo,
				Tag:   "{{ version }}",
				Auth: &QuayAuth{
					Token: "test",
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^` + regexp.QuoteMeta(test.ArgusDockerQuayRepo) + `:latest
					Get "https://example\.invalid/.*":
						dial tcp:
							lookup .*$`,
			),
		},
	}
	_quayQueryURL := quayQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps quayQueryURL.

			if tc.quayQueryURL != "" {
				quayQueryURL = tc.quayQueryURL
				t.Cleanup(func() {
					quayQueryURL = _quayQueryURL
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
