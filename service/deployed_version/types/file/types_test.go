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

package file

import (
	"testing"

	"github.com/release-argus/Argus/service/status"
)

func TestLookup_Copy(t *testing.T) {
	lookup, _ := newTestLookup(t, "1.2.3")
	wantStatus := &status.Status{}

	gotInterface := lookup.Copy(wantStatus)
	got, ok := gotInterface.(*Lookup)
	if !ok {
		t.Fatalf("Lookup.Copy() type = %T, want *Lookup", gotInterface)
	}
	if got == lookup {
		t.Fatal("Lookup.Copy() returned the original lookup")
	}
	if got.Path != lookup.Path || got.Regex != lookup.Regex || got.RegexTemplate != lookup.RegexTemplate {
		t.Errorf("Lookup.Copy() = %#v, want file settings copied from %#v", got, lookup)
	}
	if got.Status != wantStatus {
		t.Errorf("Lookup.Copy().Status = %p, want %p", got.Status, wantStatus)
	}
	if got.Options == lookup.Options {
		t.Error("Lookup.Copy().Options aliases the original options")
	}
}

func TestLookup_String(t *testing.T) {
	lookup := &Lookup{Path: "/VERSION", Regex: `v([0-9.]+)`, RegexTemplate: "$1"}

	if got, want := lookup.String(""), "path: /VERSION\nregex: v([0-9.]+)\nregex_template: $1\n"; got != want {
		t.Errorf("Lookup.String() = %q, want %q", got, want)
	}
}

func TestLookup_GetType(t *testing.T) {
	if got := (&Lookup{}).GetType(); got != Type {
		t.Errorf("Lookup.GetType() = %q, want %q", got, Type)
	}
}
