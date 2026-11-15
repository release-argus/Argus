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

//go:build unit || integration

package test

// AssertStringWithPrefixes tests String(prefix) output against an expected
// multiline string across a standard set of prefixes.
//
// For each prefix, stringify(prefix) must equal want with that prefix applied.
// call names stringify in the failure message, e.g. "Defaults.String".
func AssertStringWithPrefixes(
	t TB,
	call string,
	stringify func(prefix string) string,
	want string,
) {
	t.Helper()

	prefixes := []string{"", " ", "  ", "    ", "- "}

	for _, prefix := range prefixes {
		expected := addPrefix(want, prefix)

		got := stringify(prefix)

		if got != expected {
			t.Fatalf("%s(prefix=%q) mismatch\ngot:  %q\nwant: %q", call, prefix, got, expected)
		}
	}
}
