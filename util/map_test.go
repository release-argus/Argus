// Copyright [2024] [Argus]
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

import "testing"

func TestInitMap(t *testing.T) {
	// GIVEN a map
	tests := map[string]struct {
		input map[string]string
	}{
		"nil map": {
			input: nil,
		},
		"empty map": {
			input: map[string]string{},
		},
		"non-empty map": {
			input: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"non-empty map with same keys but differing case": {
			input: map[string]string{
				"test": "123",
				"tESt": "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			had := CopyMap(tc.input)

			// WHEN InitMap is called
			InitMap(&tc.input)

			// THEN the map is initialised correctly
			if tc.input == nil {
				t.Fatalf("map is nil")
			}
			// AND any values inside haven't changed
			if len(tc.input) != len(had) {
				t.Fatalf("want: %v\ngot:  %v",
					had, tc.input)
			}
			for i := range tc.input {
				if tc.input[i] != had[i] {
					t.Fatalf("want: %v\ngot:  %v",
						had, tc.input)
				}
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	// GIVEN two maps and a list of fields that may contain secrets
	tests := map[string]struct {
		base, overrides, want map[string]string
		fields                []string
	}{
		"empty maps": {
			base:      map[string]string{},
			overrides: map[string]string{},
			want:      map[string]string{},
		},
		"nil maps": {
			base:      nil,
			overrides: nil,
			want:      map[string]string{},
		},
		"empty base map": {
			base: map[string]string{},
			overrides: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"nil base map": {
			base: nil,
			overrides: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"empty overrides map": {
			base: map[string]string{
				"test": "123",
				"foo":  "bar"},
			overrides: map[string]string{},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"nil overrides map": {
			base: map[string]string{
				"test": "123",
				"foo":  "bar"},
			overrides: nil,
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"non-empty maps": {
			base: map[string]string{
				"test":      "123",
				"foo":       "bar",
				"bish":      "bash",
				"something": "else"},
			overrides: map[string]string{
				"test": "456",
				"foo":  "baz",
				"bish": ""},
			want: map[string]string{
				"test":      "456",
				"foo":       "baz",
				"bish":      "",
				"something": "else"},
		},
		"ref secret in base map": {
			base: map[string]string{
				"test": "123"},
			overrides: map[string]string{
				"test": SecretValue},
			want: map[string]string{
				"test": "123"},
			fields: []string{"test"},
		},
		"ref secret in base map, secret not found/empty": {
			base: map[string]string{
				"foo": ""},
			overrides: map[string]string{
				"foo":  SecretValue,
				"test": SecretValue},
			want: map[string]string{
				"foo":  SecretValue,
				"test": SecretValue},
			fields: []string{"foo", "test"},
		},
		"secret not in fields": {
			base: map[string]string{
				"test": "123",
				"foo":  "bar"},
			overrides: map[string]string{
				"test": SecretValue,
				"foo":  SecretValue},
			want: map[string]string{
				"foo":  SecretValue,
				"test": SecretValue},
			fields: []string{"other"},
		},
		"non-empty maps with secrets": {
			base: map[string]string{
				"test":      "123",
				"foo":       "bar",
				"bish":      "bash",
				"something": "else",
				"nothing":   ""},
			overrides: map[string]string{
				"test":    "456",
				"foo":     SecretValue,
				"bish":    SecretValue,
				"nothing": SecretValue},
			want: map[string]string{
				"test":      "456",
				"foo":       SecretValue,
				"bish":      "bash",
				"something": "else",
				"nothing":   SecretValue},
			fields: []string{"test", "bish"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN MergeMaps is called
			got := MergeMaps(tc.base, tc.overrides, tc.fields)

			// THEN the maps are merged correctly
			if len(got) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestCopyMap(t *testing.T) {
	// GIVEN different byte strings
	tests := map[string]struct {
		input, want map[string]string
	}{
		"empty map": {
			input: map[string]string{},
			want:  map[string]string{},
		},
		"non-empty map": {
			input: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"non-empty map with same keys but differing case": {
			input: map[string]string{
				"test": "123",
				"tESt": "bar"},
			want: map[string]string{
				"test": "123",
				"tESt": "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CopyMap is called
			got := CopyMap(tc.input)

			// THEN the map is copied correctly
			if &got == &tc.want {
				t.Error("map wasn't copied, they have the same addresses")
			}
			if len(got) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestLowercaseStringStringMap(t *testing.T) {
	// GIVEN different byte strings
	tests := map[string]struct {
		input, want map[string]string
	}{
		"empty map": {
			input: map[string]string{},
			want:  map[string]string{},
		},
		"lower-cased map": {
			input: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"lower-cased map with mixed-cased values": {
			input: map[string]string{
				"test": "123",
				"foo":  "bAr"},
			want: map[string]string{
				"test": "123",
				"foo":  "bAr"},
		},
		"upper-cased map": {
			input: map[string]string{
				"TEST": "123",
				"FOO":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"upper-cased map with mixed-case values": {
			input: map[string]string{
				"TEST": "123",
				"FOO":  "bAr"},
			want: map[string]string{
				"test": "123",
				"foo":  "bAr"},
		},
		"mixed-case map": {
			input: map[string]string{
				"tESt": "123",
				"Foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := CopyMap(tc.input)

			// WHEN LowercaseStringStringMap is called
			LowercaseStringStringMap(&got)

			// THEN the map keys are lower-cased correctly
			if len(got) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	// GIVEN a string map
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
		"z": 0}

	// WHEN SortedKeys is called on it
	sorted := SortedKeys(strMap)

	// THEN the keys of the map are returned alphabetically sorted
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
		"z"}
	for i := 1; i < 1000; i++ { // repeat due to random ordering
		for i := range sorted {
			if sorted[i] != want[i] {
				t.Errorf("want index=%d to be %q, not %q\nwant: %v\ngot:  %v",
					i, want[i], sorted[i], want, sorted)
			}
		}
	}
}
