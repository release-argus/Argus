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

package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHashToken(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		sameAs string
		diffTo string
	}{
		{
			name:  "empty string",
			token: "",
		},
		{
			name:  "typical token",
			token: "argus_deadbeefcafef00d",
		},
		{
			name:  "unicode",
			token: "tökén-é",
		},
		{
			name:   "same input hashes equal",
			token:  "repeat",
			sameAs: "repeat",
		},
		{
			name:   "different input hashes differ",
			token:  "a",
			diffTo: "b",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := HashToken(tc.token)

			sum := sha256.Sum256([]byte(tc.token))
			if want := hex.EncodeToString(sum[:]); got != want {
				t.Errorf("hash mismatch\ngot:  %q\nwant: %q", got, want)
			}

			if len(got) != sha256.Size*2 {
				t.Errorf("len(hash) mismatch\ngot:  %d\nwant: %d", len(got), sha256.Size*2)
			}

			if tc.sameAs != "" {
				if want := HashToken(tc.sameAs); got != want {
					t.Errorf("sameAs: hash mismatch\ngot:  %q\nwant: %q", got, want)
				}
			}

			if tc.diffTo != "" {
				if other := HashToken(tc.diffTo); got == other {
					t.Errorf("diffTo: hash mismatch\ngot:  %q\nwant: NOT %q", got, other)
				}
			}
		})
	}
}
