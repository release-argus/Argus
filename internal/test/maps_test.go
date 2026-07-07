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
	"fmt"
	"testing"
)

func TestMapJoin(t *testing.T) {
	// GIVEN: a map to join with a separator.
	tests := []struct {
		name string
		m    map[string]string
		sep  string
		want []string
	}{
		{
			name: "empty map",
			m:    map[string]string{},
			sep:  "_",
			want: []string{},
		},
		{
			name: "empty separator",
			m: map[string]string{
				"a": "1",
			},
			sep:  "",
			want: []string{"a1"},
		},
		{
			name: "single entry",
			m: map[string]string{
				"key": "value",
			},
			sep:  ":",
			want: []string{"key:value"},
		},
		{
			name: "multiple entires",
			m: map[string]string{
				"a": "5",
				"b": "6",
				"c": "0",
			},
			sep:  "_",
			want: []string{"a_5", "b_6", "c_0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: MapJoin is called on them.
			got := MapJoin(tc.m, tc.sep)

			prefix := fmt.Sprintf(
				"%s\nMapJoin(m=%v, sep=%q)",
				packageName, tc.m, tc.sep,
			)

			// THEN: they are joined as expected.
			if testErr := AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b string) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
