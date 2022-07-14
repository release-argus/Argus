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
	"testing"

	"github.com/release-argus/Argus/service"
)

func TestOrderingWithServices(t *testing.T) {
	// GIVEN Load is ran on a config.yml
	config := testLoad("../test/ordering_0.yml")

	// WHEN the Default Service Interval is looked at
	got := config.All

	// THEN it matches the config.yml
	want := []string{"NoDefaults", "WantDefaults", "Disabled", "Gitea"}
	for i := range got {
		if i >= len(got) || want[i] != (got)[i] {
			t.Errorf(`Order should have been %v, but got %v`, want, got)
		}
	}
}

func TestOrderingWithService(t *testing.T) {
	// GIVEN Load is ran on a config.yml
	config := testLoad("../test/ordering_1.yml")

	// WHEN the Default Service Interval is looked at
	got := config.Defaults.Service.Interval

	// THEN it matches the config.yml
	want := "123s"
	if !(want == *got) {
		t.Errorf(`config.Defaults.Service.Interval = %v, want %s`, *got, want)
	}
}

func TestGetIndentationWithIndentation(t *testing.T) {
	// GIVEN a line with 3 levels of indentation
	line := "   abc"

	// WHEN getIndentation is called on it
	got := getIndentation(line)

	// THEN we get the indentation
	want := "   "
	if got != want {
		t.Errorf("%q should have returned an indentation of %q, not %q",
			line, want, got)
	}
}

func TestGetIndentationWithNoIndentation(t *testing.T) {
	// GIVEN a line with no indentation
	line := "abc"

	// WHEN getIndentation is called on it
	got := getIndentation(line)

	// THEN we get the indentation (none)
	want := ""
	if got != want {
		t.Errorf("%q should have returned an indentation of %q, not %q",
			line, want, got)
	}
}

func TestGetIndentationWithBlankLine(t *testing.T) {
	// GIVEN a blank line
	line := ""

	// WHEN getIndentation is called on it
	got := getIndentation(line)

	// THEN we get the indentation (none)
	want := ""
	if got != want {
		t.Errorf("%q should have returned an indentation of %q, not %q",
			line, want, got)
	}
}

func TestFilterInactive(t *testing.T) {
	// GIVEN a Config with inactive and active services
	active := false
	allServices := []string{"1", "2", "3"}
	config := Config{
		Service: service.Slice{
			"1": &service.Service{},
			"2": &service.Service{Active: &active},
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
