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

package latestver

import (
	"testing"

	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_Init(t *testing.T) {
	// GIVEN a Lookup and vars for the Init.
	lookup := testLookup("github", false)
	var defaults base.Defaults
	var hardDefaults base.Defaults
	lookup.GetStatus().ServiceInfo.ID += "TestInit"
	status := status.Status{}
	status.ServiceInfo.ID = "test"
	var options opt.Options

	// WHEN Init is called on it.
	lookup.Init(
		&options,
		&status,
		&defaults, &hardDefaults)

	// THEN pointers to those vars are handed out to the Lookup:
	// 	Defaults.
	if lookup.GetDefaults() != &defaults {
		t.Errorf("%s\nDefaults were not handed to the Lookup correctly\nwant: %v\ngot:  %v",
			packageName, &defaults, lookup.GetDefaults())
	}
	// HardDefaults.
	if lookup.GetHardDefaults() != &hardDefaults {
		t.Errorf("%s\nHardDefaults were not handed to the Lookup correctly\nwant: %v\ngot:  %v",
			packageName, &hardDefaults, lookup.GetHardDefaults())
	}
	// 	Status.
	if lookup.GetStatus() != &status {
		t.Errorf("%s\nStatus was not handed to the Lookup correctly\nwant: %v\ngot:  %v",
			packageName, &status, lookup.GetStatus())
	}
	// 	Options.
	if lookup.GetOptions() != &options {
		t.Errorf("%s\nOptions were not handed to the Lookup correctly\nwant: %v\ngot:  %v",
			packageName, &options, lookup.GetOptions())
	}
}
