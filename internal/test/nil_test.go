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
)

func TestIsNil(t *testing.T) {
	tests := []struct {
		name string
		x    any
		want bool
	}{
		{
			name: "nil",
			x:    nil,
			want: true,
		},
		{
			name: "int",
			x:    1,
			want: false,
		},
		{
			name: "string",
			x:    "hello",
			want: false,
		},
		{
			name: "chan",
			x:    make(chan int),
			want: false,
		},
		{
			name: "func",
			x:    func() {},
			want: false,
		},
		{
			name: "interface",
			x:    interface{}(nil),
			want: true,
		},
		{
			name: "map",
			x:    map[string]int{},
			want: false,
		},
		{
			name: "pointer",
			x:    (*int)(nil),
			want: true,
		},
		{
			name: "slice",
			x:    []int{},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsNil is called on it.
			got := isNil(tc.x)

			// THEN: The result is as expected.
			if got != tc.want {
				t.Fatalf(
					"%s\nIsNil(%v) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.x, got, tc.want,
				)
			}
		})
	}
}
