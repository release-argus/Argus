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

package manual

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		lookupYAML         string
		wantVersion        string
		semanticVersioning bool
		errRegex           string
	}{
		"want semantic, valid": {
			lookupYAML: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
			wantVersion:        "1.2.3",
			semanticVersioning: true,
			errRegex:           `^$`,
		},
		"want semantic, fail": {
			lookupYAML: test.TrimYAML(`
				type: manual
				version: 1_2_3
			`),
			wantVersion:        "",
			semanticVersioning: true,
			errRegex:           `^failed to convert "[^"]+" to a semantic version`,
		},
		"want non-semantic, valid": {
			lookupYAML: test.TrimYAML(`
				type: manual
				version: 1_2_3
			`),
			wantVersion:        "1_2_3",
			semanticVersioning: false,
			errRegex:           `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}
			yamlNode, err := test.YAMLToNode(t, tc.lookupYAML)
			if err != nil {
				t.Fatalf("%s\nfailed to convert YAML to yaml.Node: %v",
					packageName, err)
			}
			err = lookup.UnmarshalYAML(yamlNode)
			if err != nil {
				t.Fatalf("%s\nfailed to unmarshal Lookup: %v",
					packageName, err)
			}
			defaultLookup := testLookup("", false)
			lookup.Options = defaultLookup.Options
			lookup.Options.SemanticVersioning = &tc.semanticVersioning
			lookup.Status = defaultLookup.Status
			lookup.Defaults = defaultLookup.Defaults
			lookup.HardDefaults = defaultLookup.HardDefaults

			// WHEN CheckValues is called.
			err = lookup.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
			// AND the version is set as expected.
			gotVersion := lookup.Status.DeployedVersion()
			if gotVersion != tc.wantVersion {
				t.Errorf("%s\nversion mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantVersion, gotVersion)
			}
		})
	}
}
