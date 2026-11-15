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
	defaults := &opt.Defaults{}
	hardDefaults := &opt.Defaults{}

	got := PlainOptions(opt.DefaultsConfig{Soft: defaults, Hard: hardDefaults})

	if got == nil {
		t.Fatal("PlainOptions() mismatch\ngot:  nil\nwant: non-nil")
	}

	if got.Defaults != defaults {
		t.Errorf("Options.Defaults mismatch\ngot:  %p\nwant: %p", got.Defaults, defaults)
	}

	if got.HardDefaults != hardDefaults {
		t.Errorf(
			"Options.HardDefaults mismatch\ngot:  %p\nwant: %p",
			got.HardDefaults, hardDefaults,
		)
	}
}

func TestOptions(t *testing.T) {
	got := Options(t)

	if want := "1m"; got.Interval != want {
		t.Errorf("Options.Interval mismatch\ngot:  %q\nwant: %q", got.Interval, want)
	}
}
