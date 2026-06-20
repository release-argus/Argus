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

import (
	"testing"

	opt "github.com/release-argus/Argus/service/option"
)

func TestPlainOptions(t *testing.T) {
	// GIVEN: defaults and hardDefaults.
	defaults := &opt.Defaults{}
	hardDefaults := &opt.Defaults{}

	// WHEN: PlainOptions is called with them.
	got := PlainOptions(
		t,
		opt.DefaultsConfig{
			Soft: defaults,
			Hard: hardDefaults,
		},
	)

	// THEN: got should be non-nil.
	if got == nil {
		t.Errorf("%s\nPlainOptions returned nil, expected non-nil", packageName)
	}

	// AND: the defaults should be a pointer to the defaults provided.
	if got.Defaults != defaults {
		t.Errorf(
			"%s\nPlainOptions defaults pointer mismatch on PlainDefaults()\ngot:  %p\nwant: %p",
			packageName, got.Defaults, defaults,
		)
	}

	// AND: the hardDefaults should be a pointer to the hardDefaults provided.
	if got.HardDefaults != hardDefaults {
		t.Errorf(
			"%s\nPlainOptions hardDefaults pointer mismatch on PlainDefaults()\ngot:  %p\nwant: %p",
			packageName, got.HardDefaults, hardDefaults,
		)
	}
}

func TestOptions(t *testing.T) {
	// WHEN: Options is called.
	got := Options(t)

	// THEN: the Options should be non-nil.
	if got == nil {
		t.Errorf("%s\nOptions() = nil, want non-nil", packageName)
	}
}
