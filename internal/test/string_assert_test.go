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
	"strings"
	"testing"
)

func TestAssertStringWithPrefixes(t *testing.T) {
	tests := []struct {
		name       string
		stringify  func(prefix string) string
		want       string
		shouldFail bool
	}{
		{
			name: "prefixes string",
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
			t.Parallel()

			var fakeT FakeT
			AssertStringWithPrefixes(
				&fakeT,
				"stringify",
				tc.stringify,
				tc.want,
			)

			didFail := len(fakeT.Errors)+len(fakeT.Fatals) != 0
			if len(fakeT.Fatals) > 0 && !strings.HasPrefix(fakeT.Fatals[0], `stringify(prefix="") mismatch`) {
				t.Errorf(
					"FakeT.Fatals[0] mismatch\ngot:  %q\nwant: %q",
					fakeT.Fatals[0], `stringify(prefix="") mismatch...`,
				)
			}
			if didFail != tc.shouldFail {
				t.Errorf(
					"didFail mismatch\ngot:  %t (errors=%q, fatals=%q)\nwant: %t",
					didFail, fakeT.Errors, fakeT.Fatals, tc.shouldFail,
				)
			}
		})
	}
}
