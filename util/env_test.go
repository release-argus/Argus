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

package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestFirstNonDefaultWithEnv(t *testing.T) {
	envVarNameBase := "TEST_FIRST_NON_DEFAULT_WITH_ENV_"
	// GIVEN a bunch of comparables.
	tests := map[string]struct {
		env         map[string]string
		slice       []string
		allDefault  bool
		wantIndex   int
		wantText    string
		diffAddress bool
	}{
		"no vars": {
			slice:      []string{},
			allDefault: true,
		},
		"all default vars": {
			slice: []string{
				"",
				"",
				"",
				""},
			allDefault: true,
		},
		"1 non-default var": {
			slice: []string{
				"",
				"",
				"",
				"bar"},
			wantIndex: 3,
		},
		"1 non-default var (env var)": {
			env: map[string]string{
				"ONE": "bar"},
			slice: []string{
				"",
				"",
				"",
				fmt.Sprintf(`${%s_ONE}`,
					envVarNameBase)},
			wantIndex:   3,
			wantText:    "bar",
			diffAddress: true,
		},
		"1 non-default var (env var partial)": {
			env: map[string]string{
				"TWO": "bar"},
			slice: []string{
				"",
				"",
				"",
				fmt.Sprintf(`foo${%s_TWO}`,
					envVarNameBase)},
			wantIndex:   3,
			wantText:    "foobar",
			diffAddress: true,
		},
		"2 non-default vars": {
			slice: []string{
				"foo",
				"",
				"",
				"bar"},
			wantIndex: 0,
		},
		"2 non-default vars (empty env vars ignored)": {
			env: map[string]string{
				"THREE": "",
				"FOUR":  "bar"},
			slice: []string{
				fmt.Sprintf(`${%s_THREE}`,
					envVarNameBase),
				fmt.Sprintf(`${%s_UNSET}`,
					envVarNameBase),
				"",
				fmt.Sprintf(`${%s_FOUR}`,
					envVarNameBase)},
			wantIndex: 3,
			wantText: fmt.Sprintf(`${%s_UNSET}`,
				envVarNameBase),
			diffAddress: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				envVarName := fmt.Sprintf("%s_%s",
					envVarNameBase, k)
				os.Setenv(envVarName, v)
				t.Cleanup(func() { os.Unsetenv(envVarName) })
			}

			// WHEN FirstNonDefaultWithEnv is run on a slice of slice.
			got := FirstNonDefaultWithEnv(tc.slice...)

			// THEN the correct var (or "") is returned.
			if tc.allDefault {
				if got != "" {
					t.Fatalf("%s\nwant: non-empty\ngot:  %q\nfrom: %v",
						packageName, got, tc.slice)
				}
				return
			}
			// Addresses should be the same (unless we're using an env var).
			if got != tc.slice[tc.wantIndex] &&
				!tc.diffAddress {
				t.Errorf("%s\naddress mismatch\nwant: %v\ngot:  %v",
					packageName, tc.slice[tc.wantIndex], got)
				// Addresses should only be the same.
			} else if got == tc.slice[tc.wantIndex] {
				// If we're using an env var.
				if tc.diffAddress {
					t.Errorf("%s\naddresses of pointers should differ (%v, %v)",
						packageName, tc.slice[tc.wantIndex], got)
				}
				// Should have what the env var is set to.
			} else if got != tc.wantText {
				t.Errorf("%s\nvalue mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantText, got)
			}
		})
	}
}

func TestEvalEnvVars(t *testing.T) {
	envVarNameBase := "TEST_EVAL_ENV_VARS_"
	// GIVEN a string.
	tests := map[string]struct {
		input string
		env   map[string]string
		want  string
	}{
		"no env vars": {
			input: "hello there ${not an env var}",
			want:  "hello there ${not an env var}",
		},
		"1 env var": {
			env: map[string]string{
				"ONE": "bar"},
			input: fmt.Sprintf(`hello there ${%s_ONE}`,
				envVarNameBase),
			want: "hello there bar",
		},
		"2 env vars": {
			env: map[string]string{
				"TWO":   "bar",
				"THREE": "baz"},
			input: fmt.Sprintf(`hello there ${%s_TWO} ${%s_THREE}`,
				envVarNameBase, envVarNameBase),
			want: "hello there bar baz",
		},
		"unset env var": {
			input: fmt.Sprintf(`hello there ${%s_UNSET}`,
				envVarNameBase),
			want: fmt.Sprintf(`hello there ${%s_UNSET}`,
				envVarNameBase),
		},
		"empty env var": {
			env: map[string]string{
				"FOUR": ""},
			input: fmt.Sprintf(`hello there ${%s_FOUR}`,
				envVarNameBase),
			want: "hello there ",
		},
		"nested env vars not evaluated": {
			env: map[string]string{
				"FIVE": "bar",
				"SIX": fmt.Sprintf(`${%s_SEVEN}`,
					envVarNameBase),
				"SEVEN": "qux"},
			input: fmt.Sprintf(`hello there ${%s_FIVE} ${%s_SIX}`,
				envVarNameBase, envVarNameBase),
			want: fmt.Sprintf(`hello there bar ${%s_SEVEN}`,
				envVarNameBase),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				envVarName := fmt.Sprintf("%s_%s",
					envVarNameBase, k)
				os.Setenv(envVarName, v)
				t.Cleanup(func() { os.Unsetenv(envVarName) })
			}

			// WHEN EvalEnvVars is called.
			got := EvalEnvVars(tc.input)

			// THEN the string is evaluated correctly.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestExpandEnvVariables(t *testing.T) {
	// GIVEN an env var that may or may not exist.
	envVarNameBase := "TEST_ENV_REPLACE_FUNC_"
	tests := map[string]struct {
		envVarName  string
		envVarValue *string
		want        string
	}{
		"undefined env var": {
			envVarName: fmt.Sprintf("%s_%s",
				envVarNameBase, "UNDEFINED"),
			want: fmt.Sprintf("${%s_%s}",
				envVarNameBase, "UNDEFINED"),
		},
		"empty env var": {
			envVarName: fmt.Sprintf("%s_%s",
				envVarNameBase, "EMPTY"),
			envVarValue: test.StringPtr(""),
			want:        "",
		},
		"non-empty env var": {
			envVarName: fmt.Sprintf("%s_%s",
				envVarNameBase, "NON_EMPTY"),
			envVarValue: test.StringPtr("bar"),
			want:        "bar",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.envVarValue != nil {
				os.Setenv(tc.envVarName, *tc.envVarValue)
				t.Cleanup(func() { os.Unsetenv(tc.envVarName) })
			}

			// WHEN expandEnvVariables is called.
			got := expandEnvVariables(
				fmt.Sprintf("${%s}", tc.envVarName))

			// THEN the string is evaluated correctly.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}
