// Copyright [2024] [Argus]
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

// Package util provides utility functions for the Argus project.
package util

import (
	"os"
	"regexp"
	"strings"
)

// FirstNonDefaultWithEnv returns the first non-empty variable after evaluating any environment variables.
// It returns an empty string if all variables empty.
func FirstNonDefaultWithEnv(vars ...string) string {
	for _, v := range vars {
		if v = EvalEnvVars(v); v != "" {
			return v
		}
	}
	return ""
}

// envVarRegex matches environment variables using a regular expression.
var envVarRegex = regexp.MustCompile(`\${([a-zA-Z]\w*)}`)

// EvalEnvVars will evaluate the environment variables in the string.
func EvalEnvVars(input string) string {
	// May contain an environment variable.
	if strings.Contains(input, "${") {
		return envVarRegex.ReplaceAllStringFunc(input, expandEnvVariables)
	}
	// No environment variables.
	return input
}

// expandEnvVariables replaces environment variables in a string.
func expandEnvVariables(match string) string {
	envVarName := match[2 : len(match)-1] // Remove the '${' and '}'.
	if value, ok := os.LookupEnv(envVarName); ok {
		return value
	}
	return match
}
