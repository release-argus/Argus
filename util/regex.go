// Copyright [2023] [Argus]
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

package util

import (
	"fmt"
	"regexp"
	"strings"
)

// regexCheck returns whether there is a regex match of `re` on `text`.
func RegexCheck(re string, text string) bool {
	regex := regexp.MustCompile(re)
	// Return whether there's a regex match.
	return regex.MatchString(text)
}

// regexCheckWithParams returns the result of a regex match of `re` on `text`
// after replacing "{{ version }}" with the version string.
func RegexCheckWithParams(re string, text string, version string) bool {
	re = TemplateString(re, ServiceInfo{LatestVersion: version})
	return RegexCheck(re, text)
}

// RegexTemplate on `texts[index]` with the regex `template“.
func RegexTemplate(regexMatches []string, template *string) (result string) {
	// No template, return the text at the index.
	if template == nil {
		return regexMatches[len(regexMatches)-1]
	}

	// Replace placeholders in the template with matched groups in reverse order
	// (so that '$10' isn't replace by '$1')
	result = *template
	for i := len(regexMatches) - 1; i > 0; i-- {
		placeholder := fmt.Sprintf("$%d", i)
		result = strings.ReplaceAll(result, placeholder, regexMatches[i])
	}

	return result
}
