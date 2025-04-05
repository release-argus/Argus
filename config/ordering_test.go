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

package config

import (
	"fmt"
	"os"
	"testing"
)

func TestConfig_LoadOrdering(t *testing.T) {
	// GIVEN we have configs to load.
	tests := map[string]struct {
		file  func(path string, t *testing.T)
		order []string
	}{
		"with services": {
			file:  testYAML_Ordering_0,
			order: []string{"NoDefaults", "WantDefaults", "Disabled", "Gitea"}},
		"no services": {
			file:  testYAML_Ordering_1_no_services,
			order: []string{}},
		"obscure service names": {
			file: testYAML_Ordering_2_obscure_service_names,
			order: []string{
				`123`,
				`foo bar`,
				`foo: bar`,
				`foo: "bar"`,
				`"foo: bar"`,
				`'foo: bar'`,
				`"foo bar"`,
				`'foo bar'`,
				`foo "bar"`,
				`foo: bar, baz`,
			}},
		"empty line after 'service:'": {
			file:  testYAML_Ordering_3_empty_line_after_service_line,
			order: []string{"C", "B", "A"}},
		"multiple empty lines after 'service:'": {
			file:  testYAML_Ordering_4_multiple_empty_lines_after_service_line,
			order: []string{"P", "L", "S"}},
		"eof on 'service:'": {
			file:  testYAML_Ordering_5_eof_is_service_line,
			order: []string{}},
		"no services after 'service:' - another block": {
			file:  testYAML_Ordering_6_no_services_after_service_line_another_block,
			order: []string{}},
		"no services after 'service:'": {
			file:  testYAML_Ordering_7_no_services_after_service_line,
			order: []string{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're accessing flag env vars.

			file := fmt.Sprintf("%s.yml", name)
			tc.file(file, t)

			// WHEN they are loaded.
			flags := make(map[string]bool)
			var config Config
			loadMutex.Lock() // Protect flag env vars.
			config.Load(file, &flags)
			t.Cleanup(func() {
				os.Remove(config.Settings.DataDatabaseFile())
				loadMutex.Unlock()
			})

			// THEN it gets the ordering correctly.
			gotOrder := config.Order
			for i := range gotOrder {
				if i >= len(gotOrder) ||
					tc.order[i] != (gotOrder)[i] {
					t.Fatalf("%s\n%q %s - order:\nwant: %v\ngot:  %v",
						packageName, file, name, tc.order, gotOrder)
				}
			}
		})
	}
}

func TestIndentationW(t *testing.T) {
	// GIVEN lines of varying indentation.
	tests := map[string]struct {
		input, want string
	}{
		"leading space": {
			input: "   abc",
			want:  "   "},
		"leading and trailing space": {
			input: "   abc      ",
			want:  "   "},
		"trailing space": {
			input: "abc      ",
			want:  ""},
		"no indents": {
			input: "abc",
			want:  ""},
		"empty string": {
			input: "",
			want:  ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Indentation is called on it.
			got := Indentation(tc.input)

			// THEN we get the indentation.
			if got != tc.want {
				t.Errorf("%s\n%s - %q:\nwant: %q\ngot:  %q",
					packageName, name, tc.input, tc.want, got)
			}
		})
	}
}
