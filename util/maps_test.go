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

package util

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestEnsureMap(t *testing.T) {
	// GIVEN: a map.
	tests := []struct {
		name  string
		input map[string]string
	}{
		{
			name:  "nil map",
			input: nil,
		},
		{
			name:  "empty map",
			input: map[string]string{},
		},
		{
			name: "non-empty map",
			input: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "non-empty map with same keys but differing case",
			input: map[string]string{
				"test": "123",
				"tESt": "bar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := CopyMap(tc.input)

			// WHEN: EnsureMap is called.
			result := EnsureMap(tc.input)

			prefix := fmt.Sprintf(
				"%s\nEnsureMap(%v)",
				packageName, tc.input,
			)

			// THEN: the map is initialised correctly.
			if result == nil {
				t.Fatalf("%s map is still nil", prefix)
			}

			// AND: any values inside haven't changed.
			if testErr := test.AssertMapEqual(
				t,
				result,
				had,
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	// GIVEN: two maps and a list of fields that may contain secrets.
	tests := []struct {
		name                  string
		base, overrides, want map[string]string
	}{
		{
			name:      "empty maps",
			base:      map[string]string{},
			overrides: map[string]string{},
			want:      map[string]string{},
		},
		{
			name:      "nil maps",
			base:      nil,
			overrides: nil,
			want:      map[string]string{},
		},
		{
			name: "empty base map",
			base: map[string]string{},
			overrides: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "nil base map",
			base: nil,
			overrides: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "empty overrides map",
			base: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			overrides: map[string]string{},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "nil overrides map",
			base: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			overrides: nil,
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "non-empty maps",
			base: map[string]string{
				"test":      "123",
				"foo":       "bar",
				"bish":      "bash",
				"something": "else",
			},
			overrides: map[string]string{
				"test": "456",
				"foo":  "baz",
				"bish": "",
			},
			want: map[string]string{
				"test":      "456",
				"foo":       "baz",
				"bish":      "",
				"something": "else",
			},
		},
		{
			name: "ref secret in base map",
			base: map[string]string{
				"test": "123",
			},
			overrides: map[string]string{
				"test": SecretValue,
			},
			want: map[string]string{
				"test": SecretValue,
			},
		},
		{
			name: "ref secret in base map, secret not found/empty",
			base: map[string]string{
				"foo": "",
			},
			overrides: map[string]string{
				"foo":  SecretValue,
				"test": SecretValue,
			},
			want: map[string]string{
				"foo":  SecretValue,
				"test": SecretValue,
			},
		},
		{
			name: "secret not in fields",
			base: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			overrides: map[string]string{
				"test": SecretValue,
				"foo":  SecretValue,
			},
			want: map[string]string{
				"foo":  SecretValue,
				"test": SecretValue,
			},
		},
		{
			name: "non-empty maps with secrets",
			base: map[string]string{
				"test":      "123",
				"foo":       "bar",
				"bish":      "bash",
				"something": "else",
				"nothing":   "",
			},
			overrides: map[string]string{
				"test":    "456",
				"foo":     SecretValue,
				"bish":    SecretValue,
				"nothing": SecretValue,
			},
			want: map[string]string{
				"test":      "456",
				"foo":       SecretValue,
				"bish":      SecretValue,
				"something": "else",
				"nothing":   SecretValue,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: MergeMaps is called.
			result := MergeMaps(tc.base, tc.overrides)

			prefix := fmt.Sprintf(
				"%s\nMergeMaps(m1=%v, m2=%v)",
				packageName, tc.base, tc.overrides,
			)

			// THEN: the maps are merged correctly.
			if testErr := test.AssertMapEqual(
				t,
				result,
				tc.want,
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestCopyMap(t *testing.T) {
	// GIVEN: different byte strings.
	tests := []struct {
		name        string
		input, want map[string]string
	}{
		{
			name:  "empty map",
			input: map[string]string{},
			want:  map[string]string{},
		},
		{
			name: "non-empty map",
			input: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "non-empty map with same keys but differing case",
			input: map[string]string{
				"test": "123",
				"tESt": "bar",
			},
			want: map[string]string{
				"test": "123",
				"tESt": "bar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: CopyMap is called.
			result := CopyMap(tc.input)

			prefix := fmt.Sprintf(
				"%s\nCopyMap(%v)",
				packageName, tc.input,
			)

			// THEN: the map is copied correctly.
			if &result == &tc.want {
				t.Fatalf("%s map wasn't copied as they have the same addresses", prefix)
			}
			if testErr := test.AssertMapEqual(
				t,
				result,
				tc.input,
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestLowercaseKeys(t *testing.T) {
	// GIVEN: different byte strings.
	tests := []struct {
		name        string
		input, want map[string]string
	}{
		{
			name:  "nil map",
			input: nil,
			want:  map[string]string{},
		},
		{
			name:  "empty map",
			input: map[string]string{},
			want:  map[string]string{},
		},
		{
			name: "lower-cased map",
			input: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "lower-cased map with mixed-cased values",
			input: map[string]string{
				"test": "123",
				"foo":  "bAr",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bAr",
			},
		},
		{
			name: "upper-cased map",
			input: map[string]string{
				"TEST": "123",
				"FOO":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "upper-cased map with mixed-case values",
			input: map[string]string{
				"TEST": "123",
				"FOO":  "bAr",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bAr",
			},
		},
		{
			name: "mixed-case map",
			input: map[string]string{
				"tESt": "123",
				"Foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: LowercaseKeys is called.
			result := LowercaseKeys(tc.input)

			prefix := fmt.Sprintf(
				"%s\nLowercaseKeys(%v)",
				packageName, tc.input,
			)

			// THEN: the map remains nil if it was nil.
			if tc.input == nil {
				if result != nil {
					t.Fatalf(
						"%s input map was nil but got: %v",
						prefix, result,
					)
				}
				return
			}
			// THEN: the map keys are lower-cased correctly.
			if testErr := test.AssertMapEqual(
				t,
				result,
				tc.want,
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: the map is not nil.
			if result == nil {
				t.Fatalf("%s map is nil", prefix)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	// GIVEN: a string map.
	strMap := map[string]int{
		"a": 0,
		"b": 0,
		"c": 0,
		"d": 0,
		"e": 0,
		"f": 0,
		"g": 0,
		"h": 0,
		"i": 0,
		"j": 0,
		"k": 0,
		"l": 0,
		"m": 0,
		"n": 0,
		"o": 0,
		"p": 0,
		"q": 0,
		"r": 0,
		"s": 0,
		"t": 0,
		"u": 0,
		"v": 0,
		"w": 0,
		"x": 0,
		"z": 0,
	}

	// WHEN: SortedKeys is called on it.
	sorted := SortedKeys(strMap)

	// THEN: the keys of the map are returned alphabetically sorted.
	want := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
		"h",
		"i",
		"j",
		"k",
		"l",
		"m",
		"n",
		"o",
		"p",
		"q",
		"r",
		"s",
		"t",
		"u",
		"v",
		"w",
		"x",
		"z",
	}

	prefix := fmt.Sprintf("%s\nSortedKeys(%v)", packageName, strMap)

	for i := 1; i < 1000; i++ { // repeat due to random ordering.
		if testErr := test.AssertSlicesEqualFunc(
			t,
			sorted,
			want,
			func(a, b string) bool { return a == b },
			prefix,
			"",
		); testErr != nil {
			t.Error(testErr)
		}
	}
}

func TestSortedKeys_NilMap(t *testing.T) {
	// GIVEN: a nil map.
	var strMap map[string]int

	// WHEN: SortedKeys is called on it.
	sorted := SortedKeys(strMap)

	// THEN: the returned slice is empty.
	if got := len(sorted); got != 0 {
		t.Errorf(
			"%s\nSortedKeys() length mismatch\ngot:  %d\nwant: 0",
			packageName, got,
		)
	}
}

func TestFirstNonEmptyMap(t *testing.T) {
	// GIVEN: a bunch of maps.
	tests := []struct {
		name      string
		maps      []map[string]string
		wantIndex int
	}{
		{
			name:      "no maps",
			maps:      []map[string]string{},
			wantIndex: -1,
		},
		{
			name: "all empty maps",
			maps: []map[string]string{
				{},
				{},
				{},
			},
			wantIndex: -1,
		},
		{
			name: "1 non-empty map",
			maps: []map[string]string{
				{"foo": "bar"},
				{},
				{},
			},
			wantIndex: 0,
		},
		{
			name: "2 non-empty maps",
			maps: []map[string]string{
				{"foo": "bar"},
				{"baz": "qux", "qux": "baz"},
				{},
			},
			wantIndex: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: FirstNonEmptyMap is called.
			result := FirstNonEmptyMap(tc.maps...)

			prefix := fmt.Sprintf(
				"%s\nFirstNonEmptyMap(%v)",
				packageName, tc.maps,
			)

			// THEN: the first non-empty map is returned.
			if tc.wantIndex == -1 {
				if result != nil {
					t.Fatalf("%s map is nil", prefix)
				}
				return
			}
			wantMap := tc.maps[tc.wantIndex]
			if gotLen, wantLen := len(result), len(wantMap); gotLen != wantLen {
				t.Fatalf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}
			for i := range result {
				if got, want := result[i], wantMap[i]; got != want {
					t.Fatalf(
						"%s value mismatch at %q\ngot:  %q (%+v)\nwant: %q (%+v)",
						prefix, i,
						got, result,
						want, wantMap,
					)
				}
			}

			// AND: modifications to the slice returned are reflected in the original.
			key := "TestFirstNonEmptyMap"
			result[key] = "bar"
			if got, want := tc.maps[tc.wantIndex][key], result[key]; got != want {
				t.Errorf(
					"%s modified %q in returned slice, but original slice is unchanged\ngot:  %q (%+v)\nwant: %q (%+v)",
					prefix, key,
					got, tc.maps[tc.wantIndex],
					want, result,
				)
			}
		})
	}
}
