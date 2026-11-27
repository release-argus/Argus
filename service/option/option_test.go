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
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestBase_IsZero(t *testing.T) {
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
			name: "non-empty/Interval",
			base: &Base{
				Interval: "10s",
			},
			want: false,
		},
		{
			name: "non-empty/SemanticVersioning",
			base: &Base{
				SemanticVersioning: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			base: &Base{
				Interval:           "10s",
				SemanticVersioning: new(true),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.base.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
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
			name: "non-empty/Interval",
			defaults: &Defaults{
				Interval: "10s",
			},
			want: false,
		},
		{
			name: "non-empty/SemanticVersioning",
			defaults: &Defaults{
				SemanticVersioning: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: &Defaults{
				Interval:           "10s",
				SemanticVersioning: new(true),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.defaults.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
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
				Interval:           "1m",
				SemanticVersioning: new(false),
			},
		},
	}

	// Default replaces any values already set.
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

			tc.defaults.Default()

			if tc.defaults.Interval != wants.interval {
				t.Errorf(
					"Defaults.Interval mismatch\ngot:  %q\nwant: %q",
					tc.defaults.Interval, wants.interval,
				)
			}
			if tc.defaults.SemanticVersioning == nil || *tc.defaults.SemanticVersioning != wants.semVer {
				t.Errorf(
					"Defaults.SemanticVersioning mismatch\ngot:  %s\nwant: %t",
					test.StringifyPtr(tc.defaults.SemanticVersioning), wants.semVer,
				)
			}
		})
	}
}

func TestOptions_IsZero(t *testing.T) {
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
			name: "non-empty/Interval",
			options: &Options{
				Interval: "1m",
			},
			want: false,
		},
		{
			name: "non-empty/SemanticVersioning",
			options: &Options{
				SemanticVersioning: new(false),
			},
			want: false,
		},
		{
			name: "non-empty/Active",
			options: &Options{
				Active: new(false),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			options: &Options{
				Active:             new(false),
				Interval:           "1m",
				SemanticVersioning: new(false),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.options.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestOptions_Copy(t *testing.T) {
	defaults := testDefaults(t)
	hardDefaults := testDefaults(t)

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
				Interval:           "10s",
				SemanticVersioning: new(true),
				Active:             new(true),
				Defaults:           defaults,
				HardDefaults:       hardDefaults,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newOptions := tc.options.Copy()

			if tc.options == nil {
				if newOptions != nil {
					t.Error("Options.Copy() mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}
			if newOptions == nil {
				t.Fatal("Options.Copy() mismatch\ngot:  nil\nwant: non-nil")
			}
			if newOptions == tc.options {
				t.Errorf(
					"copy aliases the original\ngot:  %p\nwant: NOT %p",
					newOptions, tc.options,
				)
			}
			if gotStr, wantStr := newOptions.String(), tc.options.String(); gotStr != wantStr {
				t.Errorf("Options.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Defaults", Got: newOptions.Defaults, Want: tc.options.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: newOptions.HardDefaults, Want: tc.options.HardDefaults, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Options"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestOptions_String(t *testing.T) {
	optCfg := plainDefaultsConfig()

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
			name:    "empty",
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
				Defaults: optCfg.Soft,
			},
			want: "{}\n",
		},
		{
			name: "all with defaults",
			options: test.Must(t, func() (*Options, error) {
				options, err := Decode(
					"yaml", []byte(test.TrimYAML(`
						active: true
						interval: 10s
						semantic_versioning: true
					`)),
					optCfg,
				)
				if err != nil {
					return nil, err
				}
				options.HardDefaults = nil
				return options, nil
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

			got := tc.options.String()

			if got != tc.want {
				t.Errorf("Options.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestOptions_GetActive(t *testing.T) {
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
			active: new(true),
			want:   true,
		},
		{
			name:   "false",
			active: new(false),
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := testOptions(t)
			options.Active = tc.active

			got := options.GetActive()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestOptions_SetDefaults(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
	}{
		{
			name:    "empty",
			options: &Options{},
		},
		{
			name: "existing defaults, hardDefaults overwritten",
			options: &Options{
				Defaults:     testDefaults(t),
				HardDefaults: testDefaults(t),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaults, hardDefaults := &Defaults{}, &Defaults{}

			tc.options.SetDefaults(defaults, hardDefaults)

			if tc.options.Defaults != defaults {
				t.Errorf(
					"Options.Defaults mismatch\ngot:  %p\nwant: %p",
					tc.options.Defaults, defaults,
				)
			}

			if tc.options.HardDefaults != hardDefaults {
				t.Errorf(
					"Options.HardDefaults mismatch\ngot:  %p\nwant: %p",
					tc.options.HardDefaults, hardDefaults,
				)
			}
		})
	}
}

func TestOptions_GetInterval(t *testing.T) {
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

			got := options.GetInterval()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestOptions_GetSemanticVersioning(t *testing.T) {
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		wantBool                                  bool
	}{
		{
			name:             "root overrides all",
			wantBool:         true,
			rootValue:        new(true),
			defaultValue:     new(false),
			hardDefaultValue: new(false),
		},
		{
			name:             "default overrides hardDefault",
			wantBool:         true,
			defaultValue:     new(true),
			hardDefaultValue: new(false),
		},
		{
			name:             "hardDefault is last resort",
			wantBool:         true,
			hardDefaultValue: new(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := testOptions(t)
			options.SemanticVersioning = tc.rootValue
			options.Defaults.SemanticVersioning = tc.defaultValue
			options.HardDefaults.SemanticVersioning = tc.hardDefaultValue

			got := options.GetSemanticVersioning()

			if got != tc.wantBool {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.wantBool)
			}
		})
	}
}

func TestOptions_SemanticVersioningDiffers(t *testing.T) {
	tests := []struct {
		name                     string
		rootValue                *bool
		softDefault, hardDefault *bool
		override                 *string
		want                     bool
	}{
		{
			name:        "nil/unchanged",
			rootValue:   new(true),
			hardDefault: new(false),
			want:        false,
		},
		{
			name:        "true/root true",
			rootValue:   new(true),
			hardDefault: new(false),
			override:    new("true"),
			want:        false,
		},
		{
			name:        "true/root false",
			rootValue:   new(false),
			hardDefault: new(true),
			override:    new("true"),
			want:        true,
		},
		{
			name:        "true/root unset, same hard default",
			hardDefault: new(true),
			override:    new("true"),
			want:        false,
		},
		{
			name:        "true/root unset, diff soft default",
			softDefault: new(false),
			hardDefault: new(true),
			override:    new("true"),
			want:        true,
		},
		{
			name:        "false/root false",
			rootValue:   new(false),
			hardDefault: new(true),
			override:    new("false"),
			want:        false,
		},
		{
			name:        "false/root true",
			rootValue:   new(true),
			hardDefault: new(false),
			override:    new("false"),
			want:        true,
		},
		{
			name:        "false/root unset, same hard default",
			hardDefault: new(false),
			override:    new("false"),
			want:        false,
		},
		{
			name:        "false/root unset, diff hard default",
			hardDefault: new(true),
			override:    new("false"),
			want:        true,
		},
		{
			name:        "null/soft default set, root matches",
			rootValue:   new(false),
			softDefault: new(false),
			hardDefault: new(true),
			override:    new("null"),
			want:        false,
		},
		{
			name:        "null/soft default set, root differs",
			rootValue:   new(true),
			softDefault: new(false),
			hardDefault: new(true),
			override:    new("null"),
			want:        true,
		},
		{
			name:        "null/soft default set, root unset",
			softDefault: new(false),
			hardDefault: new(true),
			override:    new("null"),
			want:        false,
		},
		{
			name:        "null/hard default set, root differs",
			rootValue:   new(false),
			hardDefault: new(true),
			override:    new("null"),
			want:        true,
		},
		{
			name:        "null/hard default set, root matches",
			rootValue:   new(true),
			hardDefault: new(true),
			override:    new("null"),
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				Base:         Base{SemanticVersioning: tc.rootValue},
				Defaults:     &Defaults{Base: Base{SemanticVersioning: tc.softDefault}},
				HardDefaults: &Defaults{Base: Base{SemanticVersioning: tc.hardDefault}},
			}

			got := options.SemanticVersioningDiffers(tc.override)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestOptions_SameSemanticVersioning(t *testing.T) {
	tests := []struct {
		name        string
		this, other *Options
		want        bool
	}{
		{name: "both nil Options", want: true},
		{name: "both unset", this: &Options{}, other: &Options{}, want: true},
		{name: "nil Options and unset", other: &Options{}, want: true},
		{
			name:  "same value, different pointers",
			this:  &Options{Base: Base{SemanticVersioning: new(true)}},
			other: &Options{Base: Base{SemanticVersioning: new(true)}},
			want:  true,
		},
		{
			name:  "different values",
			this:  &Options{Base: Base{SemanticVersioning: new(true)}},
			other: &Options{Base: Base{SemanticVersioning: new(false)}},
			want:  false,
		},
		{
			name:  "set and unset",
			this:  &Options{Base: Base{SemanticVersioning: new(true)}},
			other: &Options{},
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.this.SameSemanticVersioning(tc.other); got != tc.want {
				t.Errorf("Options.SameSemanticVersioning() mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestOptions_SetSemanticVersioning(t *testing.T) {
	tests := []struct {
		name      string
		rootValue *bool
		override  string
		want      *bool
		errRegex  string
	}{
		{
			name:     "true/root unset",
			override: "true",
			want:     new(true),
			errRegex: `^$`,
		},
		{
			name:      "false/root true",
			rootValue: new(true),
			override:  "false",
			want:      new(false),
			errRegex:  `^$`,
		},
		{
			name:      "null/clears root",
			rootValue: new(true),
			override:  "null",
			want:      nil,
			errRegex:  `^$`,
		},
		{
			name:      "invalid/root unchanged",
			rootValue: new(true),
			override:  `"abc"`,
			want:      new(true),
			errRegex: test.TrimYAML(`
				^semantic_versioning:
					json: .* unmarshal JSON string into Go bool$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &Options{Base: Base{SemanticVersioning: util.ClonePtr(tc.rootValue)}}
			previous := options.SemanticVersioning

			err := options.SetSemanticVersioning(tc.override)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			got, want := test.StringifyPtr(options.SemanticVersioning), test.StringifyPtr(tc.want)
			if got != want {
				t.Errorf("Options.SemanticVersioning mismatch\ngot:  %q\nwant: %q", got, want)
			}
			if previous != nil && *previous != *tc.rootValue {
				t.Errorf(
					"Options.SemanticVersioning previous target mismatch\n"+
						"got:  %t\nwant: %t (unchanged)",
					*previous, *tc.rootValue,
				)
			}
		})
	}
}

func TestOptions_VerifySemanticVersioning(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantVersion string
		errRegex    string
	}{
		{
			name:        "valid semantic version/MAJOR.MINOR.PATCH",
			version:     "1.0.0",
			wantVersion: "1.0.0",
			errRegex:    `^$`,
		},
		{
			name:        "valid semantic version/MAJOR.MINOR",
			version:     "1.0",
			wantVersion: "1.0.0", // semver.NewVersion pads the missing parts with zeros.
			errRegex:    `^$`,
		},
		{
			name:        "valid semantic version/MAJOR",
			version:     "1",
			wantVersion: "1.0.0",
			errRegex:    `^$`,
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

			got, err := options.VerifySemanticVersioning(tc.version, logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			var gotVersion string
			if got != nil {
				gotVersion = got.String()
			}
			if gotVersion != tc.wantVersion {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", gotVersion, tc.wantVersion)
			}
		})
	}
}

func TestOptions_GetIntervalPointer(t *testing.T) {
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

			got := options.GetIntervalPointer()

			if *got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", *got, tc.want)
			}
		})
	}
}

func TestOptions_GetIntervalDuration(t *testing.T) {
	options := testOptions(t)
	options.Interval = "3h2m1s"

	got := options.GetIntervalDuration()

	want := (3 * time.Hour) + (2 * time.Minute) + time.Second
	if got != want {
		t.Errorf("mismatch\ngot:  %s\nwant: %s", got, want)
	}
}

func TestOptions_CheckValues(t *testing.T) {
	optCfg := plainDefaultsConfig()

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
				tc.errRegex,
				tc.input.CheckValues,
			)

			if tc.wantInterval != "" && tc.input.Interval != tc.wantInterval {
				t.Errorf(
					"Options.Interval mismatch\ngot:  %q\nwant: %q",
					tc.input.Interval, tc.wantInterval,
				)
			}
		})
	}
}
