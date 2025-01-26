// Copyright [2025] [Argus]
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
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
)

func TestInit(t *testing.T) {
	type want struct {
		requireHasDockerDefaults bool
	}

	// GIVEN a YAML string
	tests := map[string]struct {
		overrides string
		want      want
	}{
		"no require": {
			overrides: test.TrimYAML(`
				require: null
			`),
		},
		"require with no Docker": {
			overrides: test.TrimYAML(`
				require:
					regex_version: foo
					docker: null
			`),
		},
		"require with Docker": {
			overrides: test.TrimYAML(`
				require:
					regex_version: foo
					docker:
						type: ghcr
						image: release-argus/argus
						tag: '{{ version }}'
			`),
			want: want{
				requireHasDockerDefaults: true,
			},
		},
		"URLCommands for single version": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '[0-9.]+'
						index: 1
				require:
					docker:
						type: ghcr
						image: release-argus/argus
						tag: '{{ version }}'
			`),
			want: want{
				requireHasDockerDefaults: true,
			},
		},
		"URLCommands for multiple versions": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '[0-9.]+'
			`),
			want: want{},
		},
		"require.docker and urlCommands for single version": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '[0-9.]+'
						index: 1
				require:
					docker:
						type: ghcr
						image: release-argus/argus
						tag: '{{ version }}'
			`),
			want: want{
				requireHasDockerDefaults: true,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := opt.Options{}
			status := &status.Status{}
			defaults := &base.Defaults{}
			hardDefaults := &base.Defaults{}
			hardDefaults.Default()
			lookup := &Lookup{}
			// overrides
			err := yaml.Unmarshal([]byte(tc.overrides), lookup)
			if err != nil {
				t.Fatalf("web.Lookup.Init failed to unmarshal overrides: %v", err)
			}

			// WHEN New is called with it
			lookup.Init(
				&options,
				status,
				defaults, hardDefaults,
			)

			// THEN the defaults are set as expected
			if lookup.Defaults != defaults {
				t.Errorf("web.Lookup.Defaults not set\nwant: %v\ngot:  %v",
					lookup.Defaults, defaults)
			}
			// AND the hard defaults are set as expected
			if lookup.HardDefaults != hardDefaults {
				t.Errorf("web.Lookup.HardDefaults not set\nwant: %v\ngot:  %v",
					lookup.HardDefaults, hardDefaults)
			}
			// AND the status is set as expected
			if lookup.Status != status {
				t.Errorf("web.Lookup.Status not set\nwant: %v\ngot:  %v",
					lookup.Status, status)
			}
			// AND the options are set as expected
			if lookup.Options != &options {
				t.Errorf("web.Lookup.Options not set\nwant: %v\ngot:  %v",
					lookup.Options, &options)
			}
			// AND the require is given the correct defaults
			if lookup.Require != nil && lookup.Require.Docker != nil {
				if lookup.Require.Docker.Defaults != &defaults.Require.Docker {
					t.Errorf("web.Lookup.Require.Docker.Defaults not set\nwant: %v\ngot:  %v",
						lookup.Require.Docker.Defaults, defaults.Require.Docker)
				}
			} else if tc.want.requireHasDockerDefaults {
				t.Errorf("web.Lookup.Require.Docker not set\nrequire: %v",
					lookup.Require)
			}
		})
	}
}
