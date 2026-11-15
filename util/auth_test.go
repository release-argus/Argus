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
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	username := "test"
	password := "123"

	got := BasicAuth(username, password)

	// "test:123" in base64.
	want := "dGVzdDoxMjM="
	if got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestIsHashed(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "non-hashed string",
			input: "h__foo",
			want:  false,
		},
		{
			name:  "hashed string",
			input: fmt.Sprintf("h__%x", sha256.Sum256([]byte("foo"))),
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isHashed(tc.input)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  [32]byte
	}{
		{
			name:  "empty string",
			input: "",
			want:  sha256.Sum256([]byte("")),
		},
		{
			name:  "non-empty string",
			input: "foo",
			want:  sha256.Sum256([]byte("foo")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := hash(tc.input)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %x\nwant: %x", got, tc.want)
			}
		})
	}
}

func TestHashFromString(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "empty string",
			str:  "",
		},
		{
			name: "non-empty string",
			str:  "foobar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := sha256.Sum256([]byte(tc.str))
			input := fmt.Sprintf("h__%x", want)[3:]

			got := hashFromString(input)

			if !bytes.Equal(got, want[:]) {
				t.Errorf("mismatch\ngot:  %x\nwant: %x", got, want)
			}
		})
	}
}

func TestGetHash(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		alreadyHashed bool
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "non-empty string",
			input: "foo",
		},
		{
			name:          "hashed string",
			input:         fmt.Sprintf("h__%x", sha256.Sum256([]byte("foo"))),
			alreadyHashed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := tc.input
			if !tc.alreadyHashed {
				want = FmtHash(sha256.Sum256([]byte(tc.input)))
			}

			got := GetHash(tc.input)

			gotHash := FmtHash(got)
			if gotHash != want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", gotHash, want)
			}
		})
	}
}

func TestFmtHash(t *testing.T) {
	sum := sha256.Sum256([]byte("foo"))

	got := FmtHash(sum)

	want := fmt.Sprintf("h__%x", sum)
	if got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}
