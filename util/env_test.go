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

	tests := []struct {
		name       string
		env        map[string]string
		slice      []string
		allDefault bool
		wantIndex  int
		wantText   string
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
			name: "1 non-default var/value",
			slice: []string{
				"",
				"",
				"",
				"bar",
			},
			wantIndex: 3,
		},
		{
			name: "1 non-default var/env var",
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
		},
		{
			name: "1 non-default var/env var partial",
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
			name: "2 non-default vars/empty env vars ignored",
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
			wantIndex: 1, // An unset var stays as written, so it is non-empty.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.env = prefixMapKeys(tc.env, envVarNameBase)
			test.IsolateEnv(t, tc.env)

			got := FirstNonDefaultWithEnv(tc.slice...)

			if tc.allDefault {
				if want := ""; got != want {
					t.Fatalf("value mismatch\ngot:  %q\nwant: %q", got, want)
				}
				return
			}
			want := ValueOr(tc.wantText, tc.slice[tc.wantIndex])
			if got != want {
				t.Errorf("value mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

func TestEvalEnvVars(t *testing.T) {
	envVarNameBase := "TEST_EVAL_ENV_VARS_"

	tests := []struct {
		name  string
		input string
		env   map[string]string
		want  string
	}{
		{
			name:  "no env var '${...}'",
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

			tc.env = prefixMapKeys(tc.env, envVarNameBase)
			test.IsolateEnv(t, tc.env)

			got := EvalEnvVars(tc.input)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestTryExpandEnv(t *testing.T) {
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
			expected: new("bar"),
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
			expected: new("hello-world"),
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
			expected: new("prefix-value-suffix"),
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

			test.IsolateEnv(t, tc.env)

			result := TryExpandEnv(tc.input)

			resultStr := test.StringifyPtr(result)
			expectedStr := test.StringifyPtr(tc.expected)
			if resultStr != expectedStr {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", resultStr, expectedStr)
			}
		})
	}
}

func TestExpandEnvVariables(t *testing.T) {
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
			want:       fmt.Sprintf("${%s%s}", envVarNameBase, "UNDEFINED"),
		},
		{
			name:        "empty env var",
			envVarName:  "EMPTY",
			envVarValue: new(""),
			want:        "",
		},
		{
			name:        "non-empty env var",
			envVarName:  "NON_EMPTY",
			envVarValue: new("bar"),
			want:        "bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.envVarName = envVarNameBase + tc.envVarName
			if tc.envVarValue != nil {
				env := map[string]string{tc.envVarName: *tc.envVarValue}
				test.IsolateEnv(t, env)
			}

			got := expandEnvVariables(fmt.Sprintf("${%s}", tc.envVarName))

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
