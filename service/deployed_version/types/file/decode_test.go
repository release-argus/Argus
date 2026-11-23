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

	"github.com/release-argus/Argus/service/deployed_version/types/base"
)

func TestLookup_Decode__emptyOrNull(t *testing.T) {
	for _, data := range []string{"", "null"} {
		t.Run(data, func(t *testing.T) {
			got, err := Decode("yaml", []byte(data), nil, nil, base.DefaultsConfig{})
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}
			if got != nil {
				t.Errorf("Decode() = %#v, want nil", got)
			}
		})
	}
}

func TestLookup_Decode__baseDecodeError(t *testing.T) {
	if _, err := Decode("yaml", []byte("type: []"), nil, nil, base.DefaultsConfig{}); err == nil {
		t.Fatal("Decode() unexpectedly succeeded with an invalid base field")
	}
}

func TestLookup_Decode__fileDecodeError(t *testing.T) {
	if _, err := Decode("yaml", []byte("path: []"), nil, nil, base.DefaultsConfig{}); err == nil {
		t.Fatal("Decode() unexpectedly succeeded with an invalid file field")
	}
}

func TestLookup_DecodeSelf__errorAndSuccess(t *testing.T) {
	lookup := &Lookup{Path: "/before"}
	if err := lookup.DecodeSelf("yaml", []byte("path: []")); err == nil {
		t.Fatal("DecodeSelf() unexpectedly succeeded with an invalid file field")
	}
	if lookup.Path != "/before" {
		t.Errorf("Path after failed DecodeSelf() = %q, want %q", lookup.Path, "/before")
	}

	if err := lookup.DecodeSelf("yaml", []byte("path: /after\nregex: 'v([0-9.]+)'")); err != nil {
		t.Fatalf("DecodeSelf() failed: %v", err)
	}
	if lookup.Path != "/after" || lookup.Regex != "v([0-9.]+)" {
		t.Errorf("DecodeSelf() result = %#v, want path and regex updated", lookup)
	}
}

func TestLookup_ApplyOverrides__errorAndEmptyData(t *testing.T) {
	lookup := &Lookup{Path: "/before"}
	if err := lookup.ApplyOverrides("yaml", []byte("path: []")); err == nil {
		t.Fatal("ApplyOverrides() unexpectedly succeeded with an invalid file field")
	}
	if err := lookup.ApplyOverrides("yaml", nil); err != nil {
		t.Fatalf("ApplyOverrides(empty) failed: %v", err)
	}
	if err := lookup.ApplyOverrides("yaml", []byte("path: /after")); err != nil {
		t.Fatalf("ApplyOverrides() failed: %v", err)
	}
	if lookup.Path != "/after" {
		t.Errorf("Path after ApplyOverrides() = %q, want %q", lookup.Path, "/after")
	}
}
