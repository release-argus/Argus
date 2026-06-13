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

package docker

import (
	"testing"
	"time"
)

// ####################
// # AUTH | UTILITIES #
// ####################

func TestIsUsable(t *testing.T) {
	now := time.Now().UTC()

	// GIVEN: a query token and expiration time.
	tests := []struct {
		name       string
		token      string
		validUntil time.Time
		want       bool
	}{
		{
			name:       "empty token",
			token:      "",
			validUntil: now.Add(1 * time.Hour),
			want:       false,
		},
		{
			name:       "expired token",
			token:      "abc123",
			validUntil: now.Add(-1 * time.Hour),
			want:       false,
		},
		{
			name:       "token expires <2 seconds",
			token:      "abc123",
			validUntil: now.Add(2*time.Second - time.Nanosecond),
			want:       false,
		},
		{
			name:       "~10s remaining returns token",
			token:      "abc123",
			validUntil: now.Add(10 * time.Second),
			want:       true,
		},
		{
			name:       "valid token far in-future returns token",
			token:      "longToken",
			validUntil: now.Add(24 * time.Hour),
			want:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: isUsable is called with it.
			got := isUsable(tc.token, tc.validUntil)

			// THEN: the true is returned if valid time meets requirements.
			if got != tc.want {
				t.Errorf(
					"%s\nisUsable(token=%q, validUntil=%q) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.token, tc.validUntil,
					got, tc.want,
				)
			}
		})
	}
}
