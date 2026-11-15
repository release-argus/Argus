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
	"slices"
	"testing"
)

func TestIsolateEnv(t *testing.T) {
	initial := map[string]string{
		"ARGUS_ISOLATE_TEST__IN": "in",
		"ARGUSISOLATE_TEST__IN":  "outside",
		"ARGUS_TEST__IN":         "exempt",
		"ISOLATE_TEST__OUTSIDE":  "outside",
	}
	tests := []struct {
		name       string
		vars       map[string]string
		wantDuring map[string]*string
	}{
		{
			name: "unsets ARGUS_ vars, keeps ARGUS_TEST_ and other vars",
			wantDuring: map[string]*string{
				"ARGUS_ISOLATE_TEST__IN": nil,
				"ARGUSISOLATE_TEST__IN":  new("outside"),
				"ARGUS_TEST__IN":         new("exempt"),
				"ISOLATE_TEST__OUTSIDE":  new("outside"),
			},
		},
		{
			name: "sets vars after unsetting",
			vars: map[string]string{
				"ARGUS_ISOLATE_TEST__IN":  "set",
				"ARGUS_ISOLATE_TEST__NEW": "new",
				"ARGUS_TEST__IN":          "replaced",
				"ISOLATE_TEST__OUTSIDE":   "replaced",
			},
			wantDuring: map[string]*string{
				"ARGUS_ISOLATE_TEST__IN":  new("set"),
				"ARGUS_ISOLATE_TEST__NEW": new("new"),
				"ARGUSISOLATE_TEST__IN":   new("outside"),
				"ARGUS_TEST__IN":          new("replaced"),
				"ISOLATE_TEST__OUTSIDE":   new("replaced"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			for k, v := range initial {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}
			before := SplitEnvVars(os.Environ())

			t.Run("isolated", func(t *testing.T) {
				// Serial: sets env vars.

				IsolateEnv(t, tc.vars)
				// Code under test may set vars itself.
				_ = os.Setenv("ISOLATE_TEST__SET_DURING", "during")

				for k, want := range tc.wantDuring {
					got, ok := os.LookupEnv(k)
					switch {
					case want == nil && ok:
						t.Errorf(
							"os.LookupEnv(%q) mismatch\ngot:  present (%q)\nwant: absent",
							k, got,
						)
					case want != nil && (!ok || got != *want):
						t.Errorf(
							"os.LookupEnv(%q) mismatch\ngot:  %q (%s)\nwant: %q",
							k, got, Presence(ok), *want,
						)
					}
				}
			})

			// Only key names are reported, since the environment holds sourced secrets.
			after := SplitEnvVars(os.Environ())
			var changed []string
			for k, v := range before {
				if got, ok := after[k]; !ok || got != v {
					changed = append(changed, k)
				}
			}
			for k := range after {
				if _, ok := before[k]; !ok {
					changed = append(changed, k)
				}
			}
			slices.Sort(changed)
			if len(changed) != 0 {
				t.Errorf("os.Environ() changed keys mismatch\ngot:  %q\nwant: []", changed)
			}
		})
	}
}

func TestSplitEnvVars(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		want map[string]string
	}{
		{
			name: "simple",
			env: []string{
				"FOO=bar",
				"BAZ=qux",
			},
			want: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
		},
		{
			name: "empty value",
			env:  []string{"FOO="},
			want: map[string]string{"FOO": ""},
		},
		{
			name: "no equals sign",
			env:  []string{"FOO"},
			want: map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SplitEnvVars(tc.env)

			if testErr := AssertMapEqual(
				t,
				got,
				tc.want,
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
