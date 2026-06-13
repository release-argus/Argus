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

//go:build unit

package web

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	dockertest "github.com/release-argus/Argus/service/latest_version/filter/docker/test"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestLookup_Init(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	type want struct {
		requireHasDockerDefaults bool
	}

	// GIVEN: a YAML string.
	tests := []struct {
		name      string
		overrides string
		want      want
	}{
		{
			name:      "no require",
			overrides: `require: null`,
		},
		{
			name: "require with no Docker",
			overrides: test.TrimYAML(`
				require:
					regex_version: foo
					docker: null
			`),
		},
		{
			name: "require with Docker",
			overrides: test.TrimYAML(`
				require:
					regex_version: foo
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}'
			`),
			want: want{
				requireHasDockerDefaults: true,
			},
		},
		{
			name: "URLCommands for single version",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '[0-9.]+'
						index: 1
				require:
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}'
			`),
			want: want{
				requireHasDockerDefaults: true,
			},
		},
		{
			name: "URLCommands for multiple versions",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '[0-9.]+'
			`),
			want: want{},
		},
		{
			name: "require.docker and urlCommands for single version",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '[0-9.]+'
						index: 1
				require:
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}'
			`),
			want: want{
				requireHasDockerDefaults: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &opt.Options{}
			svcStatus := &status.Status{}
			l := &Lookup{}
			// overrides.
			err := l.UnmarshalYAML([]byte(tc.overrides))
			if err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal overrides: %v",
					packageName, err,
				)
			}
			requireData, _ := polymorphic.Extract("yaml", []byte(tc.overrides), "require")
			req, err := filter.Decode(
				"yaml", requireData,
				l.Status,
				&lvCfg.Soft.Require,
			)
			if err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal require overrides: %v",
					packageName, err,
				)
			}
			l.SetRequire(req)

			// WHEN: Init is called with it.
			l.Init(options, svcStatus, lvCfg)

			prefix := fmt.Sprintf(
				"%s\nLookup.Init(options=%p, status=%p, defaults=%v)",
				packageName, options, &svcStatus, lvCfg,
			)

			// THEN: pointers to those vars are handed out to the Lookup.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: l.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: l.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: l.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: l.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: the Require is given the correct defaults.
			if l.Require != nil && l.Require.Docker != nil {
				dockerType := l.Require.Docker.GetType()
				wantDefaults, _ := dockertest.GetDefaultOfDockerType(t, dockerType, &lvCfg.Soft.Require.Docker)
				if got := l.Require.Docker.Defaults(); got != wantDefaults {
					t.Errorf(
						"%s .Require.Docker.Defaults was not handed to the Lookup correctly\ngot:  %v\nwant: %v",
						prefix, got, wantDefaults,
					)
				}
			} else if tc.want.requireHasDockerDefaults {
				t.Errorf(
					"%s .Require.Docker was not handed to the Lookup\ngot: Require=%v",
					prefix, l.Require,
				)
			}
		})
	}
}
