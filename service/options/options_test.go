// Copyright [2022] [Argus]
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

package options

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/utils"
)

func TestGetActive(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		active *bool
		want   bool
	}{
		"nil":   {active: nil, want: true},
		"true":  {active: boolPtr(true), want: true},
		"false": {active: boolPtr(false), want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testOptions()
			lookup.Active = tc.active

			// WHEN GetActive is called
			got := lookup.GetActive()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestGetInterval(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		intervalRoot        string
		intervalDefault     string
		intervalHardDefault string
		wantString          string
	}{
		"root overrides all": {wantString: "10s", intervalRoot: "10s",
			intervalDefault: "1m10s", intervalHardDefault: "1m10s"},
		"default overrides hardDefault": {wantString: "10s", intervalRoot: "",
			intervalDefault: "10s", intervalHardDefault: "1m10s"},
		"hardDefault is last resort": {wantString: "10s", intervalRoot: "", intervalDefault: "",
			intervalHardDefault: "10s"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testOptions()
			lookup.Interval = tc.intervalRoot
			lookup.Defaults.Interval = tc.intervalDefault
			lookup.HardDefaults.Interval = tc.intervalHardDefault

			// WHEN GetInterval is called
			got := lookup.GetInterval()

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Errorf("want: %q\ngot:  %q",
					tc.wantString, got)
			}
		})
	}
}

func TestGetSemanticVersioning(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		semanticVersioningRoot        *bool
		semanticVersioningDefault     *bool
		semanticVersioningHardDefault *bool
		wantBool                      bool
	}{
		"root overrides all": {wantBool: true, semanticVersioningRoot: boolPtr(true),
			semanticVersioningDefault: boolPtr(false), semanticVersioningHardDefault: boolPtr(false)},
		"default overrides hardDefault": {wantBool: true, semanticVersioningRoot: nil,
			semanticVersioningDefault: boolPtr(true), semanticVersioningHardDefault: boolPtr(false)},
		"hardDefault is last resort": {wantBool: true, semanticVersioningRoot: nil, semanticVersioningDefault: nil,
			semanticVersioningHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookup := testOptions()
			lookup.SemanticVersioning = tc.semanticVersioningRoot
			lookup.Defaults.SemanticVersioning = tc.semanticVersioningDefault
			lookup.HardDefaults.SemanticVersioning = tc.semanticVersioningHardDefault

			// WHEN GetSemanticVersioning is called
			got := lookup.GetSemanticVersioning()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("want: %t\ngot:  %t",
					tc.wantBool, got)
			}
		})
	}
}

func TestGetIntervalPointer(t *testing.T) {
	// GIVEN a Lookup
	lookup := testOptions()
	lookup.Interval = "10s"

	// WHEN GetIntervalPointer is called
	got := lookup.GetIntervalPointer()

	// THEN the function returns the correct result
	if got != &lookup.Interval {
		t.Errorf("want: %v\ngot:  %v",
			lookup.Interval, got)
	}
}

func TestGetIntervalDuration(t *testing.T) {
	// GIVEN a Lookup
	lookup := testOptions()
	lookup.Interval = "3h2m1s"

	// WHEN GetInterval is called
	got := lookup.GetIntervalDuration()

	// THEN the function returns the correct result
	want := (3 * time.Hour) + (2 * time.Minute) + time.Second
	if got != want {
		t.Errorf("want: %v\ngot:  %v",
			want, got)
	}
}

func TestPrint(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		options Options
		lines   int
	}{
		"empty/default Options":    {options: Options{}, lines: 0},
		"only active":              {options: Options{Active: boolPtr(false)}, lines: 2},
		"only interval":            {options: Options{Interval: "10s"}, lines: 2},
		"only semantic_versioning": {options: Options{SemanticVersioning: boolPtr(false)}, lines: 2},
		"all options defined":      {options: Options{Active: boolPtr(false), Interval: "10s", SemanticVersioning: boolPtr(false)}, lines: 4},
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

func TestCheckValues(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		options      Options
		wantInterval string
		errRegex     string
	}{
		"valid options": {errRegex: `^$`,
			options: Options{Active: boolPtr(false), Interval: "10s", SemanticVersioning: boolPtr(false)}},
		"invalid interval": {errRegex: `interval: .* <invalid>`,
			options: Options{Active: boolPtr(false), Interval: "10x", SemanticVersioning: boolPtr(false)}},
		"seconds get appended to pure decimal interval": {errRegex: `^$`, wantInterval: "10s",
			options: Options{Active: boolPtr(false), Interval: "10", SemanticVersioning: boolPtr(false)}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called
			err := tc.options.CheckValues("")

			// THEN it err's when expected
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}
