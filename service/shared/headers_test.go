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

// Package shared provides shared functionality for Latest Version and Deployed Version lookups.
package shared

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestHeaders_Copy(t *testing.T) {
	tests := []struct {
		name    string
		headers *Headers
	}{
		{
			name:    "nil",
			headers: nil,
		},
		{
			name:    "no headers",
			headers: &Headers{},
		},
		{
			name: "one header",
			headers: &Headers{
				{Key: "X-Test", Value: "Value1"},
			},
		},
		{
			name: "two headers",
			headers: &Headers{
				{Key: "X-Test", Value: "Value1"},
				{Key: "X-Another", Value: "Value2"},
			},
		},
		{
			name: "three headers",
			headers: &Headers{
				{Key: "X-Test", Value: "Value1"},
				{Key: "X-Another", Value: "Value2"},
				{Key: "X-Third", Value: "Value3"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.headers.Copy()

			if tc.headers == nil || len(*tc.headers) == 0 {
				if got != nil {
					t.Error("Headers.Copy() mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}

			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				*tc.headers,
				func(a, b Header) bool { return a.Key == b.Key && a.Value == b.Value },
				"Headers",
			); testErr != nil {
				t.Fatal(testErr)
			}

			if &got[0] == &(*tc.headers)[0] {
				t.Errorf("copy aliases the original\ngot:  %p\nwant: NOT %p", got, *tc.headers)
			}
		})
	}
}

func TestHeaders_InheritSecrets(t *testing.T) {
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
				{OldIndex: new(0)},
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
				{OldIndex: new(0)},
			},
			want: Headers{
				{Key: "X-Test", Value: "SecretValue"},
			},
		},
		{
			name: "don't inherit secrets with nil secretRefs",
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
			name: "don't inherit secrets with nil OldIndex",
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
			name: "don't inherit secrets when OldIndex out of range",
			h: Headers{
				{Key: "X-Test", Value: util.SecretValue},
			},
			otherHeaders: Headers{
				{Key: "X-Secret", Value: "SecretValue"},
			},
			secretRefs: []OldIntIndex{
				{OldIndex: new(2)},
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
				{OldIndex: new(0)},
				{OldIndex: new(1)},
				{OldIndex: new(2)},
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
				{OldIndex: new(1)},
				{OldIndex: new(0)},
			},
			want: Headers{
				{Key: "X-First", Value: "SecretValue2"},
				{Key: "X-Second", Value: "SecretValue1"},
				// No secretRef at index 2, so the loop stops.
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
				{OldIndex: new(0)},
				{OldIndex: new(1)},
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

			h.InheritSecrets(tc.otherHeaders, tc.secretRefs)

			if testErr := test.AssertSlicesEqualFunc(
				t,
				h,
				tc.want,
				func(a, b Header) bool { return a.Key == b.Key && a.Value == b.Value },
				"Headers",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
