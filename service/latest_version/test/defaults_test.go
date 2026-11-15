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
	lvCfg := PlainDefaultsConfig()

	if lvCfg.Soft == nil || lvCfg.Hard == nil {
		t.Fatalf(
			"DefaultsConfig mismatch\n"+
				"got:  Soft nil=%t, Hard nil=%t\nwant: Soft nil=false, Hard nil=false",
			lvCfg.Soft == nil, lvCfg.Hard == nil,
		)
	}

	want := &lvCfg.Hard.Common.Require.Docker
	if got := lvCfg.Soft.Common.Require.Docker.Defaults; got != want {
		t.Errorf(
			"DefaultsConfig.Soft.Common.Require.Docker.Defaults mismatch\ngot:  %p\nwant: %p",
			got, want,
		)
	}
}
