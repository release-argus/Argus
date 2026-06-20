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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestConfig_LoadOrdering(t *testing.T) {
	// GIVEN: we have configs to load.
	tests := []struct {
		name  string
		file  func(path string)
		order []string
	}{
		{
			name:  "with services",
			file:  testYAML_Ordering_0,
			order: []string{"NoDefaults", "WantDefaults", "Disabled", "Gitea"},
		},
		{
			name: "obscure service names",
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
			},
		},
		{
			name:  "service block/none",
			file:  testYAML_Ordering_1_no_services,
			order: []string{},
		},
		{
			name:  "service block/empty/single empty line",
			file:  testYAML_Ordering_3_empty_line_after_service_line,
			order: []string{"C", "B", "A"},
		},
		{
			name:  "service block/empty/multiple empty lines",
			file:  testYAML_Ordering_4_multiple_empty_lines_after_service_line,
			order: []string{"P", "L", "S"},
		},
		{
			name:  "service block/empty/eof",
			file:  testYAML_Ordering_5_eof_is_service_line,
			order: []string{},
		},
		{
			name:  "service block/empty/more blocks",
			file:  testYAML_Ordering_6_no_services_after_service_line_another_block,
			order: []string{},
		},
		{
			name:  "service block/empty/no more blocks",
			file:  testYAML_Ordering_7_no_services_after_service_line,
			order: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're accessing flag env vars.
			g, _ := errgroup.WithContext(t.Context())

			file := filepath.Join(t.TempDir(), "config.yml")
			tc.file(file)

			// WHEN: they are loaded.
			flags := make(map[string]bool)
			var cfg Config
			loadMu.Lock() // Protect flag env vars.
			cfg.Load(t.Context(), g, file, &flags)
			t.Cleanup(func() {
				_ = os.Remove(cfg.Settings.DataDatabaseFile())
				loadMu.Unlock()
			})

			// THEN: it gets the ordering correctly.
			gotOrder := cfg.Order
			for i := range gotOrder {
				if i >= len(gotOrder) ||
					tc.order[i] != (gotOrder)[i] {
					t.Fatalf(
						"%s\nConfig LoadOrdering %q %s - order:\ngot:  %v\nwant: %v",
						packageName, file, tc.name, gotOrder, tc.order,
					)
				}
			}
		})
	}
}

func TestIndentation(t *testing.T) {
	// GIVEN: lines of varying indentation.
	tests := []struct {
		name        string
		input, want string
	}{
		{
			name:  "leading space",
			input: "   abc",
			want:  "   ",
		},
		{
			name:  "leading and trailing space",
			input: "   abc      ",
			want:  "   ",
		},
		{
			name:  "trailing space",
			input: "abc      ",
			want:  "",
		},
		{
			name:  "no indents",
			input: "abc",
			want:  "",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Indentation is called on it.
			got := Indentation(tc.input)

			// THEN: we get the indentation.
			if got != tc.want {
				t.Errorf(
					"%s\nIndentation(%q) mismatch:\ngot:  %q\nwant: %q",
					packageName, tc.input, got, tc.want,
				)
			}
		})
	}
}
