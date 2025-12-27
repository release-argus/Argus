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

// Package shared provides shared functionality for Latest Version and Deployed Version lookups.
package shared

import (
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestInheritSecrets(t *testing.T) {
	// GIVEN a set of Headers and a list of secretRefs.
	tests := []struct {
		name         string
		h            Headers
		otherHeaders Headers
		secretRefs   []OldIntIndex
		want         Headers
	}{
		{
			name:         "no headers in either",
			h:            Headers{},
			otherHeaders: Headers{},
			secretRefs:   nil,
			want:         Headers{},
		},
		{
			name: "no secrets to inherit",
			h: Headers{
				{Key: "X-Test", Value: "Value1"},
				{Key: "X-Another", Value: "Value2"},
			},
			otherHeaders: Headers{
				{Key: "X-Secret", Value: "SecretValue"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: test.IntPtr(0)},
			},
			want: Headers{
				{Key: "X-Test", Value: "Value1"},
				{Key: "X-Another", Value: "Value2"},
			},
		},
		{
			name: "inherit secrets when valid reference",
			h: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
			otherHeaders: Headers{
				{Key: "X-Secret", Value: "SecretValue"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: test.IntPtr(0)},
			},
			want: Headers{
				{Key: "X-Test", Value: "SecretValue"},
			},
		},
		{
			name: "can't inherit secrets with nil secretRefs",
			h: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
			otherHeaders: Headers{
				{Key: "X-Secret", Value: "SecretValue"},
			},
			secretRefs: nil,
			want: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
		},
		{
			name: "cam\t inherit secrets with nil OldIndex",
			h: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
			otherHeaders: Headers{
				{Key: "X-Secret", Value: "SecretValue"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: nil},
			},
			want: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
		},
		{
			name: "can't inherit secrets when OldIndex out of range",
			h: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
			otherHeaders: Headers{
				{Key: "X-Secret", Value: "SecretValue"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: test.IntPtr(2)},
			},
			want: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
		},
		{
			name: "inherit multiple secrets",
			h: Headers{
				{Key: "X-First", Value: util.SecretValue},
				{Key: "X-Second", Value: util.SecretValue},
				{Key: "X-Third", Value: "NotASecret"},
			},
			otherHeaders: Headers{
				{Key: "X-Secret1", Value: "SecretValue1"},
				{Key: "X-Secret2", Value: "SecretValue2"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: test.IntPtr(0)},
				{OldIndex: test.IntPtr(1)},
				{OldIndex: test.IntPtr(2)},
			},
			want: Headers{
				{Key: "X-First", Value: "SecretValue1"},
				{Key: "X-Second", Value: "SecretValue2"},
				{Key: "X-Third", Value: "NotASecret"},
			},
		},
		{
			name: "inherit multiple secrets in different order",
			h: Headers{
				{Key: "X-First", Value: util.SecretValue},
				{Key: "X-Second", Value: util.SecretValue},
				{Key: "X-Third", Value: util.SecretValue},
			},
			otherHeaders: Headers{
				{Key: "X-Secret1", Value: "SecretValue1"},
				{Key: "X-Secret2", Value: "SecretValue2"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: test.IntPtr(1)},
				{OldIndex: test.IntPtr(0)},
			},
			want: Headers{
				{Key: "X-First", Value: "SecretValue2"},
				{Key: "X-Second", Value: "SecretValue1"},
				{Key: "X-Third", Value: util.SecretValue},
			},
		},
		{
			name: "extra secretRefs ignored",
			h: Headers{
				{Key: "X-First", Value: util.SecretValue},
			},
			otherHeaders: Headers{
				{Key: "X-Secret1", Value: "SecretValue1"},
				{Key: "X-Secret2", Value: "SecretValue2"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: test.IntPtr(0)},
				{OldIndex: test.IntPtr(1)},
			},
			want: Headers{
				{Key: "X-First", Value: "SecretValue1"},
			},
		},
		{
			name: "empty secretRefs and otherHeaders",
			h: Headers{
				{Key: "X-First", Value: util.SecretValue},
			},
			otherHeaders: Headers{},
			secretRefs:   nil,
			want: Headers{
				{Key: "X-First", Value: util.SecretValue},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := tc.h

			// WHEN InheritSecrets is called.
			h.InheritSecrets(tc.otherHeaders, tc.secretRefs)

			// THEN the length is as expected.
			if len(h) != len(tc.want) {
				t.Fatalf("got %v, want %v", h, tc.want)
			}
			// AND the headers are as expected.
			for i := range h {
				if h[i] != tc.want[i] {
					t.Errorf("%s\nindex %d: got %v, want %v",
						packageName, i, h[i], tc.want[i])
				}
			}
		})
	}
}
