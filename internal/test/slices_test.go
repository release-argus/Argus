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

func TestPrefixAll(t *testing.T) {
	// GIVEN: a slice and a prefix.
	tests := []struct {
		name   string
		input  []string
		prefix string
		want   []string
	}{
		{
			name:   "empty slice",
			input:  []string{},
			prefix: "hello_",
			want:   []string{},
		},
		{
			name:   "empty prefix",
			input:  []string{"a", "b"},
			prefix: "",
			want:   []string{"a", "b"},
		},
		{
			name:   "single element",
			input:  []string{"a"},
			prefix: "x_",
			want:   []string{"x_a"},
		},
		{
			name:   "prefix all elements",
			input:  []string{"a", "b", "c"},
			prefix: "hello_",
			want:   []string{"hello_a", "hello_b", "hello_c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: PrefixAll is called on them.
			got := PrefixAll(tc.input, tc.prefix)

			prefix := fmt.Sprintf(
				"%s PrefixAll(s=%v, prefix=%q)",
				packageName, tc.input, tc.prefix,
			)

			// THEN: all elements of the slice are prefixed as expected.
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
