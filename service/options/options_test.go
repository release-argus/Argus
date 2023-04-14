// Copyright [2023] [Argus]
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

package opt

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
)

func TestOptions_GetActive(t *testing.T) {
	// GIVEN Options
	tests := map[string]struct {
		active *bool
		want   bool
	}{
		"nil": {
			active: nil,
			want:   true},
		"true": {
			active: boolPtr(true),
			want:   true},
		"false": {
			active: boolPtr(false),
			want:   false},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := testOptions()
			options.Active = tc.active

			// WHEN GetActive is called
			got := options.GetActive()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestOptions_GetInterval(t *testing.T) {
	// GIVEN Options
	tests := map[string]struct {
		intervalRoot        string
		intervalDefault     string
		intervalHardDefault string
		wantString          string
	}{
		"root overrides all": {
			wantString:          "10s",
			intervalRoot:        "10s",
			intervalDefault:     "1m10s",
			intervalHardDefault: "1m10s",
		},
		"default overrides hardDefault": {
			wantString:          "10s",
			intervalRoot:        "",
			intervalDefault:     "10s",
			intervalHardDefault: "1m10s",
		},
		"hardDefault is last resort": {
			wantString:          "10s",
			intervalRoot:        "",
			intervalDefault:     "",
			intervalHardDefault: "10s",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := testOptions()
			options.Interval = tc.intervalRoot
			options.Defaults.Interval = tc.intervalDefault
			options.HardDefaults.Interval = tc.intervalHardDefault

			// WHEN GetInterval is called
			got := options.GetInterval()

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Errorf("want: %q\ngot:  %q",
					tc.wantString, got)
			}
		})
	}
}

func TestOptions_GetSemanticVersioning(t *testing.T) {
	// GIVEN Options
	tests := map[string]struct {
		semanticVersioningRoot        *bool
		semanticVersioningDefault     *bool
		semanticVersioningHardDefault *bool
		wantBool                      bool
	}{
		"root overrides all": {
			wantBool:                      true,
			semanticVersioningRoot:        boolPtr(true),
			semanticVersioningDefault:     boolPtr(false),
			semanticVersioningHardDefault: boolPtr(false),
		},
		"default overrides hardDefault": {
			wantBool:                      true,
			semanticVersioningRoot:        nil,
			semanticVersioningDefault:     boolPtr(true),
			semanticVersioningHardDefault: boolPtr(false),
		},
		"hardDefault is last resort": {
			wantBool:                      true,
			semanticVersioningRoot:        nil,
			semanticVersioningDefault:     nil,
			semanticVersioningHardDefault: boolPtr(true),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := testOptions()
			options.SemanticVersioning = tc.semanticVersioningRoot
			options.Defaults.SemanticVersioning = tc.semanticVersioningDefault
			options.HardDefaults.SemanticVersioning = tc.semanticVersioningHardDefault

			// WHEN GetSemanticVersioning is called
			got := options.GetSemanticVersioning()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("want: %t\ngot:  %t",
					tc.wantBool, got)
			}
		})
	}
}

func TestOptions_GetIntervalPointer(t *testing.T) {
	// GIVEN options
	tests := map[string]struct {
		options *Options
		want    string
	}{
		"root overrides all": {
			options: &Options{
				Interval: "10s",
				Defaults: &Options{
					Interval: "20s"},
				HardDefaults: &Options{
					Interval: "30s"}},
			want: "10s"},
		"default overrides hardDefault": {
			options: &Options{
				Defaults: &Options{
					Interval: "20s"},
				HardDefaults: &Options{
					Interval: "30s"}},
			want: "20s"},
		"hardDefault is last resort": {
			options: &Options{
				Defaults: &Options{},
				HardDefaults: &Options{
					Interval: "30s"}},
			want: "30s"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetIntervalPointer is called
			got := tc.options.GetIntervalPointer()

			// THEN the function returns the correct result
			if *got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, *got)
			}
		})
	}
}

func TestOptions_GetIntervalDuration(t *testing.T) {
	// GIVEN Options
	options := testOptions()
	options.Interval = "3h2m1s"

	// WHEN GetInterval is called
	got := options.GetIntervalDuration()

	// THEN the function returns the correct result
	want := (3 * time.Hour) + (2 * time.Minute) + time.Second
	if got != want {
		t.Errorf("want: %v\ngot:  %v",
			want, got)
	}
}

func TestOptions_Print(t *testing.T) {
	// GIVEN Options
	tests := map[string]struct {
		options Options
		lines   int
	}{
		"empty/default Options": {
			options: Options{},
			lines:   0,
		},
		"only active": {
			options: Options{
				Active: boolPtr(false)},
			lines: 2,
		},
		"only interval": {
			options: Options{
				Interval: "10s"},
			lines: 2,
		},
		"only semantic_versioning": {
			options: Options{
				SemanticVersioning: boolPtr(false)},
			lines: 2,
		},
		"all options defined": {
			options: Options{
				Active:             boolPtr(false),
				Interval:           "10s",
				SemanticVersioning: boolPtr(false)},
			lines: 4,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.options.Print("")

			// THEN it prints the expected number of lines
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
		})
	}
}

func TestOptions_CheckValues(t *testing.T) {
	// GIVEN Options
	tests := map[string]struct {
		options      Options
		wantInterval string
		errRegex     string
	}{
		"valid options": {
			errRegex: `^$`,
			options: Options{
				Active:             boolPtr(false),
				Interval:           "10s",
				SemanticVersioning: boolPtr(false)},
		},
		"invalid interval": {
			errRegex: `interval: .* <invalid>`,
			options: Options{
				Active:             boolPtr(false),
				Interval:           "10x",
				SemanticVersioning: boolPtr(false)},
		},
		"seconds get appended to pure decimal interval": {
			errRegex:     `^$`,
			wantInterval: "10s",
			options: Options{
				Active:             boolPtr(false),
				Interval:           "10",
				SemanticVersioning: boolPtr(false)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.options.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
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
			want:    "<nil>",
		},
		"empty/default Options": {
			options: &Options{},
			want:    "{}\n",
		},
		"all options defined": {
			options: &Options{
				Active: boolPtr(true), Interval: "10s", SemanticVersioning: boolPtr(true)},
			want: `
active: true
interval: 10s
semantic_versioning: true
`,
		},
		"empty with defaults": {
			options: &Options{
				Defaults: &Options{
					Active: boolPtr(true), Interval: "10s", SemanticVersioning: boolPtr(true)}},
			want: "{}\n",
		},
		"all with defaults": {
			options: &Options{
				Active: boolPtr(true), Interval: "10s", SemanticVersioning: boolPtr(true),
				Defaults: &Options{
					Active: boolPtr(false), Interval: "1h", SemanticVersioning: boolPtr(false)}},
			want: `
active: true
interval: 10s
semantic_versioning: true
`,
		},
		"empty with hardDefaults": {
			options: &Options{
				HardDefaults: &Options{
					Active: boolPtr(true), Interval: "10s", SemanticVersioning: boolPtr(true)}},
			want: "{}\n",
		},
		"all with hardDefaults": {
			options: &Options{
				Active: boolPtr(true), Interval: "10s", SemanticVersioning: boolPtr(true),
				HardDefaults: &Options{
					Active: boolPtr(false), Interval: "1h", SemanticVersioning: boolPtr(false)}},
			want: `
active: true
interval: 10s
semantic_versioning: true
`,
		},
		"empty with defaults and hardDefaults": {
			options: &Options{
				Defaults: &Options{
					Active: boolPtr(true), Interval: "10s", SemanticVersioning: boolPtr(true)},
				HardDefaults: &Options{
					Active: boolPtr(false), Interval: "1h", SemanticVersioning: boolPtr(false)}},
			want: "{}\n",
		},
		"all with defaults and hardDefaults": {
			options: &Options{
				Active: boolPtr(true), Interval: "10s", SemanticVersioning: boolPtr(true),
				Defaults: &Options{
					Active: boolPtr(false), Interval: "1h", SemanticVersioning: boolPtr(false)},
				HardDefaults: &Options{
					Active: boolPtr(true), Interval: "1m", SemanticVersioning: boolPtr(true)}},
			want: `
active: true
interval: 10s
semantic_versioning: true
`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Options is stringified with String
			got := tc.options.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
