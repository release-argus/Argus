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

package deployedver

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestApplyOverridesJSON(t *testing.T) {
	type args struct {
		lookup             Lookup
		overrides          []byte
		semanticVerDiff    bool
		semanticVersioning *string
	}
	tests := []struct {
		name     string
		args     args
		errRegex string
	}{
		{
			name: "no overrides, no semantic versioning change",
			args: args{
				lookup:             testLookup(t, "url", false, ""),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		{
			name: "invalid semantic versioning",
			args: args{
				lookup:             testLookup(t, "url", false, ""),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.Ptr("invalid"),
			},
			errRegex: test.TrimYAML(`
				^semantic_versioning:
					jsontext:
						invalid character .*$`,
			),
		},
		{
			name: "valid semantic versioning change",
			args: args{
				lookup:             testLookup(t, "url", false, ""),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.Ptr("true"),
			},
			errRegex: `^$`,
		},
		{
			name: "overrides/valid",
			args: args{
				lookup:             testLookup(t, "url", false, ""),
				overrides:          []byte(`{"url": "` + test.LookupJSON["url_valid"] + `"}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		{
			name: "overrides/invalid JSON",
			args: args{
				lookup:             testLookup(t, "url", false, ""),
				overrides:          []byte(`{"url": "}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ could not find end character.*`,
			),
		},
		{
			name: "overrides/invalid var data type",
			args: args{
				lookup:             testLookup(t, "url", false, ""),
				overrides:          []byte(`{"url": [""]}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning,
			)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\napplyOverridesJSON(%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.args.overrides,
					e, tc.errRegex,
				)
			}
		})
	}
}
