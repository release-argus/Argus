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

package option

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestBase_IsZero(t *testing.T) {
	// GIVEN: Base.
	tests := []struct {
		name string
		base *Base
		want bool
	}{
		{
			name: "empty",
			base: &Base{},
			want: true,
		},
		{
			name: "non-empty Interval",
			base: &Base{
				Interval: "10s",
			},
			want: false,
		},
		{
			name: "non-empty SemanticVersioning",
			base: &Base{
				SemanticVersioning: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "filled",
			base: &Base{
				Interval:           "10s",
				SemanticVersioning: test.Ptr(true),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.base.IsZero()

			// THEN: it should return the expected value.
			if got != tc.want {
				t.Errorf(
					"%s\nBase.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
		want     bool
	}{
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     true,
		},
		{
			name: "non-empty Interval",
			defaults: &Defaults{
				Base: Base{
					Interval: "10s",
				},
			},
			want: false,
		},
		{
			name: "non-empty SemanticVersioning",
			defaults: &Defaults{
				Base: Base{
					SemanticVersioning: test.Ptr(true),
				},
			},
			want: false,
		},
		{
			name: "filled",
			defaults: &Defaults{
				Base: Base{
					Interval:           "10s",
					SemanticVersioning: test.Ptr(true),
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: it should return the expected value.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
	}{
		{
			name:     "empty defaults",
			defaults: &Defaults{},
		},
		{
			name: "non-empty defaults",
			defaults: &Defaults{
				Base: Base{
					Interval:           "1m",
					SemanticVersioning: test.Ptr(false),
				},
			},
		},
	}

	// AND: the expected values.
	wants := struct {
		interval string
		semVer   bool
	}{
		interval: "10m",
		semVer:   true,
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Default is called.
			tc.defaults.Default()

			prefix := fmt.Sprintf("%s\nDefaults.Default()", packageName)

			// THEN: the default values are set correctly.
			if tc.defaults.Interval != wants.interval {
				t.Errorf(
					"%s .Interval value mismatch\ngot:  %q\nwant: %q",
					prefix, tc.defaults.Interval, wants.interval,
				)
			}
			if tc.defaults.SemanticVersioning == nil || *tc.defaults.SemanticVersioning != wants.semVer {
				t.Errorf(
					"%s .SemanticVersioning value mismatch\ngot:  %q\nwant: %t",
					prefix, test.StringifyPtr(tc.defaults.SemanticVersioning), wants.semVer,
				)
			}
		})
	}
}

func TestOptions_IsZero(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name    string
		options *Options
		want    bool
	}{
		{
			name:    "empty",
			options: &Options{},
			want:    true,
		},
		{
			name: "non-empty Interval",
			options: &Options{
				Base: Base{
					Interval: "1m",
				},
			},
			want: false,
		},
		{
			name: "non-empty SemanticVersioning",
			options: &Options{
				Base: Base{
					SemanticVersioning: test.Ptr(false),
				},
			},
			want: false,
		},
		{
			name: "non-empty Active",
			options: &Options{
				Active: test.Ptr(false),
			},
			want: false,
		},
		{
			name: "filled",
			options: &Options{
				Active: test.Ptr(false),
				Base: Base{
					Interval:           "1m",
					SemanticVersioning: test.Ptr(false),
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.options.IsZero()

			// THEN: it should return the expected value.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_Copy(t *testing.T) {
	defaults := testDefaults(t)
	hardDefaults := testDefaults(t)
	// GIVEN: Options.
	tests := []struct {
		name    string
		options *Options
	}{
		{
			name:    "nil options",
			options: nil,
		},
		{
			name:    "empty options",
			options: &Options{},
		},
		{
			name: "options with values",
			options: &Options{
				Base: Base{
					Interval:           "10s",
					SemanticVersioning: test.Ptr(true),
				},
				Active:       test.Ptr(true),
				Defaults:     defaults,
				HardDefaults: hardDefaults,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy is called.
			newOptions := tc.options.Copy()

			prefix := fmt.Sprintf("%s\nOptions Copy()", packageName)

			// THEN: the copied Options should match the original.
			if tc.options == nil {
				if newOptions != nil {
					t.Errorf(
						"%s copied nil\ngot:  %v\nwant: nil",
						prefix, newOptions,
					)
				}
				return
			}
			if newOptions == nil {
				t.Fatalf("%s copied non-nil\ngot:  nil\nwant: non-nil", prefix)
			}
			if gotStr, wantStr := newOptions.String(), tc.options.String(); gotStr != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: Pointers are retained as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Defaults", Got: newOptions.Defaults, Want: tc.options.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: newOptions.HardDefaults, Want: tc.options.HardDefaults, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOptions_String(t *testing.T) {
	optCfg := plainDefaultsConfig(t)

	tests := []struct {
		name    string
		options *Options
		want    string
	}{
		{
			name:    "nil",
			options: nil,
			want:    "",
		},
		{
			name:    "empty/default Options",
			options: &Options{},
			want:    "{}\n",
		},
		{
			name: "all options defined",
			options: test.Must(t, func() (*Options, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						active: true
						interval: 10s
						semantic_versioning: true
					`)),
					optCfg,
				)
			}),
			want: test.TrimYAML(`
				interval: 10s
				semantic_versioning: true
				active: true
			`),
		},
		{
			name: "empty with defaults",
			options: &Options{
				Defaults:     optCfg.Soft,
				HardDefaults: optCfg.Hard,
			},
			want: "{}\n",
		},
		{
			name: "all with defaults",
			options: test.Must(t, func() (*Options, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						active: true
						interval: 10s
						semantic_versioning: true
					`)),
					optCfg,
				)
			}),
			want: test.TrimYAML(`
				interval: 10s
				semantic_versioning: true
				active: true
			`),
		},
		{
			name: "empty with defaults and hardDefaults",
			options: test.Must(t, func() (*Options, error) {
				return Decode(
					"yaml", []byte{},
					optCfg,
				)
			}),
			want: "{}\n",
		},
		{
			name: "all with defaults and hardDefaults",
			options: test.Must(t, func() (*Options, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						active: true
						interval: 10s
						semantic_versioning: true
					`)),
					optCfg,
				)
			}),
			want: test.TrimYAML(`
				interval: 10s
				semantic_versioning: true
				active: true
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the Options is stringified with String.
			got := tc.options.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_GetActive(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name   string
		active *bool
		want   bool
	}{
		{
			name:   "nil",
			active: nil,
			want:   true,
		},
		{
			name:   "true",
			active: test.Ptr(true),
			want:   true,
		},
		{
			name:   "false",
			active: test.Ptr(false),
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := testOptions(t)
			options.Active = tc.active

			// WHEN: GetActive is called.
			got := options.GetActive()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.GetActive() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_SetDefaults(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name    string
		options *Options
	}{
		{
			name:    "empty",
			options: &Options{},
		},
		{
			name: "existing defaults/hardDefaults overwritten",
			options: &Options{
				Defaults:     testDefaults(t),
				HardDefaults: testDefaults(t),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: two sets of Defaults.
			var defaults, hardDefaults *Defaults

			// WHEN: SetDefaults is called.
			tc.options.SetDefaults(
				defaults,
				hardDefaults,
			)

			prefix := fmt.Sprintf(
				"%s\nOptions SetDefaults(defaults=%p, hardDefaults=%p)",
				packageName, defaults, hardDefaults,
			)

			// THEN: the defaults are set as expected.
			if tc.options.Defaults != defaults {
				t.Errorf(
					"%s .Defaults pointer mismatch\ngot:  %v\nwant: %v",
					prefix, tc.options.Defaults, defaults,
				)
			}

			// AND: the hardDefaults are set as expected.
			if tc.options.HardDefaults != hardDefaults {
				t.Errorf(
					"%s .HardDefaults pointer mismatch\ngot:  %v\nwant: %v",
					prefix, tc.options.HardDefaults, hardDefaults,
				)
			}
		})
	}
}

func TestOptions_GetInterval(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		{
			name:             "root overrides all",
			want:             "10s",
			rootValue:        "10s",
			defaultValue:     "1m10s",
			hardDefaultValue: "1m10s",
		},
		{
			name:             "default overrides hardDefault",
			want:             "10s",
			rootValue:        "",
			defaultValue:     "10s",
			hardDefaultValue: "1m10s",
		},
		{
			name:             "hardDefault is last resort",
			want:             "10s",
			rootValue:        "",
			defaultValue:     "",
			hardDefaultValue: "10s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := testOptions(t)
			options.Interval = tc.rootValue
			options.Defaults.Interval = tc.defaultValue
			options.HardDefaults.Interval = tc.hardDefaultValue

			// WHEN: GetInterval is called.
			got := options.GetInterval()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.GetInterval() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_GetSemanticVersioning(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		wantBool                                  bool
	}{
		{
			name:             "root overrides all",
			wantBool:         true,
			rootValue:        test.Ptr(true),
			defaultValue:     test.Ptr(false),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "default overrides hardDefault",
			wantBool:         true,
			defaultValue:     test.Ptr(true),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "hardDefault is last resort",
			wantBool:         true,
			hardDefaultValue: test.Ptr(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := testOptions(t)
			options.SemanticVersioning = tc.rootValue
			options.Defaults.SemanticVersioning = tc.defaultValue
			options.HardDefaults.SemanticVersioning = tc.hardDefaultValue

			// WHEN: GetSemanticVersioning is called.
			got := options.GetSemanticVersioning()

			// THEN: the function returns the correct result.
			if got != tc.wantBool {
				t.Errorf(
					"%s\nOptions.GetSemanticVersioning() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.wantBool,
				)
			}
		})
	}
}

func TestOptions_VerifySemanticVersioning(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name     string
		version  string
		errRegex string
	}{
		{
			name:     "valid semantic version - MAJOR.MINOR.PATCH",
			version:  "1.0.0",
			errRegex: `^$`,
		},
		{
			name:     "valid semantic version - MAJOR.MINOR",
			version:  "1.0",
			errRegex: `^$`,
		},
		{
			name:     "valid semantic version - MAJOR",
			version:  "1",
			errRegex: `^$`,
		},
		{
			name:     "invalid semantic version",
			version:  "1_0_0",
			errRegex: `^failed to convert "1_0_0" to a semantic version.*$`,
		},
		{
			name:     "non-numeric version",
			version:  "major.minor.patch",
			errRegex: `^failed to convert "major.minor.patch" to a semantic version.*$`,
		},
	}
	options := testOptions(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: VerifySemanticVersioning is called.
			_, err := options.VerifySemanticVersioning(tc.version, logx.LogFrom{})

			// THEN: the function returns the correct result.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nOptions.VerifySemanticVersioning(%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.version,
					e, tc.errRegex,
				)
			}
		})
	}
}

func TestOptions_GetIntervalPointer(t *testing.T) {
	// GIVEN: options.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue string
		want                                      string
	}{
		{
			name:             "root overrides all",
			rootValue:        "10s",
			defaultValue:     "20s",
			hardDefaultValue: "30s",
			want:             "10s",
		},
		{
			name:             "default overrides hardDefault",
			defaultValue:     "20s",
			hardDefaultValue: "30s",
			want:             "20s",
		},
		{
			name:             "hardDefault is last resort",
			hardDefaultValue: "30s",
			want:             "30s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := testOptions(t)
			options.Interval = tc.rootValue
			options.Defaults.Interval = tc.defaultValue
			options.HardDefaults.Interval = tc.hardDefaultValue

			// WHEN: GetIntervalPointer is called.
			got := options.GetIntervalPointer()

			// THEN: the function returns the correct result.
			if *got != tc.want {
				t.Errorf(
					"%s\nOptions.GetIntervalPointer() value mismatch\ngot:  %q\nwant: %q",
					packageName, *got, tc.want,
				)
			}
		})
	}
}

func TestOptions_GetIntervalDuration(t *testing.T) {
	// GIVEN: Options.
	options := testOptions(t)
	options.Interval = "3h2m1s"

	// WHEN: GetInterval is called.
	got := options.GetIntervalDuration()

	// THEN: the function returns the correct result.
	want := (3 * time.Hour) + (2 * time.Minute) + time.Second
	if got != want {
		t.Errorf(
			"%s\nOptions.GetIntervalDuration() value mismatch\ngot:  %v\nwant: %v",
			packageName, got, want,
		)
	}
}

func TestOptions_CheckValues(t *testing.T) {
	optCfg := plainDefaultsConfig(t)

	// GIVEN: Options.
	tests := []struct {
		name         string
		input        *Options
		wantInterval string
		errRegex     string
	}{
		{
			name:     "valid options",
			errRegex: `^$`,
			input: test.Must(t, func() (*Options, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						active: false
						interval: 10s
						semantic_versioning: false
					`)),
					optCfg,
				)
			}),
		},
		{
			name:     "invalid interval",
			errRegex: `interval: .* <invalid>`,
			input: test.Must(t, func() (*Options, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						active: false
						interval: 10x
						semantic_versioning: false
					`)),
					optCfg,
				)
			}),
		},
		{
			name:         "seconds get appended to pure decimal interval",
			errRegex:     `^$`,
			wantInterval: "10s",
			input: test.Must(t, func() (*Options, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						active: false
						interval: 10
						semantic_versioning: false
					`)),
					optCfg,
				)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}
