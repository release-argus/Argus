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

package test

import (
	"fmt"
	"testing"
)

func TestAssertCheckValuesWithErrorAndChanged(t *testing.T) {
	// GIVEN: A function to check values, and a function to generate the input struct.
	tests := []struct {
		name        string
		checkValues func() (error, bool)
		errRegex    string
		wantErr     bool
		wantChanged bool
	}{
		{
			name: "no error, unchanged",
			checkValues: func() (error, bool) {
				return nil, false
			},
			errRegex:    `^$`,
			wantChanged: false,
		},
		{
			name: "no error, changed",
			checkValues: func() (error, bool) {
				return nil, true
			},
			errRegex:    `^$`,
			wantChanged: true,
		},
		{
			name: "expected error, unchanged",
			checkValues: func() (error, bool) {
				return fmt.Errorf("error happened"), true
			},
			errRegex:    `^error happened`,
			wantChanged: true,
		},
		{
			name: "expected error, unchanged",
			checkValues: func() (error, bool) {
				return fmt.Errorf("error happened"), false
			},
			errRegex:    `^error happened`,
			wantChanged: false,
		},
		{
			name: "error line mismatch, unchanged",
			checkValues: func() (error, bool) {
				return fmt.Errorf("error happened"), false
			},
			errRegex: TrimYAML(`
				^error happened
				123`,
			),
			wantErr:     true,
			wantChanged: true,
		},
		{
			name: "expected error, unexpected changed",
			checkValues: func() (error, bool) {
				return fmt.Errorf("error happened"), false
			},
			errRegex:    `^error happened`,
			wantErr:     true,
			wantChanged: true,
		},
		{
			name: "error, unchanged",
			checkValues: func() (error, bool) {
				return fmt.Errorf("error happened"), false
			},
			errRegex:    `^foo`,
			wantErr:     true,
			wantChanged: false,
		},
		{
			name: "error, changed",
			checkValues: func() (error, bool) {
				return fmt.Errorf("error happened"), true
			},
			errRegex:    `^foo`,
			wantErr:     true,
			wantChanged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ft := &FakeT{}
			// WHEN: AssertCheckValuesWithErrorAndChanged is called.
			_, _ = AssertCheckValuesWithErrorAndChanged(
				ft,
				packageName,
				tc.errRegex,
				tc.wantChanged,
				tc.checkValues,
			)

			// THEN: It errors when expected.
			gotErr := len(ft.Errors) != 0
			if gotErr != tc.wantErr {
				t.Errorf(
					"%s\nAssertCheckValuesWithErrorAndChanged() checkValues didn't pass/fail as expected\ngot  fail: %v\nwant fail: %t",
					packageName, ft.Errors, tc.wantErr,
				)
			}
		})
	}
}

func TestAssertCheckValuesWithError(t *testing.T) {
	// GIVEN: A function to check values, and a function to generate the input struct.
	tests := []struct {
		name        string
		checkValues func() error
		errRegex    string
		wantErr     bool
	}{
		{
			name: "no error",
			checkValues: func() error {
				return nil
			},
			errRegex: `^$`,
		},
		{
			name: "expected error",
			checkValues: func() error {
				return fmt.Errorf("error happened")
			},
			errRegex: `^error happened`,
		},
		{
			name: "error",
			checkValues: func() error {
				return fmt.Errorf("error happened")
			},
			errRegex: `^foo`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ft := &FakeT{}
			// WHEN: AssertCheckValuesWithError is called.
			_ = AssertCheckValuesWithError(
				ft,
				packageName,
				tc.errRegex,
				tc.checkValues,
			)

			gotErr := len(ft.Errors) != 0

			// THEN: it errors when expected.
			if gotErr != tc.wantErr {
				t.Errorf(
					"%s\nAssertCheckValuesWithError() checkValues didn't pass/fail as expected\ngot  fail: %v\nwant fail: %t",
					packageName, ft.Errors, tc.wantErr,
				)
			}
		})
	}
}
