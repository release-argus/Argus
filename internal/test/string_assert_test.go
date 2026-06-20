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

func TestAssertStringWithPrefixes(t *testing.T) {
	// GIVEN: a function that prefixes a string to another string.
	tests := []struct {
		name       string
		stringify  func(prefix string) string
		want       string
		shouldFail bool
	}{
		{
			name: "correctly prefixes string",
			stringify: func(prefix string) string {
				return prefix + "hello\n"
			},
			want: "hello\n",
		},
		{
			name: "always wrong",
			stringify: func(prefix string) string {
				return "wrong"
			},
			want:       "hello",
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			// WHEN: AssertStringWithPrefixes is called.
			var fakeT FakeT
			AssertStringWithPrefixes(
				&fakeT,
				"pkg",
				tc.stringify,
				tc.want,
			)
			didFail := len(fakeT.Errors) != 0

			if tc.shouldFail != didFail {
				t.Errorf(
					"%s\nAssertStringWithPrefixes() didn't pass/fail as expected\ngot  fail: %v\nwant fail: %t",
					packageName, didFail, tc.shouldFail,
				)
			}
		})
	}
}
