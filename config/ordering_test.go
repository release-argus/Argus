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

	"github.com/release-argus/Argus/util"
)

func TestConfig_LoadOrdering(t *testing.T) {
	tests := []struct {
		name  string
		file  func(t *testing.T, path string)
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
			name:  "services named like blocks, and a nil service",
			file:  testYAML_Ordering_8_services_named_like_blocks,
			order: []string{"settings", "defaults"},
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
			// Serial: shares the flag vars.

			g, _ := errgroup.WithContext(t.Context())

			file := filepath.Join(t.TempDir(), "config.yml")
			tc.file(t, file)

			flags := make(map[string]bool)
			var cfg Config
			// loadMu guards the flag vars.
			loadMu.Lock()
			cfg.Load(t.Context(), g, file, &flags)
			t.Cleanup(func() {
				_ = os.Remove(cfg.Settings.DataDatabaseFile())
				loadMu.Unlock()
			})

			if !util.AreSlicesEqual(cfg.Order, tc.order) {
				t.Fatalf("Config.Order mismatch\ngot:  %q\nwant: %q", cfg.Order, tc.order)
			}
		})
	}
}

func TestIndentation(t *testing.T) {
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

			got := Indentation(tc.input)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
