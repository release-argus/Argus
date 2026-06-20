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

package manual

import "testing"

func TestLookup_GetType(t *testing.T) {
	// GIVEN: a Lookup with a Type.
	tests := []struct {
		name  string
		lType string
	}{
		{name: "empty", lType: ""},
		{name: "test", lType: "test"},
		{name: "x", lType: "x"},
		{name: "y", lType: "y"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &Lookup{}
			l.Type = tc.lType

			// WHEN: GetType is called.
			got := l.GetType()

			wantType := "manual"
			// THEN: the Type is returned.
			if got != wantType {
				t.Errorf(
					"%s\nLookup.GetType() mismatch\ngot:  %q\nwant: %q",
					packageName, got, wantType,
				)
			}
		})
	}
}
