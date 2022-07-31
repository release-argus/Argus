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
	"os"
	"testing"

	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/options"
)

func TestOrdering(t *testing.T) {
	// GIVEN we have configs to load
	tests := map[string]struct {
		file  string
		all   []string
		order []string
	}{
		"with services": {file: "../test/ordering_0.yml",
			all:   []string{"NoDefaults", "WantDefaults", "Disabled", "Gitea"},
			order: []string{"NoDefaults", "WantDefaults", "Gitea"}},
		"no services": {file: "../test/ordering_1.yml",
			all:   []string{},
			order: []string{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN they are loaded
			config := testLoad(tc.file)

			// THEN it gets the ordering correctly
			gotAll := config.All
			for i := range gotAll {
				if i >= len(gotAll) || tc.all[i] != (gotAll)[i] {
					t.Fatalf("%q %s - all:\nwant:%v\ngot:  %v", tc.file, name, tc.all, gotAll)
				}
			}
			gotOrder := *config.Order
			for i := range gotOrder {
				if i >= len(gotOrder) || tc.order[i] != (gotOrder)[i] {
					t.Fatalf("%q %s - order:\nwant: %v\ngot:  %v", tc.file, name, tc.order, gotOrder)
				}
			}
			os.Remove(*config.Settings.GetDataDatabaseFile())
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
			input: "   abc", want: "   "},
		"leading and trailing space": {
			input: "   abc      ", want: "   "},
		"trailing space": {
			input: "abc      ", want: ""},
		"no indents": {
			input: "abc", want: ""},
		"empty string": {
			input: "", want: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

func TestFilterInactive(t *testing.T) {
	// GIVEN a Config with inactive and active services
	active := false
	allServices := []string{"1", "2", "3"}
	config := Config{
		Service: service.Slice{
			"1": &service.Service{},
			"2": &service.Service{Options: options.Options{Active: &active}},
			"3": &service.Service{},
		},
		All:   allServices,
		Order: &allServices,
	}

	// WHEN filterInactive is called on this Config
	config.filterInactive()

	// THEN the inactive Service is removed from Order
	if len(*config.Order) != 2 || (*config.Order)[1] != "3" {
		t.Fatalf("Service %q should have been removed from Order - %v",
			"2", config.Order)
	}
}
