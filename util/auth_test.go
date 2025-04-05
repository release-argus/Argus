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
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	// GIVEN a username and password.
	username := "test"
	password := "123"

	// WHEN BasicAuth is called with this.
	got := BasicAuth(username, password)

	// THEN username:password is returned in base64.
	want := "dGVzdDoxMjM="
	if got != want {
		t.Errorf("%s\nFailed encoding\nwant: %q\ngot:  %q",
			packageName, want, got)
	}
}

func TestIsHashed(t *testing.T) {
	// GIVEN a string.
	tests := map[string]struct {
		input string
		want  bool
	}{
		"empty string": {
			input: "",
			want:  false,
		},
		"non-hashed string": {
			input: "h__foo",
			want:  false,
		},
		"hashed string": {
			input: fmt.Sprintf("h__%x", sha256.Sum256([]byte("foo"))),
			want:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN isHashed is called on it.
			got := isHashed(tc.input)

			// THEN the hash is detected correctly.
			if got != tc.want {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, tc.want, got)
			}
		})
	}
}

func TestHash(t *testing.T) {
	// GIVEN a string.
	tests := map[string]struct {
		input string
		want  [32]byte
	}{
		"empty string": {
			input: "",
			want:  sha256.Sum256([]byte("")),
		},
		"non-empty string": {
			input: "foo",
			want:  sha256.Sum256([]byte("foo")),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN hash is called on it.
			got := hash(tc.input)

			// THEN the string is hashed correctly.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestHashFromString(t *testing.T) {
	// GIVEN a string that contains a hash.
	tests := map[string]string{
		"empty string":     "",
		"non-empty string": "foobar",
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := sha256.Sum256([]byte(tc))
			input := fmt.Sprintf("h__%x", want)

			// WHEN hashFromString is called on it.
			got := hashFromString(input[3:])

			// THEN the string is hashed correctly.
			var got32 [32]byte
			copy(got32[:], got[:])
			if got32 != want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, want, got32)
			}
		})
	}
}

func TestGetHash(t *testing.T) {
	// GIVEN a string that may or may not be hashed.
	tests := map[string]struct {
		input         string
		alreadyHashed bool
	}{
		"empty string": {
			input: "",
		},
		"non-empty string": {
			input: "foo",
		},
		"hashed string": {
			input:         fmt.Sprintf("h__%x", sha256.Sum256([]byte("foo"))),
			alreadyHashed: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := tc.input
			if !tc.alreadyHashed {
				want = FmtHash(sha256.Sum256([]byte(tc.input)))
			}

			// WHEN GetHash is called on it.
			got := GetHash(tc.input)

			// THEN the string is hashed correctly.
			gotHash := FmtHash(got)
			if gotHash != want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, want, gotHash)
			}
		})
	}
}

func TestFmtHash(t *testing.T) {
	// GIVEN a hash.
	hash := sha256.Sum256([]byte("foo"))

	// WHEN FmtHash is called on it.
	got := FmtHash(hash)

	// THEN the hash is formatted correctly.
	want := fmt.Sprintf("h__%x", hash)
	if got != want {
		t.Errorf("%s\nwant: %q\ngot:  %q",
			packageName, want, got)
	}
}
