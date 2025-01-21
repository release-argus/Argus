// Copyright [2025] [Argus]
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

	"github.com/release-argus/Argus/test"
)

func TestFirstNonNilPtr(t *testing.T) {
	// GIVEN a bunch of pointers
	tests := map[string]struct {
		pointers  []*string
		allNil    bool
		wantIndex int
	}{
		"no pointers": {
			pointers: []*string{},
			allNil:   true,
		},
		"all nil pointers": {
			pointers: []*string{
				nil,
				nil,
				nil,
				nil},
			allNil: true,
		},
		"1 non-nil pointer": {
			pointers: []*string{
				nil,
				nil,
				nil,
				test.StringPtr("bar")},
			wantIndex: 3,
		},
		"2 non-nil pointers": {
			pointers: []*string{
				test.StringPtr("foo"),
				nil,
				nil,
				test.StringPtr("bar")},
			wantIndex: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN FirstNonNilPtr is run on a slice of pointers
			got := FirstNonNilPtr(tc.pointers...)

			// THEN the correct pointer (or nil) is returned
			if tc.allNil {
				if got != nil {
					t.Fatalf("got:  %v\nfrom: %v",
						got, tc.pointers)
				}
				return
			}
			if got != tc.pointers[tc.wantIndex] {
				t.Errorf("want: %v\ngot:  %v",
					tc.pointers[tc.wantIndex], got)
			}
		})
	}
}

func TestFirstNonDefault(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		slice      []string
		allDefault bool
		wantIndex  int
	}{
		"no vars": {
			slice:      []string{},
			allDefault: true,
		},
		"all default vars": {
			slice: []string{
				"",
				"",
				"",
				""},
			allDefault: true,
		},
		"1 non-default var": {
			slice: []string{
				"",
				"",
				"",
				"bar"},
			wantIndex: 3,
		},
		"2 non-default vars": {
			slice: []string{
				"foo",
				"",
				"",
				"bar"},
			wantIndex: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN FirstNonDefault is run on a slice of slice
			got := FirstNonDefault(tc.slice...)

			// THEN the correct var (or "") is returned
			if tc.allDefault {
				if got != "" {
					t.Fatalf("got:  %v\nfrom: %v",
						got, tc.slice)
				}
				return
			}
			if got != tc.slice[tc.wantIndex] {
				t.Errorf("want: %v\ngot:  %v",
					tc.slice[tc.wantIndex], got)
			}
		})
	}
}

func TestAreSlicesEqual(t *testing.T) {
	// GIVEN different slices.
	tests := map[string]struct {
		slice1, slice2 []string
		want           bool
	}{
		"both empty": {
			slice1: []string{},
			slice2: []string{},
			want:   true,
		},
		"one empty": {
			slice1: []string{"foo"},
			slice2: []string{},
			want:   false,
		},
		"same length, same elements": {
			slice1: []string{"foo", "bar"},
			slice2: []string{"foo", "bar"},
			want:   true,
		},
		"different elements": {
			slice1: []string{"foo", "bar"},
			slice2: []string{"bar", "foo"},
			want:   false,
		},
		"different lengths": {
			slice1: []string{"foo", "bar"},
			slice2: []string{"foo"},
			want:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN AreSlicesEqual is called.
			got := AreSlicesEqual(tc.slice1, tc.slice2)

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("want: %v\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

func TestNormaliseNewlines(t *testing.T) {
	// GIVEN different byte strings
	tests := map[string]struct {
		input, want []byte
	}{
		"string with no newlines": {
			input: []byte("hello there"),
			want:  []byte("hello there")},
		"string with linux newlines": {
			input: []byte("hello\nthere"),
			want:  []byte("hello\nthere")},
		"string with multiple linux newlines": {
			input: []byte("hello\nthere\n"),
			want:  []byte("hello\nthere\n")},
		"string with windows newlines": {
			input: []byte("hello\r\nthere"),
			want:  []byte("hello\nthere")},
		"string with multiple windows newlines": {
			input: []byte("hello\r\nthere\r\n"),
			want:  []byte("hello\nthere\n")},
		"string with mac newlines": {
			input: []byte("hello\r\nthere"),
			want:  []byte("hello\nthere")},
		"string with multiple mac newlines": {
			input: []byte("hello\r\nthere\r\n"),
			want:  []byte("hello\nthere\n")},
		"string with multiple mac and windows newlines": {
			input: []byte("\rhello\r\nthere\r\n. hi\r"),
			want:  []byte("\nhello\nthere\n. hi\n")},
		"string with multiple mac, windows and linux newlines": {
			input: []byte("\rhello\r\nthere\r\n. hi\r. foo\nbar\n"),
			want:  []byte("\nhello\nthere\n. hi\n. foo\nbar\n")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN NormaliseNewlines is called
			got := NormaliseNewlines(tc.input)

			// THEN the newlines are normalised correctly
			if string(got) != string(tc.want) {
				t.Errorf("want: %q\ngot:  %q",
					string(tc.want), string(got))
			}
		})
	}
}
