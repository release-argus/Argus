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

// Package web provides a web-based lookup type.
package web

import (
	"testing"
)

func TestLookup_GetType(t *testing.T) {
	// GIVEN: a Lookup with a Type.
	tests := []struct {
		name  string
		lType string
	}{
		{name: "empty", lType: ""},
		{name: "test", lType: "test"},
		{name: "x", lType: "x"},
		{name: "y", lType: "y"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &Lookup{}
			l.Type = tc.lType

			// WHEN: GetType is called.
			got := l.GetType()

			wantType := Type
			// THEN: the Type is returned.
			if got != wantType {
				t.Errorf(
					"%s\nLookup.GetType() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, wantType,
				)
			}
		})
	}
}

func TestLookup_AllowInvalidCerts(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		{
			name:             "root overrides all",
			want:             true,
			rootValue:        new(true),
			defaultValue:     new(false),
			hardDefaultValue: new(false),
		},
		{
			name:             "default overrides hardDefault",
			want:             true,
			defaultValue:     new(true),
			hardDefaultValue: new(false),
		},
		{
			name:             "hardDefault is last resort",
			want:             true,
			hardDefaultValue: new(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.AllowInvalidCerts = tc.rootValue
			lookup.typeDefaults.AllowInvalidCerts = tc.defaultValue
			lookup.typeHardDefaults.AllowInvalidCerts = tc.hardDefaultValue

			// WHEN: allowInvalidCerts is called.
			got := lookup.allowInvalidCerts()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.allowInvalidCerts() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}
