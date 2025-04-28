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

// Package util provides utility functions for the Argus project.
package util

import (
	"fmt"
	"regexp"
	"strings"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

// RegexCheck returns true if a regex match of `re` matches `text`.
func RegexCheck(re, text string) bool {
	regex := regexp.MustCompile(re)
	// Return regex match case.
	return regex.MatchString(text)
}

// RegexCheckWithVersion returns true if a regex match of `re` occurs on `text`
// after replacing "{{ version }}" with the version string.
func RegexCheckWithVersion(re, text, version string) bool {
	re = TemplateString(re, serviceinfo.ServiceInfo{LatestVersion: version})
	return RegexCheck(re, text)
}

// RegexTemplate on `texts[index]` with the regex `template`.
func RegexTemplate(regexMatches []string, template string) string {
	// No template, return the text at the last index.
	if template == "" {
		return regexMatches[len(regexMatches)-1]
	}

	// Replace placeholders with matched groups in reverse order
	// (to prevent replacing '$10' with '$1').
	result := template
	for i := len(regexMatches) - 1; i > 0; i-- {
		placeholder := fmt.Sprintf("$%d", i)
		result = strings.ReplaceAll(result, placeholder, regexMatches[i])
	}

	return result
}
