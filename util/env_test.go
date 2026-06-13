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

package util

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func prefixMapKeys(m map[string]string, prefix string) map[string]string {
	prefixed := make(map[string]string, len(m))
	for k, v := range m {
		prefixed[fmt.Sprintf("%s_%s", prefix, k)] = v
	}
	return prefixed
}

func TestFirstNonDefaultWithEnv(t *testing.T) {
	envVarNameBase := "TEST_FIRST_NON_DEFAULT_WITH_ENV_"
	// GIVEN: a bunch of comparables.
	tests := []struct {
		name       string
		env        map[string]string
		slice      []string
		allDefault bool
		wantIndex  int
		wantText   string
		diffValue  bool
	}{
		{
			name:       "no vars",
			slice:      []string{},
			allDefault: true,
		},
		{
			name: "all default vars",
			slice: []string{
				"",
				"",
				"",
				"",
			},
			allDefault: true,
		},
		{
			name: "1 non-default var",
			slice: []string{
				"",
				"",
				"",
				"bar",
			},
			wantIndex: 3,
		},
		{
			name: "1 non-default var (env var)",
			env: map[string]string{
				"ONE": "bar",
			},
			slice: []string{
				"",
				"",
				"",
				fmt.Sprintf(`${%s_ONE}`, envVarNameBase),
			},
			wantIndex: 3,
			wantText:  "bar",
			diffValue: true,
		},
		{
			name: "1 non-default var (env var partial)",
			env: map[string]string{
				"TWO": "bar",
			},
			slice: []string{
				"",
				"",
				"",
				fmt.Sprintf(`foo${%s_TWO}`, envVarNameBase),
			},
			wantIndex: 3,
			wantText:  "foobar",
			diffValue: true,
		},
		{
			name: "2 non-default vars",
			slice: []string{
				"foo",
				"",
				"",
				"bar",
			},
			wantIndex: 0,
		},
		{
			name: "2 non-default vars (empty env vars ignored)",
			env: map[string]string{
				"THREE": "",
				"FOUR":  "bar",
			},
			slice: []string{
				fmt.Sprintf(`${%s_THREE}`, envVarNameBase),
				fmt.Sprintf(`${%s_UNSET}`, envVarNameBase),
				"",
				fmt.Sprintf(`${%s_FOUR}`, envVarNameBase),
			},
			wantIndex: 3,
			wantText:  fmt.Sprintf(`${%s_UNSET}`, envVarNameBase),
			diffValue: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: the env vars are set.
			tc.env = prefixMapKeys(tc.env, envVarNameBase)
			test.SetEnv(t, tc.env)

			// WHEN: FirstNonDefaultWithEnv is run on a slice of slice.
			got := FirstNonDefaultWithEnv(tc.slice...)

			prefix := fmt.Sprintf(
				"%s\nFirstNonDefaultWithEnv(%v)",
				packageName, tc.slice,
			)

			// THEN: the correct var (or "") is returned.
			if tc.allDefault {
				if got != "" {
					t.Fatalf("%s mismatch\ngot:  %q\nwant: \"\"", packageName, got)
				}
				return
			}
			// Values differ when they contain an env var.
			if tc.diffValue {
				// Value unchanged.
				if got == tc.slice[tc.wantIndex] {
					t.Errorf(
						"%s values should differ (got %v, slice %v)",
						prefix, got, tc.slice[tc.wantIndex],
					)

					// Value isn't changed as expected.
				} else if got != tc.wantText {
					t.Errorf(
						"%s value mismatch\ngot:  %q\nwant: %q",
						prefix, got, tc.wantText,
					)
				}

				return
			}
			// Values should match slice
			if got != tc.slice[tc.wantIndex] {
				t.Errorf(
					"%s value mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.slice[tc.wantIndex],
				)
			}
		})
	}
}

func TestEvalEnvVars(t *testing.T) {
	envVarNameBase := "TEST_EVAL_ENV_VARS_"
	// GIVEN: a string.
	tests := []struct {
		name  string
		input string
		env   map[string]string
		want  string
	}{
		{
			name:  "no env var '${...}",
			input: "hello there",
			want:  "hello there",
		},
		{
			name:  "no env vars",
			input: "hello there ${not an env var}",
			want:  "hello there ${not an env var}",
		},
		{
			name: "1 env var",
			env: map[string]string{
				"ONE": "bar",
			},
			input: fmt.Sprintf(`hello there ${%s_ONE}`, envVarNameBase),
			want:  "hello there bar",
		},
		{
			name: "2 env vars",
			env: map[string]string{
				"TWO":   "bar",
				"THREE": "baz",
			},
			input: fmt.Sprintf(
				`hello there ${%s_TWO} ${%s_THREE}`,
				envVarNameBase, envVarNameBase,
			),
			want: "hello there bar baz",
		},
		{
			name:  "unset env var",
			input: fmt.Sprintf(`hello there ${%s_UNSET}`, envVarNameBase),
			want:  fmt.Sprintf(`hello there ${%s_UNSET}`, envVarNameBase),
		},
		{
			name: "empty env var",
			env: map[string]string{
				"FOUR": "",
			},
			input: fmt.Sprintf(`hello there ${%s_FOUR}`, envVarNameBase),
			want:  "hello there ",
		},
		{
			name: "nested env vars not evaluated",
			env: map[string]string{
				"FIVE":  "bar",
				"SIX":   fmt.Sprintf(`${%s_SEVEN}`, envVarNameBase),
				"SEVEN": "qux",
			},
			input: fmt.Sprintf(
				`hello there ${%s_FIVE} ${%s_SIX}`,
				envVarNameBase, envVarNameBase,
			),
			want: fmt.Sprintf(`hello there bar ${%s_SEVEN}`, envVarNameBase),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: the env vars are set.
			tc.env = prefixMapKeys(tc.env, envVarNameBase)
			test.SetEnv(t, tc.env)

			// WHEN: EvalEnvVars is called.
			got := EvalEnvVars(tc.input)

			// THEN: the string is evaluated correctly.
			if got != tc.want {
				t.Errorf(
					"%s\nEvalEnvVars(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.input,
					got, tc.want,
				)
			}
		})
	}
}

func TestTryExpandEnv(t *testing.T) {
	// GIVEN: environment variables and a string that may contain environment variables.
	tests := []struct {
		name     string
		input    string
		env      map[string]string
		expected *string
	}{
		{
			name:     "no environment variables",
			input:    "plain_text",
			env:      nil,
			expected: nil,
		},
		{
			name:  "single environment variable",
			input: "${TEST_EXPAND_ENV_FOO_1}",
			env: map[string]string{
				"TEST_EXPAND_ENV_FOO_1": "bar",
			},
			expected: test.Ptr("bar"),
		},
		{
			name:  "environment variable requires curly brackets",
			input: "$TEST_EXPAND_ENV_FOO_2",
			env: map[string]string{
				"TEST_EXPAND_ENV_FOO_2": "bar",
			},
			expected: nil,
		},
		{
			name:  "multiple environment variables",
			input: "${TEST_EXPAND_ENV_FOO_3}-${TEST_EXPAND_ENV_BAR_1}",
			env: map[string]string{
				"TEST_EXPAND_ENV_FOO_3": "hello",
				"TEST_EXPAND_ENV_BAR_1": "world",
			},
			expected: test.Ptr("hello-world"),
		},
		{
			name:     "environment variable not set",
			input:    "${TEST_EXPAND_ENV_FOO_4}",
			env:      nil,
			expected: nil,
		},
		{
			name:  "mixed text and environment variables",
			input: "prefix-${TEST_EXPAND_ENV_FOO_5}-suffix",
			env: map[string]string{
				"TEST_EXPAND_ENV_FOO_5": "value",
			},
			expected: test.Ptr("prefix-value-suffix"),
		},
		{
			name:     "no expansion needed",
			input:    "NO_EXPANSION_NEEDED",
			env:      nil,
			expected: nil,
		},
		{
			name:     "empty data string",
			input:    "",
			env:      nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: the env vars are set.
			test.SetEnv(t, tc.env)

			// WHEN: TryExpandEnv is called on the string.
			result := TryExpandEnv(tc.input)

			// THEN: the result is as expected.
			want := DerefOr(tc.expected, "<nil>")
			got := DerefOr(result, "<nil>")
			if want != got {
				t.Errorf(
					"%s\nTryExpandEnv(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.input,
					got, want,
				)
			}
		})
	}
}

func TestExpandEnvVariables(t *testing.T) {
	// GIVEN: an env var that may or may not exist.
	envVarNameBase := "TEST_ENV_REPLACE_FUNC_"
	tests := []struct {
		name        string
		envVarName  string
		envVarValue *string
		want        string
	}{
		{
			name:       "undefined env var",
			envVarName: "UNDEFINED",
			want: fmt.Sprintf(
				"${%s%s}",
				envVarNameBase, "UNDEFINED",
			),
		},
		{
			name:        "empty env var",
			envVarName:  "EMPTY",
			envVarValue: test.Ptr(""),
			want:        "",
		},
		{
			name:        "non-empty env var",
			envVarName:  "NON_EMPTY",
			envVarValue: test.Ptr("bar"),
			want:        "bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: the env vars are set.
			tc.envVarName = envVarNameBase + tc.envVarName
			if tc.envVarValue != nil {
				env := map[string]string{tc.envVarName: *tc.envVarValue}
				test.SetEnv(t, env)
			}

			// WHEN: expandEnvVariables is called.
			got := expandEnvVariables(fmt.Sprintf("${%s}", tc.envVarName))

			// THEN: the string is evaluated correctly.
			if got != tc.want {
				t.Errorf(
					"%s\nexpandEnvVariables(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.envVarName,
					got, tc.want,
				)
			}
		})
	}
}
