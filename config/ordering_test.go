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
