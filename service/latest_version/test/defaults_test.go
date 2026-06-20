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

package test

import "testing"

func TestPlainDefaultsConfig(t *testing.T) {
	// WHEN: PlainDefaultsConfig is called.
	lvCfg := PlainDefaultsConfig(t)

	// THEN: It returns a set of defaults.
	if lvCfg.Soft == nil || lvCfg.Hard == nil {
		t.Fatalf(
			"%s\nPlainDefaultsConfig() returned nil, defaults: %v, hardDefaults: %v",
			packageName, lvCfg.Soft, lvCfg.Hard,
		)
	}

	// AND: the defaults of the defaults.Require are the hardDefaults.Require.
	want := &lvCfg.Hard.Require.Docker
	if got := lvCfg.Soft.Require.Docker.Defaults; got != want {
		t.Errorf(
			"%s\nPlainDefaultsConfig() defaults.Require.Default() not set to expected HardDefaults mismatch\ngot:  %p\nwant %p",
			packageName, got, want,
		)
	}
}
