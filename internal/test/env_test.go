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
	"os"
	"testing"
)

func TestSetEnv(t *testing.T) {
	// GIVEN: a map of env vars to set.
	tests := []struct {
		name    string
		initial map[string]*string
		input   map[string]string
	}{
		{
			name: "restore unset",
			initial: map[string]*string{
				"ARGUS_TEST__FOO": nil,
			},
			input: map[string]string{
				"ARGUS_TEST__FOO": "bar",
			},
		},
		{
			name: "restore existing",
			initial: map[string]*string{
				"ARGUS_TEST__FOO": Ptr("orig"),
			},
			input: map[string]string{
				"ARGUS_TEST__FOO": "new",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// AND: the initial state of the env vars is set as specified.
			for k, v := range tc.initial {
				if v == nil {
					_ = os.Unsetenv(k)
				} else {
					_ = os.Setenv(k, *v)
				}
			}

			// THEN: they are restored to their initial state after the test.
			// t.Cleanup is LIFO, so queue this before the SetEnv cleanup.
			t.Cleanup(func() {
				for k, expected := range tc.initial {
					got, ok := os.LookupEnv(k)

					if expected == nil {
						if ok {
							t.Errorf(
								"%s\nafter SetEnv(): got %q, want %s to be unset,",
								packageName, got, k,
							)
						}
					} else {
						if !ok || got != *expected {
							t.Errorf(
								"%s\nafter SetEnv: got %q (exists=%v), want %s=%q",
								packageName, got, ok, k, *expected,
							)
						}
					}
				}
			})

			// WHEN: SetEnv is called with the input map.
			SetEnv(t, tc.input)

			// THEN: the env vars are set as expected.
			for k, expected := range tc.input {
				got, ok := os.LookupEnv(k)
				if !ok || got != expected {
					t.Fatalf(
						"%s\nduring SetEnv: got %q (exists=%v), want %s=%q",
						packageName, got, ok, k, expected,
					)
				}
			}
		})
	}
}
