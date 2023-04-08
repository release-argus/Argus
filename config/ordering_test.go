// Copyright [2022] [Argus]
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
	"sync"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestConfig_LoadOrdering(t *testing.T) {
	// GIVEN we have configs to load
	tests := map[string]struct {
		file  func(path string, t *testing.T)
		order []string
	}{
		"with services": {file: testYAML_Ordering_0,
			order: []string{"NoDefaults", "WantDefaults", "Disabled", "Gitea"}},
		"no services": {file: testYAML_Ordering_1,
			order: []string{}},
	}

	var lock sync.Mutex
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file := fmt.Sprintf("%s.yml", name)
			tc.file(file, t)

			// WHEN they are loaded
			flags := make(map[string]bool)
			log := util.NewJLog("WARN", true)
			var config Config
			// Lock as jLog = log would DATA RACE
			lock.Lock()
			config.Load(file, &flags, log)
			lock.Unlock()
			defer os.Remove(*config.Settings.GetDataDatabaseFile())

			// THEN it gets the ordering correctly
			gotOrder := config.Order
			for i := range gotOrder {
				if i >= len(gotOrder) || tc.order[i] != (gotOrder)[i] {
					t.Fatalf("%q %s - order:\nwant: %v\ngot:  %v",
						file, name, tc.order, gotOrder)
				}
			}
		})
	}
}

func TestGetIndentationW(t *testing.T) {
	// GIVEN lines of varying indentation
	tests := map[string]struct {
		input string
		want  string
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN getIndentation is called on it
			got := getIndentation(tc.input)

			// THEN we get the indentation
			if got != tc.want {
				t.Errorf("%s - %q:\nwant: %q\ngot:  %q",
					name, tc.input, tc.want, got)
			}
		})
	}
}
