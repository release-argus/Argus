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
	"errors"
	"testing"
)

func TestAssertCheckValuesWithErrorAndChanged(t *testing.T) {
	tests := []struct {
		name        string
		checkValues func() (error, bool)
		errRegex    string
		wantErr     bool
		wantChanged bool
	}{
		{
			name: "no error/unchanged",
			checkValues: func() (error, bool) {
				return nil, false
			},
			errRegex:    `^$`,
			wantChanged: false,
		},
		{
			name: "no error/changed",
			checkValues: func() (error, bool) {
				return nil, true
			},
			errRegex:    `^$`,
			wantChanged: true,
		},
		{
			name: "expected error/changed",
			checkValues: func() (error, bool) {
				return errors.New("error happened"), true
			},
			errRegex:    `^error happened`,
			wantChanged: true,
		},
		{
			name: "expected error/unchanged",
			checkValues: func() (error, bool) {
				return errors.New("error happened"), false
			},
			errRegex:    `^error happened`,
			wantChanged: false,
		},
		{
			name: "error line count mismatch/unchanged",
			checkValues: func() (error, bool) {
				return errors.New("error happened"), false
			},
			errRegex: TrimYAML(`
				^error happened
				123`,
			),
			wantErr:     true,
			wantChanged: false,
		},
		{
			name: "expected error/unexpected changed",
			checkValues: func() (error, bool) {
				return errors.New("error happened"), false
			},
			errRegex:    `^error happened`,
			wantErr:     true,
			wantChanged: true,
		},
		{
			name: "error/unchanged",
			checkValues: func() (error, bool) {
				return errors.New("error happened"), false
			},
			errRegex:    `^foo`,
			wantErr:     true,
			wantChanged: false,
		},
		{
			name: "error/changed",
			checkValues: func() (error, bool) {
				return errors.New("error happened"), true
			},
			errRegex:    `^foo`,
			wantErr:     true,
			wantChanged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fakeT := &FakeT{}

			_, _ = AssertCheckValuesWithErrorAndChanged(
				fakeT,
				tc.errRegex,
				tc.wantChanged,
				tc.checkValues,
			)

			didFail := len(fakeT.Errors)+len(fakeT.Fatals) != 0
			if didFail != tc.wantErr {
				t.Errorf(
					"didFail mismatch\ngot:  %t (errors=%q, fatals=%q)\nwant: %t",
					didFail, fakeT.Errors, fakeT.Fatals, tc.wantErr,
				)
			}
		})
	}
}

func TestAssertCheckValuesWithError(t *testing.T) {
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
				return errors.New("error happened")
			},
			errRegex: `^error happened`,
		},
		{
			name: "error",
			checkValues: func() error {
				return errors.New("error happened")
			},
			errRegex: `^foo`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fakeT := &FakeT{}

			_ = AssertCheckValuesWithError(
				fakeT,
				tc.errRegex,
				tc.checkValues,
			)

			didFail := len(fakeT.Errors)+len(fakeT.Fatals) != 0
			if didFail != tc.wantErr {
				t.Errorf(
					"didFail mismatch\ngot:  %t (errors=%q, fatals=%q)\nwant: %t",
					didFail, fakeT.Errors, fakeT.Fatals, tc.wantErr,
				)
			}
		})
	}
}
