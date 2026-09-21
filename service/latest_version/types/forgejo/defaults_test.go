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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

import (
	"testing"
)

func TestDefaults_Default(t *testing.T) {
	// GIVEN: a Defaults.
	defaults := Defaults{}

	// WHEN: Default is called.
	defaults.Default()

	// THEN: prereleases are excluded.
	if defaults.Common.UsePreRelease == nil || *defaults.Common.UsePreRelease {
		t.Fatalf(
			"%s\nDefaults.Default() use_prerelease mismatch\ngot:  %v\nwant: false",
			packageName, defaults.Common.UsePreRelease,
		)
	}
}
