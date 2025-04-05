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

package option

import (
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestOptions_GetActive(t *testing.T) {
	// GIVEN Options.
	tests := map[string]struct {
		active *bool
		want   bool
	}{
		"nil": {
			active: nil,
			want:   true},
		"true": {
			active: test.BoolPtr(true),
			want:   true},
		"false": {
			active: test.BoolPtr(false),
			want:   false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := testOptions()
			options.Active = tc.active

			// WHEN GetActive is called.
			got := options.GetActive()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestOptions_GetInterval(t *testing.T) {
	// GIVEN Options.
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		"root overrides all": {
			want:             "10s",
			rootValue:        "10s",
			defaultValue:     "1m10s",
			hardDefaultValue: "1m10s",
		},
		"default overrides hardDefault": {
			want:             "10s",
			rootValue:        "",
			defaultValue:     "10s",
			hardDefaultValue: "1m10s",
		},
		"hardDefault is last resort": {
			want:             "10s",
			rootValue:        "",
			defaultValue:     "",
			hardDefaultValue: "10s",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := testOptions()
			options.Interval = tc.rootValue
			options.Defaults.Interval = tc.defaultValue
			options.HardDefaults.Interval = tc.hardDefaultValue

			// WHEN GetInterval is called.
			got := options.GetInterval()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestOptions_GetSemanticVersioning(t *testing.T) {
	// GIVEN Options.
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue *bool
		wantBool                                  bool
	}{
		"root overrides all": {
			wantBool:         true,
			rootValue:        test.BoolPtr(true),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false),
		},
		"default overrides hardDefault": {
			wantBool:         true,
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(false),
		},
		"hardDefault is last resort": {
			wantBool:         true,
			hardDefaultValue: test.BoolPtr(true),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := testOptions()
			options.SemanticVersioning = tc.rootValue
			options.Defaults.SemanticVersioning = tc.defaultValue
			options.HardDefaults.SemanticVersioning = tc.hardDefaultValue

			// WHEN GetSemanticVersioning is called.
			got := options.GetSemanticVersioning()

			// THEN the function returns the correct result.
			if got != tc.wantBool {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.wantBool, got)
			}
		})
	}
}

func TestOptions_GetIntervalPointer(t *testing.T) {
	// GIVEN options.
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		"root overrides all": {
			rootValue:        "10s",
			defaultValue:     "20s",
			hardDefaultValue: "30s",
			want:             "10s"},
		"default overrides hardDefault": {
			defaultValue:     "20s",
			hardDefaultValue: "30s",
			want:             "20s"},
		"hardDefault is last resort": {
			hardDefaultValue: "30s",
			want:             "30s"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := testOptions()
			options.Interval = tc.rootValue
			options.Defaults.Interval = tc.defaultValue
			options.HardDefaults.Interval = tc.hardDefaultValue

			// WHEN GetIntervalPointer is called.
			got := options.GetIntervalPointer()

			// THEN the function returns the correct result.
			if *got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, *got)
			}
		})
	}
}

func TestOptions_GetIntervalDuration(t *testing.T) {
	// GIVEN Options.
	options := testOptions()
	options.Interval = "3h2m1s"

	// WHEN GetInterval is called.
	got := options.GetIntervalDuration()

	// THEN the function returns the correct result.
	want := (3 * time.Hour) + (2 * time.Minute) + time.Second
	if got != want {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, want, got)
	}
}

func TestOptions_CheckValues(t *testing.T) {
	// GIVEN Options.
	tests := map[string]struct {
		options      *Options
		wantInterval string
		errRegex     string
	}{
		"valid options": {
			errRegex: `^$`,
			options: New(
				test.BoolPtr(false), "10s", test.BoolPtr(false),
				nil, nil),
		},
		"invalid interval": {
			errRegex: `interval: .* <invalid>`,
			options: New(
				test.BoolPtr(false), "10x", test.BoolPtr(false),
				nil, nil),
		},
		"seconds get appended to pure decimal interval": {
			errRegex:     `^$`,
			wantInterval: "10s",
			options: New(
				test.BoolPtr(false), "10", test.BoolPtr(false),
				nil, nil),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called.
			err := tc.options.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestOptions_String(t *testing.T) {
	tests := map[string]struct {
		options *Options
		want    string
	}{
		"nil": {
			options: nil,
			want:    "",
		},
		"empty/default Options": {
			options: &Options{},
			want:    "{}\n",
		},
		"all options defined": {
			options: New(
				test.BoolPtr(true), "10s", test.BoolPtr(true),
				nil, nil),
			want: test.TrimYAML(`
				interval: 10s
				semantic_versioning: true
				active: true
			`),
		},
		"empty with defaults": {
			options: &Options{
				Defaults: NewDefaults(
					"10s", test.BoolPtr(true))},
			want: "{}\n",
		},
		"all with defaults": {
			options: New(
				test.BoolPtr(true), "10s", test.BoolPtr(true),
				nil,
				NewDefaults(
					"1h", test.BoolPtr(false))),
			want: test.TrimYAML(`
				interval: 10s
				semantic_versioning: true
				active: true
			`),
		},
		"empty with hardDefaults": {
			options: &Options{
				HardDefaults: NewDefaults(
					"10s", test.BoolPtr(true))},
			want: "{}\n",
		},
		"all with hardDefaults": {
			options: New(
				test.BoolPtr(true), "10s", test.BoolPtr(true),
				nil,
				NewDefaults(
					"1h", test.BoolPtr(false))),
			want: test.TrimYAML(`
				interval: 10s
				semantic_versioning: true
				active: true
			`),
		},
		"empty with defaults and hardDefaults": {
			options: New(
				nil, "", nil,
				NewDefaults(
					"10s", test.BoolPtr(true)),
				NewDefaults("1h", test.BoolPtr(false))),
			want: "{}\n",
		},
		"all with defaults and hardDefaults": {
			options: New(
				test.BoolPtr(true), "10s", test.BoolPtr(true),
				NewDefaults(
					"20s", test.BoolPtr(true)),
				NewDefaults("30s", test.BoolPtr(false))),
			want: test.TrimYAML(`
				interval: 10s
				semantic_versioning: true
				active: true
			`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Options is stringified with String.
			got := tc.options.String()

			// THEN the result is as expected.
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestOptions_Copy(t *testing.T) {
	// GIVEN Options.
	tests := map[string]struct {
		options *Options
	}{
		"nil options": {
			options: nil,
		},
		"empty options": {
			options: &Options{},
		},
		"options with values": {
			options: &Options{
				Base: Base{
					Interval:           "10s",
					SemanticVersioning: test.BoolPtr(true),
				},
				Active:       test.BoolPtr(true),
				Defaults:     NewDefaults("20s", test.BoolPtr(false)),
				HardDefaults: NewDefaults("30s", test.BoolPtr(true)),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Copy is called.
			got := tc.options.Copy()

			// THEN the copied Options should match the original.
			if tc.options == nil {
				if got != nil {
					t.Errorf("%s\ncopied nil\nwant: nil\ngot:  %v",
						packageName, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("%s\nwant: non-nil\ngot:  nil",
					packageName)
			}
			if got.String() != tc.options.String() {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, tc.options.String(), got.String())
			}
			// AND the Defaults should reference the same pointer.
			if got.Defaults != tc.options.Defaults {
				t.Errorf("%s\nDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, tc.options.Defaults, got.Defaults)
			}
			// AND the HardDefaults should reference the same pointer.
			if got.HardDefaults != tc.options.HardDefaults {
				t.Errorf("%s\nHardDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, tc.options.HardDefaults, got.HardDefaults)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	// GIVEN Defaults.
	tests := map[string]struct {
		optionsDefaults *Defaults
	}{
		"empty defaults": {
			optionsDefaults: &Defaults{},
		},
		"non-empty defaults": {
			optionsDefaults: &Defaults{
				Base: Base{
					Interval:           "1m",
					SemanticVersioning: test.BoolPtr(false),
				}},
		},
	}
	// AND the expected values.
	wants := struct {
		interval string
		semVer   bool
	}{
		interval: "10m",
		semVer:   true,
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Default is called.
			tc.optionsDefaults.Default()

			// THEN the default values are set correctly.
			if tc.optionsDefaults.Interval != wants.interval {
				t.Errorf("%s\nInterval mismatch\nwant: %q\ngot:  %q",
					packageName, wants.interval, tc.optionsDefaults.Interval)
			}
			if tc.optionsDefaults.SemanticVersioning == nil || *tc.optionsDefaults.SemanticVersioning != wants.semVer {
				t.Errorf("%s\nSemanticVersioning mismatch\nwant: %t\ngot:  %s",
					packageName, wants.semVer, test.StringifyPtr(tc.optionsDefaults.SemanticVersioning))
			}
		})
	}
}

func TestOptions_VerifySemanticVersioning(t *testing.T) {
	// GIVEN Options.
	tests := map[string]struct {
		version  string
		errRegex string
	}{
		"valid semantic version - MAJOR.MINOR.PATCH": {
			version:  "1.0.0",
			errRegex: `^$`,
		},
		"valid semantic version - MAJOR.MINOR": {
			version:  "1.0",
			errRegex: `^$`,
		},
		"valid semantic version - MAJOR": {
			version:  "1",
			errRegex: `^$`,
		},
		"invalid semantic version": {
			version:  "1_0_0",
			errRegex: `^failed to convert "1_0_0" to a semantic version.*$`,
		},
		"non-numeric version": {
			version:  "major.minor.patch",
			errRegex: `^failed to convert "major.minor.patch" to a semantic version.*$`,
		},
	}
	options := testOptions()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN VerifySemanticVersioning is called.
			_, err := options.VerifySemanticVersioning(tc.version, logutil.LogFrom{})

			// THEN the function returns the correct result.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}
