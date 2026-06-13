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

package filter

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestRequire_DockerTagCheck(t *testing.T) {
	defaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	defaults.SetDefaults(hardDefaults)

	// GIVEN: a Require with a docker registry.
	tests := []struct {
		name     string
		yaml     string
		version  string
		errRegex string
	}{
		{
			name: "GHCR/tag found",
			yaml: test.TrimYAML(`
				docker:
					type: ghcr
					image: ` + test.ArgusDockerGHCRRepo + `
					tag: "{{ version }}"
					auth:
						token: ` + test.GitHubToken(t) + `
			`),
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "GHCR/tag not found",
			yaml: test.TrimYAML(`
				docker:
					type: ghcr
					image: ` + test.ArgusDockerGHCRRepo + `
					tag: "{{ version }}-unknown"
					auth:
						token: ` + test.GitHubToken(t) + `
			`),
			version:  "latest",
			errRegex: `tag not found`,
		},
		{
			name: "Hub/tag found",
			yaml: test.TrimYAML(`
				docker:
					type: hub
					image: ` + test.ArgusDockerHubRepo + `
					tag: "{{ version }}"
					auth:
						username: ` + test.DockerHubUsername(t) + `
						token: ` + test.DockerHubToken(t) + `
			`),
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "Hub/tag not found",
			yaml: test.TrimYAML(`
				docker:
					type: hub
					image: ` + test.ArgusDockerHubRepo + `
					tag: "{{ version }}-unknown"
					auth:
						username: ` + test.DockerHubUsername(t) + `
						token: ` + test.DockerHubToken(t) + `
					`),
			version:  "latest",
			errRegex: `tag not found`,
		},
		{
			name: "Quay/tag found",
			yaml: test.TrimYAML(`
				docker:
					type: quay
					image: ` + test.ArgusDockerQuayRepo + `
					tag: "{{ version }}"
					auth:
						token: ` + test.DockerQuayToken(t) + `
			`),
			version:  "latest",
			errRegex: `^$`,
		},
		{
			name: "Quay/tag not found",
			yaml: test.TrimYAML(`
				docker:
					type: quay
					image: ` + test.ArgusDockerQuayRepo + `
					tag: "{{ version }}-unknown"
					auth:
						token: ` + test.DockerQuayToken(t) + `
			`),
			version:  "latest",
			errRegex: `tag not found`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require := test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode("yaml", []byte(tc.yaml), svcStatus, defaults)
			})

			// WHEN: DockerTagCheck is called.
			err := require.DockerTagCheck(tc.version)

			prefix := fmt.Sprintf(
				"%s\nRequire.DockerTagCheck(%q)",
				packageName, tc.version,
			)

			// THEN: the error matches expectation.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}
