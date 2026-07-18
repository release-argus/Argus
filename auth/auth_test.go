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
	"fmt"
	"testing"
)

func TestHashToken(t *testing.T) {
	// GIVEN: tokens to hash.
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

			prefix := fmt.Sprintf(
				"%s\nHashToken(%q)",
				packageName, tc.token,
			)

			// WHEN: the token is hashed.
			got := HashToken(tc.token)

			// THEN: it is the hex-encoded SHA-256 of the token.
			sum := sha256.Sum256([]byte(tc.token))
			if want := hex.EncodeToString(sum[:]); got != want {
				t.Errorf(
					"%s hash mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: it is a fixed 64-character hex string (32 bytes).
			if len(got) != sha256.Size*2 {
				t.Errorf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, len(got), sha256.Size*2,
				)
			}

			// AND: hashing is deterministic for the same input.
			if tc.sameAs != "" && got != HashToken(tc.sameAs) {
				t.Errorf(
					"%s should hash equal to %q",
					prefix, tc.sameAs,
				)
			}

			// AND: distinct inputs hash to distinct values.
			if tc.diffTo != "" && got == HashToken(tc.diffTo) {
				t.Errorf(
					"%s should not hash equal to %q",
					prefix, tc.diffTo,
				)
			}
		})
	}
}
