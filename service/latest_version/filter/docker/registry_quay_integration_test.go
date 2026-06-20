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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestQuayRegistry_Check(t *testing.T) {
	// GIVEN: a QuayRegistry, and version to check for.
	tests := []struct {
		name     string
		registry QuayRegistry
		version  string
		errRegex string
	}{
		{
			name: "no auth, known image+tag",
			registry: QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerQuayRepo,
						Tag:   "{{ version }}",
					},
					Auth: &QuayAuth{},
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth/known image+tag",
			registry: QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerQuayRepo,
						Tag:   "{{ version }}",
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: test.DockerQuayToken(t),
						},
					},
				},
			},
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "auth/known image, unknown tag",
			registry: QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerQuayRepo,
						Tag:   "{{ version }}-unknown",
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: test.DockerQuayToken(t),
						},
					},
				},
			},
			version:  "latest",
			errRegex: `^` + test.ArgusDockerQuayRepo + `:latest-unknown - tag not found$`,
		},
		{
			name: "auth/unknown image",
			registry: QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerQuayRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: test.DockerQuayToken(t),
						},
					},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				` + test.ArgusDockerQuayRepo + `-unknown:latest
					{"detail": "[^"]+".*}$`,
			),
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
					"%s\nQuayRegistry.Check() error mismatch\ngot:  %q\nwant: %q",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}

func TestQuayRegistry_Check__errors(t *testing.T) {
	// GIVEN: a QuayRegistry, and version to check for.
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerQuayRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: "test",
						},
					},
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
			quayQueryURL: "https://example.invalid",
			registry: QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: test.ArgusDockerGHCRRepo + "-unknown",
						Tag:   "{{ version }}",
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: "test",
						},
					},
				},
			},
			version: "latest",
			errRegex: test.TrimYAML(`
				^parse "https://.*
					.*invalid URL escape .*$`,
			),
		},
	}
	_quayQueryURL := quayQueryURL

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

			if tc.quayQueryURL != "" {
				quayQueryURL = tc.quayQueryURL
				t.Cleanup(func() {
					quayQueryURL = _quayQueryURL
				})
			}

			// WHEN: Check is called with this version.
			err := tc.registry.Check(tc.version)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nQuayRegistry.Check() error mismatch\ngot:  %q\nwant: %q",
					tc.name, e, tc.errRegex,
				)
			}
		})
	}
}
