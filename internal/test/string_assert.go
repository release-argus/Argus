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
// GIVEN: a function that takes a prefix and returns a string with that prefix applied.
//
// WHEN: String(prefix) is called with each prefix.
//
// THEN: the rendered output must match exactly for each prefix.
func AssertStringWithPrefixes(
	t tLogger,
	packageName string,
	stringify func(prefix string) string,
	want string,
) {
	t.Helper()

	prefixes := []string{"", " ", "  ", "    ", "- "}

	for _, prefix := range prefixes {
		expected := addPrefix(want, prefix)

		// WHEN: the value is stringified with the given prefix.
		got := stringify(prefix)

		// THEN: the output matches exactly.
		if got != expected {
			t.Fatalf(
				"%s\nStringified mismatch (prefix=%q)\ngot:  %q\nwant: %q",
				packageName, prefix,
				got, expected,
			)
		}
	}
}
