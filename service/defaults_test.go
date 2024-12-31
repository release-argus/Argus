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

package service

import "testing"

func TestDefaults_Default(t *testing.T) {
	// GIVEN a Defaults struct
	d := &Defaults{}

	// WHEN Default is called
	d.Default()

	// THEN the struct is populated with default values
	if d.Options.Interval != "10m" {
		t.Errorf("invalid Options.Interval:\nwant: 5m\ngot:  %s", d.Options.Interval)
	}
	if d.Dashboard.AutoApprove == nil {
		t.Error("invalid Dashboard.AutoApprove:\nwant: non-nil\ngot:  nil")
	}
	// AND the X.Options vars are pointing to the Options struct
	if d.LatestVersion.Options != &d.Options {
		t.Errorf("invalid LatestVersion.Options:\nwant: %p\ngot:  %p", &d.Options, d.LatestVersion.Options)
	}
	if d.DeployedVersionLookup.Options != &d.Options {
		t.Errorf("invalid DeployedVersionLookup.Options:\nwant: %p\ngot:  %p", &d.Options, d.DeployedVersionLookup.Options)
	}
}