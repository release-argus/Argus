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
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestEnsureMap(t *testing.T) {
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

			result := EnsureMap(tc.input)

			if result == nil {
				t.Fatal("result mismatch\ngot:  nil\nwant: non-nil")
			}

			if testErr := test.AssertMapEqual(
				t,
				result,
				had,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name         string
		m1, m2, want map[string]string
	}{
		{
			name: "empty maps",
			m1:   map[string]string{},
			m2:   map[string]string{},
			want: map[string]string{},
		},
		{
			name: "nil maps",
			m1:   nil,
			m2:   nil,
			want: map[string]string{},
		},
		{
			name: "empty m1 map",
			m1:   map[string]string{},
			m2: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "nil m1 map",
			m1:   nil,
			m2: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "empty m2 map",
			m1: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			m2: map[string]string{},
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "nil m2 map",
			m1: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			m2: nil,
			want: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
		},
		{
			name: "non-empty maps",
			m1: map[string]string{
				"test":      "123",
				"foo":       "bar",
				"bish":      "bash",
				"something": "else",
			},
			m2: map[string]string{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := MergeMaps(tc.m1, tc.m2)

			if testErr := test.AssertMapEqual(
				t,
				result,
				tc.want,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestCopyMap(t *testing.T) {
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

			result := CopyMap(tc.input)

			if testErr := test.AssertMapEqual(
				t,
				result,
				tc.want,
				"",
			); testErr != nil {
				t.Error(testErr)
			}

			key := "TestCopyMap"
			result[key] = "mutated"
			if got, found := tc.input[key]; found {
				t.Errorf("input[key] mismatch\ngot:  %q\nwant: absent", got)
			}
		})
	}
}

func TestLowercaseKeys(t *testing.T) {
	tests := []struct {
		name        string
		input, want map[string]string
	}{
		{
			name:  "nil map",
			input: nil,
			want:  nil,
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

			result := LowercaseKeys(tc.input)

			if tc.input == nil {
				if result != nil {
					t.Fatalf("result mismatch\ngot:  %v\nwant: nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("result mismatch\ngot:  nil\nwant: non-nil")
			}

			if testErr := test.AssertMapEqual(
				t,
				result,
				tc.want,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
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

	for i := range 1000 { // Repeat, as map iteration order is random.
		sorted := SortedKeys(strMap)

		if testErr := test.AssertSlicesEqualFunc(
			t,
			sorted,
			want,
			func(a, b string) bool { return a == b },
			"",
		); testErr != nil {
			t.Fatalf("attempt %d: %v", i, testErr)
		}
	}
}

func TestSortedKeys_NilMap(t *testing.T) {
	var strMap map[string]int

	sorted := SortedKeys(strMap)

	if got, want := len(sorted), 0; got != want {
		t.Errorf("length mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestFirstNonEmptyMap(t *testing.T) {
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

			result := FirstNonEmptyMap(tc.maps...)

			if tc.wantIndex == -1 {
				if result != nil {
					t.Fatalf("result mismatch\ngot:  %v\nwant: nil", result)
				}
				return
			}
			wantMap := tc.maps[tc.wantIndex]
			if testErr := test.AssertMapEqual(t, result, wantMap, ""); testErr != nil {
				t.Fatal(testErr)
			}

			// Returned, not copied, so writes reach the original.
			key := "TestFirstNonEmptyMap"
			result[key] = "bar"
			if got, want := tc.maps[tc.wantIndex][key], result[key]; got != want {
				t.Errorf("original[key] mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}
