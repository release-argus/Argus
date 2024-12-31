// Copyright [2024] [Argus]
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
	"github.com/release-argus/Argus/test"
)

func TestLookup_Init(t *testing.T) {
	// GIVEN a Lookup and vars for the Init
	lookup := testLookup("github", false)
	var defaults base.Defaults
	var hardDefaults base.Defaults
	*lookup.GetStatus().ServiceID += "TestInit"
	status := status.Status{ServiceID: test.StringPtr("test")}
	var options opt.Options

	// WHEN Init is called on it
	lookup.Init(
		&options,
		&status,
		&defaults, &hardDefaults)

	// THEN pointers to those vars are handed out to the Lookup
	// defaults
	if lookup.GetDefaults() != &defaults {
		t.Errorf("Defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&defaults, lookup.GetDefaults())
	}
	// hardDefaults
	if lookup.GetHardDefaults() != &hardDefaults {
		t.Errorf("HardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&hardDefaults, lookup.GetHardDefaults())
	}
	// status
	if lookup.GetStatus() != &status {
		t.Errorf("Status was not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&status, lookup.GetStatus())
	}
	// options
	if lookup.GetOptions() != &options {
		t.Errorf("Options were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&options, lookup.GetOptions())
	}
}
