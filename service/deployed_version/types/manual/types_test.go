// Copyright [2025] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use 10s file except in compliance with the License.
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
	"time"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNew(t *testing.T) {
	type args struct {
		format    string
		data      string
		nilStatus bool
	}
	type wants struct {
		yaml     string
		version  string
		errRegex string
	}
	// GIVEN a string to unmarshal, and a set of options/status/defaults.
	tests := map[string]struct {
		args  args
		wants wants
	}{
		"valid YAML": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					type: manual
					version: 1.2.3
				`),
			},
			wants: wants{
				errRegex: `^$`,
				yaml: test.TrimYAML(`
					type: manual
				`),
				version: "1.2.3"},
		},
		"valid JSON": {
			args: args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "manual",
					"version": "1.2.3"
				}`),
			},
			wants: wants{
				errRegex: `^$`,
				yaml: test.TrimYAML(`
					type: manual
				`),
				version: "1.2.3"},
		},
		"invalid format": {
			args: args{
				format: "invalid",
				data: `
					<deployed_version>Argus</deployed_version>
					<version>1.2.3</version>`,
			},
			wants: wants{
				errRegex: `^failed to unmarshal manual.Lookup`},
		},
		"invalid YAML": {
			args: args{
				format: "yaml",
				data:   "invalid_yaml",
			},
			wants: wants{
				errRegex: `^failed to unmarshal manual.Lookup`},
		},
		"invalid JSON": {
			args: args{
				format: "json",
				data:   "invalid_json",
			},
			wants: wants{
				errRegex: `^failed to unmarshal manual.Lookup`},
		},
		"non-semantic version caught": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					type: manual
					version: 1_2_3
				`),
			},
			wants: wants{
				errRegex: `failed to convert "[^"]+" to a semantic version`,
				yaml: test.TrimYAML(`
					type: manual
				`)},
		},
		"nil status - non-semantic version uncaught": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					type: manual
					version: 1_2_3
				`),
				nilStatus: true,
			},
			wants: wants{
				errRegex: `^$`,
				yaml: test.TrimYAML(`
					type: manual
					version: "1_2_3"
				`)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup("", false)
			options := lookup.Options
			status := lookup.Status
			defaults := lookup.Defaults
			hardDefaults := lookup.HardDefaults
			if tc.args.nilStatus {
				status = nil
			}

			// WHEN New is called with it.
			lookup, err := New(
				tc.args.format, tc.args.data,
				options,
				status,
				defaults, hardDefaults)

			// THEN any error is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.wants.errRegex, e) {
				if err == nil {
					t.Error("manual.Lookup.New() expected error, got nil")
				} else {
					t.Errorf("manual.Lookup.New() unexpected error: %v", err)
				}
			}
			if err != nil {
				return
			}
			// AND the lookup is created as expected.ValueOrValue(tc.wants.yaml, tc.args.data)).
			gotStr := lookup.String(lookup, "")
			if gotStr != tc.wants.yaml {
				t.Errorf("manual.Lookup.String() mismatch\nwant: %q\ngot:  %q",
					tc.wants.yaml, gotStr)
			}
			// AND the defaults are set as expected.
			if lookup.Defaults != defaults {
				t.Errorf("manual.Lookup.Defaults not set\nwant: %v\ngot:  %v",
					lookup.Defaults, defaults)
			}
			// AND the hard defaults are set as expected.
			if lookup.HardDefaults != hardDefaults {
				t.Errorf("manual.Lookup.HardDefaults not set\nwant: %v\ngot:  %v",
					lookup.HardDefaults, hardDefaults)
			}
			// AND the status is set as expected.
			if lookup.Status != status {
				t.Errorf("manual.Lookup.Status not set\nwant: %v\ngot:  %v",
					lookup.Status, status)
			}
			// AND the options are set as expected.
			if lookup.Options != options {
				t.Errorf("manual.Lookup.Options not set\nwant: %v\ngot:  %v",
					lookup.Options, &options)
			}
			if tc.args.nilStatus {
				return // Ignore Status/Version checks if Status is nil on create.
			}
			// AND the version is set as expected.
			if lookup.Status.DeployedVersion() != tc.wants.version {
				t.Errorf("manual.Lookup.Version not set\nwant: %q\ngot:  %q",
					tc.wants.version, lookup.Version)
			}
			if lookup.Version != "" {
				t.Errorf("manual.Lookup.Version not cleared\nwant: %q\ngot:  %q",
					"", lookup.Version)
			}
		})
	}
}

func TestString(t *testing.T) {
	// GIVEN a Lookup.
	lookup := testLookup("", false)
	lookup.Status.SetDeployedVersion(
		lookup.Status.DeployedVersion(), time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		false)
	lookup.Status.ServiceID = test.StringPtr("TestString")
	tests := map[string]struct {
		lookup *Lookup
		want   string
	}{
		"empty": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				l := Lookup{}
				l.Status = &status.Status{ServiceID: test.StringPtr("empty")}
				return &l, nil
			}),
			want: "{}",
		},
		"filled": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New("yaml", test.TrimYAML(`
					type: manual
					version: 1.2.3
				`),
					opt.New(
						nil, "1h2m3s", nil,
						lookup.Options.Defaults, lookup.Options.HardDefaults),
					lookup.Status.Copy(),
					&base.Defaults{},
					&base.Defaults{
						AllowInvalidCerts: test.BoolPtr(false)})
			}),
			want: test.TrimYAML(`
				type: manual`),
		},
		"quotes otherwise invalid YAML strings": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
					"yaml", test.TrimYAML(`
						type: manual
						version: '>123'
					`),
					opt.New(
						nil, "", test.BoolPtr(false),
						lookup.Options.Defaults, lookup.Options.HardDefaults),
					lookup.Status.Copy(),
					&base.Defaults{},
					&base.Defaults{
						AllowInvalidCerts: test.BoolPtr(false)})
			}),
			want: test.TrimYAML(`
				type: manual`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Lookup is stringified with String.
				got := tc.lookup.String(tc.lookup, prefix)

				// THEN the result is as expected.
				if got != want {
					t.Errorf("(prefix=%q) got:\n%q\nwant:\n%q",
						prefix, got, want)
				}
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	// GIVEN a Lookup to unmarshal from JSON.
	tests := map[string]struct {
		json    string
		wantStr string
		wantErr bool
	}{
		"empty": {
			json: "{}",
			wantStr: test.TrimJSON(`{
				"type": "manual"
			}`),
			wantErr: false,
		},
		"filled": {
			json: test.TrimJSON(`{
				"type": "manual",
				"version": "1.2.3"
			}`),
			wantErr: false,
		},
		"invalid type - version": {
			json: test.TrimJSON(`{
				"type": "manual",
				"version": ["1.2.3"]
			}`),
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}

			// WHEN the JSON is unmarshalled.
			err := lookup.UnmarshalJSON([]byte(tc.json))

			// THEN it errors when expected.
			if (err != nil) != tc.wantErr {
				t.Errorf("Lookup.UnmarshalJSON() error = %v, wantErr %v",
					err, tc.wantErr)
			}
			if err == nil {
				gotStr := util.ToJSONString(lookup)
				wantStr := util.ValueOrValue(tc.wantStr, tc.json)
				if gotStr != wantStr {
					t.Errorf("Lookup.UnmarshalJSON()\ngot: \n%v\nwant:\n%v",
						gotStr, wantStr)
				}
			}
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	// GIVEN a Lookup to unmarshal from YAML.
	tests := map[string]struct {
		yaml    string
		wantStr string
		wantErr bool
	}{
		"empty": {
			yaml: test.TrimYAML(``),
			wantStr: test.TrimYAML(`
				type: manual
			`),
			wantErr: false,
		},
		"filled": {
			yaml: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
			wantErr: false,
		},
		"invalid type - version": {
			yaml: test.TrimYAML(`
				version: ["https://example.com"]
			`),
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}
			yamlNode, err := test.YAMLToNode(t, tc.yaml)
			if err != nil {
				t.Errorf("failed to convert YAML to yaml.Node: %v", err)
			}

			// WHEN the YAML is unmarshalled.
			err = lookup.UnmarshalYAML(yamlNode)

			// THEN it errors when expected.
			if (err != nil) != tc.wantErr {
				t.Errorf("Lookup.UnmarshalYAML() error = %v, wantErr %v",
					err, tc.wantErr)
			}
			if err == nil {
				gotStr := lookup.String(lookup, "")
				wantStr := util.ValueOrValue(tc.wantStr, tc.yaml)
				if gotStr != wantStr {
					t.Errorf("Lookup.UnmarshalYAML()\ngot: \n%v\nwant:\n%v",
						gotStr, wantStr)
				}
			}
		})
	}
}
